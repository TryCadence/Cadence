package formats

import (
	"testing"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"gopkg.in/yaml.v3"
)

func TestYAMLReporter_FormatAnalysis(t *testing.T) {
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
				ID:         "test-yaml-001",
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
				ID:         "test-yaml-002",
				SourceType: analysis.SourceTypeWeb,
				SourceID:   "https://example.com",
				AnalyzedAt: startTime,
				Duration:   3 * time.Second,
				Timing: analysis.TimingInfo{
					StartedAt:   startTime,
					CompletedAt: completedTime,
					Duration:    3 * time.Second,
				},
				SourceMetrics: analysis.SourceMetrics{
					ItemsAnalyzed:  200,
					ItemsFlagged:   2,
					AverageScore:   0.72,
					CoverageRate:   0.01,
					StrategiesUsed: 20,
					StrategiesHit:  2,
					Extra:          map[string]interface{}{"wordCount": 1500},
				},
				Detections: []analysis.Detection{
					{
						Strategy:    "generic-language",
						Detected:    true,
						Severity:    "medium",
						Score:       0.65,
						Confidence:  0.7,
						Category:    "linguistic",
						Description: "Generic phrasing detected",
						Examples:    []string{"It is important to note that"},
					},
					{
						Strategy:    "overused-phrases",
						Detected:    true,
						Severity:    "high",
						Score:       0.80,
						Confidence:  0.85,
						Category:    "pattern",
						Description: "High frequency of overused AI phrases",
						Examples:    []string{"delve", "landscape"},
					},
				},
				OverallScore:        72,
				Assessment:          "Suspicious Activity Detected",
				SuspicionRate:       1.0,
				TotalDetections:     2,
				DetectionCount:      2,
				PassedDetections:    0,
				HighSeverityCount:   1,
				MediumSeverityCount: 1,
				Metrics:             map[string]interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &YAMLReporter{}
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

			// Verify output is valid YAML
			var result map[string]interface{}
			if err := yaml.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Output is not valid YAML: %v", err)
			}

			// Check basic fields exist
			if _, ok := result["id"]; !ok {
				t.Error("Missing 'id' field in YAML output")
			}
			if _, ok := result["source_type"]; !ok {
				t.Error("Missing 'source_type' field in YAML output")
			}
			if _, ok := result["detections"]; !ok {
				t.Error("Missing 'detections' field in YAML output")
			}

			// Check timing fields
			timing, ok := result["timing"].(map[string]interface{})
			if !ok {
				t.Error("Missing or invalid 'timing' object in YAML output")
			} else {
				if _, ok := timing["started_at"]; !ok {
					t.Error("Missing 'timing.started_at' field")
				}
				if _, ok := timing["completed_at"]; !ok {
					t.Error("Missing 'timing.completed_at' field")
				}
				if _, ok := timing["duration_ms"]; !ok {
					t.Error("Missing 'timing.duration_ms' field")
				}
			}

			// Check source metrics
			sm, ok := result["source_metrics"].(map[string]interface{})
			if !ok {
				t.Error("Missing or invalid 'source_metrics' object in YAML output")
			} else {
				if _, ok := sm["items_analyzed"]; !ok {
					t.Error("Missing 'source_metrics.items_analyzed' field")
				}
				if _, ok := sm["strategies_used"]; !ok {
					t.Error("Missing 'source_metrics.strategies_used' field")
				}
			}
		})
	}
}
