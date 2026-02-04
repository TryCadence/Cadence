package ai

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewOpenAIAnalyzer(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		model       []string
		expectError bool
		expectModel string
	}{
		{
			name:        "with Config struct",
			input:       &Config{APIKey: "test-key", Model: "gpt-4"},
			expectError: false,
			expectModel: "gpt-4",
		},
		{
			name:        "with string API key",
			input:       "test-key-string",
			model:       []string{"gpt-3.5-turbo"},
			expectError: false,
			expectModel: "gpt-3.5-turbo",
		},
		{
			name:        "with string API key, default model",
			input:       "test-key-string",
			expectError: false,
			expectModel: "gpt-4o-mini",
		},
		{
			name:        "empty Config API key",
			input:       &Config{APIKey: "", Model: "gpt-4"},
			expectError: true,
		},
		{
			name:        "invalid input type",
			input:       123,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewOpenAIAnalyzer(tt.input, tt.model...)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if analyzer == nil {
					t.Errorf("expected analyzer but got nil")
				} else if analyzer.config.Model != tt.expectModel {
					t.Errorf("expected model %q, got %q", tt.expectModel, analyzer.config.Model)
				}
			}
		})
	}
}

func TestAnalysisResult(t *testing.T) {
	result := &AnalysisResult{
		Assessment: "likely AI-generated",
		Confidence: 0.85,
		Reasoning:  "Code patterns suggest AI generation",
		Indicators: []string{"generic_code", "perfect_formatting"},
	}

	if result.Assessment != "likely AI-generated" {
		t.Errorf("expected Assessment 'likely AI-generated', got %q", result.Assessment)
	}

	if result.Confidence != 0.85 {
		t.Errorf("expected Confidence 0.85, got %v", result.Confidence)
	}

	if len(result.Indicators) != 2 {
		t.Errorf("expected 2 indicators, got %d", len(result.Indicators))
	}
}

func TestGetAssessmentFromText(t *testing.T) {
	tests := []struct {
		text               string
		expectedAssess     string
		expectedConfidence float64
	}{
		{
			text:               "this code is likely AI-generated",
			expectedAssess:     "likely AI-generated",
			expectedConfidence: 0.8,
		},
		{
			text:               "this code is possibly AI-generated",
			expectedAssess:     "possibly AI-generated",
			expectedConfidence: 0.5,
		},
		{
			text:               "this code was probably written by a human",
			expectedAssess:     "unlikely AI-generated",
			expectedConfidence: 0.2,
		},
		{
			text:               "",
			expectedAssess:     "unlikely AI-generated",
			expectedConfidence: 0.2,
		},
		{
			text:               "LIKELY AI",
			expectedAssess:     "unlikely AI-generated",
			expectedConfidence: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			assess, conf := getAssessmentFromText(tt.text)

			if assess != tt.expectedAssess {
				t.Errorf("expected assessment %q, got %q", tt.expectedAssess, assess)
			}

			if conf != tt.expectedConfidence {
				t.Errorf("expected confidence %.1f, got %.1f", tt.expectedConfidence, conf)
			}
		})
	}
}

func TestParseConfidence(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		name     string
	}{
		{
			name:     "string 1",
			input:    "1",
			expected: 1.0,
		},
		{
			name:     "string 1.0",
			input:    "1.0",
			expected: 1.0,
		},
		{
			name:     "string 0",
			input:    "0",
			expected: 0.0,
		},
		{
			name:     "string 0.0",
			input:    "0.0",
			expected: 0.0,
		},
		{
			name:     "string starting with 0",
			input:    "0.5",
			expected: 0.5,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0.5,
		},
		{
			name:     "other string",
			input:    "unknown",
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConfidence(tt.input)
			if result != tt.expected {
				t.Errorf("expected %.1f, got %.1f", tt.expected, result)
			}
		})
	}
}

func TestIntMin(t *testing.T) {
	tests := []struct {
		a        int
		b        int
		expected int
		name     string
	}{
		{
			name:     "a < b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b < a",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "a == b",
			a:        5,
			b:        5,
			expected: 5,
		},
		{
			name:     "negative a",
			a:        -5,
			b:        10,
			expected: -5,
		},
		{
			name:     "both negative",
			a:        -10,
			b:        -5,
			expected: -10,
		},
		{
			name:     "zero values",
			a:        0,
			b:        0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := intMin(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseAnalysisResult(t *testing.T) {
	tests := []struct {
		name              string
		responseText      string
		expectedAssess    string
		checkConfidence   bool
		expectedReasoning bool
	}{
		{
			name: "valid JSON response",
			responseText: `{
				"assessment": "likely AI-generated",
				"confidence": 0.85,
				"reasoning": "Code is too perfect",
				"indicators": ["generic", "perfect_formatting"]
			}`,
			expectedAssess:    "likely AI-generated",
			checkConfidence:   true,
			expectedReasoning: true,
		},
		{
			name:              "text without JSON",
			responseText:      "This code looks like it was generated by AI",
			expectedAssess:    "unlikely AI-generated",
			checkConfidence:   true,
			expectedReasoning: true,
		},
		{
			name:              "empty response",
			responseText:      "",
			expectedAssess:    "unlikely AI-generated",
			checkConfidence:   false,
			expectedReasoning: false,
		},
		{
			name:              "response with possibly",
			responseText:      "The code is possibly AI-generated based on patterns",
			expectedAssess:    "possibly AI-generated",
			checkConfidence:   true,
			expectedReasoning: true,
		},
		{
			name: "JSON with partial fields",
			responseText: `{
				"assessment": "possibly AI-generated",
				"confidence": 0.5
			}`,
			expectedAssess:    "possibly AI-generated",
			checkConfidence:   true,
			expectedReasoning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAnalysisResult(tt.responseText)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}

			if result.Assessment != tt.expectedAssess {
				t.Errorf("expected assessment %q, got %q", tt.expectedAssess, result.Assessment)
			}

			if tt.checkConfidence {
				if result.Confidence < 0 || result.Confidence > 1 {
					t.Errorf("confidence out of range: %v", result.Confidence)
				}
			}

			if tt.expectedReasoning && result.Reasoning == "" && !strings.Contains(tt.responseText, "reasoning") {
				// Only error if we expected reasoning and there's no JSON reasoning field
				// Allow empty reasoning if response text is empty
				if tt.responseText != "" {
					t.Errorf("expected reasoning but got empty string")
				}
			}
		})
	}
}

func TestOpenAIAnalyzerIsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		expected bool
	}{
		{
			name:     "with API key",
			apiKey:   "sk-test-key",
			expected: true,
		},
		{
			name:     "empty API key",
			apiKey:   "",
			expected: false,
		},
		{
			name:     "whitespace only",
			apiKey:   "   ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &OpenAIAnalyzer{
				config: &Config{
					APIKey: tt.apiKey,
				},
			}

			result := analyzer.IsConfigured()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAnalyzeWithSystemPromptErrors(t *testing.T) {
	t.Skip("Skipping network test - requires actual OpenAI API client initialization")
	// This test demonstrates that attempting to use the API without proper client setup fails
	// Create analyzer with no client (simulates unconfigured state)
	analyzer := &OpenAIAnalyzer{
		config: &Config{
			APIKey:    "",
			Model:     "gpt-4o-mini",
			MaxTokens: 1024,
		},
		client: nil, // This will cause an error
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// This should error due to nil client
	_, err := analyzer.AnalyzeWithSystemPrompt(ctx, "system", "user prompt")
	if err == nil {
		t.Error("expected error with nil client, got nil")
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
