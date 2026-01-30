package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

type ExcessiveStructureStrategy struct{}

func NewExcessiveStructureStrategy() *ExcessiveStructureStrategy {
	return &ExcessiveStructureStrategy{}
}

func (s *ExcessiveStructureStrategy) Name() string {
	return "excessive_structure"
}

func (s *ExcessiveStructureStrategy) Detect(content string, wordCount int) *DetectionResult {
	lines := strings.Split(content, "\n")
	listItemCount := 0
	headingCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") ||
			strings.HasPrefix(trimmed, "• ") || regexp.MustCompile(`^\d+\.\s`).MatchString(trimmed) {
			listItemCount++
		}
		if strings.HasPrefix(trimmed, "#") || (len(trimmed) > 0 && len(trimmed) < 60 && strings.ToUpper(trimmed) == trimmed) {
			headingCount++
		}
	}

	if wordCount > 100 && listItemCount > wordCount/30 {
		severity := float64(listItemCount) / (float64(wordCount) / 25.0)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Excessive use of lists and bullet points (%d items) - AI tends to over-structure", listItemCount),
			Examples:    []string{fmt.Sprintf("%d list items detected", listItemCount)},
		}
	}

	return nil
}

type PerfectGrammarStrategy struct{}

func NewPerfectGrammarStrategy() *PerfectGrammarStrategy {
	return &PerfectGrammarStrategy{}
}

func (s *PerfectGrammarStrategy) Name() string {
	return "perfect_grammar"
}

func (s *PerfectGrammarStrategy) Detect(content string, wordCount int) *DetectionResult {
	sentences := regexp.MustCompile(`[.!?]+`).Split(content, -1)
	if len(sentences) < 5 {
		return nil
	}

	perfectCount := 0
	for _, sentence := range sentences {
		trimmed := strings.TrimSpace(sentence)
		if len(trimmed) > 20 {
			words := strings.Fields(trimmed)
			if len(words) >= 8 && len(words) <= 25 {
				perfectCount++
			}
		}
	}

	perfectRatio := float64(perfectCount) / float64(len(sentences))
	if perfectRatio > 0.7 && len(sentences) > 10 {
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    perfectRatio,
			Description: fmt.Sprintf("Suspiciously consistent sentence structure (%.0f%% are 8-25 words)", perfectRatio*100),
			Examples:    []string{fmt.Sprintf("%d of %d sentences have 'perfect' length", perfectCount, len(sentences))},
		}
	}

	return nil
}

type BoilerplateTextStrategy struct{}

func NewBoilerplateTextStrategy() *BoilerplateTextStrategy {
	return &BoilerplateTextStrategy{}
}

func (s *BoilerplateTextStrategy) Name() string {
	return "boilerplate_text"
}

func (s *BoilerplateTextStrategy) Detect(content string, wordCount int) *DetectionResult {
	boilerplateIndicators := []string{
		"welcome to", "this article", "this post", "in this guide",
		"let's dive", "let's explore", "let's take a look",
		"it's important to understand", "by the end of this",
		"in this section", "as we've seen", "as discussed",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	found := make([]string, 0)

	for _, indicator := range boilerplateIndicators {
		if strings.Contains(lowerContent, indicator) {
			count++
			if len(found) < 3 {
				found = append(found, indicator)
			}
		}
	}

	if count >= 2 {
		severity := float64(count) / 10.0
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Contains common boilerplate phrases (%d detected)", count),
			Examples:    found,
		}
	}

	return nil
}

type RepetitivePatternsStrategy struct{}

func NewRepetitivePatternsStrategy() *RepetitivePatternsStrategy {
	return &RepetitivePatternsStrategy{}
}

func (s *RepetitivePatternsStrategy) Name() string {
	return "repetitive_patterns"
}

func (s *RepetitivePatternsStrategy) Detect(content string, wordCount int) *DetectionResult {
	sentences := regexp.MustCompile(`[.!?]+\s+`).Split(content, -1)
	if len(sentences) < 5 {
		return nil
	}

	repetitiveCount := 0
	for i := 0; i < len(sentences)-1; i++ {
		words1 := strings.Fields(strings.ToLower(sentences[i]))
		words2 := strings.Fields(strings.ToLower(sentences[i+1]))

		if len(words1) > 3 && len(words2) > 3 {
			if words1[0] == words2[0] && len(words1) == len(words2) {
				repetitiveCount++
			}
		}
	}

	if repetitiveCount > len(sentences)/8 {
		severity := float64(repetitiveCount) / float64(len(sentences)/4)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Repetitive sentence patterns detected (%d similar consecutive sentences)", repetitiveCount),
			Examples:    []string{fmt.Sprintf("%d repetitive patterns in %d sentences", repetitiveCount, len(sentences))},
		}
	}

	return nil
}

