package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/TryCadence/Cadence/internal/ai/skills"
)

// mockProvider implements Provider for testing purposes.
type mockProvider struct {
	name         string
	defaultModel string
	available    bool
	response     string
	err          error
}

func (m *mockProvider) Name() string         { return m.name }
func (m *mockProvider) DefaultModel() string { return m.defaultModel }
func (m *mockProvider) IsAvailable() bool    { return m.available }
func (m *mockProvider) Complete(_ context.Context, _ CompletionRequest) (string, error) {
	return m.response, m.err
}

func TestNewAnalyzerDisabled(t *testing.T) {
	// Register a dummy provider so the lookup doesn't fail
	ResetProviders()
	RegisterProvider("openai", func(cfg *Config) (Provider, error) {
		return &mockProvider{name: "openai", defaultModel: "gpt-4o-mini", available: true}, nil
	})

	tests := []struct {
		name   string
		config *Config
		isNoop bool
	}{
		{
			name:   "disabled returns NoOp",
			config: &Config{Enabled: false, Provider: "openai", APIKey: "sk-test"},
			isNoop: true,
		},
		{
			name:   "empty API key returns NoOp",
			config: &Config{Enabled: true, Provider: "openai", APIKey: ""},
			isNoop: true,
		},
		{
			name:   "enabled with key creates DefaultAnalyzer",
			config: &Config{Enabled: true, Provider: "openai", APIKey: "sk-test"},
			isNoop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewAnalyzer(tt.config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, isNoop := analyzer.(*NoOpAnalyzer)
			if isNoop != tt.isNoop {
				t.Errorf("expected NoOp=%v, got NoOp=%v", tt.isNoop, isNoop)
			}
		})
	}
}

func TestNewAnalyzerUnknownProvider(t *testing.T) {
	ResetProviders()

	cfg := &Config{Enabled: true, Provider: "unknown", APIKey: "key"}
	_, err := NewAnalyzer(cfg)
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewAnalyzerDefaultsToOpenAI(t *testing.T) {
	ResetProviders()
	RegisterProvider("openai", func(cfg *Config) (Provider, error) {
		return &mockProvider{name: "openai", defaultModel: "gpt-4o-mini", available: true}, nil
	})

	cfg := &Config{Enabled: true, Provider: "", APIKey: "sk-test"}
	analyzer, err := NewAnalyzer(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	da, ok := analyzer.(*DefaultAnalyzer)
	if !ok {
		t.Fatal("expected DefaultAnalyzer")
	}

	if da.ProviderName() != "openai" {
		t.Errorf("expected provider 'openai', got %q", da.ProviderName())
	}
}

func TestDefaultAnalyzerIsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
		expected bool
	}{
		{
			name:     "available provider",
			provider: &mockProvider{available: true},
			expected: true,
		},
		{
			name:     "unavailable provider",
			provider: &mockProvider{available: false},
			expected: false,
		},
		{
			name:     "nil provider",
			provider: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DefaultAnalyzer{provider: tt.provider, config: &Config{}}
			if a.IsConfigured() != tt.expected {
				t.Errorf("expected IsConfigured=%v, got %v", tt.expected, a.IsConfigured())
			}
		})
	}
}

func TestDefaultAnalyzerProviderName(t *testing.T) {
	a := &DefaultAnalyzer{
		provider: &mockProvider{name: "test-provider"},
		config:   &Config{},
	}
	if a.ProviderName() != "test-provider" {
		t.Errorf("expected 'test-provider', got %q", a.ProviderName())
	}

	nilProvider := &DefaultAnalyzer{provider: nil, config: &Config{}}
	if nilProvider.ProviderName() != "" {
		t.Errorf("expected empty string for nil provider, got %q", nilProvider.ProviderName())
	}
}

func TestDefaultAnalyzerAnalyzeWithSystemPrompt(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response:     "This is a test response",
	}

	a := NewDefaultAnalyzer(mock, &Config{Model: "custom-model", MaxTokens: 512})

	result, err := a.AnalyzeWithSystemPrompt(context.Background(), "system prompt", "user prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "This is a test response" {
		t.Errorf("expected 'This is a test response', got %q", result)
	}
}

func TestDefaultAnalyzerAnalyzeWithSystemPromptError(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		available: true,
		err:       fmt.Errorf("API error"),
	}

	a := NewDefaultAnalyzer(mock, &Config{Model: "test"})

	_, err := a.AnalyzeWithSystemPrompt(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDefaultAnalyzerAnalyzeSuspiciousCode(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response: `{
			"assessment": "likely AI-generated",
			"confidence": 0.85,
			"reasoning": "Generic patterns detected",
			"indicators": ["template_code"]
		}`,
	}

	a := NewDefaultAnalyzer(mock, &Config{Model: "test", MaxTokens: 512})

	result, err := a.AnalyzeSuspiciousCode(context.Background(), "abc123def456", "func hello() { return }")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == "" {
		t.Fatal("expected non-empty result")
	}

	// Should contain the assessment
	if !contains(result, "likely AI-generated") && !contains(result, "confidence") {
		t.Errorf("result should contain assessment info, got: %s", result)
	}
}

