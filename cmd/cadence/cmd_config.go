package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/TryCadence/Cadence/internal/config"
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
	fmt.Print(config.SampleConfigTemplate)
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
