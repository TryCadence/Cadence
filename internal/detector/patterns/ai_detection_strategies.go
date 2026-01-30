package patterns

import (
	"fmt"

	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

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

type RatioStrategy struct {
	maxAdditionRatio float64
	minDeletionRatio float64
	minCommitSize    int64
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

	if s.maxAdditionRatio > 0 && addRatio > s.maxAdditionRatio && total > s.minCommitSize {
		return true, fmt.Sprintf(
			"Skewed addition ratio: %.1f%% additions (threshold: %.1f%%) - possible generated code",
			addRatio*100, s.maxAdditionRatio*100,
		)
	}

	if s.minDeletionRatio > 0 && delRatio > s.minDeletionRatio && total > s.minCommitSize {
		return true, fmt.Sprintf(
			"Skewed deletion ratio: %.1f%% deletions (threshold: %.1f%%) - possible generated code",
			delRatio*100, s.minDeletionRatio*100,
		)
	}

	return false, ""
}

type PrecisionStrategy struct {
	minConsistencyScore float64
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
	if pair.Stats.Additions > 50 && pair.Stats.Deletions > 50 {
		diff := float64(pair.Stats.Additions - pair.Stats.Deletions)
		total := float64(pair.Stats.Additions + pair.Stats.Deletions)
		ratio := diff / total

		if ratio < 0.05 {
			return true, fmt.Sprintf(
				"Suspiciously balanced change ratio: %.1f%% difference - possible generated refactoring",
				ratio*100,
			)
		}
	}

	if pair.Stats.FilesChanged > 0 && pair.Stats.Additions+pair.Stats.Deletions > 100 {
		avgPerFile := float64(pair.Stats.Additions+pair.Stats.Deletions) / float64(pair.Stats.FilesChanged)
		if avgPerFile > 100 && avgPerFile < 150 {
			return true, fmt.Sprintf(
				"Suspiciously consistent file change sizes: ~%.0f LOC per file - possible generated code",
				avgPerFile,
			)
		}
	}

	return false, ""
}
