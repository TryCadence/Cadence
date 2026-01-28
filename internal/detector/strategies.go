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

// Backward compatibility: These are now in patterns package
type VelocityStrategy = patterns.VelocityStrategy
type SizeStrategy = patterns.SizeStrategy
type TimingStrategy = patterns.TimingStrategy
type MergeCommitStrategy = patterns.MergeCommitStrategy
type DispersionStrategy = patterns.DispersionStrategy
type RatioStrategy = patterns.RatioStrategy
type PrecisionStrategy = patterns.PrecisionStrategy

// Constructor functions for backward compatibility
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
