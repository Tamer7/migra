package cli

import (
	"fmt"

	"github.com/migra/migra/internal/config"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Validate the migra.yaml configuration file without executing migrations.`,
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Printf("✓ Configuration is valid\n")
	fmt.Printf("  Services: %d\n", len(cfg.Services))
	fmt.Printf("  Strategy: %s\n", cfg.Execution.Strategy)
	if cfg.Tenancy != nil && cfg.Tenancy.Enabled {
		fmt.Printf("  Tenancy: enabled (%s)\n", cfg.Tenancy.Mode)
	}

	return nil
}
