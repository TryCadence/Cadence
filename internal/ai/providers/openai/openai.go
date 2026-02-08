// Package openai implements the AI Provider interface for OpenAI's chat API.
package openai

import (
	"context"
	"fmt"

	"github.com/TryCadence/Cadence/internal/ai"
	openaisdk "github.com/sashabaranov/go-openai"
)

const (
	providerName = "openai"
	defaultModel = "gpt-4o-mini"
)

func init() {
	ai.RegisterProvider(providerName, New)
}

// Provider implements ai.Provider for OpenAI.
type Provider struct {
	client *openaisdk.Client
	config *ai.Config
}

// New creates a new OpenAI provider from the given configuration.
func New(cfg *ai.Config) (ai.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai: API key is required")
	}

	client := openaisdk.NewClient(cfg.APIKey)
	return &Provider{
		client: client,
		config: cfg,
	}, nil
}

// Name returns "openai".
func (p *Provider) Name() string {
	return providerName
}

// DefaultModel returns "gpt-4o-mini".
func (p *Provider) DefaultModel() string {
	return defaultModel
}

// Complete sends a chat completion request to the OpenAI API.
func (p *Provider) Complete(ctx context.Context, req ai.CompletionRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = defaultModel
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	resp, err := p.client.CreateChatCompletion(ctx, openaisdk.ChatCompletionRequest{
		Model: model,
		Messages: []openaisdk.ChatCompletionMessage{
			{
				Role:    openaisdk.ChatMessageRoleSystem,
				Content: req.SystemPrompt,
			},
			{
				Role:    openaisdk.ChatMessageRoleUser,
				Content: req.UserPrompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		return "", fmt.Errorf("openai: API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("openai: no response choices returned")
	}

	return resp.Choices[0].Message.Content, nil
}

// IsAvailable reports whether the provider has a valid API key.
func (p *Provider) IsAvailable() bool {
	return p.config != nil && p.config.APIKey != ""
}
