package formats

import (
	"encoding/json"

	"github.com/TryCadence/Cadence/internal/analysis"
)

type JSONReporter struct{}

func (r *JSONReporter) FormatAnalysis(report *analysis.AnalysisReport) (string, error) {
	type jsonDetection struct {
		Strategy    string   `json:"strategy"`
		Detected    bool     `json:"detected"`
		Severity    string   `json:"severity"`
		Score       float64  `json:"score"`
		Confidence  float64  `json:"confidence"`
		Category    string   `json:"category"`
		Description string   `json:"description"`
		Examples    []string `json:"examples,omitempty"`
	}

	type jsonPhaseTiming struct {
		Name       string  `json:"name"`
		StartedAt  string  `json:"startedAt"`
		DurationMs float64 `json:"durationMs"`
	}

	type jsonTiming struct {
		StartedAt   string            `json:"startedAt"`
		CompletedAt string            `json:"completedAt"`
		DurationMs  float64           `json:"durationMs"`
		DurationSec float64           `json:"durationSeconds"`
		Phases      []jsonPhaseTiming `json:"phases,omitempty"`
	}

	type jsonSourceMetrics struct {
		ItemsAnalyzed  int                    `json:"itemsAnalyzed"`
		ItemsFlagged   int                    `json:"itemsFlagged"`
		UniqueAuthors  int                    `json:"uniqueAuthors,omitempty"`
		AverageScore   float64                `json:"averageScore"`
		CoverageRate   float64                `json:"coverageRate"`
		StrategiesUsed int                    `json:"strategiesUsed"`
		StrategiesHit  int                    `json:"strategiesHit"`
		Extra          map[string]interface{} `json:"extra,omitempty"`
	}

	type jsonAnalysisReport struct {
		ID                  string                 `json:"id"`
		SourceType          string                 `json:"sourceType"`
		SourceID            string                 `json:"sourceId"`
		Timing              jsonTiming             `json:"timing"`
		SourceMetrics       jsonSourceMetrics      `json:"sourceMetrics"`
		Detections          []jsonDetection        `json:"detections"`
		OverallScore        float64                `json:"overallScore"`
		Assessment          string                 `json:"assessment"`
		SuspicionRate       float64                `json:"suspicionRate"`
		TotalDetections     int                    `json:"totalDetections"`
		PassedDetections    int                    `json:"passedDetections"`
		HighSeverityCount   int                    `json:"highSeverityCount"`
		MediumSeverityCount int                    `json:"mediumSeverityCount"`
		LowSeverityCount    int                    `json:"lowSeverityCount"`
		Metrics             map[string]interface{} `json:"metrics,omitempty"`
		Error               string                 `json:"error,omitempty"`
	}

	detections := make([]jsonDetection, len(report.Detections))
	for i, d := range report.Detections {
		detections[i] = jsonDetection{
			Strategy:    d.Strategy,
			Detected:    d.Detected,
			Severity:    d.Severity,
			Score:       d.Score,
			Confidence:  d.Confidence,
			Category:    d.Category,
			Description: d.Description,
			Examples:    d.Examples,
		}
	}

	// Build phase timings
	phases := make([]jsonPhaseTiming, len(report.Timing.Phases))
	for i, p := range report.Timing.Phases {
		phases[i] = jsonPhaseTiming{
			Name:       p.Name,
			StartedAt:  p.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs: float64(p.Duration.Milliseconds()),
		}
	}

	jr := jsonAnalysisReport{
		ID:         report.ID,
		SourceType: string(report.SourceType),
		SourceID:   report.SourceID,
		Timing: jsonTiming{
			StartedAt:   report.Timing.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			CompletedAt: report.Timing.CompletedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs:  float64(report.Timing.Duration.Milliseconds()),
			DurationSec: report.Duration.Seconds(),
			Phases:      phases,
		},
		SourceMetrics: jsonSourceMetrics{
			ItemsAnalyzed:  report.SourceMetrics.ItemsAnalyzed,
			ItemsFlagged:   report.SourceMetrics.ItemsFlagged,
			UniqueAuthors:  report.SourceMetrics.UniqueAuthors,
			AverageScore:   report.SourceMetrics.AverageScore,
			CoverageRate:   report.SourceMetrics.CoverageRate,
			StrategiesUsed: report.SourceMetrics.StrategiesUsed,
			StrategiesHit:  report.SourceMetrics.StrategiesHit,
			Extra:          report.SourceMetrics.Extra,
		},
		Detections:          detections,
		OverallScore:        report.OverallScore,
		Assessment:          report.Assessment,
		SuspicionRate:       report.SuspicionRate,
		TotalDetections:     report.TotalDetections,
		PassedDetections:    report.PassedDetections,
		HighSeverityCount:   report.HighSeverityCount,
		MediumSeverityCount: report.MediumSeverityCount,
		LowSeverityCount:    report.LowSeverityCount,
		Metrics:             report.Metrics,
		Error:               report.Error,
	}

	data, err := json.MarshalIndent(jr, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
