package formats

import (
	"fmt"
	"strings"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
)

type TextReporter struct{}

func formatDuration(d time.Duration) string {
	minutes := d.Minutes()
	return fmt.Sprintf("%.0f minutes", minutes)
}

func formatDurationPrecise(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	mins := int(d.Minutes())
	secs := d.Seconds() - float64(mins*60)
	return fmt.Sprintf("%dm %.1fs", mins, secs)
}

func truncate(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (r *TextReporter) FormatAnalysis(report *analysis.AnalysisReport) (string, error) {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("CADENCE ANALYSIS REPORT - %s\n", report.SourceType))
	sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("Source ID:      %s\n", report.SourceID))
	sb.WriteString(fmt.Sprintf("Analysis ID:    %s\n", report.ID))
	sb.WriteString(fmt.Sprintf("Started At:     %s\n", report.Timing.StartedAt.Format("2006-01-02 15:04:05.000 MST")))
	sb.WriteString(fmt.Sprintf("Completed At:   %s\n", report.Timing.CompletedAt.Format("2006-01-02 15:04:05.000 MST")))
	sb.WriteString(fmt.Sprintf("Duration:       %s\n\n", formatDurationPrecise(report.Timing.Duration)))

	// Phase timing breakdown
	if len(report.Timing.Phases) > 0 {
		sb.WriteString("Phase Breakdown:\n")
		for _, p := range report.Timing.Phases {
			sb.WriteString(fmt.Sprintf("  ├─ %-12s %s\n", p.Name+":", formatDurationPrecise(p.Duration)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString("ASSESSMENT\n")
	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Overall Score:  %.1f%%\n", report.OverallScore))
	sb.WriteString(fmt.Sprintf("Assessment:     %s\n", report.Assessment))
	sb.WriteString(fmt.Sprintf("Suspicion Rate: %.1f%%\n\n", report.SuspicionRate*100))

	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString("STATISTICS\n")
	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Total Detections:     %d\n", report.TotalDetections))
	sb.WriteString(fmt.Sprintf("Detected:             %d\n", report.DetectionCount))
	sb.WriteString(fmt.Sprintf("Passed:               %d\n", report.PassedDetections))
	sb.WriteString(fmt.Sprintf("  ├─ High Severity:   %d\n", report.HighSeverityCount))
	sb.WriteString(fmt.Sprintf("  ├─ Medium Severity: %d\n", report.MediumSeverityCount))
	sb.WriteString(fmt.Sprintf("  └─ Low Severity:    %d\n\n", report.LowSeverityCount))

	// Cross-source metrics
	sm := report.SourceMetrics
	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString("SOURCE METRICS\n")
	sb.WriteString("─────────────────────────────────────────────────────────────\n")
	sb.WriteString(fmt.Sprintf("Items Analyzed:       %d\n", sm.ItemsAnalyzed))
	sb.WriteString(fmt.Sprintf("Items Flagged:        %d\n", sm.ItemsFlagged))
	if sm.UniqueAuthors > 0 {
		sb.WriteString(fmt.Sprintf("Unique Authors:       %d\n", sm.UniqueAuthors))
	}
	sb.WriteString(fmt.Sprintf("Average Score:        %.2f\n", sm.AverageScore))
	sb.WriteString(fmt.Sprintf("Coverage Rate:        %.1f%%\n", sm.CoverageRate*100))
	sb.WriteString(fmt.Sprintf("Strategies Used:      %d\n", sm.StrategiesUsed))
	sb.WriteString(fmt.Sprintf("Strategies Triggered: %d\n", sm.StrategiesHit))
	if len(sm.Extra) > 0 {
		for key, value := range sm.Extra {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}
	sb.WriteString("\n")

	highSev := report.GetDetectionsBySeverity("high")
	mediumSev := report.GetDetectionsBySeverity("medium")
	lowSev := report.GetDetectionsBySeverity("low")

	if len(highSev) > 0 {
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString("HIGH SEVERITY DETECTIONS\n")
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		for _, d := range highSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf("• %s [%s] (%.0f%% score, %.0f%% weight)\n", d.Strategy, d.Category, d.Score*100, d.Confidence*100))
				sb.WriteString(fmt.Sprintf("  %s\n", d.Description))
				if len(d.Examples) > 0 {
					examplesStr := strings.Join(d.Examples[:min(len(d.Examples), 2)], ", ")
					sb.WriteString(fmt.Sprintf("  Examples: %s\n", examplesStr))
				}
				sb.WriteString("\n")
			}
		}
	}

	if len(mediumSev) > 0 {
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString("MEDIUM SEVERITY DETECTIONS\n")
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		for _, d := range mediumSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf("• %s [%s] (%.0f%% score, %.0f%% weight)\n", d.Strategy, d.Category, d.Score*100, d.Confidence*100))
				sb.WriteString(fmt.Sprintf("  %s\n", d.Description))
				if len(d.Examples) > 0 {
					examplesStr := strings.Join(d.Examples[:min(len(d.Examples), 2)], ", ")
					sb.WriteString(fmt.Sprintf("  Examples: %s\n", examplesStr))
				}
				sb.WriteString("\n")
			}
		}
	}

	if len(lowSev) > 0 {
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString("LOW SEVERITY DETECTIONS\n")
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		for _, d := range lowSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf("• %s [%s] (%.0f%% score, %.0f%% weight)\n", d.Strategy, d.Category, d.Score*100, d.Confidence*100))
				sb.WriteString(fmt.Sprintf("  %s\n\n", d.Description))
			}
		}
	}

	if len(report.Metrics) > 0 {
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString("ADDITIONAL METRICS\n")
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		for key, value := range report.Metrics {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	if report.Error != "" {
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString("ERROR\n")
		sb.WriteString("─────────────────────────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", report.Error))
	}

	sb.WriteString("═══════════════════════════════════════════════════════════\n")

	return sb.String(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
