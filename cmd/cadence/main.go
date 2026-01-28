package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configFile string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "cadence",
	Short: "Detect potential AI-generated code in git repositories",
	Long: `Cadence analyzes git repositories to detect potential AI-generated code
by examining commit patterns, code velocity, and statistical anomalies`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	rootCmd.AddCommand(analyzeCmd, configCmd, versionCmd, webhookCmd)
}
