package patterns

import (
	"fmt"
	"strings"
)

type SpecialCharactersStrategy struct{}

func NewSpecialCharactersStrategy() *SpecialCharactersStrategy {
	return &SpecialCharactersStrategy{}
}

func (s *SpecialCharactersStrategy) Name() string        { return "special_characters" }
func (s *SpecialCharactersStrategy) Category() string    { return "pattern" }
func (s *SpecialCharactersStrategy) Confidence() float64 { return 0.4 }
func (s *SpecialCharactersStrategy) Description() string {
	return "Detects excessive special character patterns"
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

func (s *SpecialCharactersStrategy) Detect(content string, wordCount int) *DetectionResult {
	textLength := len([]rune(content))
	if textLength == 0 {
		return &DetectionResult{Detected: false}
	}

	hyphenCount := countCharacter(content, '-')
	asteriskCount := countCharacter(content, '*')
	underscoreCount := countCharacter(content, '_')
	totalSpecialChars := hyphenCount + asteriskCount + underscoreCount

	specialCharRatio := float64(totalSpecialChars) / float64(textLength)

	if hyphenCount >= 10 {
		return &DetectionResult{
			Detected:    true,
			Type:        "special_characters",
			Severity:    0.7,
			Description: fmt.Sprintf("Excessive hyphens detected (%d occurrences)", hyphenCount),
			Examples:    []string{fmt.Sprintf("Found %d hyphens in text", hyphenCount)},
		}
	}

	if strings.Count(content, "---") > 2 || strings.Count(content, "---") > 1 && wordCount < 100 {
		return &DetectionResult{
			Detected:    true,
			Type:        "special_characters",
			Severity:    0.65,
			Description: "Multiple consecutive hyphen sequences (---) detected",
			Examples:    []string{"Unusual dash patterns found in text"},
		}
	}

	if asteriskCount >= 6 {
		return &DetectionResult{
			Detected:    true,
			Type:        "special_characters",
			Severity:    0.6,
			Description: fmt.Sprintf("Excessive asterisks detected (%d occurrences)", asteriskCount),
			Examples:    []string{fmt.Sprintf("Found %d asterisks in text", asteriskCount)},
		}
	}

	if underscoreCount >= 6 {
		return &DetectionResult{
			Detected:    true,
			Type:        "special_characters",
			Severity:    0.6,
			Description: fmt.Sprintf("Excessive underscores detected (%d occurrences)", underscoreCount),
			Examples:    []string{fmt.Sprintf("Found %d underscores in text", underscoreCount)},
		}
	}

	if specialCharRatio > 0.15 {
		return &DetectionResult{
			Detected:    true,
			Type:        "special_characters",
			Severity:    0.55,
			Description: fmt.Sprintf("High special character density (%.1f%% of text)", specialCharRatio*100),
			Examples:    []string{"Unusual pattern of special characters"},
		}
	}

	return &DetectionResult{Detected: false}
}
