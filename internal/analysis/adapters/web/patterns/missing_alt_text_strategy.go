package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type MissingAltTextStrategy struct{}

func NewMissingAltTextStrategy() *MissingAltTextStrategy {
	return &MissingAltTextStrategy{}
}

func (s *MissingAltTextStrategy) Name() string        { return "missing_alt_text" }
func (s *MissingAltTextStrategy) Category() string    { return "accessibility" }
func (s *MissingAltTextStrategy) Confidence() float64 { return 0.3 }
func (s *MissingAltTextStrategy) Description() string {
	return "Detects images missing alt text attributes"
}

func (s *MissingAltTextStrategy) Detect(content string, wordCount int) *DetectionResult {
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	images := imgRegex.FindAllString(content, -1)

	if len(images) == 0 {
		return &DetectionResult{Detected: false}
	}

	missingAlt := 0
	emptyAlt := 0

	for _, img := range images {
		if !strings.Contains(strings.ToLower(img), "alt=") {
			missingAlt++
		} else {
			altRegex := regexp.MustCompile(`alt=["']([^"']*)["']`)
			matches := altRegex.FindStringSubmatch(img)
			if len(matches) > 1 && strings.TrimSpace(matches[1]) == "" {
				emptyAlt++
			}
		}
	}

	totalMissing := missingAlt + emptyAlt
	if totalMissing > 0 {
		severity := float64(totalMissing) / float64(len(images))
		if severity > 1 {
			severity = 1
		}

		return &DetectionResult{
			Detected:    true,
			Type:        "missing_alt_text",
			Severity:    severity,
			Description: fmt.Sprintf("%d of %d images missing or have empty alt text (accessibility issue)", totalMissing, len(images)),
			Examples:    []string{fmt.Sprintf("Found %d images without alt attributes", missingAlt)},
		}
	}

	return &DetectionResult{Detected: false}
}
