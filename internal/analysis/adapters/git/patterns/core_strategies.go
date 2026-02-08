package patterns

import (
	"fmt"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/metrics"
)

// DetectionStrategy is the interface for all git-based detection strategies.
// Each strategy must provide metadata (Category, Confidence, Description)
// in addition to its detection logic.
type DetectionStrategy interface {
	Name() string
	Category() string
	Confidence() float64
	Description() string
	Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string)
}

type VelocityStrategy struct {
	maxAdditionsPerMin float64
	maxDeletionsPerMin float64
}

func NewVelocityStrategy(additionsPerMin, deletionsPerMin float64) *VelocityStrategy {
	return &VelocityStrategy{
		maxAdditionsPerMin: additionsPerMin,
		maxDeletionsPerMin: deletionsPerMin,
	}
}

func (s *VelocityStrategy) Name() string        { return "velocity_analysis" }
func (s *VelocityStrategy) Category() string    { return "velocity" }
func (s *VelocityStrategy) Confidence() float64 { return 0.7 }
func (s *VelocityStrategy) Description() string { return "Detects code velocity exceeding human norms" }

func (s *VelocityStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if pair.TimeDelta <= 0 {
		return false, ""
	}

	if s.maxAdditionsPerMin > 0 {
		addVelocity, err := metrics.CalculateVelocityPerMinute(pair.Stats.Additions, pair.TimeDelta)
		if err == nil && addVelocity > s.maxAdditionsPerMin {
			return true, fmt.Sprintf(
				"Addition velocity too high: %.1f additions/min (threshold: %.1f additions/min)",
				addVelocity, s.maxAdditionsPerMin,
			)
		}
	}

	if s.maxDeletionsPerMin > 0 {
		delVelocity, err := metrics.CalculateVelocityPerMinute(pair.Stats.Deletions, pair.TimeDelta)
		if err == nil && delVelocity > s.maxDeletionsPerMin {
			return true, fmt.Sprintf(
				"Deletion velocity too high: %.1f deletions/min (threshold: %.1f deletions/min)",
				delVelocity, s.maxDeletionsPerMin,
			)
		}
	}

	return false, ""
}

type SizeStrategy struct {
	suspiciousAdditions int64
	suspiciousDeletions int64
}

func NewSizeStrategy(additions, deletions int64) *SizeStrategy {
	return &SizeStrategy{
		suspiciousAdditions: additions,
		suspiciousDeletions: deletions,
	}
}

func (s *SizeStrategy) Name() string        { return "size_analysis" }
func (s *SizeStrategy) Category() string    { return "structural" }
func (s *SizeStrategy) Confidence() float64 { return 0.6 }
func (s *SizeStrategy) Description() string { return "Detects unusually large commit sizes" }

func (s *SizeStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if s.suspiciousAdditions > 0 && pair.Stats.Additions > s.suspiciousAdditions {
		return true, fmt.Sprintf(
			"Suspicious commit size: %d additions (threshold: %d lines)",
			pair.Stats.Additions, s.suspiciousAdditions,
		)
	}

	if s.suspiciousDeletions > 0 && pair.Stats.Deletions > s.suspiciousDeletions {
		return true, fmt.Sprintf(
			"Suspicious commit size: %d deletions (threshold: %d lines)",
			pair.Stats.Deletions, s.suspiciousDeletions,
		)
	}

	return false, ""
}

type TimingStrategy struct {
	minTimeDeltaSeconds int64
}

func NewTimingStrategy(minSeconds int64) *TimingStrategy {
	return &TimingStrategy{
		minTimeDeltaSeconds: minSeconds,
	}
}

func (s *TimingStrategy) Name() string        { return "timing_analysis" }
func (s *TimingStrategy) Category() string    { return "behavioral" }
func (s *TimingStrategy) Confidence() float64 { return 0.6 }
func (s *TimingStrategy) Description() string {
	return "Detects suspiciously short time between commits"
}

func (s *TimingStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if s.minTimeDeltaSeconds > 0 {
		if pair.TimeDelta.Seconds() < float64(s.minTimeDeltaSeconds) {
			return true, fmt.Sprintf(
				"Time between commits too short: %.1f seconds (threshold: %d seconds)",
				pair.TimeDelta.Seconds(), s.minTimeDeltaSeconds,
			)
		}
	}

	return false, ""
}

type MergeCommitStrategy struct {
	flagAsSuspicious bool
}

func NewMergeCommitStrategy(flag bool) *MergeCommitStrategy {
	return &MergeCommitStrategy{
		flagAsSuspicious: flag,
	}
}

func (s *MergeCommitStrategy) Name() string        { return "merge_commit_filter" }
func (s *MergeCommitStrategy) Category() string    { return "structural" }
func (s *MergeCommitStrategy) Confidence() float64 { return 0.3 }
func (s *MergeCommitStrategy) Description() string {
	return "Analyzes merge commits for potential history rewrites"
}

func (s *MergeCommitStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if len(pair.Current.Parents) > 1 {
		if s.flagAsSuspicious {
			return true, fmt.Sprintf(
				"Merge commit with %d parents (potential history rewrite)",
				len(pair.Current.Parents),
			)
		}
		return false, ""
	}
	return false, ""
}
