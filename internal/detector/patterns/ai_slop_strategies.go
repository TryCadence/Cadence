package patterns

import (
	"fmt"
	"strings"

	"github.com/codemeapixel/cadence/internal/git"
	"github.com/codemeapixel/cadence/internal/metrics"
)

type CommitMessageStrategy struct {
	enabled bool
}

func NewCommitMessageStrategy() *CommitMessageStrategy {
	return &CommitMessageStrategy{enabled: true}
}

func (s *CommitMessageStrategy) Name() string {
	return "commit_message_analysis"
}

func (s *CommitMessageStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	msg := strings.ToLower(pair.Current.Message)

	aiPatterns := []string{
		"implement",
		"add functionality",
		"update code",
		"refactor code",
		"improve implementation",
		"enhance functionality",
		"optimize performance",
		"fix issues",
		"improve code quality",
		"add new features",
		"update implementation",
		"add support for",
	}

	genericPatterns := []string{
		"initial commit",
		"update readme",
		"update dependencies",
		"minor fixes",
		"code cleanup",
		"bug fixes",
		"improvements",
		"updates",
		"changes",
		"modifications",
	}

	aiScore := 0
	genericScore := 0

	for _, pattern := range aiPatterns {
		if strings.Contains(msg, pattern) {
			aiScore++
		}
	}

	for _, pattern := range genericPatterns {
		if strings.Contains(msg, pattern) {
			genericScore++
		}
	}

	if aiScore >= 2 || genericScore >= 1 {
		return true, fmt.Sprintf(
			"Suspicious commit message patterns - generic/AI-like phrasing (AI patterns: %d, generic: %d)",
			aiScore, genericScore,
		)
	}

	words := strings.Fields(msg)
	if len(words) > 8 && (strings.Contains(msg, "implement") || strings.Contains(msg, "functionality")) {
		return true, "Overly verbose yet generic commit message - typical of AI generation"
	}

	return false, ""
}

type NamingPatternStrategy struct {
	enabled bool
}

func NewNamingPatternStrategy() *NamingPatternStrategy {
	return &NamingPatternStrategy{enabled: true}
}

func (s *NamingPatternStrategy) Name() string {
	return "naming_pattern_analysis"
}

func (s *NamingPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	if pair.DiffContent != "" {
		return s.analyzeCodeContent(pair.DiffContent)
	}

	msg := strings.ToLower(pair.Current.Message)
	genericNames := []string{
		"variable", "function", "method", "class", "object", "instance",
		"data", "result", "value", "item", "element", "component",
		"helper", "utility", "manager", "handler", "service",
	}

	nameCount := 0
	for _, name := range genericNames {
		if strings.Contains(msg, name) {
			nameCount++
		}
	}

	if nameCount >= 2 {
		return true, fmt.Sprintf(
			"Commit message contains multiple generic naming terms (%d) - may indicate AI-generated variable names",
			nameCount,
		)
	}

	return false, ""
}

