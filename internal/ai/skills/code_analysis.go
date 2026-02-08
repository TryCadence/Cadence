package skills

import (
	"fmt"

	"github.com/TryCadence/Cadence/internal/ai/prompts"
)

func init() {
	Register(&CodeAnalysis{})
}

// CodeAnalysisInput holds the input parameters for the code_analysis skill.
type CodeAnalysisInput struct {
	CommitHash string // Short commit hash or identifier
	Code       string // Code snippet to analyze
}

// CodeAnalysis detects AI-generated patterns in a code snippet.
// This is the core skill that backs Analyzer.AnalyzeSuspiciousCode.
type CodeAnalysis struct{}

func (s *CodeAnalysis) Name() string { return "code_analysis" }
func (s *CodeAnalysis) Description() string {
	return "Detect AI-generated code patterns in a code snippet"
}
func (s *CodeAnalysis) Category() string { return "detection" }
func (s *CodeAnalysis) MaxTokens() int   { return 1024 }

func (s *CodeAnalysis) SystemPrompt() string {
	return prompts.AnalysisSystemPrompt
}

func (s *CodeAnalysis) FormatInput(input interface{}) (string, error) {
	switch v := input.(type) {
	case CodeAnalysisInput:
		code := v.Code
		if len(code) > 2000 {
			code = code[:2000] + "...[truncated]"
		}
		hash := v.CommitHash
		if len(hash) > 8 {
			hash = hash[:8]
		}
		return fmt.Sprintf(prompts.UserPromptTemplate, hash, code), nil
	case string:
		// Convenience: pass raw code directly
		return fmt.Sprintf(prompts.UserPromptTemplate, "unknown", input.(string)), nil
	default:
		return "", fmt.Errorf("code_analysis: expected CodeAnalysisInput or string, got %T", input)
	}
}

func (s *CodeAnalysis) ParseOutput(raw string) (interface{}, error) {
	return prompts.ParseAnalysisResult(raw)
}
