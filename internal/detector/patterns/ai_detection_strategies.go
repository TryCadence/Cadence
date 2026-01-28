package patterns

import (
	"fmt"

	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

// DispersionStrategy detects commits that change too many files (sign of batch/automated changes)
type DispersionStrategy struct {
	maxFilesThreshold int
}

func NewDispersionStrategy(maxFiles int) *DispersionStrategy {
	return &DispersionStrategy{
		maxFilesThreshold: maxFiles,
	}
}

func (s *DispersionStrategy) Name() string {
	return "file_dispersion_analysis"
}

func (s *DispersionStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if s.maxFilesThreshold > 0 && pair.Stats.FilesChanged > s.maxFilesThreshold {
		return true, fmt.Sprintf(
			"Commit touches too many files: %d files (threshold: %d files) - possible batch automation",
			pair.Stats.FilesChanged, s.maxFilesThreshold,
		)
	}
	return false, ""
}

// RatioStrategy detects unusual code addition/deletion ratios (AI patterns)
type RatioStrategy struct {
	maxAdditionRatio float64 // e.g., 0.95 = flag if >95% additions
	minDeletionRatio float64 // e.g., 0.95 = flag if >95% deletions
	minCommitSize    int64   // Don't flag trivial commits
}

func NewRatioStrategy(maxAdd, minDel float64, minSize int64) *RatioStrategy {
	return &RatioStrategy{
		maxAdditionRatio: maxAdd,
		minDeletionRatio: minDel,
		minCommitSize:    minSize,
	}
}

func (s *RatioStrategy) Name() string {
	return "ratio_analysis"
}

func (s *RatioStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	total := pair.Stats.Additions + pair.Stats.Deletions
	if total < s.minCommitSize {
		return false, ""
	}

	if total == 0 {
		return false, ""
	}

	addRatio := float64(pair.Stats.Additions) / float64(total)
	delRatio := float64(pair.Stats.Deletions) / float64(total)

	// Flag predominantly additions (>90%) in large commits
	if s.maxAdditionRatio > 0 && addRatio > s.maxAdditionRatio && total > s.minCommitSize {
		return true, fmt.Sprintf(
			"Skewed addition ratio: %.1f%% additions (threshold: %.1f%%) - possible generated code",
			addRatio*100, s.maxAdditionRatio*100,
		)
	}

	// Flag predominantly deletions (>90%) in large commits
	if s.minDeletionRatio > 0 && delRatio > s.minDeletionRatio && total > s.minCommitSize {
		return true, fmt.Sprintf(
			"Skewed deletion ratio: %.1f%% deletions (threshold: %.1f%%) - possible generated code",
			delRatio*100, s.minDeletionRatio*100,
		)
	}

	return false, ""
}

// PrecisionStrategy detects suspiciously "perfect" or consistent changes
// AI code often has very consistent patterns, perfect indentation, etc.
type PrecisionStrategy struct {
	minConsistencyScore float64 // 0.0-1.0: how consistent the pattern must be to flag
}

func NewPrecisionStrategy(consistency float64) *PrecisionStrategy {
	return &PrecisionStrategy{
		minConsistencyScore: consistency,
	}
}

func (s *PrecisionStrategy) Name() string {
	return "precision_analysis"
}

func (s *PrecisionStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	// Flag when additions and deletions are very similar (sign of refactoring/generation)
	if pair.Stats.Additions > 50 && pair.Stats.Deletions > 50 {
		diff := float64(pair.Stats.Additions - pair.Stats.Deletions)
		total := float64(pair.Stats.Additions + pair.Stats.Deletions)
		ratio := diff / total

		if ratio < 0.05 { // Almost equal additions and deletions
			return true, fmt.Sprintf(
				"Suspiciously balanced change ratio: %.1f%% difference - possible generated refactoring",
				ratio*100,
			)
		}
	}

	// Flag files changed relative to size (too consistent = not human)
	if pair.Stats.FilesChanged > 0 && pair.Stats.Additions+pair.Stats.Deletions > 100 {
		avgPerFile := float64(pair.Stats.Additions+pair.Stats.Deletions) / float64(pair.Stats.FilesChanged)
		if avgPerFile > 100 && avgPerFile < 150 { // Very consistent file sizes (human variance is larger)
			return true, fmt.Sprintf(
				"Suspiciously consistent file change sizes: ~%.0f LOC per file - possible generated code",
				avgPerFile,
			)
		}
	}

	return false, ""
}