func TestDefaultAnalyzerTruncatesLongCode(t *testing.T) {
	// Create code longer than 2000 chars
	longCode := ""
	for i := 0; i < 300; i++ {
		longCode += "func foo() {} "
	}

	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response:     `{"assessment": "unlikely AI-generated", "confidence": 0.2}`,
	}

	a := NewDefaultAnalyzer(mock, &Config{Model: "test", MaxTokens: 512})

	_, err := a.AnalyzeSuspiciousCode(context.Background(), "abc123", longCode)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNoOpAnalyzer(t *testing.T) {
	noop := &NoOpAnalyzer{}

	result, err := noop.AnalyzeSuspiciousCode(context.Background(), "hash", "code")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}

	result, err = noop.AnalyzeWithSystemPrompt(context.Background(), "system", "user")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty result, got %q", result)
	}

	if noop.IsConfigured() {
		t.Error("NoOpAnalyzer should not be configured")
	}

	if noop.ProviderName() != "" {
		t.Errorf("expected empty provider name, got %q", noop.ProviderName())
	}
}

func TestConfigStruct(t *testing.T) {
	config := &Config{
		Enabled:   true,
		Provider:  "openai",
		APIKey:    "sk-test",
		Model:     "gpt-4",
		MaxTokens: 2048,
	}

	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
	if config.Provider != "openai" {
		t.Errorf("expected Provider 'openai', got %q", config.Provider)
	}
	if config.MaxTokens != 2048 {
		t.Errorf("expected MaxTokens 2048, got %d", config.MaxTokens)
	}
}

func TestProviderRegistry(t *testing.T) {
	ResetProviders()

	// Registry should be empty after reset
	if len(RegisteredProviders()) != 0 {
		t.Errorf("expected empty registry after reset, got %v", RegisteredProviders())
	}

	// Register providers
	RegisterProvider("alpha", func(cfg *Config) (Provider, error) {
		return &mockProvider{name: "alpha"}, nil
	})
	RegisterProvider("beta", func(cfg *Config) (Provider, error) {
		return &mockProvider{name: "beta"}, nil
	})

	providers := RegisteredProviders()
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
	if providers[0] != "alpha" || providers[1] != "beta" {
		t.Errorf("expected [alpha, beta], got %v", providers)
	}
}

func TestProviderRegistryPanicsOnDuplicate(t *testing.T) {
	ResetProviders()

	RegisterProvider("dupe", func(cfg *Config) (Provider, error) {
		return &mockProvider{}, nil
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()

	RegisterProvider("dupe", func(cfg *Config) (Provider, error) {
		return &mockProvider{}, nil
	})
}

func TestProviderRegistryPanicsOnNilFactory(t *testing.T) {
	ResetProviders()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil factory")
		}
	}()

	RegisterProvider("nil-factory", nil)
}

func TestNewProvider(t *testing.T) {
	ResetProviders()
	RegisterProvider("test-provider", func(cfg *Config) (Provider, error) {
		return &mockProvider{name: "test-provider", available: true}, nil
	})

	p, err := NewProvider(&Config{Provider: "test-provider", APIKey: "key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "test-provider" {
		t.Errorf("expected 'test-provider', got %q", p.Name())
	}
}

func TestNewProviderUnknown(t *testing.T) {
	ResetProviders()

	_, err := NewProvider(&Config{Provider: "nonexistent", APIKey: "key"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestDefaultAnalyzerRunSkill(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response:     `{"assessment": "likely AI-generated", "confidence": 0.9, "reasoning": "test"}`,
	}

	a := NewDefaultAnalyzer(mock, &Config{Model: "test", MaxTokens: 512})

	result, err := a.RunSkill(context.Background(), "code_analysis", skills.CodeAnalysisInput{
		CommitHash: "abc123",
		Code:       "func hello() {}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skill != "code_analysis" {
		t.Errorf("expected skill 'code_analysis', got %q", result.Skill)
	}
	if result.Parsed == nil {
		t.Error("expected parsed output")
	}
}

func TestDefaultAnalyzerRunSkillUnknown(t *testing.T) {
	mock := &mockProvider{name: "mock", available: true}
	a := NewDefaultAnalyzer(mock, &Config{Model: "test"})

	_, err := a.RunSkill(context.Background(), "nonexistent_skill", nil)
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestNoOpAnalyzerRunSkill(t *testing.T) {
	noop := &NoOpAnalyzer{}
	_, err := noop.RunSkill(context.Background(), "code_analysis", nil)
	if err == nil {
		t.Fatal("expected error from NoOpAnalyzer.RunSkill")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
