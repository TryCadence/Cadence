package ai

import (
	"os"
)

// Config holds the configuration for AI-powered code analysis.
type Config struct {
	Enabled   bool
	Provider  string // "openai", "anthropic", or empty (defaults to "openai")
	APIKey    string
	Model     string // Provider-specific model name; uses provider default if empty
	MaxTokens int
}

// LoadConfig creates a Config from environment variables.
func LoadConfig() *Config {
	return &Config{
		Enabled:   os.Getenv("CADENCE_AI_ENABLED") == "true",
		Provider:  os.Getenv("CADENCE_AI_PROVIDER"),
		APIKey:    os.Getenv("CADENCE_AI_KEY"),
		Model:     getEnvOrDefault("CADENCE_AI_MODEL", ""),
		MaxTokens: 500,
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
