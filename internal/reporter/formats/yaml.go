package formats

import (
	"github.com/TryCadence/Cadence/internal/analysis"
	"gopkg.in/yaml.v3"
)

type YAMLReporter struct{}

func (r *YAMLReporter) FormatAnalysis(report *analysis.AnalysisReport) (string, error) {
	type yamlDetection struct {
		Strategy    string   `yaml:"strategy"`
		Detected    bool     `yaml:"detected"`
		Severity    string   `yaml:"severity"`
		Score       float64  `yaml:"score"`
		Confidence  float64  `yaml:"confidence"`
		Category    string   `yaml:"category"`
		Description string   `yaml:"description"`
		Examples    []string `yaml:"examples,omitempty"`
	}

	type yamlPhaseTiming struct {
		Name       string  `yaml:"name"`
		StartedAt  string  `yaml:"started_at"`
		DurationMs float64 `yaml:"duration_ms"`
	}

	type yamlTiming struct {
		StartedAt   string            `yaml:"started_at"`
		CompletedAt string            `yaml:"completed_at"`
		DurationMs  float64           `yaml:"duration_ms"`
		DurationSec float64           `yaml:"duration_seconds"`
		Phases      []yamlPhaseTiming `yaml:"phases,omitempty"`
	}

	type yamlSourceMetrics struct {
		ItemsAnalyzed  int                    `yaml:"items_analyzed"`
		ItemsFlagged   int                    `yaml:"items_flagged"`
		UniqueAuthors  int                    `yaml:"unique_authors,omitempty"`
		AverageScore   float64                `yaml:"average_score"`
		CoverageRate   float64                `yaml:"coverage_rate"`
		StrategiesUsed int                    `yaml:"strategies_used"`
		StrategiesHit  int                    `yaml:"strategies_hit"`
		Extra          map[string]interface{} `yaml:"extra,omitempty"`
	}

	type yamlAnalysisReport struct {
		ID                  string                 `yaml:"id"`
		SourceType          string                 `yaml:"source_type"`
		SourceID            string                 `yaml:"source_id"`
		Timing              yamlTiming             `yaml:"timing"`
		SourceMetrics       yamlSourceMetrics      `yaml:"source_metrics"`
		Detections          []yamlDetection        `yaml:"detections"`
		OverallScore        float64                `yaml:"overall_score"`
		Assessment          string                 `yaml:"assessment"`
		SuspicionRate       float64                `yaml:"suspicion_rate"`
		TotalDetections     int                    `yaml:"total_detections"`
		PassedDetections    int                    `yaml:"passed_detections"`
		HighSeverityCount   int                    `yaml:"high_severity_count"`
		MediumSeverityCount int                    `yaml:"medium_severity_count"`
		LowSeverityCount    int                    `yaml:"low_severity_count"`
		Metrics             map[string]interface{} `yaml:"metrics,omitempty"`
		Error               string                 `yaml:"error,omitempty"`
	}

	detections := make([]yamlDetection, len(report.Detections))
	for i, d := range report.Detections {
		detections[i] = yamlDetection{
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

	phases := make([]yamlPhaseTiming, len(report.Timing.Phases))
	for i, p := range report.Timing.Phases {
		phases[i] = yamlPhaseTiming{
			Name:       p.Name,
			StartedAt:  p.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs: float64(p.Duration.Milliseconds()),
		}
	}

	yr := yamlAnalysisReport{
		ID:         report.ID,
		SourceType: string(report.SourceType),
		SourceID:   report.SourceID,
		Timing: yamlTiming{
			StartedAt:   report.Timing.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			CompletedAt: report.Timing.CompletedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs:  float64(report.Timing.Duration.Milliseconds()),
			DurationSec: report.Duration.Seconds(),
			Phases:      phases,
		},
		SourceMetrics: yamlSourceMetrics{
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

	data, err := yaml.Marshal(yr)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
