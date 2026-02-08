package skills

import (
	"encoding/json"
	"fmt"
	"strings"
)

func init() {
	Register(&ReportSummary{})
}

// ReportSummaryInput holds the input parameters for the report_summary skill.
type ReportSummaryInput struct {
	SourceType      string             // "git" or "web"
	SourceID        string             // Repository URL or website URL
	OverallScore    float64            // Overall AI-generation score
	Assessment      string             // Overall assessment string
	DetectionCount  int                // Number of detections that fired
	TotalDetections int                // Total number of strategies that ran
	HighSeverity    int                // Count of high-severity detections
	MediumSeverity  int                // Count of medium-severity detections
	LowSeverity     int                // Count of low-severity detections
	TopDetections   []DetectionSummary // Top N most significant detections
}

// DetectionSummary is a compact representation of a single detection for the summary.
type DetectionSummary struct {
	Strategy    string
	Severity    string
	Score       float64
	Description string
}

// ReportSummaryResult holds the structured output from the report_summary skill.
type ReportSummaryResult struct {
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	KeyFindings []string `json:"key_findings"`
	RiskLevel   string   `json:"risk_level"`
	NextSteps   []string `json:"next_steps"`
}

// ReportSummary generates a human-readable narrative summary of an analysis report.
// It interprets the detection results in natural language instead of raw numbers.
type ReportSummary struct{}

func (s *ReportSummary) Name() string { return "report_summary" }
func (s *ReportSummary) Description() string {
	return "Generate a human-readable summary of analysis results"
}
func (s *ReportSummary) Category() string { return "reporting" }
func (s *ReportSummary) MaxTokens() int   { return 1024 }

const reportSummarySystemPrompt = `You are an expert at interpreting AI-generated code detection results and writing clear, actionable summaries for developers and project maintainers.

Given analysis statistics and top detections, produce a concise summary that:
1. Explains the overall risk level in plain language
2. Highlights the most significant findings
3. Provides actionable next steps

Be direct, factual, and avoid unnecessary hedging. Use bullet points for clarity.

Respond in JSON format:
{
  "title": "One-line summary headline",
  "summary": "2-3 sentence narrative summary of the analysis",
  "key_findings": ["finding 1", "finding 2", "finding 3"],
  "risk_level": "none|low|medium|high|critical",
  "next_steps": ["actionable step 1", "actionable step 2"]
}`

func (s *ReportSummary) SystemPrompt() string {
	return reportSummarySystemPrompt
}

func (s *ReportSummary) FormatInput(input interface{}) (string, error) {
	v, ok := input.(ReportSummaryInput)
	if !ok {
		return "", fmt.Errorf("report_summary: expected ReportSummaryInput, got %T", input)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Source: %s (%s)\n", v.SourceID, v.SourceType))
	b.WriteString(fmt.Sprintf("Overall Score: %.1f%%\n", v.OverallScore*100))
	if v.Assessment != "" {
		b.WriteString(fmt.Sprintf("Assessment: %s\n", v.Assessment))
	}
	b.WriteString(fmt.Sprintf("Detections: %d/%d strategies fired\n", v.DetectionCount, v.TotalDetections))
	b.WriteString(fmt.Sprintf("Severity breakdown: %d high, %d medium, %d low\n", v.HighSeverity, v.MediumSeverity, v.LowSeverity))

	if len(v.TopDetections) > 0 {
		b.WriteString("\nTop detections:\n")
		for _, d := range v.TopDetections {
			b.WriteString(fmt.Sprintf("  - [%s] %s (score: %.2f): %s\n", d.Severity, d.Strategy, d.Score, d.Description))
		}
	}

	b.WriteString("\nGenerate a summary in the JSON format specified.")
	return b.String(), nil
}

func (s *ReportSummary) ParseOutput(raw string) (interface{}, error) {
	jsonStart := strings.Index(raw, "{")
	jsonEnd := strings.LastIndex(raw, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return &ReportSummaryResult{
			Summary: raw,
		}, nil
	}

	var result ReportSummaryResult
	if err := json.Unmarshal([]byte(raw[jsonStart:jsonEnd+1]), &result); err != nil {
		return &ReportSummaryResult{
			Summary: raw,
		}, nil
	}
	return &result, nil
}
