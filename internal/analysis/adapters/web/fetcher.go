package web

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	cerrors "github.com/TryCadence/Cadence/internal/errors"
)

type PageContent struct {
	URL         string
	Title       string
	Description string
	Body        string
	AllText     string
	Headers     map[string][]string
	StatusCode  int
	FetchedAt   time.Time
	MainContent string
	WordCount   int
	MetaTags    map[string]string
	Headings    []string
}

type Fetcher struct {
	client     *http.Client
	timeout    time.Duration
	maxRetries int
}

func NewFetcher(timeout time.Duration) *Fetcher {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Fetcher{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout:    timeout,
		maxRetries: 3,
	}
}

// isRetryableStatus returns true for HTTP status codes that indicate a
// transient failure that may succeed on retry.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func (f *Fetcher) Fetch(url string) (*PageContent, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	var lastErr error
	for attempt := 0; attempt <= f.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s, 2s
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * 500 * time.Millisecond
			time.Sleep(backoff)
		}

		content, err := f.doFetch(url)
		if err == nil {
			return content, nil
		}
		lastErr = err

		// Only retry on retryable errors (network or retryable status codes)
		if cerr, ok := err.(*cerrors.CadenceError); ok {
			if strings.Contains(cerr.Details, "returned") {
				if !isRetryableStatusFromDetails(cerr.Details) {
					return nil, err // Non-retryable status (e.g., 404), fail immediately
				}
			}
		}
	}

	return nil, cerrors.IOError("failed to fetch URL after retries").WithDetails(
		fmt.Sprintf("%s (attempts: %d)", url, f.maxRetries+1),
	).Wrap(lastErr)
}

// isRetryableStatusFromDetails checks if an error detail string indicates a retryable HTTP status.
func isRetryableStatusFromDetails(details string) bool {
	retryableCodes := []string{"429", "500", "502", "503", "504"}
	for _, code := range retryableCodes {
		if strings.Contains(details, code) {
			return true
		}
	}
	return false
}

func (f *Fetcher) doFetch(url string) (*PageContent, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, cerrors.IOError("failed to fetch URL").WithDetails(url).Wrap(err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, cerrors.IOError("unexpected status code").WithDetails(fmt.Sprintf("%s returned %d", url, resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, cerrors.IOError("failed to read response body").Wrap(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, cerrors.IOError("failed to parse HTML").Wrap(err)
	}

	content := &PageContent{
		URL:        url,
		StatusCode: resp.StatusCode,
		FetchedAt:  time.Now(),
		Headers:    resp.Header,
		MetaTags:   make(map[string]string),
		Headings:   make([]string, 0),
	}

	content.Title = strings.TrimSpace(doc.Find("title").First().Text())

	metaDesc, _ := doc.Find("meta[name='description']").Attr("content")
	content.Description = metaDesc
	doc.Find("meta").Each(func(_ int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if metaContent, exists := s.Attr("content"); exists {
				content.MetaTags[name] = metaContent
			}
		}
		if property, exists := s.Attr("property"); exists {
			if metaContent, exists := s.Attr("content"); exists {
				content.MetaTags[property] = metaContent
			}
		}
	})

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			content.Headings = append(content.Headings, text)
		}
	})

	doc.Find("script, style, nav, header, footer, aside, .ad, .advertisement, .sidebar, .menu, .navigation, #comments, .comment").Remove()

	allText := doc.Find("body").Text()
	content.Body = strings.TrimSpace(allText)
	content.AllText = extractStructuredText(doc)
	content.MainContent = extractMainContent(doc)
	content.WordCount = len(strings.Fields(content.MainContent))

	return content, nil
}

func extractStructuredText(doc *goquery.Document) string {
	var texts []string

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			texts = append(texts, text)
		}
	})

	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 20 {
			texts = append(texts, text)
		}
	})

	doc.Find("li").Each(func(_ int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			texts = append(texts, text)
		}
	})

	return strings.Join(texts, "\n")
}

func extractMainContent(doc *goquery.Document) string {
	var texts []string

	mainContent := doc.Find("main, article, [role='main'], .main-content, #main-content, .content, #content")
	if mainContent.Length() > 0 {
		mainContent.Find("h1, h2, h3, h4, h5, h6, p, li, blockquote").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 15 {
				texts = append(texts, text)
			}
		})
	}

	if len(texts) == 0 {
		doc.Find("h1, h2, h3, p").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 20 {
				texts = append(texts, text)
			}
		})
	}

	return strings.Join(texts, "\n\n")
}

func (pc *PageContent) GetMainContent() string {
	if pc.MainContent != "" {
		return pc.MainContent
	}
	if pc.AllText != "" {
		return pc.AllText
	}
	return pc.Body
}

func (pc *PageContent) GetContentQuality() float64 {
	content := pc.GetMainContent()
	if content == "" {
		return 0.0
	}

	score := 1.0

	if pc.WordCount < 50 {
		score *= 0.5
	} else if pc.WordCount < 100 {
		score *= 0.7
	}

	if len(pc.Headings) > 0 {
		score *= 1.1
	}

	lines := strings.Split(content, "\n")
	shortLines := 0
	for _, line := range lines {
		if len(strings.TrimSpace(line)) < 30 {
			shortLines++
		}
	}
	if len(lines) > 0 && float64(shortLines)/float64(len(lines)) > 0.7 {
		score *= 0.6
	}

	return score
}
