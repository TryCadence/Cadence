package git

import (
	"math"
	"time"
)

// Analyzer provides git-specific analysis operations
type Analyzer struct {
	repo Repository
}

// NewAnalyzer creates a new git analyzer instance
func NewAnalyzer(repo Repository) *Analyzer {
	return &Analyzer{
		repo: repo,
	}
}

// AnalyzeCommitHistory analyzes the commit history for patterns
type CommitHistoryAnalysis struct {
	TotalCommits      int
	DateRange         time.Duration
	CommitFrequency   float64 // commits per day
	AuthorDiversity   float64 // 0.0-1.0: how diverse the author set is
	BranchingPatterns int     // Number of merge commits
	RegularityScore   float64 // 0.0-1.0: how regular commit patterns are
	LargeCommitCount  int     // Number of commits over 500 LOC
	BurstActivity     int     // Number of rapid-fire commit bursts
}

// AnalyzeCommitPatterns analyzes commit history for anomalies
func (a *Analyzer) AnalyzeCommitPatterns(commits []*Commit) *CommitHistoryAnalysis {
	analysis := &CommitHistoryAnalysis{}

	if len(commits) == 0 {
		return analysis
	}

	analysis.TotalCommits = len(commits)

	// Calculate date range
	if len(commits) > 1 {
		lastIdx := len(commits) - 1
		firstTime := commits[0].Timestamp
		lastTime := commits[lastIdx].Timestamp
		analysis.DateRange = firstTime.Sub(lastTime)
		if analysis.DateRange < 0 {
			analysis.DateRange = -analysis.DateRange
		}

		if analysis.DateRange.Hours() > 0 {
			analysis.CommitFrequency = float64(len(commits)) / analysis.DateRange.Hours() * 24
		}
	}

	// Calculate time deltas for regularity
	var timeDelta []time.Duration
	for i := 0; i < len(commits)-1; i++ {
		delta := commits[i].Timestamp.Sub(commits[i+1].Timestamp)
		if delta < 0 {
			delta = -delta
		}
		timeDelta = append(timeDelta, delta)
	}

	analysis.RegularityScore = calculateRegularityScore(timeDelta)
	analysis.BranchingPatterns = countMergeCommits(commits)

	return analysis
}

// AnalyzeAuthorBehavior analyzes patterns specific to each author
type AuthorBehavior struct {
	Author          string
	Email           string
	CommitCount     int
	AverageSize     int64
	MaxSize         int64
	MinSize         int64
	AverageVelocity float64
	TimeActive      time.Duration
	IsConsistent    bool
	AnomalyFlags    []string
}

// AnalyzeAuthorBehaviors analyzes individual author patterns
func (a *Analyzer) AnalyzeAuthorBehaviors(pairs []*CommitPair) map[string]*AuthorBehavior {
	behaviors := make(map[string]*AuthorBehavior)

	for _, pair := range pairs {
		author := pair.Current.Email
		if _, exists := behaviors[author]; !exists {
			behaviors[author] = &AuthorBehavior{
				Author:  pair.Current.Author,
				Email:   author,
				MaxSize: -1,
			}
		}

		behavior := behaviors[author]
		behavior.CommitCount++

		totalSize := pair.Stats.Additions + pair.Stats.Deletions
		behavior.AverageSize += totalSize
		if totalSize > behavior.MaxSize {
			behavior.MaxSize = totalSize
		}
		if behavior.MinSize == -1 || totalSize < behavior.MinSize {
			behavior.MinSize = totalSize
		}
	}

	// Finalize calculations
	for _, behavior := range behaviors {
		if behavior.CommitCount > 0 {
			behavior.AverageSize /= int64(behavior.CommitCount)
		}
		behavior.IsConsistent = isAuthorConsistent(behavior)
	}

	return behaviors
}

// Helper functions

func calculateRegularityScore(timeDelta []time.Duration) float64 {
	if len(timeDelta) < 2 {
		return 0.5
	}

	// Calculate standard deviation of time deltas
	var sum, sumSq float64
	for _, td := range timeDelta {
		hours := td.Hours()
		sum += hours
	}
	mean := sum / float64(len(timeDelta))

	for _, td := range timeDelta {
		hours := td.Hours()
		sumSq += (hours - mean) * (hours - mean)
	}
	variance := sumSq / float64(len(timeDelta))
	if variance == 0 {
		return 1.0 // Perfect regularity
	}

	// Lower coefficient of variation = more regular
	stdDev := variance / (mean + 1.0) // +1 to avoid division by zero
	if stdDev > 1.0 {
		stdDev = 1.0
	}
	return 1.0 - stdDev
}

func countMergeCommits(commits []*Commit) int {
	count := 0
	for _, commit := range commits {
		if len(commit.Parents) > 1 {
			count++
		}
	}
	return count
}

func isAuthorConsistent(behavior *AuthorBehavior) bool {
	if behavior.CommitCount < 3 {
		return false
	}

	// If max and min differ significantly, not consistent
	if behavior.MaxSize > 0 && behavior.MinSize > 0 {
		ratio := float64(behavior.MaxSize) / float64(behavior.MinSize)
		return ratio < 5.0 // Commits vary less than 5x
	}

	return true
}

// CalculateAuthorDiversity calculates how diverse the author set is
func CalculateAuthorDiversity(commits []*Commit) float64 {
	if len(commits) == 0 {
		return 0
	}

	authorCounts := make(map[string]int)
	for _, commit := range commits {
		authorCounts[commit.Email]++
	}

	// Shannon entropy
	entropy := 0.0
	for _, count := range authorCounts {
		p := float64(count) / float64(len(commits))
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}

	// Normalize to 0-1
	maxEntropy := math.Log2(float64(len(authorCounts)))
	if maxEntropy == 0 {
		return 0
	}

	return entropy / maxEntropy
}
