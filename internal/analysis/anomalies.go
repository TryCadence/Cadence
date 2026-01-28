package analysis

import (
	"math"

	"github.com/codemeapixel/cadence/internal/git"
)

type AnomalyType string

const (
	AnomalyZScoreAdditions AnomalyType = "z_score_additions"
	AnomalyZScoreDeletions AnomalyType = "z_score_deletions"
	AnomalyOutlierSize     AnomalyType = "outlier_size"
	AnomalyUnusualRatio    AnomalyType = "unusual_ratio"
	AnomalyEntropy         AnomalyType = "entropy_anomaly"
	AnomalyIQROutlier      AnomalyType = "iqr_outlier"
	AnomalyTimingCluster   AnomalyType = "timing_cluster"
	AnomalyAuthorBehavior  AnomalyType = "author_behavior_change"
	AnomalyFileDispersion  AnomalyType = "file_dispersion_anomaly"
)

type StatisticalAnomaly struct {
	Type          AnomalyType
	CommitHash    string
	Score         float64
	Description   string
	BaselineValue float64
	ObservedValue float64
	IsSignificant bool
}

type CommitStatistics struct {
	Additions    int64
	Deletions    int64
	FilesChanged int
	TotalLines   int64
}

// RepositoryBaseline represents statistical baselines for a repository
type RepositoryBaseline struct {
	AvgAdditions     float64
	StdDevAdditions  float64
	AvgDeletions     float64
	StdDevDeletions  float64
	AvgFilesChanged  float64
	AvgAdditionRatio float64 // additions / (additions + deletions)
	MedianCommitSize int64
	Q1CommitSize     int64 // 25th percentile
	Q3CommitSize     int64 // 75th percentile
}

// CalculateBaseline computes statistical baseline from commit pairs
func CalculateBaseline(pairs []*git.CommitPair) *RepositoryBaseline {
	if len(pairs) == 0 {
		return &RepositoryBaseline{}
	}

	baseline := &RepositoryBaseline{}

	// Calculate mean values
	var sumAdditions, sumDeletions, sumFiles float64
	var additionRatios []float64
	var commitSizes []int64

	for _, pair := range pairs {
		stats := pair.Stats
		if stats.Additions == 0 && stats.Deletions == 0 {
			continue
		}

		sumAdditions += float64(stats.Additions)
		sumDeletions += float64(stats.Deletions)
		sumFiles += float64(stats.FilesChanged)

		// Calculate addition ratio
		total := float64(stats.Additions + stats.Deletions)
		if total > 0 {
			ratio := float64(stats.Additions) / total
			additionRatios = append(additionRatios, ratio)
		}

		commitSize := stats.Additions + stats.Deletions
		commitSizes = append(commitSizes, commitSize)
	}

	n := float64(len(pairs))
	baseline.AvgAdditions = sumAdditions / n
	baseline.AvgDeletions = sumDeletions / n
	baseline.AvgFilesChanged = sumFiles / n

	// Calculate standard deviations
	var sumAddSqDev, sumDelSqDev float64
	for _, pair := range pairs {
		if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
			continue
		}
		sumAddSqDev += math.Pow(float64(pair.Stats.Additions)-baseline.AvgAdditions, 2)
		sumDelSqDev += math.Pow(float64(pair.Stats.Deletions)-baseline.AvgDeletions, 2)
	}

	baseline.StdDevAdditions = math.Sqrt(sumAddSqDev / n)
	baseline.StdDevDeletions = math.Sqrt(sumDelSqDev / n)

	// Calculate average ratio
	if len(additionRatios) > 0 {
		sum := 0.0
		for _, r := range additionRatios {
			sum += r
		}
		baseline.AvgAdditionRatio = sum / float64(len(additionRatios))
	}

	// Calculate percentiles for commit size
	if len(commitSizes) > 0 {
		baseline.MedianCommitSize = calculatePercentileValue(commitSizes, 50)
		baseline.Q1CommitSize = calculatePercentileValue(commitSizes, 25)
		baseline.Q3CommitSize = calculatePercentileValue(commitSizes, 75)
	}

	return baseline
}

