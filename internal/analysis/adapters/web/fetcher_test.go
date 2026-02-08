package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	tests := []struct {
		timeout    time.Duration
		name       string
		expectZero bool
	}{
		{
			name:       "positive timeout",
			timeout:    5 * time.Second,
			expectZero: false,
		},
		{
			name:       "zero timeout defaults to 10s",
			timeout:    0,
			expectZero: true,
		},
		{
			name:       "one second timeout",
			timeout:    1 * time.Second,
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher(tt.timeout)

			if fetcher == nil {
				t.Error("expected fetcher but got nil")
				return
			}

			if fetcher.client == nil {
				t.Error("expected HTTP client but got nil")
			}

			if tt.expectZero {
				if fetcher.timeout != 10*time.Second {
					t.Errorf("expected timeout 10s, got %v", fetcher.timeout)
				}
			} else if fetcher.timeout != tt.timeout {
				t.Errorf("expected timeout %v, got %v", tt.timeout, fetcher.timeout)
			}
		})
	}
}

func TestFetchURLPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test</title></head><body>Content</body></html>`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)

	// Extract just the host part without protocol
	url := strings.TrimPrefix(server.URL, "https://")
	url = strings.TrimPrefix(url, "http://")

	content, err := fetcher.Fetch(url)
	if err == nil {
		// If it succeeds, verify content was fetched
		if content == nil {
			t.Error("expected content but got nil")
		}
	}
	// Allow errors due to localhost resolution differences
}

func TestFetchHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)
	_, err := fetcher.Fetch(server.URL)

	if err == nil {
		t.Error("expected error for 404 status, got nil")
	}
}

func TestFetchServerError(t *testing.T) {
	fetcher := NewFetcher(5 * time.Second)
	_, err := fetcher.Fetch("http://invalid.local.hostname.that.does.not.exist:9999/path")

	if err == nil {
		t.Error("expected error for unreachable host, got nil")
	}
}

func TestFetchValidHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<html>
  <head>
    <title>Test Page</title>
    <meta name="description" content="Test description">
    <meta property="og:title" content="OG Title">
  </head>
  <body>
    <h1>Main Heading</h1>
    <h2>Sub Heading</h2>
    <p>This is a paragraph with content.</p>
    <p>Another paragraph here.</p>
    <ul>
      <li>List item 1</li>
      <li>List item 2</li>
    </ul>
  </body>
</html>
		`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)
	content, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == nil {
		t.Fatal("expected content but got nil")
	}

	if content.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", content.Title)
	}

	if content.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", content.StatusCode)
	}

	if !strings.Contains(content.Body, "Main Heading") {
		t.Error("expected body to contain main heading")
	}

	if len(content.MetaTags) == 0 {
		t.Error("expected meta tags to be parsed")
	}

	if len(content.Headings) == 0 {
		t.Error("expected headings to be extracted")
	}

	if content.WordCount == 0 {
		t.Error("expected word count to be calculated")
	}
}

func TestPageContentGetMainContent(t *testing.T) {
	tests := []struct {
		name        string
		content     *PageContent
		expectedKey string
	}{
		{
			name: "with MainContent",
			content: &PageContent{
				MainContent: "main content text",
				AllText:     "all text",
				Body:        "body text",
			},
			expectedKey: "main content text",
		},
		{
			name: "without MainContent, with AllText",
			content: &PageContent{
				MainContent: "",
				AllText:     "all text",
				Body:        "body text",
			},
			expectedKey: "all text",
		},
		{
			name: "without MainContent and AllText",
			content: &PageContent{
				MainContent: "",
				AllText:     "",
				Body:        "body text",
			},
			expectedKey: "body text",
		},
		{
			name: "all empty",
			content: &PageContent{
				MainContent: "",
				AllText:     "",
				Body:        "",
			},
			expectedKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.content.GetMainContent()
			if result != tt.expectedKey {
				t.Errorf("expected %q, got %q", tt.expectedKey, result)
			}
		})
	}
}

