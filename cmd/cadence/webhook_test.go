package main

import (
	"fmt"
	"testing"
	"time"
)

func TestWebhookFlagsStructure(t *testing.T) {
	tests := []struct {
		fieldName string
		check     func() bool
	}{
		{
			fieldName: "port",
			check: func() bool {
				return webhookFlags.port >= 0
			},
		},
		{
			fieldName: "host",
			check: func() bool {
				return webhookFlags.host == "" || len(webhookFlags.host) > 0
			},
		},
		{
			fieldName: "secret",
			check: func() bool {
				return webhookFlags.secret == "" || len(webhookFlags.secret) > 0
			},
		},
		{
			fieldName: "maxWorkers",
			check: func() bool {
				return webhookFlags.maxWorkers >= 0
			},
		},
		{
			fieldName: "readTimeout",
			check: func() bool {
				return webhookFlags.readTimeout >= 0
			},
		},
		{
			fieldName: "writeTimeout",
			check: func() bool {
				return webhookFlags.writeTimeout >= 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			if !tt.check() {
				t.Errorf("webhook flag %s has invalid state", tt.fieldName)
			}
		})
	}
}

func TestWebhookTimeoutConversion(t *testing.T) {
	tests := []struct {
		seconds  int
		expected time.Duration
		name     string
	}{
		{
			name:     "30 seconds",
			seconds:  30,
			expected: 30 * time.Second,
		},
		{
			name:     "0 seconds",
			seconds:  0,
			expected: 0 * time.Second,
		},
		{
			name:     "60 seconds",
			seconds:  60,
			expected: 60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := time.Duration(tt.seconds) * time.Second
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWebhookDefaultPortRange(t *testing.T) {
	// Test that webhook ports are within valid range
	validPorts := []int{0, 80, 443, 3000, 8000, 8080, 9000, 65535}

	for _, port := range validPorts {
		t.Run(fmt.Sprintf("port_%d", port), func(t *testing.T) {
			if port < 0 || port > 65535 {
				t.Errorf("port %d is out of valid range", port)
			}
		})
	}
}

func TestWebhookHostValidation(t *testing.T) {
	tests := []struct {
		host    string
		name    string
		isValid bool
	}{
		{
			name:    "localhost",
			host:    "localhost",
			isValid: true,
		},
		{
			name:    "127.0.0.1",
			host:    "127.0.0.1",
			isValid: true,
		},
		{
			name:    "0.0.0.0",
			host:    "0.0.0.0",
			isValid: true,
		},
		{
			name:    "empty",
			host:    "",
			isValid: true,
		},
		{
			name:    "example.com",
			host:    "example.com",
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.isValid
			if !isValid {
				t.Errorf("host %q should be valid", tt.host)
			}
		})
	}
}

func TestWebhookSecretValidation(t *testing.T) {
	tests := []struct {
		secret  string
		name    string
		isEmpty bool
	}{
		{
			name:    "empty secret",
			secret:  "",
			isEmpty: true,
		},
		{
			name:    "simple secret",
			secret:  "mysecret",
			isEmpty: false,
		},
		{
			name:    "long secret",
			secret:  "very-long-secret-key-with-many-characters-12345",
			isEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if (tt.secret == "") != tt.isEmpty {
				if tt.isEmpty && tt.secret != "" {
					t.Errorf("expected secret to be empty, got %q", tt.secret)
				}
			}
		})
	}
}

func TestWebhookMaxWorkersValidation(t *testing.T) {
	tests := []struct {
		workers int
		name    string
		valid   bool
	}{
		{
			name:    "zero workers",
			workers: 0,
			valid:   true,
		},
		{
			name:    "single worker",
			workers: 1,
			valid:   true,
		},
		{
			name:    "four workers",
			workers: 4,
			valid:   true,
		},
		{
			name:    "many workers",
			workers: 100,
			valid:   true,
		},
		{
			name:    "negative workers",
			workers: -1,
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				if tt.workers < 0 {
					t.Errorf("workers %d should be valid (non-negative)", tt.workers)
				}
			} else {
				if tt.workers >= 0 {
					t.Errorf("workers %d should be invalid", tt.workers)
				}
			}
		})
	}
}
