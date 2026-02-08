package skills

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TryCadence/Cadence/internal/ai/prompts"
)

// ---------------------------------------------------------------------------
// Registry tests
// ---------------------------------------------------------------------------

func TestRegistryBuiltinSkills(t *testing.T) {
	// The init() functions should have registered all 4 built-in skills.
	names := RegisteredSkills()
	expected := []string{"code_analysis", "commit_review", "pattern_explain", "report_summary"}

	if len(names) < len(expected) {
		t.Fatalf("expected at least %d skills, got %d: %v", len(expected), len(names), names)
	}

	for _, exp := range expected {
		found := false
		for _, n := range names {
			if n == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected skill %q to be registered, got: %v", exp, names)
		}
	}
}

func TestRegistryGet(t *testing.T) {
	skill, err := Get("code_analysis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skill.Name() != "code_analysis" {
		t.Errorf("expected name 'code_analysis', got %q", skill.Name())
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	_, err := Get("nonexistent_skill")
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
}

func TestRegistryAll(t *testing.T) {
	all := All()
	if len(all) < 4 {
		t.Errorf("expected at least 4 skills, got %d", len(all))
	}
	// Should be sorted by name
	for i := 1; i < len(all); i++ {
		if all[i-1].Name() > all[i].Name() {
			t.Errorf("skills not sorted: %q > %q", all[i-1].Name(), all[i].Name())
		}
	}
}

// ---------------------------------------------------------------------------
// CodeAnalysis skill tests
// ---------------------------------------------------------------------------

func TestCodeAnalysisMetadata(t *testing.T) {
	s := &CodeAnalysis{}
	if s.Name() != "code_analysis" {
		t.Errorf("expected name 'code_analysis', got %q", s.Name())
	}
	if s.Category() != "detection" {
		t.Errorf("expected category 'detection', got %q", s.Category())
	}
	if s.MaxTokens() != 1024 {
		t.Errorf("expected MaxTokens 1024, got %d", s.MaxTokens())
	}
	if s.SystemPrompt() == "" {
		t.Error("expected non-empty system prompt")
	}
}

func TestCodeAnalysisFormatInput(t *testing.T) {
	s := &CodeAnalysis{}

	// With struct input
	prompt, err := s.FormatInput(CodeAnalysisInput{
		CommitHash: "abc123def456",
		Code:       "func hello() {}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "abc123de") {
		t.Error("expected truncated hash in prompt")
	}
	if !strings.Contains(prompt, "func hello()") {
		t.Error("expected code in prompt")
	}

	// With string input
	prompt, err = s.FormatInput("raw code snippet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "raw code snippet") {
		t.Error("expected raw code in prompt")
	}

	// With invalid input
	_, err = s.FormatInput(42)
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestCodeAnalysisTruncation(t *testing.T) {
	s := &CodeAnalysis{}
	longCode := strings.Repeat("x", 3000)
	prompt, err := s.FormatInput(CodeAnalysisInput{CommitHash: "abc", Code: longCode})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "...[truncated]") {
		t.Error("expected truncation marker in prompt")
	}
}

func TestCodeAnalysisParseOutput(t *testing.T) {
	s := &CodeAnalysis{}
	result, err := s.ParseOutput(`{"assessment": "likely AI-generated", "confidence": 0.9}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ar, ok := result.(*prompts.AnalysisResult)
	if !ok {
		t.Fatalf("expected *prompts.AnalysisResult, got %T", result)
	}
	if ar.Assessment != "likely AI-generated" {
		t.Errorf("expected 'likely AI-generated', got %q", ar.Assessment)
	}
}

// ---------------------------------------------------------------------------
// CommitReview skill tests
// ---------------------------------------------------------------------------

func TestCommitReviewMetadata(t *testing.T) {
	s := &CommitReview{}
	if s.Name() != "commit_review" {
		t.Errorf("expected name 'commit_review', got %q", s.Name())
	}
	if s.Category() != "detection" {
		t.Errorf("expected category 'detection', got %q", s.Category())
	}
}

func TestCommitReviewFormatInput(t *testing.T) {
	s := &CommitReview{}
	prompt, err := s.FormatInput(CommitReviewInput{
		CommitHash:    "abc123",
		Author:        "user@example.com",
		Message:       "Add feature X",
		Additions:     "func newFeature() {}",
		FilesChanged:  []string{"main.go", "util.go"},
		AdditionCount: 50,
		DeletionCount: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "abc123") {
		t.Error("expected commit hash in prompt")
	}
	if !strings.Contains(prompt, "user@example.com") {
		t.Error("expected author in prompt")
	}
	if !strings.Contains(prompt, "Add feature X") {
		t.Error("expected commit message in prompt")
	}
	if !strings.Contains(prompt, "+50 -10") {
		t.Error("expected stats in prompt")
	}
}

func TestCommitReviewFormatInputInvalid(t *testing.T) {
	s := &CommitReview{}
	_, err := s.FormatInput("not a struct")
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestCommitReviewManyFiles(t *testing.T) {
	s := &CommitReview{}
	files := make([]string, 25)
	for i := range files {
		files[i] = fmt.Sprintf("file_%d.go", i)
	}
	prompt, err := s.FormatInput(CommitReviewInput{
		CommitHash:   "abc",
		FilesChanged: files,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "and 5 more") {
		t.Error("expected file truncation message")
	}
}

// ---------------------------------------------------------------------------
// PatternExplain skill tests
// ---------------------------------------------------------------------------

func TestPatternExplainMetadata(t *testing.T) {
	s := &PatternExplain{}
	if s.Name() != "pattern_explain" {
		t.Errorf("expected name 'pattern_explain', got %q", s.Name())
	}
	if s.Category() != "explanation" {
		t.Errorf("expected category 'explanation', got %q", s.Category())
	}
	if s.MaxTokens() != 512 {
		t.Errorf("expected MaxTokens 512, got %d", s.MaxTokens())
	}
}

func TestPatternExplainFormatInput(t *testing.T) {
	s := &PatternExplain{}
	prompt, err := s.FormatInput(PatternExplainInput{
		Strategy:    "velocity_anomaly",
		Category:    "velocity",
		Severity:    "high",
		Score:       0.9,
		Description: "Unusual commit velocity detected",
		Examples:    []string{"commit abc: 500 additions in 1 minute"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "velocity_anomaly") {
		t.Error("expected strategy name in prompt")
	}
	if !strings.Contains(prompt, "0.90") {
		t.Error("expected score in prompt")
	}
	if !strings.Contains(prompt, "500 additions") {
		t.Error("expected example in prompt")
	}
}

func TestPatternExplainFormatInputInvalid(t *testing.T) {
	s := &PatternExplain{}
	_, err := s.FormatInput(123)
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestPatternExplainParseOutput(t *testing.T) {
	s := &PatternExplain{}

	// Valid JSON
	result, err := s.ParseOutput(`{
		"explanation": "The velocity anomaly strategy detected an unusually high rate of code additions.",
		"why_it_matters": "Human developers rarely add 500 lines per minute.",
		"false_positive_likelihood": "low",
		"suggestions": ["Review the commit timestamps", "Check for bulk imports"]
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	per, ok := result.(*PatternExplainResult)
	if !ok {
		t.Fatalf("expected *PatternExplainResult, got %T", result)
	}
	if per.FalsePositive != "low" {
		t.Errorf("expected false_positive 'low', got %q", per.FalsePositive)
	}
	if len(per.Suggestions) != 2 {
		t.Errorf("expected 2 suggestions, got %d", len(per.Suggestions))
	}

	// Plain text fallback
	result2, err := s.ParseOutput("This is a plain text explanation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	per2, ok := result2.(*PatternExplainResult)
	if !ok {
		t.Fatalf("expected *PatternExplainResult, got %T", result2)
	}
	if per2.Explanation != "This is a plain text explanation" {
		t.Errorf("expected plain text in Explanation, got %q", per2.Explanation)
	}
}

func TestPatternExplainExampleTruncation(t *testing.T) {
	s := &PatternExplain{}
	examples := make([]string, 5)
	for i := range examples {
		examples[i] = fmt.Sprintf("example %d", i)
	}
	prompt, err := s.FormatInput(PatternExplainInput{
		Strategy: "test",
		Category: "test",
		Severity: "low",
		Examples: examples,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "and 2 more") {
		t.Error("expected example truncation message")
	}
}

// ---------------------------------------------------------------------------
// ReportSummary skill tests
// ---------------------------------------------------------------------------

func TestReportSummaryMetadata(t *testing.T) {
	s := &ReportSummary{}
	if s.Name() != "report_summary" {
		t.Errorf("expected name 'report_summary', got %q", s.Name())
	}
	if s.Category() != "reporting" {
		t.Errorf("expected category 'reporting', got %q", s.Category())
	}
	if s.MaxTokens() != 1024 {
		t.Errorf("expected MaxTokens 1024, got %d", s.MaxTokens())
	}
}

func TestReportSummaryFormatInput(t *testing.T) {
	s := &ReportSummary{}
	prompt, err := s.FormatInput(ReportSummaryInput{
		SourceType:      "git",
		SourceID:        "https://github.com/example/repo",
		OverallScore:    0.75,
		Assessment:      "likely AI-generated",
		DetectionCount:  5,
		TotalDetections: 18,
		HighSeverity:    2,
		MediumSeverity:  1,
		LowSeverity:     2,
		TopDetections: []DetectionSummary{
			{Strategy: "velocity", Severity: "high", Score: 0.95, Description: "Fast commits"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(prompt, "example/repo") {
		t.Error("expected source ID in prompt")
	}
	if !strings.Contains(prompt, "75.0%") {
		t.Error("expected overall score in prompt")
	}
	if !strings.Contains(prompt, "5/18") {
		t.Error("expected detection counts in prompt")
	}
	if !strings.Contains(prompt, "velocity") {
		t.Error("expected top detection in prompt")
	}
}

func TestReportSummaryFormatInputInvalid(t *testing.T) {
	s := &ReportSummary{}
	_, err := s.FormatInput("not a struct")
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestReportSummaryParseOutput(t *testing.T) {
	s := &ReportSummary{}

	result, err := s.ParseOutput(`{
		"title": "High AI-generation risk detected",
		"summary": "Analysis found significant indicators.",
		"key_findings": ["High velocity commits", "Generic naming"],
		"risk_level": "high",
		"next_steps": ["Review flagged commits", "Talk to the author"]
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rsr, ok := result.(*ReportSummaryResult)
	if !ok {
		t.Fatalf("expected *ReportSummaryResult, got %T", result)
	}
	if rsr.RiskLevel != "high" {
		t.Errorf("expected risk_level 'high', got %q", rsr.RiskLevel)
	}
	if len(rsr.KeyFindings) != 2 {
		t.Errorf("expected 2 key findings, got %d", len(rsr.KeyFindings))
	}
	if len(rsr.NextSteps) != 2 {
		t.Errorf("expected 2 next steps, got %d", len(rsr.NextSteps))
	}

	// Plain text fallback
	result2, err := s.ParseOutput("Plain text summary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rsr2 := result2.(*ReportSummaryResult)
	if rsr2.Summary != "Plain text summary" {
		t.Errorf("expected plain text in Summary, got %q", rsr2.Summary)
	}
}
