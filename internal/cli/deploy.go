package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/migra/migra/internal/adapter"
	"github.com/migra/migra/internal/config"
	"github.com/migra/migra/internal/engine"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/migra/migra/pkg/migra"
	"github.com/spf13/cobra"
)

var (
	deployServiceFilter string
	deployDryRun        bool
	deployParallel      bool
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy migrations across all services",
	Long: `Execute migrations for all configured services according to the
execution strategy defined in the configuration file.`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVar(&deployServiceFilter, "service", "", "filter by service name")
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", false, "dry run without executing migrations")
	deployCmd.Flags().BoolVar(&deployParallel, "parallel", false, "override execution strategy to use parallel")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	start := time.Now()

	// Load configuration
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup logger
	logLevel := logger.ParseLevel(cfg.Logging.Level)
	log := logger.NewLogger(cfg.Logging.Format, logLevel, verbose, quiet)

	log.Info("Starting migration deployment")

	// Setup state manager
	workDir, _ := os.Getwd()
	stateManager := state.NewManager(workDir)
	if err := stateManager.Load(); err != nil {
		log.Warn("Failed to load state, starting fresh", logger.F("error", err.Error()))
	}

	// Setup adapter registry
	registry := adapter.NewDefaultRegistry()

	// Filter services if needed
	services := cfg.Services
	if deployServiceFilter != "" {
		filtered := make([]migra.Service, 0)
		for _, svc := range cfg.Services {
			if svc.Name == deployServiceFilter {
				filtered = append(filtered, svc)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("service '%s' not found", deployServiceFilter)
		}
		services = filtered
		log.Info(fmt.Sprintf("Filtered to service: %s", deployServiceFilter))
	}

	// Determine execution strategy
	strategy := cfg.Execution.Strategy
	if deployParallel {
		strategy = config.StrategyParallel
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Warn("Received interrupt signal, stopping...")
		cancel()
	}()

	// Create execution engine
	var eng engine.Engine
	if strategy == config.StrategyParallel {
		parallelLimit := cfg.Execution.ParallelLimit
		if parallelLimit == 0 {
			parallelLimit = cfg.ParallelLimit
		}
		if parallelLimit == 0 {
			parallelLimit = config.DefaultParallelLimit
		}
		log.Info(fmt.Sprintf("Using parallel execution (limit: %d)", parallelLimit))
		eng = engine.NewParallelEngine(registry, stateManager, log, cfg.Execution.StopOnFailure, deployDryRun, parallelLimit)
	} else {
		log.Info("Using sequential execution")
		eng = engine.NewSequentialEngine(registry, stateManager, log, cfg.Execution.StopOnFailure, deployDryRun)
	}

	// Execute migrations
	log.Info(fmt.Sprintf("Executing migrations for %d service(s)", len(services)))
	results, err := eng.Execute(ctx, services, migra.OperationDeploy)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	// Summarize results
	summary := engine.SummarizeResults(results, time.Since(start))

	// Print summary
	if jsonOutput {
		printJSONSummary(summary)
	} else {
		printConsoleSummary(summary, log)
	}

	// Exit with error if any failures
	if summary.TotalFailure > 0 {
		return fmt.Errorf("deployment completed with %d failure(s)", summary.TotalFailure)
	}

	log.Info("Deployment completed successfully")
	return nil
}

func printConsoleSummary(summary *engine.Result, log logger.Logger) {
	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("MIGRATION SUMMARY")
	fmt.Println(separator)
	fmt.Printf("Total Services:  %d\n", len(summary.Services))
	fmt.Printf("Successful:      %d\n", summary.TotalSuccess)
	fmt.Printf("Failed:          %d\n", summary.TotalFailure)
	fmt.Printf("Total Duration:  %s\n", summary.Duration)
	fmt.Println(separator)

	if summary.TotalFailure > 0 {
		fmt.Println("\nFailed Services:")
		for _, svc := range summary.Services {
			if !svc.Success {
				fmt.Printf("  - %s: %s\n", svc.ServiceName, svc.Error)
			}
		}
	}
}

func printJSONSummary(summary *engine.Result) {
	// This would encode the summary as JSON
	fmt.Printf(`{"total":%d,"success":%d,"failure":%d,"duration":"%s"}`, 
		len(summary.Services), summary.TotalSuccess, summary.TotalFailure, summary.Duration)
	fmt.Println()
}
