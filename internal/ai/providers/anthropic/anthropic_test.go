package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
			config:      &ai.Config{APIKey: "sk-ant-test", Model: "claude-sonnet-4-20250514"},
			expectError: false,
		},
		{
			name:        "empty API key",
			config:      &ai.Config{APIKey: "", Model: "claude-sonnet-4-20250514"},
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
	p, err := New(&ai.Config{APIKey: "sk-ant-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "anthropic" {
		t.Errorf("expected 'anthropic', got %q", p.Name())
	}
}

func TestProviderDefaultModel(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "sk-ant-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DefaultModel() != "claude-sonnet-4-20250514" {
		t.Errorf("expected 'claude-sonnet-4-20250514', got %q", p.DefaultModel())
	}
}

func TestProviderIsAvailable(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "sk-ant-test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.IsAvailable() {
		t.Error("expected IsAvailable to be true")
	}
}

func TestProviderComplete(t *testing.T) {
	// Set up a mock Anthropic API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key 'test-key', got %q", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != defaultAPIVersion {
			t.Errorf("expected anthropic-version %q, got %q", defaultAPIVersion, r.Header.Get("anthropic-version"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}

		// Verify request body
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Model != "claude-sonnet-4-20250514" {
			t.Errorf("expected model 'claude-sonnet-4-20250514', got %q", req.Model)
		}
		if req.System != "You are a test assistant" {
			t.Errorf("unexpected system prompt: %q", req.System)
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "Hello" {
			t.Errorf("unexpected messages: %v", req.Messages)
		}

		// Return mock response
		resp := messagesResponse{
			ID:   "msg_test",
			Type: "message",
			Role: "assistant",
			Content: []contentBlock{
				{Type: "text", Text: "Hello from Claude!"},
			},
			Model: "claude-sonnet-4-20250514",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &Provider{
		apiKey:     "test-key",
		baseURL:    server.URL,
		apiVersion: defaultAPIVersion,
		httpClient: server.Client(),
		config:     &ai.Config{APIKey: "test-key"},
	}

	result, err := p.Complete(context.Background(), ai.CompletionRequest{
		SystemPrompt: "You are a test assistant",
		UserPrompt:   "Hello",
		Model:        "claude-sonnet-4-20250514",
		MaxTokens:    1024,
		Temperature:  0.3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "Hello from Claude!" {
		t.Errorf("expected 'Hello from Claude!', got %q", result)
	}
}

func TestProviderCompleteAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"type":    "authentication_error",
				"message": "Invalid API key",
			},
		})
	}))
	defer server.Close()

	p := &Provider{
		apiKey:     "bad-key",
		baseURL:    server.URL,
		apiVersion: defaultAPIVersion,
		httpClient: server.Client(),
		config:     &ai.Config{APIKey: "bad-key"},
	}

	_, err := p.Complete(context.Background(), ai.CompletionRequest{
		UserPrompt: "Hello",
		MaxTokens:  100,
	})
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestProviderCompleteNoTextContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := messagesResponse{
			ID:      "msg_test",
			Type:    "message",
			Role:    "assistant",
			Content: []contentBlock{}, // No content
			Model:   "claude-sonnet-4-20250514",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &Provider{
		apiKey:     "test-key",
		baseURL:    server.URL,
		apiVersion: defaultAPIVersion,
		httpClient: server.Client(),
		config:     &ai.Config{APIKey: "test-key"},
	}

	_, err := p.Complete(context.Background(), ai.CompletionRequest{
		UserPrompt: "Hello",
		MaxTokens:  100,
	})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestProviderImplementsInterface(t *testing.T) {
	p, err := New(&ai.Config{APIKey: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var _ ai.Provider = p
}
