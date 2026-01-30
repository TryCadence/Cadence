package patterns

import (
	"fmt"

	"github.com/codemeapixel/cadence/internal/analysis"
	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

type StatisticalAnomalyStrategy struct {
	baseline        *analysis.RepositoryBaseline
	baselinePairs   []*git.CommitPair
	baselineUpdated bool
	enabled         bool
}

func NewStatisticalAnomalyStrategy() *StatisticalAnomalyStrategy {
	return &StatisticalAnomalyStrategy{
		enabled: true,
	}
}

func (s *StatisticalAnomalyStrategy) Name() string {
	return "StatisticalAnomaly"
}

func (s *StatisticalAnomalyStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled || pair == nil || pair.Stats == nil {
		return false, ""
	}

	if s.baseline == nil {
		return false, ""
	}

	anomalies := analysis.DetectStatisticalAnomalies(pair, s.baseline)

	if len(anomalies) == 0 {
		return false, ""
	}

	var significantAnomalies []*analysis.StatisticalAnomaly
	for _, anomaly := range anomalies {
		if anomaly.IsSignificant {
			significantAnomalies = append(significantAnomalies, anomaly)
		}
	}

	if len(significantAnomalies) > 0 {
		return true, fmt.Sprintf("Statistical anomalies detected: %s (z-score: %.2f, baseline: %.0f, observed: %.0f)",
			significantAnomalies[0].Description,
			significantAnomalies[0].Score,
			significantAnomalies[0].BaselineValue,
			significantAnomalies[0].ObservedValue,
		)
	}

	if len(anomalies) > 0 {
		return true, fmt.Sprintf("Moderate statistical deviation: %s (z-score: %.2f)",
			anomalies[0].Description,
			anomalies[0].Score,
		)
	}

	return false, ""
}

func (s *StatisticalAnomalyStrategy) SetBaseline(pairs []*git.CommitPair) {
	s.baselinePairs = pairs
	if len(pairs) > 0 {
		s.baseline = analysis.CalculateBaseline(pairs)
		s.baselineUpdated = true
	}
}

type TimingAnomalyStrategy struct {
	enabled bool
}

func NewTimingAnomalyStrategy() *TimingAnomalyStrategy {
	return &TimingAnomalyStrategy{
		enabled: true,
	}
}

func (s *TimingAnomalyStrategy) Name() string {
	return "TimingAnomaly"
}

func (s *TimingAnomalyStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled || pair == nil {
		return false, ""
	}

	timeDeltaMinutes := pair.TimeDelta.Minutes()

	if timeDeltaMinutes < 1.0 && (pair.Stats.Additions > 50 || pair.Stats.Deletions > 50) {
		return true, fmt.Sprintf("Unusually short time since last commit (%.1f seconds) with significant changes",
			pair.TimeDelta.Seconds())
	}

	if timeDeltaMinutes > 1440 && (pair.Stats.Additions > 500 || pair.Stats.Deletions > 500) {
		return true, fmt.Sprintf("Very long gap since last commit (%.1f hours) followed by large changes",
			pair.TimeDelta.Hours())
	}

	return false, ""
}
