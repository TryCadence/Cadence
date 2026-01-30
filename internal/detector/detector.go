package detector

import (
	"fmt"
	"time"

	"github.com/codemeapixel/cadence/internal/detector/patterns"
	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

type SuspiciousCommit struct {
	Pair             *git.CommitPair
	AdditionVelocity *metrics.VelocityMetrics
	DeletionVelocity *metrics.VelocityMetrics
	Reasons          []string
	Score            float64
	AIAnalysis       string
}

type Detector struct {
	thresholds *Thresholds
	strategies []DetectionStrategy
}

func New(thresholds *Thresholds) (*Detector, error) {
	if err := thresholds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid thresholds: %w", err)
	}

	strategies := make([]DetectionStrategy, 0)

	if thresholds.SuspiciousAdditions > 0 || thresholds.SuspiciousDeletions > 0 {
		strategies = append(strategies, NewSizeStrategy(thresholds.SuspiciousAdditions, thresholds.SuspiciousDeletions))
	}

	if thresholds.MaxAdditionsPerMin > 0 || thresholds.MaxDeletionsPerMin > 0 {
		strategies = append(strategies, NewVelocityStrategy(thresholds.MaxAdditionsPerMin, thresholds.MaxDeletionsPerMin))
	}

	if thresholds.MinTimeDeltaSeconds > 0 {
		strategies = append(strategies, NewTimingStrategy(thresholds.MinTimeDeltaSeconds))
	}

	if thresholds.MaxFilesPerCommit > 0 {
		strategies = append(strategies, NewDispersionStrategy(thresholds.MaxFilesPerCommit))
	}

	if thresholds.MaxAdditionRatio > 0 || thresholds.MinDeletionRatio > 0 {
		strategies = append(strategies, NewRatioStrategy(thresholds.MaxAdditionRatio, thresholds.MinDeletionRatio, thresholds.MinCommitSizeRatio))
	}

	if thresholds.EnablePrecisionAnalysis {
		strategies = append(strategies, NewPrecisionStrategy(0.85))
	}

	strategies = append(strategies, NewCommitMessageStrategy())
	strategies = append(strategies, NewNamingPatternStrategy())
	strategies = append(strategies, NewStructuralConsistencyStrategy())
	strategies = append(strategies, NewBurstPatternStrategy(10))
	strategies = append(strategies, NewErrorHandlingPatternStrategy())
	strategies = append(strategies, NewTemplatePatternStrategy())
	strategies = append(strategies, NewFileExtensionPatternStrategy())
	strategies = append(strategies, NewStatisticalAnomalyStrategy())
	strategies = append(strategies, NewTimingAnomalyStrategy())

	return &Detector{
		thresholds: thresholds,
		strategies: strategies,
	}, nil
}

func (d *Detector) DetectSuspicious(pairs []*git.CommitPair, repoStats *metrics.RepositoryStats) []*SuspiciousCommit {
	if pairs == nil {
		return []*SuspiciousCommit{}
	}

	for _, strategy := range d.strategies {
		if statStrategy, ok := strategy.(*patterns.StatisticalAnomalyStrategy); ok {
			statStrategy.SetBaseline(pairs)
			break
		}
	}

	suspicious := make([]*SuspiciousCommit, 0)

	for _, pair := range pairs {
		if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
			continue
		}

		if len(pair.Current.Parents) > 1 {
			continue
		}

		reasons := make([]string, 0)
		detectionCount := 0

		for _, strategy := range d.strategies {
			detected, reason := strategy.Detect(pair, repoStats)
			if detected {
				reasons = append(reasons, reason)
				detectionCount++
			}
		}

		if len(reasons) > 0 {
			var additionVelocity, deletionVelocity *metrics.VelocityMetrics
			if pair.TimeDelta > 0 {
				var err error
				additionVelocity, err = metrics.CalculateVelocity(pair.Stats.Additions, pair.TimeDelta)
				if err != nil {
					additionVelocity = nil
				}
				deletionVelocity, err = metrics.CalculateVelocity(pair.Stats.Deletions, pair.TimeDelta)
				if err != nil {
					deletionVelocity = nil
				}
			}

			score := float64(detectionCount) / float64(len(d.strategies))

			suspicious = append(suspicious, &SuspiciousCommit{
				Pair:             pair,
				AdditionVelocity: additionVelocity,
				DeletionVelocity: deletionVelocity,
				Reasons:          reasons,
				Score:            score,
			})
		}
	}

	return suspicious
}

func FormatTimeDelta(d time.Duration) string {
	minutes := d.Minutes()
	if minutes < 1 {
		return fmt.Sprintf("%.2f minutes", minutes)
	}
	return fmt.Sprintf("%.1f minutes", minutes)
}
