package detectors

import (
	"context"
	"fmt"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git/patterns"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/web"
)

type WebDetector struct {
}

func NewWebDetector() *WebDetector {
	return &WebDetector{}
}

func (w *WebDetector) Detect(ctx context.Context, data *analysis.SourceData) ([]analysis.Detection, error) {
	if data.Type != "web" {
		return nil, fmt.Errorf("WebDetector only supports web sources")
	}

	page, ok := data.RawContent.(*web.PageContent)
	if !ok {
		return nil, fmt.Errorf("invalid RawContent type for web source")
	}

	slopAnalyzer := patterns.NewTextSlopAnalyzer()
	slopResult, err := slopAnalyzer.AnalyzeContent(page.AllText)
	if err != nil {
		data.Metadata["analysis_error"] = err.Error()
		return []analysis.Detection{}, nil
	}

	detections := make([]analysis.Detection, 0)

	for _, pattern := range slopResult.Patterns {
		severity := "low"
		if pattern.Severity >= 0.7 {
			severity = "high"
		} else if pattern.Severity >= 0.4 {
			severity = "medium"
		}

		detection := analysis.Detection{
			Strategy:    pattern.Type,
			Detected:    true,
			Severity:    severity,
			Score:       pattern.Severity,
			Confidence:  pattern.Confidence,
			Category:    "web-pattern",
			Description: pattern.Description,
			Examples:    pattern.Examples,
		}
		detections = append(detections, detection)
	}

	for _, pattern := range slopResult.PassedPatterns {
		detection := analysis.Detection{
			Strategy:    pattern.Type,
			Detected:    false,
			Severity:    "none",
			Score:       0,
			Confidence:  pattern.Confidence,
			Category:    "web-pattern",
			Description: pattern.Description,
			Examples:    pattern.Examples,
		}
		detections = append(detections, detection)
	}

	data.Metadata["slop_suspicion_rate"] = slopResult.SuspicionRate
	data.Metadata["slop_word_count"] = slopResult.WordCount

	return detections, nil
}
