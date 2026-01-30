package patterns

import (
	"fmt"
	"strings"
)

type GenericLanguageStrategy struct{}

func NewGenericLanguageStrategy() *GenericLanguageStrategy {
	return &GenericLanguageStrategy{}
}

func (s *GenericLanguageStrategy) Name() string {
	return "generic_language"
}

func (s *GenericLanguageStrategy) Detect(content string, wordCount int) *DetectionResult {
	genericTerms := []string{
		"the user", "the customer", "the client",
		"provide value", "add value", "deliver value",
		"various", "multiple", "diverse",
		"ensure", "maximize", "optimize",
		"unique", "unique solution",
		"stakeholder", "utilize", "implement",
		"enhance", "empower", "enable",
		"streamline", "facilitate",
		"comprehensive", "robust", "scalable",
		"seamless", "intuitive",
		"world-class", "best-in-class",
		"mission-critical",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	foundExamples := make([]string, 0)

	for _, term := range genericTerms {
		occurrences := strings.Count(lowerContent, strings.ToLower(term))
		if occurrences > 0 {
			count += occurrences
			if len(foundExamples) < 5 {
				foundExamples = append(foundExamples, term)
			}
		}
	}

	if wordCount > 0 && count > wordCount/100 {
		severity := float64(count) / (float64(wordCount) / 80.0)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Excessive use of generic business language (%d instances in %d words)", count, wordCount),
			Examples:    foundExamples,
		}
	}

	return nil
}
