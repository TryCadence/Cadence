package formats

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
)

type HTMLReporter struct{}

func formatHTMLDuration(d time.Duration) string {
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

func (r *HTMLReporter) FormatAnalysis(report *analysis.AnalysisReport) (string, error) {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cadence Analysis Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: #333; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 12px; box-shadow: 0 20px 60px rgba(0,0,0,0.3); overflow: hidden; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 40px; }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header .meta { font-size: 0.95em; opacity: 0.9; margin-top: 10px; }
        .content { padding: 40px; }
        .section { margin-bottom: 40px; border-bottom: 1px solid #eee; padding-bottom: 30px; }
        .section:last-child { border-bottom: none; }
        .section h2 { font-size: 1.6em; color: #667eea; margin-bottom: 20px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin: 20px 0; }
        .stat-card { background: #f8f9fa; border-left: 4px solid #667eea; padding: 20px; border-radius: 6px; }
        .stat-card .label { font-size: 0.85em; color: #666; text-transform: uppercase; letter-spacing: 0.5px; }
        .stat-card .value { font-size: 1.8em; font-weight: bold; color: #667eea; margin-top: 8px; }
        .score-high { color: #ef4444; }
        .score-medium { color: #f59e0b; }
        .score-low { color: #10b981; }
        .assessment { background: #f8f9fa; padding: 20px; border-radius: 6px; border-left: 4px solid #667eea; margin: 20px 0; }
        .assessment .label { font-size: 0.9em; color: #666; text-transform: uppercase; }
        .assessment .text { font-size: 1.1em; margin-top: 10px; color: #333; }
        .detection-list { list-style: none; }
        .detection-item { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 6px; border-left: 4px solid #667eea; }
        .detection-item.high { border-left-color: #ef4444; }
        .detection-item.medium { border-left-color: #f59e0b; }
        .detection-item.low { border-left-color: #10b981; }
        .detection-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
        .detection-strategy { font-weight: 600; color: #333; }
        .detection-score { font-weight: bold; }
        .detection-description { color: #666; font-size: 0.95em; }
        .detection-examples { color: #999; font-size: 0.85em; margin-top: 8px; font-style: italic; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        thead { background: #f8f9fa; }
        th { padding: 15px; text-align: left; font-weight: 600; color: #333; border-bottom: 2px solid #ddd; }
        td { padding: 15px; border-bottom: 1px solid #eee; }
        tr:hover { background: #f8f9fa; }
        .footer { background: #f8f9fa; padding: 20px; text-align: center; color: #666; font-size: 0.9em; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 0.8em; font-weight: 600; }
        .badge-high { background: #fee2e2; color: #991b1b; }
        .badge-medium { background: #fef3c7; color: #b45309; }
        .badge-low { background: #dcfce7; color: #166534; }
        .error-box { background: #fee2e2; border-left: 4px solid #ef4444; padding: 15px; border-radius: 6px; margin: 20px 0; }
        .error-box p { color: #991b1b; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Cadence Analysis Report</h1>
            <div class="meta">`)

	if report.SourceType != "" {
		sb.WriteString(fmt.Sprintf("Source: %s | ", html.EscapeString(string(report.SourceType))))
	}
	if report.SourceID != "" {
		sb.WriteString(fmt.Sprintf("ID: %s | ", html.EscapeString(report.SourceID)))
	}
	sb.WriteString(fmt.Sprintf("Started: %s | Completed: %s",
		report.Timing.StartedAt.Format("2006-01-02 15:04:05 MST"),
		report.Timing.CompletedAt.Format("2006-01-02 15:04:05 MST")))

	sb.WriteString(`</div>
        </div>
        <div class="content">
`)

	// Assessment Section
	sb.WriteString(`            <section class="section">
                <h2>Assessment</h2>
`)

	scoreClass := "score-low"
	if report.OverallScore > 0.7 {
		scoreClass = "score-high"
	} else if report.OverallScore > 0.4 {
		scoreClass = "score-medium"
	}

	sb.WriteString(`                <div class="grid">
`)
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Overall Score</div>
                        <div class="value %s">%.1f%%</div>
                    </div>
`, scoreClass, report.OverallScore))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Suspicion Rate</div>
                        <div class="value">%.1f%%</div>
                    </div>
`, report.SuspicionRate*100))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Duration</div>
                        <div class="value">%s</div>
                    </div>
`, html.EscapeString(formatHTMLDuration(report.Timing.Duration))))
	sb.WriteString(`                </div>
`)
	sb.WriteString(fmt.Sprintf(`                <div class="assessment">
                    <div class="label">Assessment Result</div>
                    <div class="text">%s</div>
                </div>
`, html.EscapeString(report.Assessment)))
	sb.WriteString(`            </section>
`)

	// Timing Section
	sb.WriteString(`            <section class="section">
                <h2>Timing</h2>
                <div class="grid">
`)
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Started At</div>
                        <div class="value" style="font-size:1em">%s</div>
                    </div>
`, html.EscapeString(report.Timing.StartedAt.Format("15:04:05.000"))))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Completed At</div>
                        <div class="value" style="font-size:1em">%s</div>
                    </div>
`, html.EscapeString(report.Timing.CompletedAt.Format("15:04:05.000"))))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Total Duration</div>
                        <div class="value" style="font-size:1.2em">%s</div>
                    </div>
`, html.EscapeString(formatHTMLDuration(report.Timing.Duration))))
	sb.WriteString(`                </div>
`)
	if len(report.Timing.Phases) > 0 {
		sb.WriteString(`                <table>
                    <thead><tr><th>Phase</th><th>Duration</th></tr></thead>
                    <tbody>
`)
		for _, p := range report.Timing.Phases {
			sb.WriteString(fmt.Sprintf(`                        <tr><td>%s</td><td>%s</td></tr>
`, html.EscapeString(p.Name), html.EscapeString(formatHTMLDuration(p.Duration))))
		}
		sb.WriteString(`                    </tbody>
                </table>
`)
	}
	sb.WriteString(`            </section>
`)

	// Source Metrics Section
	sm := report.SourceMetrics
	sb.WriteString(`            <section class="section">
                <h2>Source Metrics</h2>
                <div class="grid">
`)
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Items Analyzed</div>
                        <div class="value">%d</div>
                    </div>
`, sm.ItemsAnalyzed))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Items Flagged</div>
                        <div class="value">%d</div>
                    </div>
`, sm.ItemsFlagged))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Strategies Used</div>
                        <div class="value">%d</div>
                    </div>
`, sm.StrategiesUsed))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Strategies Triggered</div>
                        <div class="value">%d</div>
                    </div>
`, sm.StrategiesHit))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Avg Score</div>
                        <div class="value">%.2f</div>
                    </div>
`, sm.AverageScore))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Coverage Rate</div>
                        <div class="value">%.1f%%</div>
                    </div>
`, sm.CoverageRate*100))
	if sm.UniqueAuthors > 0 {
		sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Unique Authors</div>
                        <div class="value">%d</div>
                    </div>
`, sm.UniqueAuthors))
	}
	sb.WriteString(`                </div>
`)
	if len(sm.Extra) > 0 {
		sb.WriteString(`                <table>
                    <thead><tr><th>Metric</th><th>Value</th></tr></thead>
                    <tbody>
`)
		for key, value := range sm.Extra {
			sb.WriteString(fmt.Sprintf(`                        <tr><td><strong>%s</strong></td><td>%v</td></tr>
`, html.EscapeString(key), html.EscapeString(fmt.Sprint(value))))
		}
		sb.WriteString(`                    </tbody>
                </table>
`)
	}
	sb.WriteString(`            </section>
`)

	// Statistics Section
	sb.WriteString(`            <section class="section">
                <h2>Statistics</h2>
                <div class="grid">
`)
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Total Detections</div>
                        <div class="value">%d</div>
                    </div>
`, report.TotalDetections))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Detected Issues</div>
                        <div class="value">%d</div>
                    </div>
`, report.DetectionCount))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Passed</div>
                        <div class="value">%d</div>
                    </div>
`, report.PassedDetections))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">High Severity</div>
                        <div class="value">%d</div>
                    </div>
`, report.HighSeverityCount))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Medium Severity</div>
                        <div class="value">%d</div>
                    </div>
`, report.MediumSeverityCount))
	sb.WriteString(fmt.Sprintf(`                    <div class="stat-card">
                        <div class="label">Low Severity</div>
                        <div class="value">%d</div>
                    </div>
`, report.LowSeverityCount))
	sb.WriteString(`                </div>
            </section>
`)

	// High Severity Detections
	highSev := report.GetDetectionsBySeverity("high")
	if len(highSev) > 0 {
		sb.WriteString(`            <section class="section">
                <h2>High Severity Detections</h2>
                <ul class="detection-list">
`)
		for _, d := range highSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf(`                    <li class="detection-item high">
                        <div class="detection-header">
                            <span class="detection-strategy">%s</span>
                            <span class="detection-score"><span class="badge badge-high">%.0f%% confidence</span></span>
                        </div>
                        <div class="detection-description">%s</div>
`, html.EscapeString(d.Strategy), d.Score*100, html.EscapeString(d.Description)))
				if len(d.Examples) > 0 {
					sb.WriteString(fmt.Sprintf(`                        <div class="detection-examples">Examples: %s</div>
`, html.EscapeString(strings.Join(d.Examples[:min(len(d.Examples), 2)], ", "))))
				}
				sb.WriteString(`                    </li>
`)
			}
		}
		sb.WriteString(`                </ul>
            </section>
`)
	}

	// Medium Severity Detections
	mediumSev := report.GetDetectionsBySeverity("medium")
	if len(mediumSev) > 0 {
		sb.WriteString(`            <section class="section">
                <h2>Medium Severity Detections</h2>
                <ul class="detection-list">
`)
		for _, d := range mediumSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf(`                    <li class="detection-item medium">
                        <div class="detection-header">
                            <span class="detection-strategy">%s</span>
                            <span class="detection-score"><span class="badge badge-medium">%.0f%% confidence</span></span>
                        </div>
                        <div class="detection-description">%s</div>
`, html.EscapeString(d.Strategy), d.Score*100, html.EscapeString(d.Description)))
				if len(d.Examples) > 0 {
					sb.WriteString(fmt.Sprintf(`                        <div class="detection-examples">Examples: %s</div>
`, html.EscapeString(strings.Join(d.Examples[:min(len(d.Examples), 2)], ", "))))
				}
				sb.WriteString(`                    </li>
`)
			}
		}
		sb.WriteString(`                </ul>
            </section>
`)
	}

	// Low Severity Detections
	lowSev := report.GetDetectionsBySeverity("low")
	if len(lowSev) > 0 {
		sb.WriteString(`            <section class="section">
                <h2>Low Severity Detections</h2>
                <ul class="detection-list">
`)
		for _, d := range lowSev {
			if d.Detected {
				sb.WriteString(fmt.Sprintf(`                    <li class="detection-item low">
                        <div class="detection-header">
                            <span class="detection-strategy">%s</span>
                            <span class="detection-score"><span class="badge badge-low">%.0f%% confidence</span></span>
                        </div>
                        <div class="detection-description">%s</div>
`, html.EscapeString(d.Strategy), d.Score*100, html.EscapeString(d.Description)))
				sb.WriteString(`                    </li>
`)
			}
		}
		sb.WriteString(`                </ul>
            </section>
`)
	}

	// Error Section
	if report.Error != "" {
		sb.WriteString(`            <section class="section">
`)
		sb.WriteString(fmt.Sprintf(`                <div class="error-box">
                    <p><strong>Error:</strong> %s</p>
                </div>
`, html.EscapeString(report.Error)))
		sb.WriteString(`            </section>
`)
	}

	// Metrics Section
	if len(report.Metrics) > 0 {
		sb.WriteString(`            <section class="section">
                <h2>Additional Metrics</h2>
                <table>
                    <thead>
                        <tr>
                            <th>Metric</th>
                            <th>Value</th>
                        </tr>
                    </thead>
                    <tbody>
`)
		for key, value := range report.Metrics {
			sb.WriteString(fmt.Sprintf(`                        <tr>
                            <td><strong>%s</strong></td>
                            <td>%v</td>
                        </tr>
`, html.EscapeString(key), html.EscapeString(fmt.Sprint(value))))
		}
		sb.WriteString(`                    </tbody>
                </table>
            </section>
`)
	}

	sb.WriteString(`        </div>
        <div class="footer">
            <p>Generated by Cadence • `)
	sb.WriteString(time.Now().Format("2006-01-02 15:04:05 MST"))
	sb.WriteString(`</p>
        </div>
    </div>
</body>
</html>`)

	return sb.String(), nil
}
