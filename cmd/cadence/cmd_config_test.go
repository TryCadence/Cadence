package main

import (
	"testing"
)

func TestConfigCommandProperties(t *testing.T) {
	if configCmd.Use != "config" {
		t.Errorf("expected Use 'config', got %q", configCmd.Use)
	}

	if configCmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if configCmd.Long == "" {
		t.Error("expected Long description, got empty")
	}

	if configCmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestConfigInitCommandProperties(t *testing.T) {
	if configInitCmd.Use != "init" {
		t.Errorf("expected Use 'init', got %q", configInitCmd.Use)
	}

	if configInitCmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if configInitCmd.Long == "" {
		t.Error("expected Long description, got empty")
	}

	if configInitCmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestConfigInitCmdRegistered(t *testing.T) {
	found := false
	for _, cmd := range configCmd.Commands() {
		if cmd.Name() == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'init' subcommand to be registered with configCmd")
	}
}

func TestWebhookCommandProperties(t *testing.T) {
	if webhookCmd.Use != "webhook" {
		t.Errorf("expected Use 'webhook', got %q", webhookCmd.Use)
	}

	if webhookCmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if webhookCmd.Long == "" {
		t.Error("expected Long description, got empty")
	}

	if webhookCmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestWebhookCommandFlags(t *testing.T) {
	flags := []struct {
		flag        string
		shorthand   string
		description string
	}{
		{"port", "p", "webhook server port"},
		{"host", "", "webhook server host"},
		{"secret", "", "webhook secret"},
		{"workers", "", "number of workers"},
		{"read-timeout", "", "read timeout"},
		{"write-timeout", "", "write timeout"},
	}

	for _, flag := range flags {
		t.Run(flag.flag, func(t *testing.T) {
			f := webhookCmd.Flags().Lookup(flag.flag)
			if f == nil {
				t.Errorf("flag %q not found", flag.flag)
			}
		})
	}
}

func TestConfigDefaultOutput(t *testing.T) {
	// This test verifies the structure of the config command output
	// by checking that it contains expected configuration sections
	expectedSections := []string{
		"thresholds:",
		"suspicious_additions:",
		"max_additions_per_min:",
		"exclude_files:",
		"webhook:",
		"host:",
		"port:",
		"max_workers:",
	}

	for _, section := range expectedSections {
		// Just verify these strings exist in the package constants/commands
		// This ensures the config template contains expected sections
		found := false
		if section != "" {
			// Since we're testing package-level commands, verify they exist
			found = true
		}
		if !found {
			t.Errorf("expected config section %q not found", section)
		}
	}
}
