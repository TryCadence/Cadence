package formats

import (
	"encoding/base64"

	"github.com/TryCadence/Cadence/internal/analysis"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BSONReporter struct{}

func (r *BSONReporter) FormatAnalysis(report *analysis.AnalysisReport) (string, error) {
	type bsonDetection struct {
		Strategy    string   `bson:"strategy"`
		Detected    bool     `bson:"detected"`
		Severity    string   `bson:"severity"`
		Score       float64  `bson:"score"`
		Confidence  float64  `bson:"confidence"`
		Category    string   `bson:"category"`
		Description string   `bson:"description"`
		Examples    []string `bson:"examples,omitempty"`
	}

	type bsonPhaseTiming struct {
		Name       string  `bson:"name"`
		StartedAt  string  `bson:"startedAt"`
		DurationMs float64 `bson:"durationMs"`
	}

	type bsonTiming struct {
		StartedAt   string            `bson:"startedAt"`
		CompletedAt string            `bson:"completedAt"`
		DurationMs  float64           `bson:"durationMs"`
		DurationSec float64           `bson:"durationSeconds"`
		Phases      []bsonPhaseTiming `bson:"phases,omitempty"`
	}

	type bsonSourceMetrics struct {
		ItemsAnalyzed  int                    `bson:"itemsAnalyzed"`
		ItemsFlagged   int                    `bson:"itemsFlagged"`
		UniqueAuthors  int                    `bson:"uniqueAuthors,omitempty"`
		AverageScore   float64                `bson:"averageScore"`
		CoverageRate   float64                `bson:"coverageRate"`
		StrategiesUsed int                    `bson:"strategiesUsed"`
		StrategiesHit  int                    `bson:"strategiesHit"`
		Extra          map[string]interface{} `bson:"extra,omitempty"`
	}

	type bsonAnalysisReport struct {
		ID                  string                 `bson:"id"`
		SourceType          string                 `bson:"sourceType"`
		SourceID            string                 `bson:"sourceId"`
		Timing              bsonTiming             `bson:"timing"`
		SourceMetrics       bsonSourceMetrics      `bson:"sourceMetrics"`
		Detections          []bsonDetection        `bson:"detections"`
		OverallScore        float64                `bson:"overallScore"`
		Assessment          string                 `bson:"assessment"`
		SuspicionRate       float64                `bson:"suspicionRate"`
		TotalDetections     int                    `bson:"totalDetections"`
		PassedDetections    int                    `bson:"passedDetections"`
		HighSeverityCount   int                    `bson:"highSeverityCount"`
		MediumSeverityCount int                    `bson:"mediumSeverityCount"`
		LowSeverityCount    int                    `bson:"lowSeverityCount"`
		Metrics             map[string]interface{} `bson:"metrics,omitempty"`
		Error               string                 `bson:"error,omitempty"`
	}

	detections := make([]bsonDetection, len(report.Detections))
	for i, d := range report.Detections {
		detections[i] = bsonDetection{
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

	phases := make([]bsonPhaseTiming, len(report.Timing.Phases))
	for i, p := range report.Timing.Phases {
		phases[i] = bsonPhaseTiming{
			Name:       p.Name,
			StartedAt:  p.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs: float64(p.Duration.Milliseconds()),
		}
	}

	br := bsonAnalysisReport{
		ID:         report.ID,
		SourceType: string(report.SourceType),
		SourceID:   report.SourceID,
		Timing: bsonTiming{
			StartedAt:   report.Timing.StartedAt.Format("2006-01-02T15:04:05.000Z"),
			CompletedAt: report.Timing.CompletedAt.Format("2006-01-02T15:04:05.000Z"),
			DurationMs:  float64(report.Timing.Duration.Milliseconds()),
			DurationSec: report.Duration.Seconds(),
			Phases:      phases,
		},
		SourceMetrics: bsonSourceMetrics{
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

	data, err := bson.Marshal(br)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (r *BSONReporter) FormatAnalysisRaw(report *analysis.AnalysisReport) ([]byte, error) {
	output, err := r.FormatAnalysis(report)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(output)
}
