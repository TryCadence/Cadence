package git

import (
	"math"
)

// CalculateAuthorDiversity calculates how diverse the author set is
// This is used in metrics package to analyze repository health
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
