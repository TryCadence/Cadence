package analysis

import "testing"

func TestStrategyInfo_Fields(t *testing.T) {
	info := StrategyInfo{
		Name:        "test",
		Category:    CategoryPattern,
		Confidence:  0.7,
		Description: "Test strategy",
		SourceTypes: []string{"git", "web"},
	}

	if info.Name != "test" {
		t.Errorf("Name = %q, want %q", info.Name, "test")
	}
	if info.Category != CategoryPattern {
		t.Errorf("Category = %q, want %q", info.Category, CategoryPattern)
	}
	if info.Confidence != 0.7 {
		t.Errorf("Confidence = %f, want 0.7", info.Confidence)
	}
	if info.Description != "Test strategy" {
		t.Errorf("Description = %q, want %q", info.Description, "Test strategy")
	}
	if len(info.SourceTypes) != 2 {
		t.Errorf("len(SourceTypes) = %d, want 2", len(info.SourceTypes))
	}
}

func TestStrategyCategories_AreDistinct(t *testing.T) {
	categories := []string{
		CategoryVelocity,
		CategoryStructural,
		CategoryBehavioral,
		CategoryStatistical,
		CategoryPattern,
		CategoryLinguistic,
		CategoryAccessibility,
	}

	seen := make(map[string]bool)
	for _, c := range categories {
		if c == "" {
			t.Error("Category constant is empty")
		}
		if seen[c] {
			t.Errorf("Duplicate category constant: %q", c)
		}
		seen[c] = true
	}

	if len(categories) != 7 {
		t.Errorf("Expected 7 category constants, got %d", len(categories))
	}
}

func TestDetection_ConfidenceField(t *testing.T) {
	d := Detection{
		Strategy:   "test",
		Detected:   true,
		Severity:   "high",
		Score:      0.9,
		Confidence: 0.8,
		Category:   "pattern",
	}

	if d.Confidence != 0.8 {
		t.Errorf("Confidence = %f, want 0.8", d.Confidence)
	}

	// Zero value should work
	d2 := Detection{Strategy: "test2"}
	if d2.Confidence != 0 {
		t.Errorf("Default Confidence = %f, want 0", d2.Confidence)
	}
}
