package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	client  *http.Client
	timeout time.Duration
}

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
