package patterns

import (
	"fmt"
	"strings"
)

type OverusedPhrasesStrategy struct{}

func NewOverusedPhrasesStrategy() *OverusedPhrasesStrategy {
	return &OverusedPhrasesStrategy{}
}

func (s *OverusedPhrasesStrategy) Name() string        { return "overused_phrases" }
func (s *OverusedPhrasesStrategy) Category() string    { return "linguistic" }
func (s *OverusedPhrasesStrategy) Confidence() float64 { return 0.8 }
func (s *OverusedPhrasesStrategy) Description() string {
	return "Detects common AI-generated filler phrases"
}

func (s *OverusedPhrasesStrategy) Detect(content string, wordCount int) *DetectionResult {
	phrases := []string{
		"it is important to note that",
		"it's worth noting that",
		"in conclusion",
		"furthermore",
		"in today's world",
		"in today's digital age",
		"as you know",
		"the power of",
		"next level",
		"best practices",
		"moving forward",
		"at the end of the day",
		"synergy",
		"leveraging",
		"paradigm shift",
		"transformative",
		"revolutionary",
		"innovative approach",
		"seamlessly integrated",
		"cutting-edge",
		"state-of-the-art",
		"game-changing",
		"it's no secret that",
		"the fact of the matter is",
		"at the forefront",
		"driving innovation",
		"unlock the potential",
		"harness the power",
		"in an ever-changing",
		"ever-evolving",
		"comprehensive solution",
		"holistic approach",
		"robust solution",
		"scalable solution",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	foundExamples := make([]string, 0)

	for _, phrase := range phrases {
		occurrences := strings.Count(lowerContent, strings.ToLower(phrase))
		if occurrences > 0 {
			count += occurrences
			if len(foundExamples) < 5 {
				foundExamples = append(foundExamples, phrase)
			}
		}
	}

	if wordCount > 0 && count > 0 && count > wordCount/150 {
		severity := float64(count) / (float64(wordCount) / 100.0)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Excessive use of generic AI phrases (%d instances in %d words)", count, wordCount),
			Examples:    foundExamples,
		}
	}

	return nil
}
