package webhook

import (
	"net/http"
	"testing"
)

func TestWebhookHandlers_HealthCheck(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)
	_ = NewWebhookHandlers("test-secret", queue, nil)

	server, err := NewServer(&ServerConfig{
		Host:          "localhost",
		Port:          9999,
		WebhookSecret: "test-secret",
		MaxWorkers:    2,
	}, processor)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	app := server.GetApp()

	t.Run("health check", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test() unexpected error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})
}

func TestWebhookHandlers_Routes(t *testing.T) {
	processor := NewDefaultProcessor()
	server, err := NewServer(&ServerConfig{
		Host:          "localhost",
		Port:          9999,
		WebhookSecret: "test-secret",
		MaxWorkers:    2,
	}, processor)
	if err != nil {
		t.Fatalf("NewServer() failed: %v", err)
	}

	app := server.GetApp()

	t.Run("github webhook endpoint exists", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/webhooks/github", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test() unexpected error = %v", err)
		}
		// Should return 401 because signature is missing
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("list jobs endpoint exists", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/jobs", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Test() unexpected error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})
}

func TestVerifySignature(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)
	handlers := NewWebhookHandlers("test-secret", queue, nil)

	t.Run("valid signature format", func(t *testing.T) {
		body := []byte(`{"test": "data"}`)
		// This would need a valid HMAC-SHA256 signature
		// For now, just test that invalid format is rejected
		err := handlers.verifySignature(body, "invalid-format")
		if err == nil {
			t.Error("verifySignature() expected error for invalid format")
		}
	})

	t.Run("invalid signature format", func(t *testing.T) {
		body := []byte(`{"test": "data"}`)
		err := handlers.verifySignature(body, "noseparator")
		if err == nil {
			t.Error("verifySignature() expected error for missing separator")
		}
	})

	t.Run("invalid hex encoding", func(t *testing.T) {
		body := []byte(`{"test": "data"}`)
		err := handlers.verifySignature(body, "sha256=notahex")
		if err == nil {
			t.Error("verifySignature() expected error for invalid hex")
		}
	})
}
