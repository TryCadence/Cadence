package formats

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBSONReporter_FormatAnalysis(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-5 * time.Second)
	completedTime := now

	tests := []struct {
		name    string
		report  *analysis.AnalysisReport
		wantErr bool
	}{
		{
			name: "formats analysis report with no detections",
			report: &analysis.AnalysisReport{
				ID:         "test-bson-001",
				SourceType: analysis.SourceTypeGit,
				SourceID:   "/repo/path",
				AnalyzedAt: startTime,
				Duration:   5 * time.Second,
				Timing: analysis.TimingInfo{
					StartedAt:   startTime,
					CompletedAt: completedTime,
					Duration:    5 * time.Second,
					Phases: []analysis.PhaseTiming{
						{Name: "validate", StartedAt: startTime, Duration: 10 * time.Millisecond},
						{Name: "fetch", StartedAt: startTime.Add(10 * time.Millisecond), Duration: 3 * time.Second},
					},
				},
				SourceMetrics: analysis.SourceMetrics{
					ItemsAnalyzed:  50,
					ItemsFlagged:   0,
					StrategiesUsed: 18,
					StrategiesHit:  0,
				},
				OverallScore:    0,
				Assessment:      "No suspicious activity detected",
				SuspicionRate:   0,
				TotalDetections: 0,
				DetectionCount:  0,
				Detections:      []analysis.Detection{},
				Metrics:         map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "formats analysis report with detections",
			report: &analysis.AnalysisReport{
				ID:         "test-bson-002",
				SourceType: analysis.SourceTypeGit,
				SourceID:   "/repo/path",
				AnalyzedAt: startTime,
				Duration:   5 * time.Second,
				Timing: analysis.TimingInfo{
					StartedAt:   startTime,
					CompletedAt: completedTime,
					Duration:    5 * time.Second,
				},
				SourceMetrics: analysis.SourceMetrics{
					ItemsAnalyzed:  100,
					ItemsFlagged:   1,
					UniqueAuthors:  3,
					AverageScore:   0.85,
					CoverageRate:   0.01,
					StrategiesUsed: 18,
					StrategiesHit:  1,
				},
				Detections: []analysis.Detection{
					{
						Strategy:    "velocity-analysis",
						Detected:    true,
						Severity:    "high",
						Score:       0.85,
						Confidence:  0.9,
						Category:    "velocity",
						Description: "Unusually high code velocity",
						Examples:    []string{"500 LOC added in 1 minute"},
					},
				},
				OverallScore:      85,
				Assessment:        "Suspicious activity detected",
				SuspicionRate:     1.0,
				TotalDetections:   1,
				DetectionCount:    1,
				PassedDetections:  0,
				HighSeverityCount: 1,
				Metrics:           map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &BSONReporter{}
			output, err := reporter.FormatAnalysis(tt.report)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatAnalysis() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if output == "" {
				t.Fatal("FormatAnalysis() returned empty output")
			}

			// Output should be valid base64
			rawBytes, err := base64.StdEncoding.DecodeString(output)
			if err != nil {
				t.Fatalf("Output is not valid base64: %v", err)
			}

			if len(rawBytes) == 0 {
				t.Fatal("Decoded BSON bytes are empty")
			}

			// Verify bytes are valid BSON by unmarshaling
			var result map[string]interface{}
			if err := bson.Unmarshal(rawBytes, &result); err != nil {
				t.Fatalf("Output is not valid BSON: %v", err)
			}

			// Check basic fields exist
			if _, ok := result["id"]; !ok {
				t.Error("Missing 'id' field in BSON output")
			}
			if _, ok := result["sourceType"]; !ok {
				t.Error("Missing 'sourceType' field in BSON output")
			}
			if _, ok := result["detections"]; !ok {
				t.Error("Missing 'detections' field in BSON output")
			}

			// Check timing field exists (BSON unmarshals embedded docs as bson.D)
			if _, ok := result["timing"]; !ok {
				t.Error("Missing 'timing' field in BSON output")
			}

			// Check source metrics field exists
			if _, ok := result["sourceMetrics"]; !ok {
				t.Error("Missing 'sourceMetrics' field in BSON output")
			}
		})
	}
}

func TestBSONReporter_FormatAnalysisRaw(t *testing.T) {
	reporter := &BSONReporter{}
	report := &analysis.AnalysisReport{
		ID:         "test-raw-001",
		SourceType: analysis.SourceTypeGit,
		SourceID:   "/repo/path",
		AnalyzedAt: time.Now(),
		Duration:   1 * time.Second,
		Timing: analysis.TimingInfo{
			StartedAt:   time.Now().Add(-time.Second),
			CompletedAt: time.Now(),
			Duration:    1 * time.Second,
		},
		Detections: []analysis.Detection{},
		Metrics:    map[string]interface{}{},
	}

	raw, err := reporter.FormatAnalysisRaw(report)
	if err != nil {
		t.Fatalf("FormatAnalysisRaw() error = %v", err)
	}

	if len(raw) == 0 {
		t.Fatal("FormatAnalysisRaw() returned empty bytes")
	}

	// Verify raw bytes round-trip
	var result map[string]interface{}
	if err := bson.Unmarshal(raw, &result); err != nil {
		t.Fatalf("Raw bytes are not valid BSON: %v", err)
	}

	if result["id"] != "test-raw-001" {
		t.Errorf("Expected id 'test-raw-001', got %v", result["id"])
	}
}
