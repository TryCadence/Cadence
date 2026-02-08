package config

import (
	"fmt"
	"os"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/git/patterns"
	"github.com/spf13/viper"
)

// SampleConfigTemplate is the template for generating sample configuration files
// This is the single source of truth for the default Cadence configuration
const SampleConfigTemplate = `# Cadence Configuration - AI-Generated Code Detection
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
  - "dist/**"
  - "build/**"
  - "out/**"
  - "bin/**"
  - ".next/**"
  - "vendor/**"
  - ".git/**"
  - "*.png"
  - "*.jpg"
  - "*.jpeg"
  - "*.gif"
  - "*.svg"
  - "*.ico"
  - "*.woff"
  - "*.woff2"
  - "*.ttf"
  - "*.eot"
  - "*.otf"

# WEBHOOK SERVER CONFIGURATION
webhook:
  # Enable/disable webhook server
  enabled: false
  
  # Server host and port
  host: "0.0.0.0"
  port: 8000
  
  # Webhook secret for signature verification (set this!)
  secret: "your-webhook-secret-key-here"
  
  # Number of concurrent workers for processing webhook events
  max_workers: 4
  
  # Request timeouts in seconds
  read_timeout: 30
  write_timeout: 30

# AI ANALYSIS CONFIGURATION (Optional - requires API key)
ai:
  # Enable/disable AI-powered code analysis
  enabled: false
  
  # AI provider ("openai" or "anthropic")
  provider: "openai"
  
  # API key for the selected provider (or set via CADENCE_AI_KEY environment variable)
  api_key: ""
  
  # Model name (provider-specific; leave empty for provider default)
  # OpenAI default: gpt-4o-mini | Anthropic default: claude-sonnet-4-20250514
  model: ""

# STRATEGY CONFIGURATION (Optional - control which detection strategies are active)
strategies:
  # Set any strategy to false to disable it. All strategies are enabled by default.
  # commit_message_analysis: true
  # naming_pattern_analysis: true
  # structural_consistency: true
  # burst_pattern: true
  # error_handling_pattern: true
  # template_pattern: true
  # file_extension_pattern: true
  # statistical_anomaly: true
  # timing_anomaly: true
`

type Config struct {
	Thresholds   patterns.Thresholds
	ExcludeFiles []string
	Webhook      WebhookConfig
	AI           AIConfig
	Strategies   StrategyConfig
}

// WebhookConfig holds webhook server configuration
type WebhookConfig struct {
	Enabled      bool
	Host         string
	Port         int
	Secret       string
	MaxWorkers   int
	ReadTimeout  int
	WriteTimeout int
}

// AIConfig holds AI analysis configuration
type AIConfig struct {
	Enabled  bool
	Provider string
	APIKey   string
	Model    string
}

// StrategyConfig controls which detection strategies are active.
// All strategies default to enabled (true). Set a strategy to false to disable it.
type StrategyConfig struct {
	DisabledStrategies map[string]bool // strategy name -> disabled
}

// IsEnabled returns whether a strategy is enabled. Defaults to true if not explicitly disabled.
func (sc *StrategyConfig) IsEnabled(name string) bool {
	if sc.DisabledStrategies == nil {
		return true
	}
	disabled, exists := sc.DisabledStrategies[name]
	if !exists {
		return true
	}
	return !disabled
}

func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("thresholds.suspicious_additions", 500)
	v.SetDefault("thresholds.suspicious_deletions", 1000)
	v.SetDefault("thresholds.max_additions_per_min", 100)
	v.SetDefault("thresholds.max_deletions_per_min", 500)
	v.SetDefault("thresholds.min_time_delta_seconds", 60)
	v.SetDefault("thresholds.max_files_per_commit", 50)
	v.SetDefault("thresholds.max_addition_ratio", 0.95)
	v.SetDefault("thresholds.min_deletion_ratio", 0.95)
	v.SetDefault("thresholds.min_commit_size_ratio", 100)
	v.SetDefault("thresholds.enable_precision_analysis", true)

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	v.SetEnvPrefix("CADENCE")
	v.AutomaticEnv()

	config := &Config{}

	config.Thresholds.SuspiciousAdditions = v.GetInt64("thresholds.suspicious_additions")
	config.Thresholds.SuspiciousDeletions = v.GetInt64("thresholds.suspicious_deletions")
	config.Thresholds.MaxAdditionsPerMin = v.GetFloat64("thresholds.max_additions_per_min")
	config.Thresholds.MaxDeletionsPerMin = v.GetFloat64("thresholds.max_deletions_per_min")
	config.Thresholds.MinTimeDeltaSeconds = v.GetInt64("thresholds.min_time_delta_seconds")
	config.Thresholds.MaxFilesPerCommit = v.GetInt("thresholds.max_files_per_commit")
	config.Thresholds.MaxAdditionRatio = v.GetFloat64("thresholds.max_addition_ratio")
	config.Thresholds.MinDeletionRatio = v.GetFloat64("thresholds.min_deletion_ratio")
	config.Thresholds.MinCommitSizeRatio = v.GetInt64("thresholds.min_commit_size_ratio")
	config.Thresholds.EnablePrecisionAnalysis = v.GetBool("thresholds.enable_precision_analysis")

	config.ExcludeFiles = v.GetStringSlice("exclude_files")

	// Load webhook configuration
	config.Webhook.Enabled = v.GetBool("webhook.enabled")
	config.Webhook.Host = v.GetString("webhook.host")
	if config.Webhook.Host == "" {
		config.Webhook.Host = "0.0.0.0"
	}
	config.Webhook.Port = v.GetInt("webhook.port")
	if config.Webhook.Port == 0 {
		config.Webhook.Port = 8000
	}
	config.Webhook.Secret = v.GetString("webhook.secret")
	config.Webhook.MaxWorkers = v.GetInt("webhook.max_workers")
	if config.Webhook.MaxWorkers == 0 {
		config.Webhook.MaxWorkers = 4
	}
	config.Webhook.ReadTimeout = v.GetInt("webhook.read_timeout")
	if config.Webhook.ReadTimeout == 0 {
		config.Webhook.ReadTimeout = 30
	}
	config.Webhook.WriteTimeout = v.GetInt("webhook.write_timeout")
	if config.Webhook.WriteTimeout == 0 {
		config.Webhook.WriteTimeout = 30
	}

	// Load AI configuration
	config.AI.Enabled = v.GetBool("ai.enabled")
	config.AI.Provider = v.GetString("ai.provider")
	config.AI.APIKey = v.GetString("ai.api_key")
	config.AI.Model = v.GetString("ai.model")
	// Model defaults are handled by the provider — leave empty to use provider default

	// Load strategy configuration
	config.Strategies.DisabledStrategies = make(map[string]bool)
	strategyNames := []string{
		"commit_message_analysis",
		"naming_pattern_analysis",
		"structural_consistency",
		"burst_pattern",
		"error_handling_pattern",
		"template_pattern",
		"file_extension_pattern",
		"statistical_anomaly",
		"timing_anomaly",
	}
	for _, name := range strategyNames {
		key := "strategies." + name
		if v.IsSet(key) && !v.GetBool(key) {
			config.Strategies.DisabledStrategies[name] = true
		}
	}

	return config, nil
}

func GenerateSampleConfig(path string) error {
	return os.WriteFile(path, []byte(SampleConfigTemplate), 0o600)
}
