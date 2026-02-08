package detectors

import (
	"context"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git/patterns"
	"github.com/TryCadence/Cadence/internal/config"
	cerrors "github.com/TryCadence/Cadence/internal/errors"
	"github.com/TryCadence/Cadence/internal/metrics"
)

type GitDetector struct {
	Thresholds     *patterns.Thresholds
	StrategyConfig *config.StrategyConfig
}

func NewGitDetector(thresholds *patterns.Thresholds) *GitDetector {
	if thresholds == nil {
		thresholds = &patterns.Thresholds{
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
	}
	return &GitDetector{Thresholds: thresholds}
}

// NewGitDetectorWithConfig creates a GitDetector with strategy enable/disable support.
func NewGitDetectorWithConfig(thresholds *patterns.Thresholds, stratCfg *config.StrategyConfig) *GitDetector {
	d := NewGitDetector(thresholds)
	d.StrategyConfig = stratCfg
	return d
}

func (g *GitDetector) Detect(ctx context.Context, data *analysis.SourceData) ([]analysis.Detection, error) {
	if data.Type != "git" {
		return nil, cerrors.ValidationError("GitDetector only supports git sources")
	}

	pairs, ok := data.RawContent.([]*git.CommitPair)
	if !ok {
		return nil, cerrors.ValidationError("invalid RawContent type for git source")
	}

	strategies, err := g.buildStrategies()
	if err != nil {
		return nil, cerrors.AnalysisError("failed to build strategies").Wrap(err)
	}

	repoStats := &metrics.RepositoryStats{}

	for _, strategy := range strategies {
		if statStrategy, ok := strategy.(*patterns.StatisticalAnomalyStrategy); ok {
			statStrategy.SetBaseline(pairs)
			break
		}
	}

	detections := make([]analysis.Detection, 0)

	for _, pair := range pairs {
		if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
			continue
		}

		if len(pair.Current.Parents) > 1 {
			continue
		}

		type strategyHit struct {
			reason     string
			category   string
			confidence float64
		}

		hits := make([]strategyHit, 0)

		for _, strategy := range strategies {
			detected, reason := strategy.Detect(pair, repoStats)
			if detected {
				hits = append(hits, strategyHit{
					reason:     reason,
					category:   strategy.Category(),
					confidence: strategy.Confidence(),
				})
			}
		}

		if len(hits) > 0 {
			score := float64(len(hits)) / float64(len(strategies))

			// Weight score by average confidence of triggered strategies
			totalConfidence := 0.0
			for _, h := range hits {
				totalConfidence += h.confidence
			}
			avgConfidence := totalConfidence / float64(len(hits))

			severity := "low"
			if score >= 0.7 {
				severity = "high"
			} else if score >= 0.4 {
				severity = "medium"
			}

			examples := make([]string, 0, len(hits)+1)
			examples = append(examples, pair.Current.Hash)
			for _, h := range hits {
				examples = append(examples, h.reason)
			}

			// Use the most common category from triggered strategies
			categoryCounts := make(map[string]int)
			for _, h := range hits {
				categoryCounts[h.category]++
			}
			topCategory := "git-analysis"
			topCount := 0
			for cat, count := range categoryCounts {
				if count > topCount {
					topCategory = cat
					topCount = count
				}
			}

			detection := analysis.Detection{
				Strategy:    "git-velocity-analysis",
				Detected:    true,
				Severity:    severity,
				Score:       score,
				Confidence:  avgConfidence,
				Category:    topCategory,
				Description: pair.Current.Message,
				Examples:    examples,
			}
			detections = append(detections, detection)
		}
	}

	data.Metadata["suspicious_count"] = len(detections)

	return detections, nil
}

func (g *GitDetector) buildStrategies() ([]patterns.DetectionStrategy, error) {
	if err := g.Thresholds.Validate(); err != nil {
		return nil, cerrors.ValidationError("invalid thresholds").Wrap(err)
	}

	strategies := make([]patterns.DetectionStrategy, 0)

	if g.Thresholds.SuspiciousAdditions > 0 || g.Thresholds.SuspiciousDeletions > 0 {
		strategies = append(strategies, patterns.NewSizeStrategy(g.Thresholds.SuspiciousAdditions, g.Thresholds.SuspiciousDeletions))
	}

	if g.Thresholds.MaxAdditionsPerMin > 0 || g.Thresholds.MaxDeletionsPerMin > 0 {
		strategies = append(strategies, patterns.NewVelocityStrategy(g.Thresholds.MaxAdditionsPerMin, g.Thresholds.MaxDeletionsPerMin))
	}

	if g.Thresholds.MinTimeDeltaSeconds > 0 {
		strategies = append(strategies, patterns.NewTimingStrategy(g.Thresholds.MinTimeDeltaSeconds))
	}

	if g.Thresholds.MaxFilesPerCommit > 0 {
		strategies = append(strategies, patterns.NewDispersionStrategy(g.Thresholds.MaxFilesPerCommit))
	}

	if g.Thresholds.MaxAdditionRatio > 0 || g.Thresholds.MinDeletionRatio > 0 {
		strategies = append(strategies, patterns.NewRatioStrategy(g.Thresholds.MaxAdditionRatio, g.Thresholds.MinDeletionRatio, g.Thresholds.MinCommitSizeRatio))
	}

	if g.Thresholds.EnablePrecisionAnalysis {
		strategies = append(strategies, patterns.NewPrecisionStrategy(0.85))
	}

	strategies = append(strategies,
		patterns.NewCommitMessageStrategy(),
		patterns.NewNamingPatternStrategy(),
		patterns.NewStructuralConsistencyStrategy(),
		patterns.NewBurstPatternStrategy(10),
		patterns.NewErrorHandlingPatternStrategy(),
		patterns.NewTemplatePatternStrategy(),
		patterns.NewFileExtensionPatternStrategy(),
		patterns.NewStatisticalAnomalyStrategy(),
		patterns.NewTimingAnomalyStrategy(),
	)

	// Filter out strategies disabled via config
	if g.StrategyConfig != nil {
		filtered := make([]patterns.DetectionStrategy, 0, len(strategies))
		for _, s := range strategies {
			if g.StrategyConfig.IsEnabled(s.Name()) {
				filtered = append(filtered, s)
			}
		}
		strategies = filtered
	}

	return strategies, nil
}
