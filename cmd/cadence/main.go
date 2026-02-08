package main

import (
	"fmt"
	"os"

	// Register AI providers
	_ "github.com/TryCadence/Cadence/internal/ai/providers/anthropic"
	_ "github.com/TryCadence/Cadence/internal/ai/providers/openai"

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
	Short: "Detect AI-generated content in git repositories and websites",
	Long: `Cadence analyzes git repositories and websites to detect AI-generated content.

Capabilities:
  • Git commits: Detects suspicious commits via patterns, velocity, and anomalies
  • Websites: Analyzes page content for AI-generated text patterns
  • Optional AI validation: Uses OpenAI GPT-4o-mini for expert analysis

Use 'cadence --help' to see all available commands.`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file path")
	rootCmd.AddCommand(analyzeCmd, webCmd, configCmd, versionCmd, webhookCmd)
}