func (s *NamingPatternStrategy) analyzeCodeContent(diffContent string) (bool, string) {
	lines := strings.Split(diffContent, "\n")
	addedLines := make([]string, 0)

	// Extract only added lines (starting with +)
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, strings.TrimPrefix(line, "+"))
		}
	}

	if len(addedLines) == 0 {
		return false, ""
	}

	suspiciousPatterns := 0
	totalPatterns := 0

	codeContent := strings.Join(addedLines, "\n")

	genericVarPatterns := []string{
		"var1", "var2", "temp", "data", "result", "value", "item",
		"element", "obj", "instance", "helper", "utility", "manager",
	}
	for _, pattern := range genericVarPatterns {
		if strings.Contains(strings.ToLower(codeContent), pattern) {
			suspiciousPatterns++
		}
		totalPatterns++
	}

	todoCount := strings.Count(strings.ToLower(codeContent), "todo")
	fixmeCount := strings.Count(strings.ToLower(codeContent), "fixme")
	if todoCount > 2 || fixmeCount > 1 {
		suspiciousPatterns++
	}

	words := strings.Fields(codeContent)
	perfectCamelCaseCount := 0
	for _, word := range words {
		if len(word) > 4 && isPerfectCamelCase(word) {
			perfectCamelCaseCount++
		}
	}
	if len(words) > 10 && float64(perfectCamelCaseCount)/float64(len(words)) > 0.3 {
		suspiciousPatterns++
	}

	if strings.Contains(codeContent, "catch") || strings.Contains(codeContent, "except") {
		errorHandlingCount := strings.Count(codeContent, "catch") + strings.Count(codeContent, "except") + strings.Count(codeContent, "try")
		if errorHandlingCount > len(addedLines)/20 {
			suspiciousPatterns++
		}
	}

	if suspiciousPatterns >= 2 {
		return true, fmt.Sprintf(
			"Code contains multiple AI-slop patterns (%d detected) - generic names, TODO comments, perfect patterns",
			suspiciousPatterns,
		)
	}

	return false, ""
}

func isPerfectCamelCase(word string) bool {
	if len(word) < 2 {
		return false
	}

	if !(word[0] >= 'a' && word[0] <= 'z') {
		return false
	}

	uppercaseCount := 0
	for _, r := range word {
		if r >= 'A' && r <= 'Z' {
			uppercaseCount++
		}
	}

	return uppercaseCount == 1
}

type StructuralConsistencyStrategy struct {
	enabled bool
}

func NewStructuralConsistencyStrategy() *StructuralConsistencyStrategy {
	return &StructuralConsistencyStrategy{enabled: true}
}

func (s *StructuralConsistencyStrategy) Name() string {
	return "structural_consistency_analysis"
}

func (s *StructuralConsistencyStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	if pair.Stats.Additions > 100 && pair.Stats.Deletions > 100 {
		ratio := float64(pair.Stats.Additions) / float64(pair.Stats.Deletions)

		if ratio >= 0.9 && ratio <= 1.1 {
			return true, fmt.Sprintf(
				"Suspiciously balanced addition/deletion ratio: %.2f - may indicate automated refactoring",
				ratio,
			)
		}

		// Or very consistent ratios (e.g., exactly 2:1, 3:1)
		if isNearInteger(ratio, 0.05) || isNearInteger(1.0/ratio, 0.05) {
			return true, fmt.Sprintf(
				"Suspiciously consistent addition/deletion ratio: %.2f - may indicate template-based generation",
				ratio,
			)
		}
	}

	return false, ""
}

type BurstPatternStrategy struct {
	maxCommitsPerHour int
	enabled           bool
}

func NewBurstPatternStrategy(maxPerHour int) *BurstPatternStrategy {
	return &BurstPatternStrategy{
		maxCommitsPerHour: maxPerHour,
		enabled:           true,
	}
}

func (s *BurstPatternStrategy) Name() string {
	return "burst_pattern_analysis"
}

func (s *BurstPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled || s.maxCommitsPerHour <= 0 {
		return false, ""
	}

	if pair.TimeDelta.Seconds() < 300 && pair.Stats.Additions > 50 {
		return true, fmt.Sprintf(
			"Rapid commit pattern: %.1f seconds between substantial commits - may indicate batch processing",
			pair.TimeDelta.Seconds(),
		)
	}

	return false, ""
}

type ErrorHandlingPatternStrategy struct {
	enabled bool
}

func NewErrorHandlingPatternStrategy() *ErrorHandlingPatternStrategy {
	return &ErrorHandlingPatternStrategy{enabled: true}
}

func (s *ErrorHandlingPatternStrategy) Name() string {
	return "error_handling_analysis"
}

