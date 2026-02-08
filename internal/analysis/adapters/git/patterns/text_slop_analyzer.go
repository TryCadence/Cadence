package patterns

import (
	"fmt"
	"strings"

	webpatterns "github.com/TryCadence/Cadence/internal/analysis/adapters/web/patterns"
)

type TextSlopAnalyzer struct {
	enabled  bool
	registry *webpatterns.WebPatternRegistry
}

func NewTextSlopAnalyzer() *TextSlopAnalyzer {
	return &TextSlopAnalyzer{
		enabled:  true,
		registry: webpatterns.NewWebPatternRegistry(),
	}
}

func (a *TextSlopAnalyzer) AnalyzeContent(content string) (*TextSlopResult, error) {
	if content == "" {
		return nil, fmt.Errorf("empty content")
	}

	wordCount := len(strings.Fields(content))
	if wordCount < 50 {
		return nil, fmt.Errorf("content too short for reliable analysis (minimum 50 words, got %d)", wordCount)
	}

	result := &TextSlopResult{
		Patterns:       make([]Pattern, 0),
		PassedPatterns: make([]Pattern, 0),
		SuspicionRate:  0,
		WordCount:      wordCount,
	}

	// Build confidence lookup from registered strategies
	confidenceLookup := make(map[string]float64)
	for _, strategy := range a.registry.GetStrategies() {
		confidenceLookup[strategy.Name()] = strategy.Confidence()
	}

	// Get all results including passed ones
	allResults := a.registry.DetectAllWithPassed(content, wordCount)

	for _, dr := range allResults {
		confidence := confidenceLookup[dr.Type]
		if confidence == 0 {
			confidence = 0.5 // default for unknown strategies
		}

		pattern := Pattern{
			Type:        dr.Type,
			Severity:    dr.Severity,
			Confidence:  confidence,
			Description: dr.Description,
			Examples:    dr.Examples,
		}

		if dr.Detected {
			result.Patterns = append(result.Patterns, pattern)
		} else {
			result.PassedPatterns = append(result.PassedPatterns, pattern)
		}
	}

	if len(result.Patterns) > 0 {
		totalSeverity := 0.0
		for _, p := range result.Patterns {
			totalSeverity += p.Severity
		}
		result.SuspicionRate = totalSeverity / float64(len(result.Patterns))
		if result.SuspicionRate > 1.0 {
			result.SuspicionRate = 1.0
		}
	}

	return result, nil
}

func (a *TextSlopAnalyzer) RegisterStrategy(strategy webpatterns.WebPatternStrategy) {
	a.registry.Register(strategy)
}

func (a *TextSlopAnalyzer) GetRegistry() *webpatterns.WebPatternRegistry {
	return a.registry
}

type Pattern struct {
	Type        string
	Severity    float64
	Confidence  float64
	Description string
	Examples    []string
}

type TextSlopResult struct {
	Patterns       []Pattern
	PassedPatterns []Pattern
	SuspicionRate  float64
	Summary        string
	WordCount      int
}

func (r *TextSlopResult) GetConfidenceScore() int {
	if r.SuspicionRate == 0 {
		return 0
	}
	score := int(r.SuspicionRate * 100)
	if score > 100 {
		score = 100
	}
	return score
}

func (r *TextSlopResult) GetSummary() string {
	if len(r.Patterns) == 0 {
		return "No AI-generated content indicators detected"
	}

	score := r.GetConfidenceScore()
	var summary string

	switch {
	case score >= 70:
		summary = "LIKELY AI-GENERATED: Strong indicators of AI-generated text"
	case score >= 50:
		summary = "POSSIBLY AI-GENERATED: Multiple indicators suggest AI generation"
	case score >= 30:
		summary = "SUSPICIOUS: Some AI-like patterns detected"
	default:
		summary = "LIKELY HUMAN-WRITTEN: Minimal AI indicators"
	}

	summary += fmt.Sprintf(" (confidence: %d%%)\n\nDetected patterns:\n", score)

	for i, p := range r.Patterns {
		summary += fmt.Sprintf("%d. %s (severity: %.1f%%)\n   %s\n", i+1, p.Type, p.Severity*100, p.Description)
	}

	return summary
}
