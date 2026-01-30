package ai

import (
	"context"
	"os"

	"github.com/sashabaranov/go-openai"
)

// Config holds AI service configuration
type Config struct {
	Enabled   bool
	Provider  string // "openai" or empty to disable
	APIKey    string
	Model     string
	MaxTokens int
}

// LoadConfig loads AI configuration from environment or defaults
func LoadConfig() *Config {
	return &Config{
		Enabled:   os.Getenv("CADENCE_AI_ENABLED") == "true",
		Provider:  os.Getenv("CADENCE_AI_PROVIDER"),
		APIKey:    os.Getenv("CADENCE_AI_KEY"),
		Model:     getEnvOrDefault("CADENCE_AI_MODEL", "gpt-4o-mini"),
		MaxTokens: 500,
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// Analyzer provides AI-powered analysis
type Analyzer interface {
	AnalyzeSuspiciousCode(ctx context.Context, commitHash string, additions string) (string, error)
	IsConfigured() bool
}

// NewAnalyzer creates a new AI analyzer based on configuration
func NewAnalyzer(cfg *Config) (Analyzer, error) {
	if !cfg.Enabled || cfg.APIKey == "" {
		return &NoOpAnalyzer{}, nil
	}

	switch cfg.Provider {
	case "openai":
		client := &openai.Client{}
		return &OpenAIAnalyzer{
			client: client,
			config: cfg,
		}, nil
	default:
		return &NoOpAnalyzer{}, nil
	}
}

// NoOpAnalyzer is a no-op implementation when AI is disabled
type NoOpAnalyzer struct{}

func (n *NoOpAnalyzer) AnalyzeSuspiciousCode(ctx context.Context, commitHash string, additions string) (string, error) {
	return "", nil
}

func (n *NoOpAnalyzer) IsConfigured() bool {
	return false
}
