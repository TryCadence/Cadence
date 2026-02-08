package analysis

import (
	"sync"
)

type StrategyRegistry struct {
	mu         sync.RWMutex
	strategies map[string]StrategyInfo
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{
		strategies: make(map[string]StrategyInfo),
	}
}

func (r *StrategyRegistry) Register(info StrategyInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies[info.Name] = info
}

func (r *StrategyRegistry) Get(name string) (StrategyInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.strategies[name]
	return info, ok
}

func (r *StrategyRegistry) All() []StrategyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]StrategyInfo, 0, len(r.strategies))
	for _, info := range r.strategies {
		infos = append(infos, info)
	}
	return infos
}

func (r *StrategyRegistry) ByCategory(category string) []StrategyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []StrategyInfo
	for _, info := range r.strategies {
		if info.Category == category {
			result = append(result, info)
		}
	}
	return result
}

func (r *StrategyRegistry) BySourceType(sourceType string) []StrategyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []StrategyInfo
	for _, info := range r.strategies {
		for _, st := range info.SourceTypes {
			if st == sourceType {
				result = append(result, info)
				break
			}
		}
	}
	return result
}

func (r *StrategyRegistry) AboveConfidence(threshold float64) []StrategyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []StrategyInfo
	for _, info := range r.strategies {
		if info.Confidence >= threshold {
			result = append(result, info)
		}
	}
	return result
}

func (r *StrategyRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.strategies)
}

func DefaultGitRegistry() *StrategyRegistry {
	r := NewStrategyRegistry()

	gitStrategies := []StrategyInfo{
		{Name: "velocity_analysis", Category: CategoryVelocity, Confidence: 0.7, Description: "Detects code velocity exceeding human norms", SourceTypes: []string{"git"}},
		{Name: "size_analysis", Category: CategoryStructural, Confidence: 0.6, Description: "Detects unusually large commit sizes", SourceTypes: []string{"git"}},
		{Name: "timing_analysis", Category: CategoryBehavioral, Confidence: 0.6, Description: "Detects suspiciously short time between commits", SourceTypes: []string{"git"}},
		{Name: "merge_commit_filter", Category: CategoryStructural, Confidence: 0.3, Description: "Analyzes merge commits for potential history rewrites", SourceTypes: []string{"git"}},
		{Name: "file_dispersion_analysis", Category: CategoryStructural, Confidence: 0.5, Description: "Detects too many files changed in a single commit", SourceTypes: []string{"git"}},
		{Name: "ratio_analysis", Category: CategoryStatistical, Confidence: 0.6, Description: "Detects skewed addition/deletion ratios indicating generated code", SourceTypes: []string{"git"}},
		{Name: "precision_analysis", Category: CategoryStatistical, Confidence: 0.7, Description: "Detects suspiciously precise and balanced code changes", SourceTypes: []string{"git"}},
		{Name: "commit_message_analysis", Category: CategoryBehavioral, Confidence: 0.8, Description: "Detects AI-typical commit message patterns and phrasing", SourceTypes: []string{"git"}},
		{Name: "naming_pattern_analysis", Category: CategoryPattern, Confidence: 0.7, Description: "Detects generic or AI-typical variable and function naming", SourceTypes: []string{"git"}},
		{Name: "structural_consistency_analysis", Category: CategoryStatistical, Confidence: 0.6, Description: "Detects suspiciously balanced addition/deletion ratios", SourceTypes: []string{"git"}},
		{Name: "burst_pattern_analysis", Category: CategoryBehavioral, Confidence: 0.7, Description: "Detects rapid-fire commit patterns suggesting batch processing", SourceTypes: []string{"git"}},
		{Name: "error_handling_analysis", Category: CategoryPattern, Confidence: 0.6, Description: "Detects missing or excessive error handling typical of AI code", SourceTypes: []string{"git"}},
		{Name: "template_pattern_analysis", Category: CategoryPattern, Confidence: 0.7, Description: "Detects template/boilerplate code patterns from AI generation", SourceTypes: []string{"git"}},
		{Name: "file_extension_analysis", Category: CategoryStructural, Confidence: 0.5, Description: "Detects suspicious bulk file creation patterns", SourceTypes: []string{"git"}},
		{Name: "StatisticalAnomaly", Category: CategoryStatistical, Confidence: 0.8, Description: "Detects statistical deviations from repository baseline", SourceTypes: []string{"git"}},
		{Name: "TimingAnomaly", Category: CategoryBehavioral, Confidence: 0.7, Description: "Detects unusual timing patterns between commits", SourceTypes: []string{"git"}},
		{Name: "emoji_pattern_analysis", Category: CategoryPattern, Confidence: 0.4, Description: "Detects excessive emoji usage in commit messages", SourceTypes: []string{"git"}},
		{Name: "special_character_pattern_analysis", Category: CategoryPattern, Confidence: 0.4, Description: "Detects unusual special character patterns in commits", SourceTypes: []string{"git"}},
	}

	for _, s := range gitStrategies {
		r.Register(s)
	}
	return r
}

