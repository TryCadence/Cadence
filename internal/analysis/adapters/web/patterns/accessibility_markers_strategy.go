package patterns

import (
	"fmt"
	"strings"
)

type AccessibilityMarkersStrategy struct{}

func NewAccessibilityMarkersStrategy() *AccessibilityMarkersStrategy {
	return &AccessibilityMarkersStrategy{}
}

func (s *AccessibilityMarkersStrategy) Name() string        { return "accessibility_markers" }
func (s *AccessibilityMarkersStrategy) Category() string    { return "accessibility" }
func (s *AccessibilityMarkersStrategy) Confidence() float64 { return 0.3 }
func (s *AccessibilityMarkersStrategy) Description() string {
	return "Detects missing accessibility markers and ARIA attributes"
}

func (s *AccessibilityMarkersStrategy) Detect(content string, wordCount int) *DetectionResult {
	lowerContent := strings.ToLower(content)

	hasAriaLabels := strings.Contains(lowerContent, "aria-label")
	hasAriaDescribed := strings.Contains(lowerContent, "aria-describedby")
	hasAriaHidden := strings.Contains(lowerContent, "aria-hidden")
	hasRole := strings.Contains(lowerContent, `role="`)
	hasLang := strings.Contains(lowerContent, `lang="`)

	accessibilityMarkers := 0
	if hasAriaLabels {
		accessibilityMarkers++
	}
	if hasAriaDescribed {
		accessibilityMarkers++
	}
	if hasAriaHidden {
		accessibilityMarkers++
	}
	if hasRole {
		accessibilityMarkers++
	}
	if hasLang {
		accessibilityMarkers++
	}

	if accessibilityMarkers == 0 {
		return &DetectionResult{
			Detected:    true,
			Type:        "accessibility_markers",
			Severity:    0.7,
			Description: "No accessibility markers found (aria-labels, roles, lang attributes) - potential accessibility issues",
			Examples:    []string{"Missing aria-label attributes", "Missing role attributes", "Missing lang attribute on <html>"},
		}
	}

	if accessibilityMarkers < 2 {
		return &DetectionResult{
			Detected:    true,
			Type:        "accessibility_markers",
			Severity:    0.5,
			Description: fmt.Sprintf("Limited accessibility markers detected (%d/5 types found)", accessibilityMarkers),
			Examples:    []string{"Consider adding more aria attributes and role definitions"},
		}
	}

	return &DetectionResult{Detected: false}
}
