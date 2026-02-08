package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type FormIssuesStrategy struct{}

func NewFormIssuesStrategy() *FormIssuesStrategy {
	return &FormIssuesStrategy{}
}

func (s *FormIssuesStrategy) Name() string        { return "form_issues" }
func (s *FormIssuesStrategy) Category() string    { return "accessibility" }
func (s *FormIssuesStrategy) Confidence() float64 { return 0.3 }
func (s *FormIssuesStrategy) Description() string {
	return "Detects form inputs missing labels, types, or names"
}

func (s *FormIssuesStrategy) Detect(content string, wordCount int) *DetectionResult {
	lowerContent := strings.ToLower(content)

	inputRegex := regexp.MustCompile(`<input[^>]*>`)
	inputs := inputRegex.FindAllString(lowerContent, -1)

	if len(inputs) == 0 {
		return &DetectionResult{Detected: false}
	}

	inputsWithoutLabels := 0
	inputsWithoutType := 0
	inputsWithoutName := 0

	for _, input := range inputs {
		if !strings.Contains(input, "label") && !strings.Contains(input, "placeholder") {
			inputsWithoutLabels++
		}
		if !strings.Contains(input, `type="`) {
			inputsWithoutType++
		}
		if !strings.Contains(input, `name="`) {
			inputsWithoutName++
		}
	}

	issues := 0
	details := []string{}

	if inputsWithoutLabels > len(inputs)/2 {
		issues++
		details = append(details, fmt.Sprintf("%d inputs without labels", inputsWithoutLabels))
	}

	if inputsWithoutType > 0 {
		issues++
		details = append(details, fmt.Sprintf("%d inputs without type attribute", inputsWithoutType))
	}

	if inputsWithoutName > 0 {
		issues++
		details = append(details, fmt.Sprintf("%d inputs without name attribute", inputsWithoutName))
	}

	if issues > 0 {
		return &DetectionResult{
			Detected:    true,
			Type:        "form_issues",
			Severity:    0.7,
			Description: fmt.Sprintf("Form accessibility issues detected (%d issues)", issues),
			Examples:    details,
		}
	}

	return &DetectionResult{Detected: false}
}