// DetectStatisticalAnomalies identifies commits that deviate significantly from baseline
// A score of |z| > 2 is considered noteworthy; |z| > 3 is very significant
func DetectStatisticalAnomalies(pair *git.CommitPair, baseline *RepositoryBaseline) []*StatisticalAnomaly {
	anomalies := make([]*StatisticalAnomaly, 0)

	if baseline.StdDevAdditions == 0 && baseline.StdDevDeletions == 0 {
		return anomalies // Cannot calculate z-scores without deviation
	}

	// Check additions anomaly
	if baseline.StdDevAdditions > 0 {
		zAdditions := (float64(pair.Stats.Additions) - baseline.AvgAdditions) / baseline.StdDevAdditions
		if math.Abs(zAdditions) > 2.0 {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyZScoreAdditions,
				CommitHash:    pair.Current.Hash,
				Score:         zAdditions,
				BaselineValue: baseline.AvgAdditions,
				ObservedValue: float64(pair.Stats.Additions),
				IsSignificant: math.Abs(zAdditions) > 3.0,
				Description:   "Commit additions significantly deviate from repository average",
			})
		}
	}

	// Check deletions anomaly
	if baseline.StdDevDeletions > 0 {
		zDeletions := (float64(pair.Stats.Deletions) - baseline.AvgDeletions) / baseline.StdDevDeletions
		if math.Abs(zDeletions) > 2.0 {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyZScoreDeletions,
				CommitHash:    pair.Current.Hash,
				Score:         zDeletions,
				BaselineValue: baseline.AvgDeletions,
				ObservedValue: float64(pair.Stats.Deletions),
				IsSignificant: math.Abs(zDeletions) > 3.0,
				Description:   "Commit deletions significantly deviate from repository average",
			})
		}
	}

	// Check for unusual addition/deletion ratio
	if pair.Stats.Additions > 0 || pair.Stats.Deletions > 0 {
		total := float64(pair.Stats.Additions + pair.Stats.Deletions)
		ratio := float64(pair.Stats.Additions) / total
		// Flag if heavily skewed toward additions (>90% adds, rarely happens in maintenance)
		if ratio > 0.9 && pair.Stats.Additions > int64(baseline.AvgAdditions*2) {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyUnusualRatio,
				CommitHash:    pair.Current.Hash,
				Score:         ratio,
				BaselineValue: baseline.AvgAdditionRatio,
				ObservedValue: ratio,
				IsSignificant: true,
				Description:   "Unusual addition/deletion ratio (predominantly additions)",
			})
		}
	}

	return anomalies
}

// calculatePercentileValue returns the value at a given percentile (0-100)
func calculatePercentileValue(sortedValues []int64, percentile float64) int64 {
	if len(sortedValues) == 0 {
		return 0
	}

	if len(sortedValues) == 1 {
		return sortedValues[0]
	}

	// Simple linear interpolation percentile calculation
	index := (percentile / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return int64(float64(sortedValues[lower])*(1-weight) + float64(sortedValues[upper])*weight)
}

// AnalyzeTimingPatterns detects unusual timing patterns between commits
type TimingAnomaly struct {
	CommitHash          string
	TimeSinceLastCommit float64 // minutes
	IsAnomalous         bool
	Description         string
}

// DetectTimingAnomalies identifies commits with unusual time gaps
func DetectTimingAnomalies(pairs []*git.CommitPair) []*TimingAnomaly {
	anomalies := make([]*TimingAnomaly, 0)

	if len(pairs) < 3 {
		return anomalies
	}

	// Calculate typical inter-commit times
	var timeDeltasMinutes []float64
	for _, pair := range pairs {
		timeDeltasMinutes = append(timeDeltasMinutes, pair.TimeDelta.Minutes())
	}

	// Calculate mean and std dev
	var sum float64
	for _, td := range timeDeltasMinutes {
		sum += td
	}
	meanTime := sum / float64(len(timeDeltasMinutes))

	var sumSqDev float64
	for _, td := range timeDeltasMinutes {
		sumSqDev += math.Pow(td-meanTime, 2)
	}
	stdDev := math.Sqrt(sumSqDev / float64(len(timeDeltasMinutes)))

	// Flag commits with unusual timing (>2 sigma away)
	for _, pair := range pairs {
		timeDelta := pair.TimeDelta.Minutes()
		if stdDev > 0 {
			zScore := (timeDelta - meanTime) / stdDev
			if zScore > 2.0 { // Much longer than normal
				anomalies = append(anomalies, &TimingAnomaly{
					CommitHash:          pair.Current.Hash,
					TimeSinceLastCommit: timeDelta,
					IsAnomalous:         true,
					Description:         "Unusually long time since last commit (possible batch processing)",
				})
			} else if zScore < -2.0 && timeDelta < 1.0 { // Much shorter than normal
				anomalies = append(anomalies, &TimingAnomaly{
					CommitHash:          pair.Current.Hash,
					TimeSinceLastCommit: timeDelta,
					IsAnomalous:         true,
					Description:         "Unusually short time since last commit (rapid-fire commits)",
				})
			}
		}
	}

	return anomalies
}
