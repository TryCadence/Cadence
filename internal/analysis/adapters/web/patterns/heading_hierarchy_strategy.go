package patterns

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type HeadingHierarchyStrategy struct{}

func NewHeadingHierarchyStrategy() *HeadingHierarchyStrategy {
	return &HeadingHierarchyStrategy{}
}

func (s *HeadingHierarchyStrategy) Name() string        { return "heading_hierarchy_issues" }
func (s *HeadingHierarchyStrategy) Category() string    { return "structural" }
func (s *HeadingHierarchyStrategy) Confidence() float64 { return 0.4 }
func (s *HeadingHierarchyStrategy) Description() string {
	return "Detects improper heading level order and hierarchy"
}

func (s *HeadingHierarchyStrategy) Detect(content string, wordCount int) *DetectionResult {
	headingRegex := regexp.MustCompile(`<h([1-6])`)
	matches := headingRegex.FindAllStringSubmatch(strings.ToLower(content), -1)

	if len(matches) < 2 {
		return &DetectionResult{Detected: false}
	}

	headingLevels := make([]int, 0, len(matches))
	for _, match := range matches {
		level, _ := strconv.Atoi(match[1])
		headingLevels = append(headingLevels, level)
	}

	issues := 0

	if len(headingLevels) > 0 && headingLevels[0] != 1 {
		issues++
	}

	for i := 1; i < len(headingLevels); i++ {
		diff := headingLevels[i] - headingLevels[i-1]
		if diff > 1 {
			issues++
		}
	}

	if issues > 0 {
		return &DetectionResult{
			Detected:    true,
			Type:        "heading_hierarchy_issues",
			Severity:    0.65,
			Description: fmt.Sprintf("Improper heading hierarchy detected (%d issues found)", issues),
			Examples:    []string{"Found skipped heading levels", "Document should start with <h1>"},
		}
	}

	return &DetectionResult{Detected: false}
}
