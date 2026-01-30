package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type TextSlopAnalyzer struct {
	enabled bool
}

func NewTextSlopAnalyzer() *TextSlopAnalyzer {
	return &TextSlopAnalyzer{enabled: true}
}

func (a *TextSlopAnalyzer) AnalyzeContent(content string) (*TextSlopResult, error) {
	if content == "" {
		return nil, fmt.Errorf("empty content")
	}

	result := &TextSlopResult{
		Patterns:      make([]Pattern, 0),
		SuspicionRate: 0,
	}

	result.addPattern(a.detectOverusedPhrases(content))
	result.addPattern(a.detectGenericLanguage(content))
	result.addPattern(a.detectExcessiveStructure(content))
	result.addPattern(a.detectPerfectGrammar(content))
	result.addPattern(a.detectBoilerplateText(content))
	result.addPattern(a.detectRepetitivePatterns(content))
	result.addPattern(a.detectMissingNuance(content))
	result.addPattern(a.detectExcessiveTransitions(content))

	if len(result.Patterns) > 0 {
		result.SuspicionRate = float64(len(result.Patterns)) / 8.0 // 8 detectors
		if result.SuspicionRate > 1.0 {
			result.SuspicionRate = 1.0
		}
	}

	return result, nil
}

func (a *TextSlopAnalyzer) detectOverusedPhrases(content string) *Pattern {
	phrases := []string{
		"it is important to note that",
		"in conclusion",
		"furthermore",
		"in today's world",
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
	}

	lowerContent := strings.ToLower(content)
	count := 0
	for _, phrase := range phrases {
		count += strings.Count(lowerContent, strings.ToLower(phrase))
	}

	contentWords := len(strings.Fields(content))
	if contentWords > 0 && count > contentWords/200 {
		return &Pattern{
			Type:        "overused_phrases",
			Severity:    float64(count) / float64(contentWords),
			Description: fmt.Sprintf("Excessive use of generic AI phrases (%d instances)", count),
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectGenericLanguage(content string) *Pattern {
	genericTerms := []string{
		"the user", "the customer", "the client",
		"provide value", "add value",
		"various", "multiple", "diverse",
		"ensure", "maximize", "optimize",
		"unique", "unique solution",
		"stakeholder", "utilize", "implement",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	for _, term := range genericTerms {
		count += strings.Count(lowerContent, strings.ToLower(term))
	}

	contentWords := len(strings.Fields(content))
	if contentWords > 0 && count > contentWords/150 {
		return &Pattern{
			Type:        "generic_language",
			Severity:    float64(count) / float64(contentWords),
			Description: fmt.Sprintf("Excessive use of generic business language (%d instances)", count),
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectExcessiveStructure(content string) *Pattern {
	lines := strings.Split(content, "\n")
	bulletPoints := 0
	numberedLists := 0

	bulletPattern := regexp.MustCompile(`^[\s]*[-*•]\s+`)
	numberPattern := regexp.MustCompile(`^[\s]*\d+\.\s+`)

	for _, line := range lines {
		if bulletPattern.MatchString(line) {
			bulletPoints++
		}
		if numberPattern.MatchString(line) {
			numberedLists++
		}
	}

	structuredItems := bulletPoints + numberedLists
	if len(lines) > 10 && structuredItems > len(lines)/3 {
		return &Pattern{
			Type:        "excessive_structure",
			Severity:    float64(structuredItems) / float64(len(lines)),
			Description: fmt.Sprintf("Excessive list/bullet structure (%d items in %d lines) - typical AI formatting", structuredItems, len(lines)),
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectPerfectGrammar(content string) *Pattern {
	hasContractions := strings.Contains(content, "'ll") || strings.Contains(content, "'ve") ||
		strings.Contains(content, "'re") || strings.Contains(content, "'d") ||
		strings.Contains(content, "don't") || strings.Contains(content, "can't") ||
		strings.Contains(content, "won't")

	hasColloquialisms := strings.Contains(content, "kinda") || strings.Contains(content, "sorta") ||
		strings.Contains(content, "gonna") || strings.Contains(content, "wanna") ||
		strings.Contains(content, "ain't") || strings.Contains(content, "y'all") ||
		strings.Contains(content, "hmm") || strings.Contains(content, "err")

	sentences := strings.Split(content, ". ")
	wellCapitalizedCount := 0
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 0 && trimmed[0] >= 'A' && trimmed[0] <= 'Z' {
			wellCapitalizedCount++
		}
	}

	if len(sentences) > 5 && !hasContractions && !hasColloquialisms &&
		float64(wellCapitalizedCount)/float64(len(sentences)) > 0.95 {
		return &Pattern{
			Type:        "perfect_grammar",
			Severity:    0.6,
			Description: "Suspiciously perfect grammar and formatting - lack of natural language variation",
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectBoilerplateText(content string) *Pattern {
	boilerplate := []string{
		"our mission is",
		"we are committed to",
		"we believe in",
		"our vision is",
		"our team is dedicated",
		"with over",
		"industry-leading",
		"award-winning",
		"trusted by",
		"proven track record",
		"customer satisfaction",
		"quality assurance",
		"best in class",
		"trusted partner",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	for _, phrase := range boilerplate {
		count += strings.Count(lowerContent, strings.ToLower(phrase))
	}

	if count >= 3 {
		return &Pattern{
			Type:        "boilerplate_text",
			Severity:    float64(count) / 10.0,
			Description: fmt.Sprintf("Boilerplate/marketing language detected (%d instances)", count),
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectRepetitivePatterns(content string) *Pattern {
	sentences := strings.Split(content, ". ")
	if len(sentences) < 5 {
		return nil
	}

	startWords := make(map[string]int)
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		words := strings.Fields(trimmed)
		if len(words) > 0 {
			startWords[strings.ToLower(words[0])]++
		}
	}

	for _, count := range startWords {
		if float64(count)/float64(len(sentences)) > 0.3 {
			return &Pattern{
				Type:        "repetitive_patterns",
				Severity:    0.7,
				Description: "Repetitive sentence structures - common in generated text",
			}
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectMissingNuance(content string) *Pattern {
	vaguePhrases := []string{
		"etc", "and so on", "for example", "like", "such as",
		"basically", "essentially", "literally", "arguably",
	}

	hasQuotes := strings.Count(content, "\"") >= 2
	hasReferences := strings.Contains(content, "http") || strings.Contains(content, "www")

	lowerContent := strings.ToLower(content)
	vagueCount := 0
	for _, phrase := range vaguePhrases {
		vagueCount += strings.Count(lowerContent, strings.ToLower(phrase))
	}

	contentWords := len(strings.Fields(content))
	if contentWords > 100 && !hasQuotes && !hasReferences && vagueCount > 5 {
		return &Pattern{
			Type:        "missing_nuance",
			Severity:    0.6,
			Description: "Lack of specific examples, citations, or concrete details",
		}
	}
	return nil
}

func (a *TextSlopAnalyzer) detectExcessiveTransitions(content string) *Pattern {
	transitions := []string{
		"however,", "therefore,", "moreover,", "consequently,",
		"in addition,", "additionally,", "on the other hand,",
		"similarly,", "likewise,", "as a result,",
	}

	count := 0
	lowerContent := strings.ToLower(content)
	for _, trans := range transitions {
		count += strings.Count(lowerContent, trans)
	}

	sentences := len(strings.Split(content, "."))
	if sentences > 10 && count > sentences/5 {
		return &Pattern{
			Type:        "excessive_transitions",
			Severity:    0.5,
			Description: fmt.Sprintf("Excessive transition phrases (%d) - over-explains obvious points", count),
		}
	}
	return nil
}

type Pattern struct {
	Type        string
	Severity    float64
	Description string
}

type TextSlopResult struct {
	Patterns      []Pattern
	SuspicionRate float64
	Summary       string
}

func (r *TextSlopResult) addPattern(p *Pattern) {
	if p != nil {
		r.Patterns = append(r.Patterns, *p)
	}
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

	if score >= 70 {
		summary = "LIKELY AI-GENERATED: Strong indicators of AI-generated text"
	} else if score >= 50 {
		summary = "POSSIBLY AI-GENERATED: Multiple indicators suggest AI generation"
	} else if score >= 30 {
		summary = "SUSPICIOUS: Some AI-like patterns detected"
	} else {
		summary = "LIKELY HUMAN-WRITTEN: Minimal AI indicators"
	}

	summary += fmt.Sprintf(" (confidence: %d%%)\n\nDetected patterns:\n", score)

	for i, p := range r.Patterns {
		summary += fmt.Sprintf("%d. %s (severity: %.1f%%)\n   %s\n", i+1, p.Type, p.Severity*100, p.Description)
	}

	return summary
}
