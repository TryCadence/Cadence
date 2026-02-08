package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type HardcodedValuesStrategy struct{}

func NewHardcodedValuesStrategy() *HardcodedValuesStrategy {
	return &HardcodedValuesStrategy{}
}

func (s *HardcodedValuesStrategy) Name() string        { return "hardcoded_values" }
func (s *HardcodedValuesStrategy) Category() string    { return "pattern" }
func (s *HardcodedValuesStrategy) Confidence() float64 { return 0.5 }
func (s *HardcodedValuesStrategy) Description() string {
	return "Detects hardcoded inline styles, pixels, and color values"
}

func (s *HardcodedValuesStrategy) Detect(content string, wordCount int) *DetectionResult {
	inlineStyleRegex := regexp.MustCompile(`style="[^"]*(?:width|height|color|font-size):[^"]*"`)
	inlineStyles := inlineStyleRegex.FindAllString(strings.ToLower(content), -1)

	pixelRegex := regexp.MustCompile(`\b\d+px\b`)
	pixelValues := pixelRegex.FindAllString(content, -1)

	hardcodedColors := strings.Count(strings.ToLower(content), "color:") +
		strings.Count(strings.ToLower(content), "background-color:") +
		strings.Count(strings.ToLower(content), "background:")

	hardcodedIssues := 0

	if len(inlineStyles) > 3 {
		hardcodedIssues++
	}

	if len(pixelValues) > 10 {
		hardcodedIssues++
	}

	if hardcodedColors > 5 {
		hardcodedIssues++
	}

	if hardcodedIssues > 0 {
		severity := float64(hardcodedIssues) / 3.0
		if severity > 1 {
			severity = 1
		}

		return &DetectionResult{
			Detected:    true,
			Type:        "hardcoded_values",
			Severity:    severity,
			Description: fmt.Sprintf("Hardcoded dimensions, colors, and sizes detected (%d patterns) - not responsive", hardcodedIssues*2),
			Examples:    []string{"Inline styles with fixed pixel values", "Hardcoded colors instead of CSS variables", "No responsive breakpoints"},
		}
	}

	return &DetectionResult{Detected: false}
}
