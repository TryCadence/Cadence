package analysis

import (
	"math"
	"sort"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
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

type RepositoryBaseline struct {
	AvgAdditions     float64
	StdDevAdditions  float64
	AvgDeletions     float64
	StdDevDeletions  float64
	AvgFilesChanged  float64
	AvgAdditionRatio float64
	MedianCommitSize int64
	Q1CommitSize     int64
	Q3CommitSize     int64
}

func CalculateBaseline(pairs []*git.CommitPair) *RepositoryBaseline {
	if len(pairs) == 0 {
		return &RepositoryBaseline{}
	}

	baseline := &RepositoryBaseline{}

	// Collect raw values for trimmed statistics
	additions := make([]float64, 0, len(pairs))
	deletions := make([]float64, 0, len(pairs))
	var sumFiles float64
	var additionRatios []float64
	commitSizes := make([]int64, 0, len(pairs))

	for _, pair := range pairs {
		stats := pair.Stats
		if stats.Additions == 0 && stats.Deletions == 0 {
			continue
		}

		additions = append(additions, float64(stats.Additions))
		deletions = append(deletions, float64(stats.Deletions))
		sumFiles += float64(stats.FilesChanged)

		total := float64(stats.Additions + stats.Deletions)
		if total > 0 {
			ratio := float64(stats.Additions) / total
			additionRatios = append(additionRatios, ratio)
		}

		commitSize := stats.Additions + stats.Deletions
		commitSizes = append(commitSizes, commitSize)
	}

	n := float64(len(additions))
	if n == 0 {
		return baseline
	}

	// Use trimmed mean/stddev (exclude top/bottom 10%) to resist outlier pollution.
	// This prevents a few massive AI-generated commits from skewing the baseline
	// and making other large commits look "normal" by comparison.
	baseline.AvgAdditions, baseline.StdDevAdditions = trimmedMeanStdDev(additions, 0.10)
	baseline.AvgDeletions, baseline.StdDevDeletions = trimmedMeanStdDev(deletions, 0.10)
	baseline.AvgFilesChanged = sumFiles / n

	if len(additionRatios) > 0 {
		sum := 0.0
		for _, r := range additionRatios {
			sum += r
		}
		baseline.AvgAdditionRatio = sum / float64(len(additionRatios))
	}

	if len(commitSizes) > 0 {
		baseline.MedianCommitSize = calculatePercentileValue(commitSizes, 50)
		baseline.Q1CommitSize = calculatePercentileValue(commitSizes, 25)
		baseline.Q3CommitSize = calculatePercentileValue(commitSizes, 75)
	}

	return baseline
}

// trimmedMeanStdDev computes the mean and standard deviation after excluding
// the top and bottom trimFraction of values. For example, trimFraction=0.10
// excludes the lowest 10% and highest 10%, computing stats on the middle 80%.
// This makes the baseline robust against outliers (e.g., a few massive AI commits
// polluting the mean/stddev for the entire repository).
func trimmedMeanStdDev(values []float64, trimFraction float64) (mean, stddev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	trimCount := int(float64(len(sorted)) * trimFraction)
	// Don't trim everything — need at least 1 value
	if trimCount*2 >= len(sorted) {
		trimCount = 0
	}

	trimmed := sorted[trimCount : len(sorted)-trimCount]
	if len(trimmed) == 0 {
		return 0, 0
	}

	n := float64(len(trimmed))
	sum := 0.0
	for _, v := range trimmed {
		sum += v
	}
	mean = sum / n

	sumSqDev := 0.0
	for _, v := range trimmed {
		sumSqDev += math.Pow(v-mean, 2)
	}
	stddev = math.Sqrt(sumSqDev / n)

	return mean, stddev
}

func DetectStatisticalAnomalies(pair *git.CommitPair, baseline *RepositoryBaseline) []*StatisticalAnomaly {
	anomalies := make([]*StatisticalAnomaly, 0)

	if baseline.StdDevAdditions == 0 && baseline.StdDevDeletions == 0 {
		return anomalies
	}

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

	commitSize := pair.Stats.Additions + pair.Stats.Deletions
	if baseline.Q1CommitSize > 0 && baseline.Q3CommitSize > 0 {
		iqr := baseline.Q3CommitSize - baseline.Q1CommitSize
		upperBound := baseline.Q3CommitSize + int64(1.5*float64(iqr))

		if commitSize > upperBound {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyIQROutlier,
				CommitHash:    pair.Current.Hash,
				Score:         float64(commitSize-upperBound) / float64(iqr),
				BaselineValue: float64(baseline.MedianCommitSize),
				ObservedValue: float64(commitSize),
				IsSignificant: commitSize > upperBound+int64(3*float64(iqr)),
				Description:   "Commit size is an extreme outlier (IQR method)",
			})
		}

		// Extreme outlier for very large commits
		if commitSize > baseline.Q3CommitSize+int64(3*float64(iqr)) {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyOutlierSize,
				CommitHash:    pair.Current.Hash,
				Score:         float64(commitSize) / float64(baseline.MedianCommitSize),
				BaselineValue: float64(baseline.MedianCommitSize),
				ObservedValue: float64(commitSize),
				IsSignificant: true,
				Description:   "Extremely large commit size (>3x IQR beyond Q3)",
			})
		}
	}

	if pair.Stats.Additions > 0 || pair.Stats.Deletions > 0 {
		total := float64(pair.Stats.Additions + pair.Stats.Deletions)
		ratio := float64(pair.Stats.Additions) / total
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

	if pair.Stats.FilesChanged > 0 && baseline.AvgFilesChanged > 0 {
		filesRatio := float64(pair.Stats.FilesChanged) / baseline.AvgFilesChanged
		if filesRatio > 3.0 && commitSize > baseline.MedianCommitSize {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyFileDispersion,
				CommitHash:    pair.Current.Hash,
				Score:         filesRatio,
				BaselineValue: baseline.AvgFilesChanged,
				ObservedValue: float64(pair.Stats.FilesChanged),
				IsSignificant: filesRatio > 5.0,
				Description:   "Unusually high number of files changed in single commit",
			})
		}
	}

	return anomalies
}

