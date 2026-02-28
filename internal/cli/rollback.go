package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/migra/migra/internal/adapter"
	"github.com/migra/migra/internal/config"
	"github.com/migra/migra/internal/engine"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/migra/migra/pkg/migra"
	"github.com/spf13/cobra"
)

var (
	rollbackService string
	rollbackSteps   int
	rollbackTenant  string
)

// rollbackCmd represents the rollback command
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback migrations",
	Long:  `Rollback the most recent migration for specified services.`,
	RunE:  runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)

	rollbackCmd.Flags().StringVar(&rollbackService, "service", "", "service name (required)")
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "number of steps to rollback")
	rollbackCmd.Flags().StringVar(&rollbackTenant, "tenant", "", "tenant ID (for multi-tenant)")
	rollbackCmd.MarkFlagRequired("service")
}

func runRollback(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger
	logLevel := logger.ParseLevel(cfg.Logging.Level)
	log := logger.NewLogger(cfg.Logging.Format, logLevel, verbose, quiet)

	log.Info(fmt.Sprintf("Rolling back service: %s", rollbackService))

	// Find service
	var targetService *migra.Service
	for _, svc := range cfg.Services {
		if svc.Name == rollbackService {
			targetService = &svc
			break
		}
	}

	if targetService == nil {
		return fmt.Errorf("service '%s' not found", rollbackService)
	}

	// Setup state manager
	workDir, _ := os.Getwd()
	stateManager := state.NewManager(workDir)
	if err := stateManager.Load(); err != nil {
		log.Warn("Failed to load state", logger.F("error", err.Error()))
	}

	// Setup adapter registry
	registry := adapter.NewDefaultRegistry()

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
	eng := engine.NewSequentialEngine(registry, stateManager, log, true, false)

	// Execute rollback
	results, err := eng.Execute(ctx, []migra.Service{*targetService}, migra.OperationRollback)
	if err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	if len(results) > 0 && !results[0].Success {
		return fmt.Errorf("rollback failed: %s", results[0].Error)
	}

	log.Info("Rollback completed successfully")
	return nil
}
