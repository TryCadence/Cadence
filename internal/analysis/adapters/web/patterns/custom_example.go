package patterns

import (
	"fmt"
	"strings"
)

type CustomPatternStrategy struct {
	name      string
	keywords  []string
	threshold int
}

func NewCustomPatternStrategy(name string, keywords []string, threshold int) *CustomPatternStrategy {
	return &CustomPatternStrategy{
		name:      name,
		keywords:  keywords,
		threshold: threshold,
	}
}

func (s *CustomPatternStrategy) Name() string        { return s.name }
func (s *CustomPatternStrategy) Category() string    { return "pattern" }
func (s *CustomPatternStrategy) Confidence() float64 { return 0.5 }
func (s *CustomPatternStrategy) Description() string {
	return "Custom keyword-matching pattern strategy"
}

func (s *CustomPatternStrategy) Detect(content string, wordCount int) *DetectionResult {
	lowerContent := strings.ToLower(content)
	count := 0
	foundKeywords := make([]string, 0)

	for _, keyword := range s.keywords {
		if strings.Contains(lowerContent, keyword) {
			count++
			if len(foundKeywords) < 5 {
				foundKeywords = append(foundKeywords, keyword)
			}
		}
	}

	if count >= s.threshold {
		severity := float64(count) / float64(len(s.keywords))
		if severity > 1.0 {
			severity = 1.0
		}

		return &DetectionResult{
			Detected:    true,
			Type:        s.name,
			Severity:    severity,
			Description: fmt.Sprintf("Detected %d instances of %s pattern keywords", count, s.name),
			Examples:    foundKeywords,
		}
	}

	return nil
}
