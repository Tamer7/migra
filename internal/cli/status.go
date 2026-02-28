package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/migra/migra/internal/config"
	"github.com/migra/migra/internal/logger"
	"github.com/migra/migra/internal/state"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status across services",
	Long:  `Display the current migration status for all configured services.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Setup logger
	logLevel := logger.ParseLevel(cfg.Logging.Level)
	log := logger.NewLogger(cfg.Logging.Format, logLevel, verbose, quiet)

	// Load state
	workDir, _ := os.Getwd()
	stateManager := state.NewManager(workDir)
	if err := stateManager.Load(); err != nil {
		log.Warn("No state file found - no migrations have been run yet")
		return nil
	}

	currentState := stateManager.GetState()

	if jsonOutput {
		// JSON output would go here
		fmt.Println("{}")
		return nil
	}

	// Console output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "SERVICE\tLAST RUN\tRESULT\tSUCCESS\tFAILURES")
	fmt.Fprintln(w, "-------\t--------\t------\t-------\t--------")

	for _, service := range cfg.Services {
		svcState, ok := currentState.Services[service.Name]
		if !ok {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n", 
				service.Name, "never", "-", 0, 0)
			continue
		}

		lastRun := "never"
		if !svcState.LastRun.IsZero() {
			lastRun = formatTime(svcState.LastRun)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
			service.Name,
			lastRun,
			svcState.LastResult,
			svcState.SuccessCount,
			svcState.FailureCount,
		)
	}

	w.Flush()

	// Show tenant summary if tenancy is enabled
	if cfg.Tenancy != nil && cfg.Tenancy.Enabled && len(currentState.Tenants) > 0 {
		fmt.Printf("\nTenant Summary: %d tenant(s) processed\n", len(currentState.Tenants))
	}

	return nil
}

func formatTime(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d min ago", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}
}
