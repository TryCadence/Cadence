package ai

import (
	"context"
	"fmt"
	"testing"

	"github.com/TryCadence/Cadence/internal/ai/skills"
)

func TestSkillRunnerRun(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response: `{
			"assessment": "likely AI-generated",
			"confidence": 0.85,
			"reasoning": "Generic patterns detected",
			"indicators": ["template_code"]
		}`,
	}

	runner := NewSkillRunner(mock, &Config{Model: "test", MaxTokens: 512})

	skill, _ := skills.Get("code_analysis")
	result, err := runner.Run(context.Background(), skill, skills.CodeAnalysisInput{
		CommitHash: "abc123",
		Code:       "func hello() { return }",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skill != "code_analysis" {
		t.Errorf("expected skill 'code_analysis', got %q", result.Skill)
	}
	if result.Raw == "" {
		t.Error("expected non-empty raw response")
	}
	if result.Parsed == nil {
		t.Error("expected parsed result")
	}
	if result.Provider != "mock" {
		t.Errorf("expected provider 'mock', got %q", result.Provider)
	}
}

func TestSkillRunnerRunByName(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "test-model",
		available:    true,
		response:     `{"assessment": "unlikely AI-generated", "confidence": 0.2}`,
	}

	runner := NewSkillRunner(mock, &Config{Model: "test"})

	result, err := runner.RunByName(context.Background(), "code_analysis", skills.CodeAnalysisInput{
		CommitHash: "def456",
		Code:       "x := 1 + 2",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skill != "code_analysis" {
		t.Errorf("expected skill 'code_analysis', got %q", result.Skill)
	}
}

func TestSkillRunnerRunByNameUnknown(t *testing.T) {
	mock := &mockProvider{name: "mock", available: true}
	runner := NewSkillRunner(mock, &Config{})

	_, err := runner.RunByName(context.Background(), "nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestSkillRunnerProviderError(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		available: true,
		err:       fmt.Errorf("API error"),
	}

	runner := NewSkillRunner(mock, &Config{Model: "test"})

	_, err := runner.RunByName(context.Background(), "code_analysis", skills.CodeAnalysisInput{
		CommitHash: "abc",
		Code:       "code",
	})
	if err == nil {
		t.Fatal("expected error from provider")
	}
}

func TestSkillRunnerCodeAnalysis(t *testing.T) {
	mock := &mockProvider{
		name:         "mock",
		defaultModel: "m",
		available:    true,
		response:     `{"assessment": "possibly AI-generated", "confidence": 0.6, "reasoning": "Mixed signals"}`,
	}
	runner := NewSkillRunner(mock, &Config{Model: "m"})

	result, err := runner.RunByName(context.Background(), "code_analysis", skills.CodeAnalysisInput{
		CommitHash: "abc",
		Code:       "x := 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ar := result.Parsed.(*AnalysisResult)
	if ar.Assessment != "possibly AI-generated" {
		t.Errorf("expected 'possibly AI-generated', got %q", ar.Assessment)
	}
}

func TestSkillRunnerCommitReview(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		available: true,
		response:  `{"assessment": "unlikely AI-generated", "confidence": 0.2}`,
	}
	runner := NewSkillRunner(mock, &Config{Model: "m"})

	result, err := runner.RunByName(context.Background(), "commit_review", skills.CommitReviewInput{
		CommitHash: "abc",
		Message:    "Fix typo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Skill != "commit_review" {
		t.Errorf("expected skill 'commit_review', got %q", result.Skill)
	}
}

func TestSkillRunnerPatternExplain(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		available: true,
		response:  `{"explanation": "Test explanation", "why_it_matters": "Important", "false_positive_likelihood": "medium", "suggestions": ["Check it"]}`,
	}
	runner := NewSkillRunner(mock, &Config{Model: "m"})

	result, err := runner.RunByName(context.Background(), "pattern_explain", skills.PatternExplainInput{
		Strategy: "test_strategy",
		Category: "test",
		Severity: "medium",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	per := result.Parsed.(*skills.PatternExplainResult)
	if per.FalsePositive != "medium" {
		t.Errorf("expected 'medium', got %q", per.FalsePositive)
	}
}

func TestSkillRunnerReportSummary(t *testing.T) {
	mock := &mockProvider{
		name:      "mock",
		available: true,
		response:  `{"title": "Clean repo", "summary": "No issues found", "risk_level": "none", "key_findings": [], "next_steps": []}`,
	}
	runner := NewSkillRunner(mock, &Config{Model: "m"})

	result, err := runner.RunByName(context.Background(), "report_summary", skills.ReportSummaryInput{
		SourceType: "git",
		SourceID:   "https://github.com/example/repo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rsr := result.Parsed.(*skills.ReportSummaryResult)
	if rsr.RiskLevel != "none" {
		t.Errorf("expected 'none', got %q", rsr.RiskLevel)
	}
}

func TestSkillRunnerInvalidInput(t *testing.T) {
	mock := &mockProvider{name: "mock", available: true, response: "ok"}
	runner := NewSkillRunner(mock, &Config{Model: "m"})

	// commit_review expects CommitReviewInput, not a string
	_, err := runner.RunByName(context.Background(), "commit_review", "bad input")
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}
