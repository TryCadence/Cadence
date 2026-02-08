package patterns

import (
	"fmt"
)

type EmojiStrategy struct{}

func NewEmojiStrategy() *EmojiStrategy {
	return &EmojiStrategy{}
}

func (s *EmojiStrategy) Name() string        { return "emoji_overuse" }
func (s *EmojiStrategy) Category() string    { return "pattern" }
func (s *EmojiStrategy) Confidence() float64 { return 0.4 }
func (s *EmojiStrategy) Description() string { return "Detects excessive emoji usage in content" }

func isEmoji(r rune) bool {
	return r >= 0x1F300 && r <= 0x1F9FF ||
		r >= 0x2600 && r <= 0x27BF ||
		r >= 0x1F600 && r <= 0x1F64F ||
		r >= 0x1F900 && r <= 0x1F9FF ||
		r >= 0x2300 && r <= 0x23FF ||
		r >= 0x2000 && r <= 0x206F
}

func countEmojis(text string) int {
	count := 0
	for _, r := range text {
		if isEmoji(r) {
			count++
		}
	}
	return count
}

func (s *EmojiStrategy) Detect(content string, wordCount int) *DetectionResult {
	emojiCount := countEmojis(content)
	textLength := len([]rune(content))

	if textLength == 0 || emojiCount == 0 {
		return &DetectionResult{Detected: false}
	}

	emojiRatio := float64(emojiCount) / float64(textLength)

	if emojiCount >= 5 && emojiRatio > 0.05 {
		severity := emojiRatio * 100
		if severity > 1 {
			severity = 1
		}

		return &DetectionResult{
			Detected:    true,
			Type:        "emoji_overuse",
			Severity:    severity,
			Description: fmt.Sprintf("Excessive emoji usage detected (%d emojis, %.1f%% of content)", emojiCount, emojiRatio*100),
			Examples:    []string{fmt.Sprintf("Found %d emoji characters", emojiCount)},
		}
	}

	if emojiRatio > 0.15 {
		return &DetectionResult{
			Detected:    true,
			Type:        "emoji_overuse",
			Severity:    0.8,
			Description: fmt.Sprintf("High emoji density detected (%.1f%% of content)", emojiRatio*100),
			Examples:    []string{"Text is predominantly emoji characters"},
		}
	}

	return &DetectionResult{Detected: false}
}
