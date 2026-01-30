package detector

import (
	"github.com/codemeapixel/cadence/internal/detector/patterns"
	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

type DetectionStrategy interface {
	Name() string
	Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string)
}

type VelocityStrategy = patterns.VelocityStrategy
type SizeStrategy = patterns.SizeStrategy
type TimingStrategy = patterns.TimingStrategy
type MergeCommitStrategy = patterns.MergeCommitStrategy
type DispersionStrategy = patterns.DispersionStrategy
type RatioStrategy = patterns.RatioStrategy
type PrecisionStrategy = patterns.PrecisionStrategy

type CommitMessageStrategy = patterns.CommitMessageStrategy
type NamingPatternStrategy = patterns.NamingPatternStrategy
type StructuralConsistencyStrategy = patterns.StructuralConsistencyStrategy
type BurstPatternStrategy = patterns.BurstPatternStrategy
type ErrorHandlingPatternStrategy = patterns.ErrorHandlingPatternStrategy
type TemplatePatternStrategy = patterns.TemplatePatternStrategy
type FileExtensionPatternStrategy = patterns.FileExtensionPatternStrategy
type StatisticalAnomalyStrategy = patterns.StatisticalAnomalyStrategy
type TimingAnomalyStrategy = patterns.TimingAnomalyStrategy

func NewVelocityStrategy(additionsPerMin, deletionsPerMin float64) *VelocityStrategy {
	return patterns.NewVelocityStrategy(additionsPerMin, deletionsPerMin)
}

func NewSizeStrategy(additions, deletions int64) *SizeStrategy {
	return patterns.NewSizeStrategy(additions, deletions)
}

func NewTimingStrategy(minSeconds int64) *TimingStrategy {
	return patterns.NewTimingStrategy(minSeconds)
}

func NewMergeCommitStrategy(flag bool) *MergeCommitStrategy {
	return patterns.NewMergeCommitStrategy(flag)
}

func NewDispersionStrategy(maxFiles int) *DispersionStrategy {
	return patterns.NewDispersionStrategy(maxFiles)
}

func NewRatioStrategy(maxAdd, minDel float64, minSize int64) *RatioStrategy {
	return patterns.NewRatioStrategy(maxAdd, minDel, minSize)
}

func NewPrecisionStrategy(consistency float64) *PrecisionStrategy {
	return patterns.NewPrecisionStrategy(consistency)
}

func NewCommitMessageStrategy() *CommitMessageStrategy {
	return patterns.NewCommitMessageStrategy()
}

func NewNamingPatternStrategy() *NamingPatternStrategy {
	return patterns.NewNamingPatternStrategy()
}

func NewStructuralConsistencyStrategy() *StructuralConsistencyStrategy {
	return patterns.NewStructuralConsistencyStrategy()
}

func NewBurstPatternStrategy(maxPerHour int) *BurstPatternStrategy {
	return patterns.NewBurstPatternStrategy(maxPerHour)
}

func NewErrorHandlingPatternStrategy() *ErrorHandlingPatternStrategy {
	return patterns.NewErrorHandlingPatternStrategy()
}

func NewTemplatePatternStrategy() *TemplatePatternStrategy {
	return patterns.NewTemplatePatternStrategy()
}

func NewFileExtensionPatternStrategy() *FileExtensionPatternStrategy {
	return patterns.NewFileExtensionPatternStrategy()
}

func NewStatisticalAnomalyStrategy() *StatisticalAnomalyStrategy {
	return patterns.NewStatisticalAnomalyStrategy()
}

func NewTimingAnomalyStrategy() *TimingAnomalyStrategy {
	return patterns.NewTimingAnomalyStrategy()
}
