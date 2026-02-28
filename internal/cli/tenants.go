package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/migra/migra/internal/adapter"
	"github.com/migra/migra/internal/config"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/migra/migra/internal/tenant"
	"github.com/migra/migra/pkg/migra"
	"github.com/spf13/cobra"
)

var (
	tenantsMaxParallel   int
	tenantsStopOnFailure bool
)

// tenantsCmd represents the tenants command
var tenantsCmd = &cobra.Command{
	Use:   "tenants",
	Short: "Multi-tenant operations",
	Long:  `Commands for managing multi-tenant migrations.`,
}

// tenantsDeployCmd represents the tenants deploy command
var tenantsDeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy migrations to all tenants",
	Long:  `Execute migrations for all configured tenants across all services.`,
	RunE:  runTenantsDeploy,
}

func init() {
	rootCmd.AddCommand(tenantsCmd)
	tenantsCmd.AddCommand(tenantsDeployCmd)

	tenantsDeployCmd.Flags().IntVar(&tenantsMaxParallel, "max-parallel", 0, "maximum parallel tenant executions")
	tenantsDeployCmd.Flags().BoolVar(&tenantsStopOnFailure, "stop-on-failure", false, "stop on first tenant failure")
}

func runTenantsDeploy(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if tenancy is enabled
	if cfg.Tenancy == nil || !cfg.Tenancy.Enabled {
		return fmt.Errorf("tenancy is not enabled in configuration")
	}

	// Setup logger
	logLevel := logger.ParseLevel(cfg.Logging.Level)
	log := logger.NewLogger(cfg.Logging.Format, logLevel, verbose, quiet)

	log.Info("Starting multi-tenant migration deployment")

	// Setup state manager
	workDir, _ := os.Getwd()
	stateManager := state.NewManager(workDir)
	if err := stateManager.Load(); err != nil {
		log.Warn("Failed to load state, starting fresh", logger.F("error", err.Error()))
	}

	// Setup adapter registry
	registry := adapter.NewDefaultRegistry()

	// Create tenant source
	var source tenant.Source
	switch cfg.Tenancy.TenantSource {
	case config.TenantSourceEnv:
		source = tenant.NewEnvSource("MIGRA_TENANTS")
	case config.TenantSourceFile:
		// Assume file path is in environment or default
		filePath := os.Getenv("MIGRA_TENANTS_FILE")
		if filePath == "" {
			filePath = "tenants.json"
		}
		source = tenant.NewFileSource(filePath)
	case config.TenantSourceCommand:
		command := os.Getenv("MIGRA_TENANTS_COMMAND")
		if command == "" {
			return fmt.Errorf("MIGRA_TENANTS_COMMAND environment variable not set")
		}
		parts := strings.Split(command, " ")
		source = tenant.NewCommandSource(parts[0], parts[1:]...)
	default:
		return fmt.Errorf("unsupported tenant source: %s", cfg.Tenancy.TenantSource)
	}

	// Determine max parallel
	maxParallel := tenantsMaxParallel
	if maxParallel == 0 {
		maxParallel = cfg.Tenancy.MaxParallel
	}

	// Determine stop on failure
	stopOnFailure := tenantsStopOnFailure || cfg.Tenancy.StopOnFailure

	// Create tenant executor
	executor := tenant.NewExecutor(source, registry, stateManager, log, stopOnFailure, maxParallel)

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

	// Execute tenant migrations
	results, err := executor.Execute(ctx, cfg.Services, migra.OperationDeploy)
	if err != nil {
		return fmt.Errorf("tenant execution failed: %w", err)
	}

	// Print summary
	successCount := 0
	failureCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failureCount++
		}
	}

	separator := "============================================================"
	fmt.Println("\n" + separator)
	fmt.Println("TENANT MIGRATION SUMMARY")
	fmt.Println(separator)
	fmt.Printf("Total Tenants:   %d\n", len(results))
	fmt.Printf("Successful:      %d\n", successCount)
	fmt.Printf("Failed:          %d\n", failureCount)
	fmt.Println(separator)

	if failureCount > 0 {
		fmt.Println("\nFailed Tenants:")
		for _, r := range results {
			if !r.Success {
				fmt.Printf("  - %s: %s\n", r.TenantID, r.Error)
			}
		}
		return fmt.Errorf("deployment completed with %d tenant failure(s)", failureCount)
	}

	log.Info("Tenant deployment completed successfully")
	return nil
}
