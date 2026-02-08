// Package anthropic implements the AI Provider interface for Anthropic's Messages API.
// It uses a plain HTTP client — no external SDK dependency required.
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/TryCadence/Cadence/internal/ai"
)

const (
	providerName      = "anthropic"
	defaultModel      = "claude-sonnet-4-20250514"
	defaultBaseURL    = "https://api.anthropic.com"
	defaultAPIVersion = "2023-06-01"
)

func init() {
	ai.RegisterProvider(providerName, New)
}

type Provider struct {
	apiKey     string
	baseURL    string
	apiVersion string
	httpClient *http.Client
	config     *ai.Config
}

func New(cfg *ai.Config) (ai.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic: API key is required")
	}

	return &Provider{
		apiKey:     cfg.APIKey,
		baseURL:    defaultBaseURL,
		apiVersion: defaultAPIVersion,
		httpClient: &http.Client{},
		config:     cfg,
	}, nil
}

// Name returns "anthropic".
func (p *Provider) Name() string {
	return providerName
}

// DefaultModel returns the default Claude model.
func (p *Provider) DefaultModel() string {
	return defaultModel
}

// messagesRequest is the request body for the Anthropic Messages API.
type messagesRequest struct {
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens"`
	System      string         `json:"system,omitempty"`
	Messages    []messageParam `json:"messages"`
	Temperature float32        `json:"temperature,omitempty"`
}

type messageParam struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messagesResponse is the response body from the Anthropic Messages API.
type messagesResponse struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
	Model   string         `json:"model"`
	Error   *apiError      `json:"error,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type apiError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Complete sends a chat completion request to the Anthropic Messages API.
func (p *Provider) Complete(ctx context.Context, req ai.CompletionRequest) (string, error) {
	model := req.Model
	if model == "" {
		model = defaultModel
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	body := messagesRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []messageParam{
			{Role: "user", Content: req.UserPrompt},
		},
		Temperature: req.Temperature,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", p.apiVersion)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("anthropic: API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic: failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic: API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var msgResp messagesResponse
	if err := json.Unmarshal(respBody, &msgResp); err != nil {
		return "", fmt.Errorf("anthropic: failed to parse response: %w", err)
	}

	if msgResp.Error != nil {
		return "", fmt.Errorf("anthropic: API error (%s): %s", msgResp.Error.Type, msgResp.Error.Message)
	}

	// Extract text from the first text content block
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("anthropic: no text content in response")
}

// IsAvailable reports whether the provider has a valid API key.
func (p *Provider) IsAvailable() bool {
	return p.config != nil && p.apiKey != ""
}