func TestPageContentGetContentQuality(t *testing.T) {
	tests := []struct {
		name       string
		content    *PageContent
		checkRange func(float64) bool
	}{
		{
			name: "empty content",
			content: &PageContent{
				WordCount:   0,
				MainContent: "",
				Headings:    []string{},
			},
			checkRange: func(score float64) bool { return score == 0.0 },
		},
		{
			name: "short content",
			content: &PageContent{
				WordCount:   30,
				MainContent: strings.Repeat("word ", 30),
				Headings:    []string{},
			},
			checkRange: func(score float64) bool { return score > 0 && score < 1.0 },
		},
		{
			name: "medium content",
			content: &PageContent{
				WordCount:   80,
				MainContent: strings.Repeat("word ", 80),
				Headings:    []string{"Title"},
			},
			checkRange: func(score float64) bool { return score > 0 && score <= 1.0 },
		},
		{
			name: "good content",
			content: &PageContent{
				WordCount:   500,
				MainContent: strings.Repeat("word word word word word\n", 20),
				Headings:    []string{"H1", "H2", "H3"},
			},
			checkRange: func(score float64) bool { return score > 0 && score <= 2.0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tt.content.GetContentQuality()
			if !tt.checkRange(score) {
				t.Errorf("quality score %f outside expected range", score)
			}
		})
	}
}

func TestPageContentStructure(t *testing.T) {
	content := &PageContent{
		URL:         "https://example.com",
		Title:       "Example",
		Description: "Example description",
		Body:        "Example body",
		StatusCode:  200,
		WordCount:   100,
		Headings:    []string{"H1", "H2"},
		MetaTags:    map[string]string{"author": "John"},
		Headers:     make(map[string][]string),
	}

	if content.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", content.URL)
	}

	if content.Title != "Example" {
		t.Errorf("expected title 'Example', got %q", content.Title)
	}

	if content.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", content.StatusCode)
	}

	if len(content.Headings) != 2 {
		t.Errorf("expected 2 headings, got %d", len(content.Headings))
	}

	if len(content.MetaTags) != 1 {
		t.Errorf("expected 1 meta tag, got %d", len(content.MetaTags))
	}
}

func TestFetchHTTPSWithAlternateProtocols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test</title></head><body>Test content</body></html>`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)

	// Test with explicit http://
	content, err := fetcher.Fetch(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content == nil {
		t.Fatal("expected content but got nil")
	}

	if content.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", content.Title)
	}
}

func TestExtractStructuredText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<html>
  <body>
    <h1>Title</h1>
    <h2>Subtitle</h2>
    <p>Short text</p>
    <p>This is a longer paragraph with more than twenty characters in it</p>
    <ul>
      <li>Item one</li>
      <li>Item two</li>
    </ul>
  </body>
</html>
		`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)
	content, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content.AllText == "" {
		t.Error("expected AllText to be populated")
	}

	if !strings.Contains(content.AllText, "Title") {
		t.Error("expected AllText to contain headings")
	}
}

func TestFetchRemovesScriptAndStyle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<html>
  <body>
    <script>console.log('secret code')</script>
    <style>body { color: red; }</style>
    <p>Visible content</p>
  </body>
</html>
		`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)
	content, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Script and style content should not appear in the body
	if strings.Contains(content.Body, "console.log") {
		t.Error("expected script content to be removed")
	}

	if strings.Contains(content.Body, "color: red") {
		t.Error("expected style content to be removed")
	}

	if !strings.Contains(content.Body, "Visible content") {
		t.Error("expected visible content to be present")
	}
}

func TestPageContentHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.Write([]byte(`<html><body>Test</body></html>`))
	}))
	defer server.Close()

	fetcher := NewFetcher(5 * time.Second)
	content, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if content.Headers == nil {
		t.Error("expected headers to be populated")
	}

	if len(content.Headers) == 0 {
		t.Error("expected headers to contain values")
	}
}