type MissingNuanceStrategy struct{}

func NewMissingNuanceStrategy() *MissingNuanceStrategy {
	return &MissingNuanceStrategy{}
}

func (s *MissingNuanceStrategy) Name() string {
	return "missing_nuance"
}

func (s *MissingNuanceStrategy) Detect(content string, wordCount int) *DetectionResult {
	absoluteTerms := []string{
		"always", "never", "all", "none", "every", "completely",
		"entirely", "totally", "absolutely", "definitely", "certainly",
	}

	lowerContent := strings.ToLower(content)
	count := 0

	for _, term := range absoluteTerms {
		count += strings.Count(lowerContent, " "+term+" ")
	}

	if wordCount > 100 && count > wordCount/50 {
		severity := float64(count) / (float64(wordCount) / 40.0)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Excessive use of absolute terms without nuance (%d instances)", count),
			Examples:    []string{"always", "never", "completely", "absolutely"},
		}
	}

	return nil
}

type ExcessiveTransitionsStrategy struct{}

func NewExcessiveTransitionsStrategy() *ExcessiveTransitionsStrategy {
	return &ExcessiveTransitionsStrategy{}
}

func (s *ExcessiveTransitionsStrategy) Name() string {
	return "excessive_transitions"
}

func (s *ExcessiveTransitionsStrategy) Detect(content string, wordCount int) *DetectionResult {
	transitions := []string{
		"however", "moreover", "furthermore", "additionally",
		"consequently", "therefore", "nevertheless", "nonetheless",
		"meanwhile", "subsequently", "accordingly",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	found := make([]string, 0)

	for _, transition := range transitions {
		occurrences := strings.Count(lowerContent, transition)
		if occurrences > 0 {
			count += occurrences
			if len(found) < 5 {
				found = append(found, transition)
			}
		}
	}

	if wordCount > 100 && count > wordCount/80 {
		severity := float64(count) / (float64(wordCount) / 60.0)
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Excessive use of transition words (%d in %d words)", count, wordCount),
			Examples:    found,
		}
	}

	return nil
}

type UniformSentenceLengthStrategy struct{}

func NewUniformSentenceLengthStrategy() *UniformSentenceLengthStrategy {
	return &UniformSentenceLengthStrategy{}
}

func (s *UniformSentenceLengthStrategy) Name() string {
	return "uniform_sentence_length"
}

func (s *UniformSentenceLengthStrategy) Detect(content string, wordCount int) *DetectionResult {
	sentences := regexp.MustCompile(`[.!?]+`).Split(content, -1)
	if len(sentences) < 5 {
		return nil
	}

	lengths := make([]int, 0)
	for _, sentence := range sentences {
		words := strings.Fields(strings.TrimSpace(sentence))
		if len(words) > 3 {
			lengths = append(lengths, len(words))
		}
	}

	if len(lengths) < 5 {
		return nil
	}

	sum := 0
	for _, l := range lengths {
		sum += l
	}
	avg := float64(sum) / float64(len(lengths))

	variance := 0.0
	for _, l := range lengths {
		variance += (float64(l) - avg) * (float64(l) - avg)
	}
	variance /= float64(len(lengths))

	if variance < 15.0 && len(lengths) > 10 {
		severity := (20.0 - variance) / 20.0
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Suspiciously uniform sentence lengths (variance: %.1f, avg: %.1f words)", variance, avg),
			Examples:    []string{fmt.Sprintf("Average: %.1f words, very low variance", avg)},
		}
	}

	return nil
}

type AIVocabularyStrategy struct{}

func NewAIVocabularyStrategy() *AIVocabularyStrategy {
	return &AIVocabularyStrategy{}
}

func (s *AIVocabularyStrategy) Name() string {
	return "ai_vocabulary"
}

func (s *AIVocabularyStrategy) Detect(content string, wordCount int) *DetectionResult {
	aiVocab := []string{
		"delve into", "delve deeper", "navigating", "landscape",
		"realm", "tapestry", "intricacies", "nuances",
		"unveil", "unravel", "embark", "journey",
		"paramount", "crucial", "pivotal", "vital",
		"plethora", "myriad", "multifaceted",
	}

	lowerContent := strings.ToLower(content)
	count := 0
	found := make([]string, 0)

	for _, term := range aiVocab {
		if strings.Contains(lowerContent, term) {
			count++
			if len(found) < 5 {
				found = append(found, term)
			}
		}
	}

	if count >= 3 {
		severity := float64(count) / 8.0
		if severity > 1.0 {
			severity = 1.0
		}
		return &DetectionResult{
			Detected:    true,
			Type:        s.Name(),
			Severity:    severity,
			Description: fmt.Sprintf("Contains AI-characteristic vocabulary (%d instances)", count),
			Examples:    found,
		}
	}

	return nil
}