func (s *ErrorHandlingPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	// Analyze actual code content if available
	if pair.DiffContent != "" && pair.Stats.Additions > 50 {
		return s.analyzeErrorHandling(pair.DiffContent, pair.Stats.Additions)
	}

	if pair.Stats.Additions > 200 {
		msg := strings.ToLower(pair.Current.Message)
		hasErrorPatterns := strings.Contains(msg, "error") ||
			strings.Contains(msg, "exception") ||
			strings.Contains(msg, "try") ||
			strings.Contains(msg, "catch") ||
			strings.Contains(msg, "handle")

		if !hasErrorPatterns && pair.Stats.Additions > 300 {
			return true, fmt.Sprintf(
				"Large code addition (%d lines) with no error handling mentions - AI often omits error handling",
				pair.Stats.Additions,
			)
		}
	}

	return false, ""
}

func (s *ErrorHandlingPatternStrategy) analyzeErrorHandling(diffContent string, additions int64) (bool, string) {
	lines := strings.Split(diffContent, "\n")
	addedLines := make([]string, 0)

	// Extract only added lines
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, strings.TrimPrefix(line, "+"))
		}
	}

	if len(addedLines) < 20 {
		return false, ""
	}

	codeContent := strings.ToLower(strings.Join(addedLines, "\n"))

	errorHandlingPatterns := 0

	errorHandlingPatterns += strings.Count(codeContent, "try")
	errorHandlingPatterns += strings.Count(codeContent, "catch")
	errorHandlingPatterns += strings.Count(codeContent, "except")
	errorHandlingPatterns += strings.Count(codeContent, "throw")
	errorHandlingPatterns += strings.Count(codeContent, "throws")

	errorHandlingPatterns += strings.Count(codeContent, "error")
	errorHandlingPatterns += strings.Count(codeContent, "exception")
	errorHandlingPatterns += strings.Count(codeContent, "handle")

	errorHandlingPatterns += strings.Count(codeContent, "if err != nil")
	errorHandlingPatterns += strings.Count(codeContent, ".catch(") // JavaScript promises
	errorHandlingPatterns += strings.Count(codeContent, "rescue")  // Ruby

	// Calculate expected error handling density
	expectedErrorHandling := len(addedLines) / 30

	if additions > 100 && errorHandlingPatterns < expectedErrorHandling {
		return true, fmt.Sprintf(
			"Large code addition (%d lines) with insufficient error handling (%d patterns, expected ~%d) - typical AI omission",
			additions, errorHandlingPatterns, expectedErrorHandling,
		)
	}

	if errorHandlingPatterns > len(addedLines)/5 {
		return true, fmt.Sprintf(
			"Excessive error handling patterns (%d in %d lines) - may indicate AI over-compensation",
			errorHandlingPatterns, len(addedLines),
		)
	}

	return false, ""
}

type TemplatePatternStrategy struct {
	enabled bool
}

func NewTemplatePatternStrategy() *TemplatePatternStrategy {
	return &TemplatePatternStrategy{enabled: true}
}

func (s *TemplatePatternStrategy) Name() string {
	return "template_pattern_analysis"
}

func (s *TemplatePatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	// Analyze actual code content if available
	if pair.DiffContent != "" && pair.Stats.Additions > 50 {
		detected, reason := s.analyzeTemplatePatterns(pair.DiffContent)
		if detected {
			return detected, reason
		}
	}

	msg := strings.ToLower(pair.Current.Message)
	templatePatterns := []string{
		"boilerplate", "template", "skeleton", "scaffold", "stub", "placeholder", "todo", "fixme",
	}

	patternCount := 0
	for _, pattern := range templatePatterns {
		if strings.Contains(msg, pattern) {
			patternCount++
		}
	}

	if patternCount > 0 && pair.Stats.Additions > 100 {
		return true, fmt.Sprintf(
			"Template/boilerplate patterns detected in large commit (%d lines) - may be AI-generated scaffold",
			pair.Stats.Additions,
		)
	}

	return false, ""
}

