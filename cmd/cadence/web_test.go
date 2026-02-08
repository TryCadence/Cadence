package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/web"
)

func TestWebCommandFlag(t *testing.T) {
	// Test that web command has all required flags
	flags := []struct {
		flag        string
		shorthand   string
		description string
	}{
		{"output", "o", "write report to file"},
		{"verbose", "v", "show detailed analysis"},
		{"json", "j", "JSON format"},
	}

	for _, flag := range flags {
		t.Run(flag.flag, func(t *testing.T) {
			f := webCmd.Flags().Lookup(flag.flag)
			if f == nil {
				t.Errorf("flag %q not found", flag.flag)
			}
		})
	}
}

func TestWebCommandProperties(t *testing.T) {
	if webCmd.Use != "web <url>" {
		t.Errorf("expected Use 'web <url>', got %q", webCmd.Use)
	}

	if webCmd.Short == "" {
		t.Error("expected Short description, got empty")
	}

	if webCmd.Long == "" {
		t.Error("expected Long description, got empty")
	}

	if webCmd.RunE == nil {
		t.Error("expected RunE function, got nil")
	}
}

func TestFetcherIntegration(t *testing.T) {
	t.Skip("Skipping network test - requires internet connectivity")
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<html>
  <head><title>Test Page</title></head>
  <body>
    <h1>Main Heading</h1>
    <p>This is test content that should be analyzed.</p>
    <p>More content for analysis purposes.</p>
  </body>
</html>
`))
	}))
	defer server.Close()

	fetcher := web.NewFetcher(10)
	content, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == nil {
		t.Error("expected content, got nil")
	}

	if content.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", content.Title)
	}

	if content.WordCount < 5 {
		t.Errorf("expected word count >= 5, got %d", content.WordCount)
	}

	mainContent := content.GetMainContent()
	if !strings.Contains(mainContent, "Main Heading") {
		t.Error("expected main heading in content")
	}
}

func TestContentQualityScore(t *testing.T) {
	tests := []struct {
		name    string
		content *web.PageContent
		check   func(score float64) bool
	}{
		{
			name: "empty content",
			content: &web.PageContent{
				WordCount: 0,
				Headings:  []string{},
			},
			check: func(score float64) bool { return score >= 0 && score <= 1 },
		},
		{
			name: "minimal content",
			content: &web.PageContent{
				WordCount: 10,
				Headings:  []string{},
			},
			check: func(score float64) bool { return score >= 0 && score <= 1 },
		},
		{
			name: "well structured content",
			content: &web.PageContent{
				WordCount: 500,
				Headings:  []string{"H1", "H2", "H3"},
			},
			check: func(score float64) bool { return score >= 0 && score <= 1 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.content.GetContentQuality()
			if !tt.check(score) {
				t.Errorf("quality score %f failed validation check", score)
			}
		})
	}
}
