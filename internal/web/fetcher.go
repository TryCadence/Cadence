package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// PageContent represents fetched website content
type PageContent struct {
	URL         string
	Title       string
	Description string
	Body        string
	AllText     string
	Headers     map[string][]string
	StatusCode  int
	FetchedAt   time.Time
}

// Fetcher handles website content retrieval
type Fetcher struct {
	client  *http.Client
	timeout time.Duration
}

// NewFetcher creates a new website fetcher
func NewFetcher(timeout time.Duration) *Fetcher {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// Fetch retrieves and parses website content
func (f *Fetcher) Fetch(url string) (*PageContent, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	content := &PageContent{
		URL:        url,
		StatusCode: resp.StatusCode,
		FetchedAt:  time.Now(),
		Headers:    resp.Header,
	}

	// Extract title
	content.Title = strings.TrimSpace(doc.Find("title").First().Text())

	// Extract meta description
	metaDesc, _ := doc.Find("meta[name='description']").Attr("content")
	content.Description = metaDesc

	// Extract main body text
	// Remove scripts and styles
	doc.Find("script, style").Remove()

	// Get all text content
	allText := doc.Find("body").Text()
	content.Body = strings.TrimSpace(allText)

	// Also extract h1, h2, h3, p tags for structured analysis
	content.AllText = extractStructuredText(doc)

	return content, nil
}

// extractStructuredText extracts text from important elements
func extractStructuredText(doc *goquery.Document) string {
	var texts []string

	// Headers
	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			texts = append(texts, text)
		}
	})

	// Paragraphs
	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 20 { // Skip very short paragraphs
			texts = append(texts, text)
		}
	})

	// Lists
	doc.Find("li").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			texts = append(texts, text)
		}
	})

	return strings.Join(texts, "\n")
}

// GetMainContent returns the most relevant content for analysis
func (pc *PageContent) GetMainContent() string {
	if pc.AllText != "" {
		return pc.AllText
	}
	return pc.Body
}
