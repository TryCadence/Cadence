package patterns

import (
	"testing"
)

func TestAllGitStrategies_HaveMetadata(t *testing.T) {
	th := &Thresholds{
		SuspiciousAdditions:     1000,
		SuspiciousDeletions:     500,
		MaxAdditionsPerMin:      300,
		MaxDeletionsPerMin:      150,
		MinTimeDeltaSeconds:     10,
		MaxFilesPerCommit:       20,
		MaxAdditionRatio:        0.9,
		MinDeletionRatio:        0.1,
		MinCommitSizeRatio:      100,
		EnablePrecisionAnalysis: true,
	}

	strategies := []DetectionStrategy{
		NewVelocityStrategy(th.MaxAdditionsPerMin, th.MaxDeletionsPerMin),
		NewSizeStrategy(th.SuspiciousAdditions, th.SuspiciousDeletions),
		NewTimingStrategy(th.MinTimeDeltaSeconds),
		NewMergeCommitStrategy(false),
		NewDispersionStrategy(th.MaxFilesPerCommit),
		NewRatioStrategy(th.MaxAdditionRatio, th.MinDeletionRatio, th.MinCommitSizeRatio),
		NewPrecisionStrategy(0.85),
		NewCommitMessageStrategy(),
		NewNamingPatternStrategy(),
		NewStructuralConsistencyStrategy(),
		NewBurstPatternStrategy(10),
		NewErrorHandlingPatternStrategy(),
		NewTemplatePatternStrategy(),
		NewFileExtensionPatternStrategy(),
		NewStatisticalAnomalyStrategy(),
		NewTimingAnomalyStrategy(),
		NewEmojiPatternStrategy(),
		NewSpecialCharacterPatternStrategy(),
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

func TestGitStrategy_Categories_AreValid(t *testing.T) {
	validCategories := map[string]bool{
		"velocity":      true,
		"structural":    true,
		"behavioral":    true,
		"statistical":   true,
		"pattern":       true,
		"linguistic":    true,
		"accessibility": true,
	}

	th := &Thresholds{
		SuspiciousAdditions: 1000,
		SuspiciousDeletions: 500,
		MaxAdditionsPerMin:  300,
		MaxDeletionsPerMin:  150,
		MinTimeDeltaSeconds: 10,
	}
	strategies := []DetectionStrategy{
		NewVelocityStrategy(th.MaxAdditionsPerMin, th.MaxDeletionsPerMin),
		NewSizeStrategy(th.SuspiciousAdditions, th.SuspiciousDeletions),
		NewTimingStrategy(th.MinTimeDeltaSeconds),
		NewCommitMessageStrategy(),
		NewNamingPatternStrategy(),
		NewStatisticalAnomalyStrategy(),
	}

	for _, strategy := range strategies {
		cat := strategy.Category()
		if !validCategories[cat] {
			t.Errorf("Strategy %q has unknown category %q", strategy.Name(), cat)
		}
	}
}

func TestGitStrategy_HighConfidenceStrategies(t *testing.T) {
	// These strategies should have confidence >= 0.7
	highConfidence := []DetectionStrategy{
		NewVelocityStrategy(300, 150),
		NewPrecisionStrategy(0.85),
		NewStatisticalAnomalyStrategy(),
	}

	for _, strategy := range highConfidence {
		if strategy.Confidence() < 0.7 {
			t.Errorf("Strategy %q expected high confidence (>=0.7), got %f",
				strategy.Name(), strategy.Confidence())
		}
	}
}

func TestGitStrategy_MediumConfidenceStrategies(t *testing.T) {
	// CommitMessageStrategy was intentionally lowered to medium confidence (0.6)
	// because generic commit patterns cause too many false positives
	mediumConfidence := []DetectionStrategy{
		NewCommitMessageStrategy(),
	}

	for _, strategy := range mediumConfidence {
		if strategy.Confidence() < 0.5 || strategy.Confidence() >= 0.7 {
			t.Errorf("Strategy %q expected medium confidence (0.5-0.7), got %f",
				strategy.Name(), strategy.Confidence())
		}
	}
}
