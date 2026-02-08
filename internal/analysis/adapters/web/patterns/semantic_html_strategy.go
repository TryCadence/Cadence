package patterns

import (
	"fmt"
	"strings"
)

type SemanticHTMLStrategy struct{}

func NewSemanticHTMLStrategy() *SemanticHTMLStrategy {
	return &SemanticHTMLStrategy{}
}

func (s *SemanticHTMLStrategy) Name() string        { return "semantic_html_issues" }
func (s *SemanticHTMLStrategy) Category() string    { return "accessibility" }
func (s *SemanticHTMLStrategy) Confidence() float64 { return 0.3 }
func (s *SemanticHTMLStrategy) Description() string {
	return "Detects overuse of div tags instead of semantic HTML"
}

func (s *SemanticHTMLStrategy) Detect(content string, wordCount int) *DetectionResult {
	lowerContent := strings.ToLower(content)

	divCount := strings.Count(lowerContent, `<div`)
	totalDivs := divCount

	semanticTags := strings.Count(lowerContent, `<nav`) + strings.Count(lowerContent, `<header`) +
		strings.Count(lowerContent, `<section`) + strings.Count(lowerContent, `<article`) +
		strings.Count(lowerContent, `<footer`) + strings.Count(lowerContent, `<main`) +
		strings.Count(lowerContent, `<aside`)

	if totalDivs == 0 {
		return &DetectionResult{Detected: false}
	}

	divRatio := float64(totalDivs) / float64(totalDivs+semanticTags)

	if divRatio > 0.7 {
		return &DetectionResult{
			Detected:    true,
			Type:        "semantic_html_issues",
			Severity:    0.75,
			Description: fmt.Sprintf("Heavy use of divs (%.0f%%) instead of semantic HTML tags (nav, section, article, etc.)", divRatio*100),
			Examples:    []string{"Using <div> for navigation instead of <nav>", "Using <div> for headers instead of <header>"},
		}
	}

	if strings.Count(lowerContent, `<div class="container"`) > 2 {
		return &DetectionResult{
			Detected:    true,
			Type:        "semantic_html_issues",
			Severity:    0.6,
			Description: "Multiple generic container divs detected - consider using semantic HTML",
			Examples:    []string{"Generic <div class=\"container\"> divs found"},
		}
	}

	return &DetectionResult{Detected: false}
}
