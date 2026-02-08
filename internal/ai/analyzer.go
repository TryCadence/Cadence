package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/TryCadence/Cadence/internal/ai/prompts"
)

type Analyzer interface {
	AnalyzeSuspiciousCode(ctx context.Context, commitHash string, additions string) (string, error)

	AnalyzeWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error)

	// RunSkill executes a named skill with the given input.
	RunSkill(ctx context.Context, skillName string, input interface{}) (*SkillResult, error)

	IsConfigured() bool

	ProviderName() string
}

type DefaultAnalyzer struct {
	provider    Provider
	config      *Config
	skillRunner *SkillRunner
}

func NewDefaultAnalyzer(provider Provider, cfg *Config) *DefaultAnalyzer {
	return &DefaultAnalyzer{
		provider:    provider,
		config:      cfg,
		skillRunner: NewSkillRunner(provider, cfg),
	}
}

func NewAnalyzer(cfg *Config) (Analyzer, error) {
	if !cfg.Enabled || cfg.APIKey == "" {
		return &NoOpAnalyzer{}, nil
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}
	if cfg.Model == "" {
		cfg.Model = provider.DefaultModel()
	}

	return &DefaultAnalyzer{
		provider:    provider,
		config:      cfg,
		skillRunner: NewSkillRunner(provider, cfg),
	}, nil
}

func (a *DefaultAnalyzer) AnalyzeSuspiciousCode(ctx context.Context, commitHash, additions string) (string, error) {
	result, err := a.analyzeWithReasoning(ctx, commitHash, additions)
	if err != nil {
		return "", err
	}

	output := fmt.Sprintf("%s (confidence: %.0f%%)", result.Assessment, result.Confidence*100)
	if result.Reasoning != "" {
		output += fmt.Sprintf("\nReasoning: %s", result.Reasoning)
	}
	return output, nil
}

func (a *DefaultAnalyzer) AnalyzeWithSystemPrompt(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	model := a.config.Model
	if model == "" {
		model = a.provider.DefaultModel()
	}

	maxTokens := a.config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	return a.provider.Complete(ctx, CompletionRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        model,
		MaxTokens:    maxTokens,
		Temperature:  0.3,
	})
}

func (a *DefaultAnalyzer) analyzeWithReasoning(ctx context.Context, commitHash, additions string) (*AnalysisResult, error) {
	codeSnippet := additions
	if len(codeSnippet) > 2000 {
		codeSnippet = truncateAtLineBoundary(codeSnippet, 2000)
	}

	hashPrefix := commitHash
	if len(hashPrefix) > 8 {
		hashPrefix = hashPrefix[:8]
	}

	userPrompt := fmt.Sprintf(prompts.UserPromptTemplate, hashPrefix, codeSnippet)

	response, err := a.AnalyzeWithSystemPrompt(ctx, prompts.AnalysisSystemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	return ParseAnalysisResult(response)
}

func (a *DefaultAnalyzer) RunSkill(ctx context.Context, skillName string, input interface{}) (*SkillResult, error) {
	return a.skillRunner.RunByName(ctx, skillName, input)
}

func (a *DefaultAnalyzer) IsConfigured() bool {
	return a.provider != nil && a.provider.IsAvailable()
}

func (a *DefaultAnalyzer) ProviderName() string {
	if a.provider == nil {
		return ""
	}
	return a.provider.Name()
}

// truncateAtLineBoundary truncates a code snippet to at most maxLen characters,
// cutting at the last complete line boundary instead of mid-line. This preserves
// readable code structure for AI analysis rather than slicing mid-statement.
func truncateAtLineBoundary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Find the last newline before the limit
	truncated := s[:maxLen]
	lastNewline := strings.LastIndex(truncated, "\n")
	if lastNewline > 0 {
		truncated = truncated[:lastNewline]
	}

	lineCount := strings.Count(truncated, "\n") + 1
	totalLines := strings.Count(s, "\n") + 1
	return truncated + fmt.Sprintf("\n...[truncated: showing %d of %d lines]", lineCount, totalLines)
}

type NoOpAnalyzer struct{}

func (n *NoOpAnalyzer) AnalyzeSuspiciousCode(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (n *NoOpAnalyzer) AnalyzeWithSystemPrompt(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (n *NoOpAnalyzer) RunSkill(_ context.Context, _ string, _ interface{}) (*SkillResult, error) {
	return nil, fmt.Errorf("AI is not configured")
}

func (n *NoOpAnalyzer) IsConfigured() bool {
	return false
}

func (n *NoOpAnalyzer) ProviderName() string {
	return ""
}
