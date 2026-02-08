package openai

import (
	"testing"

	"github.com/TryCadence/Cadence/internal/ai"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name        string
		config      *ai.Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      &ai.Config{APIKey: "sk-test", Model: "gpt-4"},
			expectError: false,
		},
		{
			name:        "empty API key",
			config:      &ai.Config{APIKey: "", Model: "gpt-4"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p == nil {
				t.Fatal("expected provider but got nil")
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("expected 'openai', got %q", p.Name())
	}
}

func TestProviderDefaultModel(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DefaultModel() != "gpt-4o-mini" {
		t.Errorf("expected 'gpt-4o-mini', got %q", p.DefaultModel())
	}
}

func TestProviderIsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{name: "with key", apiKey: "sk-test", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(&ai.Config{APIKey: tt.apiKey})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.IsAvailable() != tt.expected {
				t.Errorf("expected IsAvailable=%v, got %v", tt.expected, p.IsAvailable())
			}
		})
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "sk-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Compile-time check that Provider satisfies ai.Provider
	var _ ai.Provider = p
}
