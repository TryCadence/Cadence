package patterns

import (
	"fmt"
	"strings"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/metrics"
)

type SpecialCharacterPatternStrategy struct {
	enabled bool
}

func NewSpecialCharacterPatternStrategy() *SpecialCharacterPatternStrategy {
	return &SpecialCharacterPatternStrategy{enabled: true}
}

func (s *SpecialCharacterPatternStrategy) Name() string        { return "special_character_pattern_analysis" }
func (s *SpecialCharacterPatternStrategy) Category() string    { return "pattern" }
func (s *SpecialCharacterPatternStrategy) Confidence() float64 { return 0.4 }
func (s *SpecialCharacterPatternStrategy) Description() string {
	return "Detects unusual special character patterns in commits"
}

func countCharacter(text string, char rune) int {
	count := 0
	for _, r := range text {
		if r == char {
			count++
		}
	}
	return count
}

func detectConsecutiveCharacters(text string, char rune, threshold int) bool {
	consecutive := 0
	for _, r := range text {
		if r == char {
			consecutive++
			if consecutive >= threshold {
				return true
			}
		} else {
			consecutive = 0
		}
	}
	return false
}

func getConsecutiveCount(text string, char rune) int {
	maxConsecutive := 0
	consecutive := 0
	for _, r := range text {
		if r == char {
			consecutive++
			if consecutive > maxConsecutive {
				maxConsecutive = consecutive
			}
		} else {
			consecutive = 0
		}
	}
	return maxConsecutive
}

func splitByLines(text string) []string {
	return strings.Split(text, "\n")
}

func countSpecialCharactersInLine(line string) map[rune]int {
	specialChars := make(map[rune]int)
	suspiciousChars := []rune{'-', '*', '_', '=', '+', '#', '!', '?', '.', ','}

	for _, char := range suspiciousChars {
		if count := countCharacter(line, char); count > 0 {
			specialChars[char] = count
		}
	}

	return specialChars
}

func (s *SpecialCharacterPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if !s.enabled || pair == nil || pair.Current == nil {
		return false, ""
	}

	msg := pair.Current.Message
	lines := splitByLines(msg)

	hyphenCount := countCharacter(msg, '-')
	if hyphenCount >= 5 {
		return true, fmt.Sprintf(
			"Excessive hyphens in commit message (%d occurrences)",
			hyphenCount,
		)
	}

	if detectConsecutiveCharacters(msg, '-', 3) {
		maxConsec := getConsecutiveCount(msg, '-')
		return true, fmt.Sprintf(
			"Multiple consecutive hyphens detected (%d in a row)",
			maxConsec,
		)
	}

	asteriskCount := countCharacter(msg, '*')
	if asteriskCount >= 4 {
		return true, fmt.Sprintf(
			"Excessive asterisks in commit message (%d occurrences)",
			asteriskCount,
		)
	}

	if detectConsecutiveCharacters(msg, '*', 3) {
		maxConsec := getConsecutiveCount(msg, '*')
		return true, fmt.Sprintf(
			"Multiple consecutive asterisks detected (%d in a row)",
			maxConsec,
		)
	}

	underscoreCount := countCharacter(msg, '_')
	if underscoreCount >= 4 {
		return true, fmt.Sprintf(
			"Excessive underscores in commit message (%d occurrences)",
			underscoreCount,
		)
	}

	specialCharPatterns := []string{
		"---", "***", "===", "+++", "!!!",
	}

	for _, pattern := range specialCharPatterns {
		if strings.Count(msg, pattern) > 1 {
			return true, fmt.Sprintf(
				"Special character pattern clustering detected: '%s' repeated %d times",
				pattern, strings.Count(msg, pattern),
			)
		}
	}

	for lineIdx, line := range lines {
		if line == "" {
			continue
		}

		charCounts := countSpecialCharactersInLine(line)
		totalSpecialChars := 0
		for _, count := range charCounts {
			totalSpecialChars += count
		}

		lineLength := len([]rune(line))
		if lineLength > 5 && float64(totalSpecialChars)/float64(lineLength) > 0.25 {
			return true, fmt.Sprintf(
				"High special character density on line %d (%.1f%% of line)",
				lineIdx+1, float64(totalSpecialChars)/float64(lineLength)*100,
			)
		}
	}

	return false, ""
}