func DefaultWebRegistry() *StrategyRegistry {
	r := NewStrategyRegistry()

	webStrategies := []StrategyInfo{
		{Name: "overused_phrases", Category: CategoryLinguistic, Confidence: 0.8, Description: "Detects common AI-generated filler phrases", SourceTypes: []string{"web"}},
		{Name: "generic_language", Category: CategoryLinguistic, Confidence: 0.7, Description: "Detects excessive use of generic business language", SourceTypes: []string{"web"}},
		{Name: "excessive_structure", Category: CategoryStructural, Confidence: 0.6, Description: "Detects over-structured content with excessive lists and headings", SourceTypes: []string{"web"}},
		{Name: "perfect_grammar", Category: CategoryLinguistic, Confidence: 0.5, Description: "Detects suspiciously consistent sentence lengths", SourceTypes: []string{"web"}},
		{Name: "boilerplate_text", Category: CategoryPattern, Confidence: 0.7, Description: "Detects common boilerplate and filler phrases", SourceTypes: []string{"web"}},
		{Name: "repetitive_patterns", Category: CategoryPattern, Confidence: 0.7, Description: "Detects repetitive sentence structures and patterns", SourceTypes: []string{"web"}},
		{Name: "missing_nuance", Category: CategoryLinguistic, Confidence: 0.6, Description: "Detects excessive absolute terms lacking nuance", SourceTypes: []string{"web"}},
		{Name: "excessive_transitions", Category: CategoryLinguistic, Confidence: 0.7, Description: "Detects overuse of transition words and connectors", SourceTypes: []string{"web"}},
		{Name: "uniform_sentence_length", Category: CategoryStatistical, Confidence: 0.6, Description: "Detects unnaturally uniform sentence lengths", SourceTypes: []string{"web"}},
		{Name: "ai_vocabulary", Category: CategoryLinguistic, Confidence: 0.8, Description: "Detects AI-characteristic vocabulary and word choices", SourceTypes: []string{"web"}},
		{Name: "emoji_overuse", Category: CategoryPattern, Confidence: 0.4, Description: "Detects excessive emoji usage in content", SourceTypes: []string{"web"}},
		{Name: "special_characters", Category: CategoryPattern, Confidence: 0.4, Description: "Detects excessive special character patterns", SourceTypes: []string{"web"}},
		{Name: "missing_alt_text", Category: CategoryAccessibility, Confidence: 0.3, Description: "Detects images missing alt text attributes", SourceTypes: []string{"web"}},
		{Name: "semantic_html_issues", Category: CategoryAccessibility, Confidence: 0.3, Description: "Detects overuse of div tags instead of semantic HTML", SourceTypes: []string{"web"}},
		{Name: "accessibility_markers", Category: CategoryAccessibility, Confidence: 0.3, Description: "Detects missing accessibility markers and ARIA attributes", SourceTypes: []string{"web"}},
		{Name: "heading_hierarchy_issues", Category: CategoryStructural, Confidence: 0.4, Description: "Detects improper heading level order and hierarchy", SourceTypes: []string{"web"}},
		{Name: "hardcoded_values", Category: CategoryPattern, Confidence: 0.5, Description: "Detects hardcoded inline styles, pixels, and color values", SourceTypes: []string{"web"}},
		{Name: "form_issues", Category: CategoryAccessibility, Confidence: 0.3, Description: "Detects form inputs missing labels, types, or names", SourceTypes: []string{"web"}},
		{Name: "link_text_quality", Category: CategoryAccessibility, Confidence: 0.4, Description: "Detects generic or non-descriptive link text", SourceTypes: []string{"web"}},
		{Name: "generic_styling", Category: CategoryPattern, Confidence: 0.4, Description: "Detects lack of CSS variables, theming, and overuse of inline styles", SourceTypes: []string{"web"}},
	}

	for _, s := range webStrategies {
		r.Register(s)
	}
	return r
}

func DefaultRegistry() *StrategyRegistry {
	r := NewStrategyRegistry()

	gitReg := DefaultGitRegistry()
	for _, s := range gitReg.All() {
		r.Register(s)
	}

	webReg := DefaultWebRegistry()
	for _, s := range webReg.All() {
		r.Register(s)
	}

	return r
}
