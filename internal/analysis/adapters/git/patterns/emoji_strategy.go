package patterns

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/metrics"
)

type EmojiPatternStrategy struct {
	enabled bool
}

func NewEmojiPatternStrategy() *EmojiPatternStrategy {
	return &EmojiPatternStrategy{enabled: true}
}

func (s *EmojiPatternStrategy) Name() string        { return "emoji_pattern_analysis" }
func (s *EmojiPatternStrategy) Category() string    { return "pattern" }
func (s *EmojiPatternStrategy) Confidence() float64 { return 0.4 }
func (s *EmojiPatternStrategy) Description() string {
	return "Detects excessive emoji usage in commit messages"
}

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

func isEmojiHeavy(text string) bool {
	if text == "" {
		return false
	}

	emojiCount := countEmojis(text)
	if emojiCount == 0 {
		return false
	}

	emojiRatio := float64(emojiCount) / float64(len([]rune(text)))
	return emojiRatio > 0.2
}

func hasEmojiOnlyCommit(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}

	emojiCount := countEmojis(trimmed)
	if emojiCount == 0 {
		return false
	}

	nonEmojiChars := 0
	for _, r := range trimmed {
		if !isEmoji(r) && !unicode.IsSpace(r) {
			nonEmojiChars++
		}
	}

	return emojiCount > nonEmojiChars*2
}

func (s *EmojiPatternStrategy) Detect(pair *git.CommitPair, repoStats *metrics.RepositoryStats) (isSuspicious bool, reason string) {
	if !s.enabled || pair == nil || pair.Current == nil {
		return false, ""
	}

	msg := pair.Current.Message

	emojiCount := countEmojis(msg)
	msgLength := len([]rune(msg))

	if emojiCount >= 3 {
		return true, fmt.Sprintf(
			"Excessive emoji usage in commit message (%d emojis detected)",
			emojiCount,
		)
	}

	if isEmojiHeavy(msg) {
		return true, fmt.Sprintf(
			"Commit message is predominantly emoji (%.1f%% emoji content)",
			float64(emojiCount)/float64(msgLength)*100,
		)
	}

	if hasEmojiOnlyCommit(msg) {
		return true, "Commit message is primarily emoji with minimal text content"
	}

	return false, ""
}
