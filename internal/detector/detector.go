package detector

import (
	"fmt"
	"time"

	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

type SuspiciousCommit struct {
	Pair             *git.CommitPair
	AdditionVelocity *metrics.VelocityMetrics
	DeletionVelocity *metrics.VelocityMetrics
	Reasons          []string
	// Score is a confidence metric (0.0-1.0) for how suspicious the commit is
	Score float64
}

type Detector struct {
	thresholds *Thresholds
	strategies []DetectionStrategy
}

func New(thresholds *Thresholds) (*Detector, error) {
	if err := thresholds.Validate(); err != nil {
		return nil, fmt.Errorf("invalid thresholds: %w", err)
	}

	// Initialize strategies based on configured thresholds
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

	return &Detector{
		thresholds: thresholds,
		strategies: strategies,
	}, nil
}

func (d *Detector) DetectSuspicious(pairs []*git.CommitPair, repoStats *metrics.RepositoryStats) []*SuspiciousCommit {
	if pairs == nil {
		return []*SuspiciousCommit{}
	}

	suspicious := make([]*SuspiciousCommit, 0)

	for _, pair := range pairs {
		// Skip commits with no changes
		if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
			continue
		}

		// Skip merge commits (handled by strategy if needed)
		if len(pair.Current.Parents) > 1 {
			continue
		}

		reasons := make([]string, 0)
		detectionCount := 0

		// Apply all strategies
		for _, strategy := range d.strategies {
			detected, reason := strategy.Detect(pair, repoStats)
			if detected {
				reasons = append(reasons, reason)
				detectionCount++
			}
		}

		// Only add to suspicious if at least one strategy detected something
		if len(reasons) > 0 {
			// Calculate velocity metrics for reporting
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

			// Calculate a simple confidence score based on number of criteria triggered
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

// FormatTimeDelta formats a duration in a human-readable way
func FormatTimeDelta(d time.Duration) string {
	minutes := d.Minutes()
	if minutes < 1 {
		return fmt.Sprintf("%.2f minutes", minutes)
	}
	return fmt.Sprintf("%.1f minutes", minutes)
}
