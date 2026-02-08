package ai

import (
	"context"
	"fmt"

	"github.com/TryCadence/Cadence/internal/ai/skills"
)

// SkillResult holds the output from running a skill.
type SkillResult struct {
	Skill    string      `json:"skill"`
	Raw      string      `json:"raw"`
	Parsed   interface{} `json:"parsed"`
	Provider string      `json:"provider"`
	Model    string      `json:"model"`
}

// SkillRunner executes skills using a Provider.
type SkillRunner struct {
	provider Provider
	config   *Config
}

// NewSkillRunner creates a SkillRunner backed by the given provider and config.
func NewSkillRunner(provider Provider, cfg *Config) *SkillRunner {
	return &SkillRunner{provider: provider, config: cfg}
}

// Run executes the given skill with the provided input and returns the result.
func (r *SkillRunner) Run(ctx context.Context, skill skills.Skill, input interface{}) (*SkillResult, error) {
	userPrompt, err := skill.FormatInput(input)
	if err != nil {
		return nil, fmt.Errorf("skill %q: failed to format input: %w", skill.Name(), err)
	}

	model := r.config.Model
	if model == "" {
		model = r.provider.DefaultModel()
	}

	maxTokens := skill.MaxTokens()
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	raw, err := r.provider.Complete(ctx, CompletionRequest{
		SystemPrompt: skill.SystemPrompt(),
		UserPrompt:   userPrompt,
		Model:        model,
		MaxTokens:    maxTokens,
		Temperature:  0.3,
	})
	if err != nil {
		return nil, fmt.Errorf("skill %q: provider error: %w", skill.Name(), err)
	}

	parsed, err := skill.ParseOutput(raw)
	if err != nil {
		// Return raw response even if parsing fails — the caller can still use it.
		return &SkillResult{
			Skill:    skill.Name(),
			Raw:      raw,
			Parsed:   nil,
			Provider: r.provider.Name(),
			Model:    model,
		}, fmt.Errorf("skill %q: failed to parse output (raw available): %w", skill.Name(), err)
	}

	return &SkillResult{
		Skill:    skill.Name(),
		Raw:      raw,
		Parsed:   parsed,
		Provider: r.provider.Name(),
		Model:    model,
	}, nil
}

// RunByName looks up a skill by name in the registry and runs it.
func (r *SkillRunner) RunByName(ctx context.Context, name string, input interface{}) (*SkillResult, error) {
	skill, err := skills.Get(name)
	if err != nil {
		return nil, err
	}
	return r.Run(ctx, skill, input)
}
