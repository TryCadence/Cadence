package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type LinkTextQualityStrategy struct{}

func NewLinkTextQualityStrategy() *LinkTextQualityStrategy {
	return &LinkTextQualityStrategy{}
}

func (s *LinkTextQualityStrategy) Name() string        { return "link_text_quality" }
func (s *LinkTextQualityStrategy) Category() string    { return "accessibility" }
func (s *LinkTextQualityStrategy) Confidence() float64 { return 0.4 }
func (s *LinkTextQualityStrategy) Description() string {
	return "Detects generic or non-descriptive link text"
}

func (s *LinkTextQualityStrategy) Detect(content string, wordCount int) *DetectionResult {
	linkRegex := regexp.MustCompile(`<a[^>]*href[^>]*>([^<]+)</a>`)
	matches := linkRegex.FindAllStringSubmatch(strings.ToLower(content), -1)

	if len(matches) == 0 {
		return &DetectionResult{Detected: false}
	}

	genericLinkTexts := []string{"click here", "read more", "learn more", "more", "link", "here", "this", "go", "next"}
	genericLinks := 0

	for _, match := range matches {
		if len(match) > 1 {
			linkText := strings.TrimSpace(match[1])
			for _, generic := range genericLinkTexts {
				if linkText == generic || linkText == generic+"." {
					genericLinks++
					break
				}
			}
		}
	}

	if genericLinks > 0 {
		ratio := float64(genericLinks) / float64(len(matches))
		severity := ratio * 0.8
		if severity > 1 {
			severity = 1
		}

		return &DetectionResult{
			Detected:    true,
			Type:        "link_text_quality",
			Severity:    severity,
			Description: fmt.Sprintf("Generic link text detected (%d of %d links)", genericLinks, len(matches)),
			Examples:    []string{"Found 'click here', 'read more', or other generic link text"},
		}
	}

	return &DetectionResult{Detected: false}
}
