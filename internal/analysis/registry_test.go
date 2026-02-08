package analysis

import (
	"testing"
)

func TestNewStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()
	if r == nil {
		t.Fatal("NewStrategyRegistry() returned nil")
	}
	if r.Count() != 0 {
		t.Errorf("New registry should be empty, got %d", r.Count())
	}
}

func TestStrategyRegistry_Register(t *testing.T) {
	r := NewStrategyRegistry()

	info := StrategyInfo{
		Name:        "test_strategy",
		Category:    CategoryPattern,
		Confidence:  0.7,
		Description: "A test strategy",
		SourceTypes: []string{"git"},
	}

	r.Register(info)

	if r.Count() != 1 {
		t.Errorf("Count() = %d, want 1", r.Count())
	}

	got, ok := r.Get("test_strategy")
	if !ok {
		t.Fatal("Get() returned false for registered strategy")
	}
	if got.Name != "test_strategy" {
		t.Errorf("Get().Name = %q, want %q", got.Name, "test_strategy")
	}
	if got.Category != CategoryPattern {
		t.Errorf("Get().Category = %q, want %q", got.Category, CategoryPattern)
	}
	if got.Confidence != 0.7 {
		t.Errorf("Get().Confidence = %f, want 0.7", got.Confidence)
	}
}

func TestStrategyRegistry_Get_NotFound(t *testing.T) {
	r := NewStrategyRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for unregistered strategy")
	}
}

func TestStrategyRegistry_All(t *testing.T) {
	r := NewStrategyRegistry()
	r.Register(StrategyInfo{Name: "a", Category: CategoryPattern, Confidence: 0.5, SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "b", Category: CategoryStatistical, Confidence: 0.8, SourceTypes: []string{"web"}})
	r.Register(StrategyInfo{Name: "c", Category: CategoryPattern, Confidence: 0.3, SourceTypes: []string{"git", "web"}})

	all := r.All()
	if len(all) != 3 {
		t.Errorf("All() returned %d items, want 3", len(all))
	}
}

func TestStrategyRegistry_ByCategory(t *testing.T) {
	r := NewStrategyRegistry()
	r.Register(StrategyInfo{Name: "a", Category: CategoryPattern, SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "b", Category: CategoryStatistical, SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "c", Category: CategoryPattern, SourceTypes: []string{"web"}})

	patterns := r.ByCategory(CategoryPattern)
	if len(patterns) != 2 {
		t.Errorf("ByCategory(pattern) = %d items, want 2", len(patterns))
	}

	stats := r.ByCategory(CategoryStatistical)
	if len(stats) != 1 {
		t.Errorf("ByCategory(statistical) = %d items, want 1", len(stats))
	}

	empty := r.ByCategory("nonexistent")
	if len(empty) != 0 {
		t.Errorf("ByCategory(nonexistent) = %d items, want 0", len(empty))
	}
}

func TestStrategyRegistry_BySourceType(t *testing.T) {
	r := NewStrategyRegistry()
	r.Register(StrategyInfo{Name: "a", SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "b", SourceTypes: []string{"web"}})
	r.Register(StrategyInfo{Name: "c", SourceTypes: []string{"git", "web"}})

	gitStrategies := r.BySourceType("git")
	if len(gitStrategies) != 2 {
		t.Errorf("BySourceType(git) = %d items, want 2", len(gitStrategies))
	}

	webStrategies := r.BySourceType("web")
	if len(webStrategies) != 2 {
		t.Errorf("BySourceType(web) = %d items, want 2", len(webStrategies))
	}
}

func TestStrategyRegistry_AboveConfidence(t *testing.T) {
	r := NewStrategyRegistry()
	r.Register(StrategyInfo{Name: "low", Confidence: 0.3, SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "medium", Confidence: 0.6, SourceTypes: []string{"git"}})
	r.Register(StrategyInfo{Name: "high", Confidence: 0.9, SourceTypes: []string{"git"}})

	above05 := r.AboveConfidence(0.5)
	if len(above05) != 2 {
		t.Errorf("AboveConfidence(0.5) = %d items, want 2", len(above05))
	}

	above08 := r.AboveConfidence(0.8)
	if len(above08) != 1 {
		t.Errorf("AboveConfidence(0.8) = %d items, want 1", len(above08))
	}

	all := r.AboveConfidence(0.0)
	if len(all) != 3 {
		t.Errorf("AboveConfidence(0.0) = %d items, want 3", len(all))
	}
}

func TestStrategyRegistry_RegisterOverwrite(t *testing.T) {
	r := NewStrategyRegistry()
	r.Register(StrategyInfo{Name: "test", Confidence: 0.5})
	r.Register(StrategyInfo{Name: "test", Confidence: 0.9})

	if r.Count() != 1 {
		t.Errorf("Count() = %d after overwrite, want 1", r.Count())
	}

	got, _ := r.Get("test")
	if got.Confidence != 0.9 {
		t.Errorf("Get().Confidence = %f after overwrite, want 0.9", got.Confidence)
	}
}

func TestDefaultGitRegistry(t *testing.T) {
	r := DefaultGitRegistry()
	if r.Count() < 15 {
		t.Errorf("DefaultGitRegistry() has %d strategies, expected at least 15", r.Count())
	}

	// All should be git source type
	for _, info := range r.All() {
		found := false
		for _, st := range info.SourceTypes {
			if st == "git" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Strategy %q in git registry has no 'git' source type", info.Name)
		}
	}
}

func TestDefaultWebRegistry(t *testing.T) {
	r := DefaultWebRegistry()
	if r.Count() < 15 {
		t.Errorf("DefaultWebRegistry() has %d strategies, expected at least 15", r.Count())
	}

	// All should be web source type
	for _, info := range r.All() {
		found := false
		for _, st := range info.SourceTypes {
			if st == "web" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Strategy %q in web registry has no 'web' source type", info.Name)
		}
	}
}

func TestDefaultRegistry(t *testing.T) {
	r := DefaultRegistry()
	gitCount := DefaultGitRegistry().Count()
	webCount := DefaultWebRegistry().Count()

	if r.Count() != gitCount+webCount {
		t.Errorf("DefaultRegistry() has %d strategies, expected %d (git: %d + web: %d)",
			r.Count(), gitCount+webCount, gitCount, webCount)
	}
}

func TestDefaultRegistries_HaveCategories(t *testing.T) {
	r := DefaultRegistry()
	categories := make(map[string]int)

	for _, info := range r.All() {
		if info.Category == "" {
			t.Errorf("Strategy %q has empty category", info.Name)
		}
		categories[info.Category]++
	}

	// Should have at least 4 different categories
	if len(categories) < 4 {
		t.Errorf("Expected at least 4 categories, got %d: %v", len(categories), categories)
	}
}

func TestDefaultRegistries_HaveConfidence(t *testing.T) {
	r := DefaultRegistry()

	for _, info := range r.All() {
		if info.Confidence <= 0 || info.Confidence > 1.0 {
			t.Errorf("Strategy %q has invalid confidence %f (should be 0 < c <= 1.0)",
				info.Name, info.Confidence)
		}
	}
}
