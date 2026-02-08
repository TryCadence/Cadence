package formats

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
)

func TestJSONReporter_FormatAnalysis(t *testing.T) {
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
				ID:         "test-001",
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
						{Name: "detect", StartedAt: startTime.Add(3010 * time.Millisecond), Duration: 2 * time.Second},
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
			name: "formats analysis report with high severity detection",
			report: &analysis.AnalysisReport{
				ID:         "test-002",
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
			reporter := &JSONReporter{}
			output, err := reporter.FormatAnalysis(tt.report)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatAnalysis() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			// Verify output is valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Output is not valid JSON: %v", err)
			}

			// Check basic fields exist
			if _, ok := result["id"]; !ok {
				t.Error("Missing 'id' field in JSON output")
			}
			if _, ok := result["sourceType"]; !ok {
				t.Error("Missing 'sourceType' field in JSON output")
			}
			if _, ok := result["detections"]; !ok {
				t.Error("Missing 'detections' field in JSON output")
			}

			// Check timing fields exist
			timing, ok := result["timing"].(map[string]interface{})
			if !ok {
				t.Error("Missing or invalid 'timing' object in JSON output")
			} else {
				if _, ok := timing["startedAt"]; !ok {
					t.Error("Missing 'timing.startedAt' field")
				}
				if _, ok := timing["completedAt"]; !ok {
					t.Error("Missing 'timing.completedAt' field")
				}
				if _, ok := timing["durationMs"]; !ok {
					t.Error("Missing 'timing.durationMs' field")
				}
				if _, ok := timing["durationSeconds"]; !ok {
					t.Error("Missing 'timing.durationSeconds' field")
				}
			}

			// Check source metrics fields exist
			sm, ok := result["sourceMetrics"].(map[string]interface{})
			if !ok {
				t.Error("Missing or invalid 'sourceMetrics' object in JSON output")
			} else {
				if _, ok := sm["itemsAnalyzed"]; !ok {
					t.Error("Missing 'sourceMetrics.itemsAnalyzed' field")
				}
				if _, ok := sm["strategiesUsed"]; !ok {
					t.Error("Missing 'sourceMetrics.strategiesUsed' field")
				}
			}
		})
	}
}
