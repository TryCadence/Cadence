package formats

import (
	"strings"
	"testing"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
)

func TestTextReporter_FormatAnalysis(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-5 * time.Second)
	completedTime := now

	tests := []struct {
		name         string
		report       *analysis.AnalysisReport
		wantErr      bool
		wantContains []string
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
			wantContains: []string{
				"CADENCE ANALYSIS REPORT",
				"test-001",
				"No suspicious activity detected",
				"Started At:",
				"Completed At:",
				"Duration:",
				"Phase Breakdown:",
				"validate:",
				"fetch:",
				"SOURCE METRICS",
				"Items Analyzed:",
				"Strategies Used:",
			},
		},
		{
			name: "formats analysis report with detections",
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
			wantContains: []string{
				"CADENCE ANALYSIS REPORT",
				"Suspicious activity detected",
				"velocity-analysis",
				"Started At:",
				"Completed At:",
				"SOURCE METRICS",
				"Unique Authors:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := &TextReporter{}
			output, err := reporter.FormatAnalysis(tt.report)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FormatAnalysis() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if output == "" {
				t.Error("FormatAnalysis() returned empty output")
			}

			// Check for expected content
			for _, expected := range tt.wantContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output missing expected string: %s", expected)
				}
			}
		})
	}
}
