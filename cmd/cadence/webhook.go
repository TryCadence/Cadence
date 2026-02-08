package main

import (
	"fmt"
	"time"

	"github.com/TryCadence/Cadence/internal/config"
	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/TryCadence/Cadence/internal/webhook"
	"github.com/spf13/cobra"
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Run the Cadence webhook server for repository integration",
	Long: `Start the webhook server to integrate Cadence with GitHub, GitLab, or other Git hosting platforms.

The webhook server listens for push events and triggers AI code detection analysis.
Configure the webhook settings in your config file or via environment variables.

Example:
  cadence webhook --config config.yaml
  cadence webhook --port 8000 --secret my-webhook-secret`,
	RunE: runWebhookServer,
}

var webhookFlags struct {
	port         int
	host         string
	secret       string
	maxWorkers   int
	readTimeout  int
	writeTimeout int
}

func init() {
	webhookCmd.Flags().IntVarP(&webhookFlags.port, "port", "p", 0, "webhook server port (default: 8000)")
	webhookCmd.Flags().StringVar(&webhookFlags.host, "host", "", "webhook server host (default: 0.0.0.0)")
	webhookCmd.Flags().StringVar(&webhookFlags.secret, "secret", "", "webhook secret for signature verification")
	webhookCmd.Flags().IntVar(&webhookFlags.maxWorkers, "workers", 0, "number of concurrent workers (default: 4)")
	webhookCmd.Flags().IntVar(&webhookFlags.readTimeout, "read-timeout", 0, "request read timeout in seconds (default: 30)")
	webhookCmd.Flags().IntVar(&webhookFlags.writeTimeout, "write-timeout", 0, "request write timeout in seconds (default: 30)")
}

func runWebhookServer(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	webhookCfg := cfg.Webhook
	if webhookFlags.port > 0 {
		webhookCfg.Port = webhookFlags.port
	}
	if webhookFlags.host != "" {
		webhookCfg.Host = webhookFlags.host
	}
	if webhookFlags.secret != "" {
		webhookCfg.Secret = webhookFlags.secret
	}
	if webhookFlags.maxWorkers > 0 {
		webhookCfg.MaxWorkers = webhookFlags.maxWorkers
	}
	if webhookFlags.readTimeout > 0 {
		webhookCfg.ReadTimeout = webhookFlags.readTimeout
	}
	if webhookFlags.writeTimeout > 0 {
		webhookCfg.WriteTimeout = webhookFlags.writeTimeout
	}

	// Validate configuration
	if webhookCfg.Secret == "" {
		return fmt.Errorf("webhook secret is required (set via --secret flag or webhook.secret in config)")
	}

	// Create server configuration
	serverCfg := &webhook.ServerConfig{
		Host:          webhookCfg.Host,
		Port:          webhookCfg.Port,
		WebhookSecret: webhookCfg.Secret,
		MaxWorkers:    webhookCfg.MaxWorkers,
		ReadTimeout:   time.Duration(webhookCfg.ReadTimeout) * time.Second,
		WriteTimeout:  time.Duration(webhookCfg.WriteTimeout) * time.Second,
	}

	// Create analysis processor
	processor := &webhook.AnalysisProcessor{
		DetectorThresholds: &cfg.Thresholds,
		Logger:             logging.Default().With("component", "processor"),
	}

	// Create and start server
	server, err := webhook.NewServer(serverCfg, processor)
	if err != nil {
		return fmt.Errorf("failed to create webhook server: %w", err)
	}

	log := logging.Default().With("component", "webhook_cmd")
	log.Info("starting Cadence webhook server",
		"host", webhookCfg.Host,
		"port", webhookCfg.Port,
		"workers", webhookCfg.MaxWorkers,
	)

	return server.Start()
}
