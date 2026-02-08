package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/google/uuid"
)

type DefaultDetectionRunner struct {
	logger *logging.Logger
}

func NewDefaultDetectionRunner() *DefaultDetectionRunner {
	return &DefaultDetectionRunner{
		logger: logging.Default(),
	}
}

func NewDefaultDetectionRunnerWithLogger(logger *logging.Logger) *DefaultDetectionRunner {
	if logger == nil {
		logger = logging.Default()
	}
	return &DefaultDetectionRunner{logger: logger}
}

func (r *DefaultDetectionRunner) Run(ctx context.Context, source AnalysisSource, detectors ...Detector) (*AnalysisReport, error) {
	startTime := time.Now()

	r.logger.LogAnalysis(source.Type(), "", "phase", "validating")

	phaseStart := time.Now()
	if err := source.Validate(ctx); err != nil {
		return nil, fmt.Errorf("source validation failed: %w", err)
	}
	validatePhase := PhaseTiming{Name: "validate", StartedAt: phaseStart, Duration: time.Since(phaseStart)}

	r.logger.LogAnalysis(source.Type(), "", "phase", "fetching")

	phaseStart = time.Now()
	sourceData, err := source.Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source data: %w", err)
	}
	fetchPhase := PhaseTiming{Name: "fetch", StartedAt: phaseStart, Duration: time.Since(phaseStart)}

	report := &AnalysisReport{
		ID:         uuid.New().String(),
		SourceType: SourceType(source.Type()),
		SourceID:   sourceData.ID,
		AnalyzedAt: startTime,
		Detections: make([]Detection, 0),
		Metrics:    sourceData.Metadata,
	}

	r.logger.LogAnalysis(source.Type(), sourceData.ID, "phase", "detecting", "detector_count", len(detectors))

	phaseStart = time.Now()
	for _, detector := range detectors {
		detections, err := detector.Detect(ctx, sourceData)
		if err != nil {
			return nil, fmt.Errorf("detection failed: %w", err)
		}

		report.Detections = append(report.Detections, detections...)
	}
	detectPhase := PhaseTiming{Name: "detect", StartedAt: phaseStart, Duration: time.Since(phaseStart)}

	completedAt := time.Now()
	report.Duration = completedAt.Sub(startTime)
	report.Timing = TimingInfo{
		StartedAt:   startTime,
		CompletedAt: completedAt,
		Duration:    report.Duration,
		Phases:      []PhaseTiming{validatePhase, fetchPhase, detectPhase},
	}

	calculateReportStats(report)
	calculateSourceMetrics(report)

	r.logger.LogAnalysis(source.Type(), sourceData.ID,
		"phase", "complete",
		"duration_ms", report.Duration.Milliseconds(),
		"total_detections", report.TotalDetections,
		"detected_count", report.DetectionCount,
		"overall_score", report.OverallScore,
	)

	return report, nil
}

func calculateReportStats(report *AnalysisReport) {
	report.TotalDetections = len(report.Detections)

	highCount := 0
	mediumCount := 0
	lowCount := 0
	detectedCount := 0

	for _, d := range report.Detections {
		if d.Detected {
			detectedCount++

			// Use confidence to weight the scoring contribution.
			// Higher-confidence strategies contribute more to the overall score.
			weight := d.Confidence
			if weight <= 0 {
				weight = 0.5 // default weight for strategies without confidence
			}

			switch d.Severity {
			case "high":
				highCount++
				report.OverallScore += 0.4 * weight
			case "medium":
				mediumCount++
				report.OverallScore += 0.2 * weight
			case "low":
				lowCount++
				report.OverallScore += 0.1 * weight
			}
		}
	}

	report.HighSeverityCount = highCount
	report.MediumSeverityCount = mediumCount
	report.LowSeverityCount = lowCount
	report.DetectionCount = detectedCount
	report.PassedDetections = report.TotalDetections - detectedCount

	if report.TotalDetections > 0 {
		report.SuspicionRate = float64(detectedCount) / float64(report.TotalDetections)
	}

	if report.OverallScore > 100 {
		report.OverallScore = 100
	}

	if report.OverallScore >= 70 {
		report.Assessment = "Suspicious Activity Detected"
	} else if report.OverallScore >= 40 {
		report.Assessment = "Moderate Suspicion"
	} else {
		report.Assessment = "Low Suspicion"
	}
}
