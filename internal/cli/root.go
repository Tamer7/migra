package cli

import (
	"fmt"
	"os"

	"github.com/migra/migra/pkg/migra"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	jsonOutput bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "migra",
	Short: "Migra - Migration Orchestration CLI",
	Long: `Migra is an open-source, framework-agnostic CLI tool that orchestrates 
database migrations across heterogeneous microservice architectures.

It supports multi-tenant systems, parallel execution, and multiple frameworks
including Django, Laravel, and Prisma.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "migra.yaml", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "quiet mode")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "JSON output format")

	rootCmd.Version = fmt.Sprintf("%s (built: %s)", migra.Version, migra.BuildTime)
}
