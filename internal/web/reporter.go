package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TryCadence/Cadence/internal/detector/patterns"
)

type WebReporter interface {
	Generate(data *WebReportData) (string, error)
}

type WebReportData struct {
	Content    *PageContent
	Analysis   *patterns.TextSlopResult
	AIAnalysis string
	AnalyzedAt time.Time
}

type JSONWebReporter struct{}

type JSONWebReport struct {
	URL          string               `json:"url"`
	Title        string               `json:"title"`
	StatusCode   int                  `json:"status_code"`
	AnalyzedAt   string               `json:"analyzed_at"`
	ContentStats JSONContentStats     `json:"content_stats"`
	Analysis     JSONAnalysisResult   `json:"analysis"`
	FlaggedItems []JSONFlaggedContent `json:"flagged_items"`
	AIAnalysis   string               `json:"ai_analysis,omitempty"`
}

type JSONContentStats struct {
	WordCount      int      `json:"word_count"`
	CharacterCount int      `json:"character_count"`
	HeadingCount   int      `json:"heading_count"`
	Headings       []string `json:"headings"`
	QualityScore   float64  `json:"quality_score"`
}

type JSONAnalysisResult struct {
	ConfidenceScore int     `json:"confidence_score"`
	SuspicionRate   float64 `json:"suspicion_rate"`
	PatternCount    int     `json:"pattern_count"`
	Assessment      string  `json:"assessment"`
}