func (s *TemplatePatternStrategy) analyzeTemplatePatterns(diffContent string) (bool, string) {
	lines := strings.Split(diffContent, "\n")
	addedLines := make([]string, 0)

	// Extract only added lines
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines = append(addedLines, strings.TrimPrefix(line, "+"))
		}
	}

	if len(addedLines) < 10 {
		return false, ""
	}

	codeContent := strings.Join(addedLines, "\n")
	lowerContent := strings.ToLower(codeContent)

	suspiciousPatterns := 0

	templateComments := []string{
		"todo", "fixme", "placeholder", "implement", "add code here",
		"your code here", "example", "sample", "template", "boilerplate",
	}
	for _, pattern := range templateComments {
		if strings.Contains(lowerContent, pattern) {
			suspiciousPatterns++
		}
	}

	repetitiveCount := 0
	for i := 0; i < len(addedLines)-1; i++ {
		line1 := strings.TrimSpace(addedLines[i])
		line2 := strings.TrimSpace(addedLines[i+1])

		if len(line1) > 10 && len(line2) > 10 {
			words1 := strings.Fields(line1)
			words2 := strings.Fields(line2)

			if len(words1) == len(words2) && len(words1) > 3 {
				differences := 0
				for j := 0; j < len(words1); j++ {
					if words1[j] != words2[j] {
						differences++
					}
				}
				if differences <= 2 {
					repetitiveCount++
				}
			}
		}
	}

	if repetitiveCount > len(addedLines)/8 {
		suspiciousPatterns++
	}

	indentationLevels := make(map[int]int)
	for _, line := range addedLines {
		if strings.TrimSpace(line) != "" {
			indent := 0
			for _, r := range line {
				if r == ' ' {
					indent++
				} else if r == '\t' {
					indent += 4
				} else {
					break
				}
			}
			indentationLevels[indent]++
		}
	}

	totalLines := len(addedLines)
	if totalLines > 20 {
		consistentIndentation := 0
		for _, count := range indentationLevels {
			if count > totalLines/10 {
				consistentIndentation += count
			}
		}
		if float64(consistentIndentation)/float64(totalLines) > 0.9 {
			suspiciousPatterns++
		}
	}

	if strings.Contains(lowerContent, "import") {
		importLines := make([]string, 0)
		for _, line := range addedLines {
			if strings.Contains(strings.ToLower(line), "import") {
				importLines = append(importLines, line)
			}
		}

		if len(importLines) > 5 && len(addedLines) < len(importLines)*10 {
			suspiciousPatterns++
		}
	}

	if suspiciousPatterns >= 2 {
		return true, fmt.Sprintf(
			"Template/generated code patterns detected (%d indicators) - repetitive structure, perfect formatting, template comments",
			suspiciousPatterns,
		)
	}

	return false, ""
}

func isNearInteger(value, tolerance float64) bool {
	remainder := value - float64(int(value))
	return remainder < tolerance || remainder > (1.0-tolerance)
}

type FileExtensionPatternStrategy struct {
	enabled bool
}

func NewFileExtensionPatternStrategy() *FileExtensionPatternStrategy {
	return &FileExtensionPatternStrategy{enabled: true}
}

func (s *FileExtensionPatternStrategy) Name() string {
	return "file_extension_analysis"
}

func (s *FileExtensionPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (bool, string) {
	if !s.enabled {
		return false, ""
	}

	if pair.Stats.FilesChanged > 10 && pair.Stats.Additions > 1000 {
		avgLinesPerFile := float64(pair.Stats.Additions) / float64(pair.Stats.FilesChanged)

		if avgLinesPerFile > 50 && avgLinesPerFile < 200 {
			consistency := 1.0 - (avgLinesPerFile - float64(int(avgLinesPerFile)))
			if consistency > 0.8 {
				return true, fmt.Sprintf(
					"Suspicious file creation pattern: %d files with consistent size (~%.0f lines each) - may be generated",
					pair.Stats.FilesChanged, avgLinesPerFile,
				)
			}
		}
	}

	return false, ""
}