func calculatePercentileValue(sortedValues []int64, percentile float64) int64 {
	if len(sortedValues) == 0 {
		return 0
	}

	if len(sortedValues) == 1 {
		return sortedValues[0]
	}

	index := (percentile / 100.0) * float64(len(sortedValues)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sortedValues[lower]
	}

	weight := index - float64(lower)
	return int64(float64(sortedValues[lower])*(1-weight) + float64(sortedValues[upper])*weight)
}

func DetectTimingClusters(pairs []*git.CommitPair) []*StatisticalAnomaly {
	anomalies := make([]*StatisticalAnomaly, 0)

	if len(pairs) < 5 {
		return anomalies
	}

	rapidCommitWindow := 0
	for i := 0; i < len(pairs); i++ {
		if pairs[i].TimeDelta.Minutes() < 5.0 {
			rapidCommitWindow++
		} else {
			if rapidCommitWindow >= 3 {
				anomalies = append(anomalies, &StatisticalAnomaly{
					Type:          AnomalyTimingCluster,
					CommitHash:    pairs[i].Current.Hash,
					Score:         float64(rapidCommitWindow),
					BaselineValue: 1.0,
					ObservedValue: float64(rapidCommitWindow),
					IsSignificant: rapidCommitWindow >= 5,
					Description:   "Cluster of rapid-fire commits detected (potential batch processing)",
				})
			}
			rapidCommitWindow = 0
		}
	}

	return anomalies
}

func DetectAuthorBehaviorAnomalies(pairs []*git.CommitPair) []*StatisticalAnomaly {
	anomalies := make([]*StatisticalAnomaly, 0)

	if len(pairs) < 10 {
		return anomalies
	}

	authorStats := make(map[string]*CommitStatistics)
	authorCommitCounts := make(map[string]int)

	for _, pair := range pairs {
		email := pair.Current.Email
		if _, exists := authorStats[email]; !exists {
			authorStats[email] = &CommitStatistics{}
		}

		stats := authorStats[email]
		stats.Additions += pair.Stats.Additions
		stats.Deletions += pair.Stats.Deletions
		stats.FilesChanged += pair.Stats.FilesChanged
		stats.TotalLines += pair.Stats.Additions + pair.Stats.Deletions
		authorCommitCounts[email]++
	}

	for _, pair := range pairs {
		email := pair.Current.Email
		stats := authorStats[email]
		count := authorCommitCounts[email]

		if count < 3 {
			continue
		}

		avgCommitSize := float64(stats.TotalLines) / float64(count)
		currentCommitSize := float64(pair.Stats.Additions + pair.Stats.Deletions)

		if currentCommitSize > avgCommitSize*5.0 && currentCommitSize > 500 {
			anomalies = append(anomalies, &StatisticalAnomaly{
				Type:          AnomalyAuthorBehavior,
				CommitHash:    pair.Current.Hash,
				Score:         currentCommitSize / avgCommitSize,
				BaselineValue: avgCommitSize,
				ObservedValue: currentCommitSize,
				IsSignificant: currentCommitSize > avgCommitSize*10.0,
				Description:   "Commit size significantly deviates from author's typical pattern",
			})
		}
	}

	return anomalies
}

func DetectEntropyAnomalies(pair *git.CommitPair) *StatisticalAnomaly {
	if pair.Stats.Additions == 0 && pair.Stats.Deletions == 0 {
		return nil
	}

	total := float64(pair.Stats.Additions + pair.Stats.Deletions)
	if total == 0 {
		return nil
	}

	pAdd := float64(pair.Stats.Additions) / total
	pDel := float64(pair.Stats.Deletions) / total

	entropy := 0.0
	if pAdd > 0 {
		entropy -= pAdd * math.Log2(pAdd)
	}
	if pDel > 0 {
		entropy -= pDel * math.Log2(pDel)
	}

	if entropy < 0.5 && total > 500 {
		return &StatisticalAnomaly{
			Type:          AnomalyEntropy,
			CommitHash:    pair.Current.Hash,
			Score:         1.0 - entropy,
			BaselineValue: 1.0,
			ObservedValue: entropy,
			IsSignificant: entropy < 0.3 && total > 1000,
			Description:   "Low entropy commit (highly imbalanced additions/deletions)",
		}
	}

	return nil
}

type TimingAnomaly struct {
	CommitHash          string
	TimeSinceLastCommit float64
	IsAnomalous         bool
	Description         string
}

func DetectTimingAnomalies(pairs []*git.CommitPair) []*TimingAnomaly {
	anomalies := make([]*TimingAnomaly, 0)

	if len(pairs) < 3 {
		return anomalies
	}

	timeDeltasMinutes := make([]float64, 0, len(pairs))
	for _, pair := range pairs {
		timeDeltasMinutes = append(timeDeltasMinutes, pair.TimeDelta.Minutes())
	}

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

	for _, pair := range pairs {
		timeDelta := pair.TimeDelta.Minutes()
		if stdDev > 0 {
			zScore := (timeDelta - meanTime) / stdDev
			if zScore > 2.0 {
				anomalies = append(anomalies, &TimingAnomaly{
					CommitHash:          pair.Current.Hash,
					TimeSinceLastCommit: timeDelta,
					IsAnomalous:         true,
					Description:         "Unusually long time since last commit (possible batch processing)",
				})
			} else if zScore < -2.0 && timeDelta < 1.0 {
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