type JSONFlaggedContent struct {
	PatternType string   `json:"pattern_type"`
	Severity    float64  `json:"severity"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
	Context     string   `json:"context,omitempty"`
}

func (r *JSONWebReporter) Generate(data *WebReportData) (string, error) {
	quality := data.Content.GetContentQuality()
	assessment := getAssessment(data.Analysis.GetConfidenceScore())

	report := JSONWebReport{
		URL:        data.Content.URL,
		Title:      data.Content.Title,
		StatusCode: data.Content.StatusCode,
		AnalyzedAt: data.AnalyzedAt.Format(time.RFC3339),
		ContentStats: JSONContentStats{
			WordCount:      data.Content.WordCount,
			CharacterCount: len(data.Content.GetMainContent()),
			HeadingCount:   len(data.Content.Headings),
			Headings:       data.Content.Headings,
			QualityScore:   quality,
		},
		Analysis: JSONAnalysisResult{
			ConfidenceScore: data.Analysis.GetConfidenceScore(),
			SuspicionRate:   data.Analysis.SuspicionRate,
			PatternCount:    len(data.Analysis.Patterns),
			Assessment:      assessment,
		},
		FlaggedItems: make([]JSONFlaggedContent, 0, len(data.Analysis.Patterns)),
		AIAnalysis:   data.AIAnalysis,
	}

	mainContent := data.Content.GetMainContent()
	for _, pattern := range data.Analysis.Patterns {
		flagged := JSONFlaggedContent{
			PatternType: pattern.Type,
			Severity:    pattern.Severity,
			Description: pattern.Description,
			Examples:    pattern.Examples,
		}

		if len(pattern.Examples) > 0 {
			contexts := make([]string, 0, len(pattern.Examples))
			for _, example := range pattern.Examples {
				if ctx := extractContext(mainContent, example, 100); ctx != "" {
					contexts = append(contexts, ctx)
				}
			}
			if len(contexts) > 0 {
				flagged.Context = strings.Join(contexts, " ... ")
			}
		}

		report.FlaggedItems = append(report.FlaggedItems, flagged)
	}

	bytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

type TextWebReporter struct{}

func (r *TextWebReporter) Generate(data *WebReportData) (string, error) {
	quality := data.Content.GetContentQuality()
	qualityLabel := "Good"
	if quality < 0.5 {
		qualityLabel = "Low"
	} else if quality < 0.7 {
		qualityLabel = "Moderate"
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║           WEBSITE AI CONTENT ANALYSIS REPORT              ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("URL: %s\n", data.Content.URL))
	sb.WriteString(fmt.Sprintf("Title: %s\n", data.Content.Title))
	sb.WriteString(fmt.Sprintf("Status Code: %d\n", data.Content.StatusCode))
	sb.WriteString(fmt.Sprintf("Analyzed At: %s\n", data.AnalyzedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("\n")

	sb.WriteString("CONTENT METRICS\n")
	sb.WriteString("────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Word Count: %d words\n", data.Content.WordCount))
	sb.WriteString(fmt.Sprintf("Character Count: %d\n", len(data.Content.GetMainContent())))
	sb.WriteString(fmt.Sprintf("Headings Found: %d\n", len(data.Content.Headings)))
	sb.WriteString(fmt.Sprintf("Content Quality: %s (%.2f)\n", qualityLabel, quality))
	sb.WriteString("\n")

	sb.WriteString("ANALYSIS RESULTS\n")
	sb.WriteString("────────────────────────────────────────────────────────────\n")
	score := data.Analysis.GetConfidenceScore()
	sb.WriteString(fmt.Sprintf("AI-Generated Content Confidence: %d%%\n", score))
	sb.WriteString(fmt.Sprintf("Pattern Count: %d detected\n", len(data.Analysis.Patterns)))
	sb.WriteString(fmt.Sprintf("Assessment: %s\n", getAssessment(score)))
	sb.WriteString("\n")

	if len(data.Analysis.Patterns) > 0 {
		sb.WriteString("FLAGGED CONTENT WITH REASONING\n")
		sb.WriteString("────────────────────────────────────────────────────────────\n")

		mainContent := data.Content.GetMainContent()
		for i, pattern := range data.Analysis.Patterns {
			sb.WriteString(fmt.Sprintf("\n%d. %s\n", i+1, pattern.Type))
			sb.WriteString(fmt.Sprintf("   Severity: %.0f%%\n", pattern.Severity*100))
			sb.WriteString(fmt.Sprintf("   Reason: %s\n", pattern.Description))

			if len(pattern.Examples) > 0 {
				sb.WriteString(fmt.Sprintf("   Found %d instances:\n", len(pattern.Examples)))
				for j, example := range pattern.Examples {
					if j >= 3 {
						sb.WriteString(fmt.Sprintf("   ... and %d more\n", len(pattern.Examples)-3))
						break
					}
					sb.WriteString(fmt.Sprintf("   - %q\n", truncateExample(example, 80)))

					if context := extractContext(mainContent, example, 150); context != "" {
						sb.WriteString(fmt.Sprintf("     Context: ...%s...\n", context))
					}
				}
			}
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("No AI-generation patterns detected in this content.\n\n")
	}

	if data.AIAnalysis != "" {
		sb.WriteString("AI EXPERT ANALYSIS\n")
		sb.WriteString("────────────────────────────────────────────────────────────\n")
		sb.WriteString(data.AIAnalysis)
		sb.WriteString("\n\n")
	}

	if quality < 0.5 {
		sb.WriteString("⚠ NOTE: Content quality is low. Analysis may be less reliable.\n")
		sb.WriteString("   Consider analyzing a page with more substantive text content.\n\n")
	}

	sb.WriteString("════════════════════════════════════════════════════════════\n")

	return sb.String(), nil
}

func getAssessment(score int) string {
	if score >= 70 {
		return "LIKELY AI-GENERATED"
	} else if score >= 50 {
		return "POSSIBLY AI-GENERATED"
	} else if score >= 30 {
		return "SUSPICIOUS"
	}
	return "LIKELY HUMAN-WRITTEN"
}

func extractContext(content, example string, contextChars int) string {
	lowerContent := strings.ToLower(content)
	lowerExample := strings.ToLower(example)

	index := strings.Index(lowerContent, lowerExample)
	if index == -1 {
		return ""
	}

	start := index - contextChars/2
	if start < 0 {
		start = 0
	}

	end := index + len(example) + contextChars/2
	if end > len(content) {
		end = len(content)
	}

	context := content[start:end]
	context = strings.TrimSpace(context)

	if start > 0 {
		firstSpace := strings.Index(context, " ")
		if firstSpace > 0 && firstSpace < 20 {
			context = context[firstSpace+1:]
		}
	}

	if end < len(content) {
		lastSpace := strings.LastIndex(context, " ")
		if lastSpace > len(context)-20 && lastSpace > 0 {
			context = context[:lastSpace]
		}
	}

	return context
}

func truncateExample(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
