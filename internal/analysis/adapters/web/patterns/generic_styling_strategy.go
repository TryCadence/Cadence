package patterns

import (
	"fmt"
	"strings"
)

type GenericStylingStrategy struct{}

func NewGenericStylingStrategy() *GenericStylingStrategy {
	return &GenericStylingStrategy{}
}

func (s *GenericStylingStrategy) Name() string        { return "generic_styling" }
func (s *GenericStylingStrategy) Category() string    { return "pattern" }
func (s *GenericStylingStrategy) Confidence() float64 { return 0.4 }
func (s *GenericStylingStrategy) Description() string {
	return "Detects lack of CSS variables, theming, and overuse of inline styles"
}

func (s *GenericStylingStrategy) Detect(content string, wordCount int) *DetectionResult {
	lowerContent := strings.ToLower(content)

	hasCSSVars := strings.Contains(lowerContent, "var(--") || strings.Contains(lowerContent, "css")
	hasTheme := strings.Contains(lowerContent, "theme") || strings.Contains(lowerContent, "dark") || strings.Contains(lowerContent, "light")
	hasMediaQueries := strings.Contains(lowerContent, "@media")
	hasCustomClasses := strings.Count(lowerContent, "class=") > 10

	defaultColorUsage := strings.Count(lowerContent, "color:black") +
		strings.Count(lowerContent, "color:white") +
		strings.Count(lowerContent, "color:#000") +
		strings.Count(lowerContent, "color:#fff")

	genericIndicators := 0

	if !hasCSSVars && !hasTheme {
		genericIndicators++
	}

	if !hasMediaQueries && wordCount > 500 {
		genericIndicators++
	}

	if defaultColorUsage > 5 {
		genericIndicators++
	}

	if !hasCustomClasses && strings.Count(lowerContent, `style="`) > 5 {
		genericIndicators++
	}

	if genericIndicators >= 2 {
		severity := float64(genericIndicators) / 4.0
		if severity > 1 {
			severity = 1
		}

		return &DetectionResult{
			Detected:    true,
			Type:        "generic_styling",
			Severity:    severity,
			Description: fmt.Sprintf("Generic/default styling detected (%d indicators) - no custom theming", genericIndicators),
			Examples:    []string{"Using default colors instead of CSS variables", "No responsive breakpoints", "No theme support"},
		}
	}

	return &DetectionResult{Detected: false}
}
