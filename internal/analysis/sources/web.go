package sources

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/web"
)

type WebsiteSource struct {
	URL string
}

func NewWebsiteSource(url string) *WebsiteSource {
	return &WebsiteSource{URL: url}
}

func (w *WebsiteSource) Type() string {
	return "web"
}

func (w *WebsiteSource) Validate(ctx context.Context) error {
	if w.URL == "" {
		return fmt.Errorf("website URL is required")
	}

	_, err := url.Parse(w.URL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}

func (w *WebsiteSource) Fetch(ctx context.Context) (*analysis.SourceData, error) {
	fetcher := web.NewFetcher(30 * time.Second)
	page, err := fetcher.Fetch(w.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch website: %w", err)
	}

	return &analysis.SourceData{
		ID:         w.URL,
		Type:       "web",
		RawContent: page,
		Metadata: map[string]interface{}{
			"url":             w.URL,
			"title":           page.Title,
			"word_count":      page.WordCount,
			"character_count": len(page.AllText),
			"heading_count":   len(page.Headings),
			"headings":        page.Headings,
		},
	}, nil
}
