package patterns

import (
	"testing"
)

func TestAllWebStrategies_HaveMetadata(t *testing.T) {
	registry := NewWebPatternRegistry()
	strategies := registry.GetStrategies()

	if len(strategies) < 15 {
		t.Fatalf("Expected at least 15 web strategies, got %d", len(strategies))
	}

	for _, strategy := range strategies {
		t.Run(strategy.Name(), func(t *testing.T) {
			if name := strategy.Name(); name == "" {
				t.Error("Name() returned empty string")
			}

			if cat := strategy.Category(); cat == "" {
				t.Errorf("Strategy %q has empty Category()", strategy.Name())
			}

			conf := strategy.Confidence()
			if conf <= 0 || conf > 1.0 {
				t.Errorf("Strategy %q has invalid Confidence() = %f (should be 0 < c <= 1.0)",
					strategy.Name(), conf)
			}

			if desc := strategy.Description(); desc == "" {
				t.Errorf("Strategy %q has empty Description()", strategy.Name())
			}
		})
	}
}

func TestWebStrategy_Categories_AreValid(t *testing.T) {
	validCategories := map[string]bool{
		"velocity":      true,
		"structural":    true,
		"behavioral":    true,
		"statistical":   true,
		"pattern":       true,
		"linguistic":    true,
		"accessibility": true,
	}

	registry := NewWebPatternRegistry()
	for _, strategy := range registry.GetStrategies() {
		cat := strategy.Category()
		if !validCategories[cat] {
			t.Errorf("Strategy %q has unknown category %q", strategy.Name(), cat)
		}
	}
}

func TestWebStrategy_LinguisticHighConfidence(t *testing.T) {
	// AI detection strategies should have higher confidence
	highConfidence := []WebPatternStrategy{
		NewOverusedPhrasesStrategy(),
		NewAIVocabularyStrategy(),
	}

	for _, strategy := range highConfidence {
		if strategy.Confidence() < 0.7 {
			t.Errorf("Strategy %q expected high confidence (>=0.7), got %f",
				strategy.Name(), strategy.Confidence())
		}
	}
}

func TestWebStrategy_AccessibilityLowConfidence(t *testing.T) {
	// Accessibility strategies should have lower confidence since they
	// detect code quality issues rather than AI-generated content
	lowConfidence := []WebPatternStrategy{
		NewMissingAltTextStrategy(),
		NewSemanticHTMLStrategy(),
		NewAccessibilityMarkersStrategy(),
		NewFormIssuesStrategy(),
	}

	for _, strategy := range lowConfidence {
		if strategy.Confidence() > 0.5 {
			t.Errorf("Strategy %q expected low confidence (<=0.5), got %f",
				strategy.Name(), strategy.Confidence())
		}
	}
}

func TestCustomPatternStrategy_HasMetadata(t *testing.T) {
	custom := NewCustomPatternStrategy("custom_test", []string{"test"}, 1)

	if custom.Name() != "custom_test" {
		t.Errorf("Name() = %q, want %q", custom.Name(), "custom_test")
	}
	if custom.Category() != "pattern" {
		t.Errorf("Category() = %q, want %q", custom.Category(), "pattern")
	}
	if custom.Confidence() != 0.5 {
		t.Errorf("Confidence() = %f, want 0.5", custom.Confidence())
	}
	if custom.Description() == "" {
		t.Error("Description() returned empty string")
	}
}
