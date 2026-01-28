package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/codemeapixel/cadence/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage Cadence configuration. Use "cadence config init" to create a .cadence.yaml file, or "cadence config" to print a sample configuration to stdout.`,
	RunE:  runConfigDefault,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate sample configuration file",
	Long:  `Generate a sample .cadence.yaml configuration file in the current directory.`,
	RunE:  runConfigInit,
}

func init() {
	configCmd.AddCommand(configInitCmd)
}

func runConfigDefault(cmd *cobra.Command, args []string) error {
	sample := `# Cadence Configuration - AI-Generated Code Detection
# Analyzes git repositories to detect potential AI-generated code patterns

thresholds:
  # SIZE-BASED DETECTION
  suspicious_additions: 500
  suspicious_deletions: 1000
  
  # VELOCITY-BASED DETECTION
  max_additions_per_min: 100
  max_deletions_per_min: 500
  
  # TIMING-BASED DETECTION
  min_time_delta_seconds: 60
  
  # FILE DISPERSION DETECTION
  max_files_per_commit: 50
  
  # RATIO-BASED DETECTION
  max_addition_ratio: 0.95
  min_deletion_ratio: 0.95
  min_commit_size_ratio: 100
  
  # PRECISION ANALYSIS
  enable_precision_analysis: true

# File patterns to exclude from analysis
exclude_files:
  - package-lock.json
  - yarn.lock
  - "*.min.js"
  - "*.min.css"
  - "node_modules/**"

# WEBHOOK SERVER CONFIGURATION
webhook:
  # Enable/disable webhook server
  enabled: false
  
  # Server host and port
  host: "0.0.0.0"
  port: 3000
  
  # Webhook secret for signature verification (set this!)
  secret: "your-webhook-secret-key-here"
  
  # Number of concurrent workers for processing webhook events
  max_workers: 4
  
  # Request timeouts in seconds
  read_timeout: 30
  write_timeout: 30
`
	fmt.Print(sample)
	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	configPath := ".cadence.yaml"

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	if err := config.GenerateSampleConfig(configPath); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Printf("Sample configuration file created: %s\n", configPath)
	fmt.Println("Edit this file to configure your thresholds.")

	return nil
}
