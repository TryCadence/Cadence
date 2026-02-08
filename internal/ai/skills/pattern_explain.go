package skills

import (
	"encoding/json"
	"fmt"
	"strings"
)

func init() {
	Register(&PatternExplain{})
}

// PatternExplainInput holds the input parameters for the pattern_explain skill.
type PatternExplainInput struct {
	Strategy    string   // Name of the detection strategy that fired
	Category    string   // Strategy category (velocity, structural, etc.)
	Severity    string   // Detection severity (low, medium, high)
	Score       float64  // Detection score
	Description string   // Strategy's base description
	Examples    []string // Code examples or evidence that triggered the strategy
}

// PatternExplainResult holds the structured output from the pattern_explain skill.
type PatternExplainResult struct {
	Explanation   string   `json:"explanation"`
	WhyItMatters  string   `json:"why_it_matters"`
	FalsePositive string   `json:"false_positive_likelihood"`
	Suggestions   []string `json:"suggestions"`
}

// PatternExplain generates a human-readable explanation of why a specific
// detection strategy flagged content as potentially AI-generated.
type PatternExplain struct{}

func (s *PatternExplain) Name() string { return "pattern_explain" }
func (s *PatternExplain) Description() string {
	return "Explain why a detection strategy flagged content"
}
func (s *PatternExplain) Category() string { return "explanation" }
func (s *PatternExplain) MaxTokens() int   { return 512 }

const patternExplainSystemPrompt = `You are a technical writer explaining AI code detection results to developers.

Given a detection strategy result, explain:
1. What the strategy detected and why it matters
2. How likely this is a false positive
3. What the developer should look at to verify

Be concise, specific, and helpful. Avoid jargon where possible.

Respond in JSON format:
{
  "explanation": "Clear explanation of what was detected",
  "why_it_matters": "Why this pattern suggests AI generation",
  "false_positive_likelihood": "low|medium|high",
  "suggestions": ["actionable suggestion 1", "actionable suggestion 2"]
}`

func (s *PatternExplain) SystemPrompt() string {
	return patternExplainSystemPrompt
}

func (s *PatternExplain) FormatInput(input interface{}) (string, error) {
	v, ok := input.(PatternExplainInput)
	if !ok {
		return "", fmt.Errorf("pattern_explain: expected PatternExplainInput, got %T", input)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Strategy: %s\n", v.Strategy))
	b.WriteString(fmt.Sprintf("Category: %s\n", v.Category))
	b.WriteString(fmt.Sprintf("Severity: %s\n", v.Severity))
	b.WriteString(fmt.Sprintf("Score: %.2f\n", v.Score))

	if v.Description != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", v.Description))
	}

	if len(v.Examples) > 0 {
		b.WriteString("\nEvidence:\n")
		for i, ex := range v.Examples {
			if i >= 3 {
				b.WriteString(fmt.Sprintf("  ... and %d more examples\n", len(v.Examples)-3))
				break
			}
			snippet := ex
			if len(snippet) > 500 {
				snippet = snippet[:500] + "...[truncated]"
			}
			b.WriteString(fmt.Sprintf("  Example %d: %s\n", i+1, snippet))
		}
	}

	b.WriteString("\nExplain this detection in the JSON format specified.")
	return b.String(), nil
}

func (s *PatternExplain) ParseOutput(raw string) (interface{}, error) {
	jsonStart := strings.Index(raw, "{")
	jsonEnd := strings.LastIndex(raw, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return &PatternExplainResult{
			Explanation: raw,
		}, nil
	}

	var result PatternExplainResult
	if err := json.Unmarshal([]byte(raw[jsonStart:jsonEnd+1]), &result); err != nil {
		return &PatternExplainResult{
			Explanation: raw,
		}, nil
	}
	return &result, nil
}
