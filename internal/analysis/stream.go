package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/google/uuid"
)

type StreamEventType string

const (
	EventDetection StreamEventType = "detection"
	EventProgress  StreamEventType = "progress"
	EventComplete  StreamEventType = "complete"
	EventError     StreamEventType = "error"
)

type StreamEvent struct {
	Type      StreamEventType
	Detection *Detection
	Progress  *ProgressInfo
	Report    *AnalysisReport
	Error     error
}

type ProgressInfo struct {
	Phase       string
	Current     int
	Total       int
	Message     string
	ElapsedTime time.Duration
}

type StreamingRunner struct {
	logger *logging.Logger
}

func NewStreamingRunner() *StreamingRunner {
	return &StreamingRunner{
		logger: logging.Default(),
	}
}

func NewStreamingRunnerWithLogger(logger *logging.Logger) *StreamingRunner {
	if logger == nil {
		logger = logging.Default()
	}
	return &StreamingRunner{logger: logger}
}

func (r *StreamingRunner) RunStream(ctx context.Context, source AnalysisSource, detectors ...Detector) <-chan StreamEvent {
	events := make(chan StreamEvent, 64)

	go func() {
		defer close(events)

		// Recover from panics in sources/detectors so an unrecovered panic
		// doesn't kill the entire server process during SSE streaming.
		defer func() {
			if rec := recover(); rec != nil {
				r.logger.Error("panic in streaming runner", "panic", fmt.Sprintf("%v", rec))
				r.emit(ctx, events, StreamEvent{
					Type:  EventError,
					Error: fmt.Errorf("internal analysis error: %v", rec),
				})
			}
		}()

		startTime := time.Now()
		var phases []PhaseTiming

		r.emit(ctx, events, StreamEvent{
			Type: EventProgress,
			Progress: &ProgressInfo{
				Phase:   "validating",
				Message: "Validating source input",
			},
		})

		phaseStart := time.Now()
		if err := source.Validate(ctx); err != nil {
			r.emit(ctx, events, StreamEvent{
				Type:  EventError,
				Error: err,
			})
			return
		}
		phases = append(phases, PhaseTiming{Name: "validate", StartedAt: phaseStart, Duration: time.Since(phaseStart)})

		r.emit(ctx, events, StreamEvent{
			Type: EventProgress,
			Progress: &ProgressInfo{
				Phase:       "fetching",
				Message:     "Fetching source data",
				ElapsedTime: time.Since(startTime),
			},
		})

		phaseStart = time.Now()
		sourceData, err := source.Fetch(ctx)
		if err != nil {
			r.emit(ctx, events, StreamEvent{
				Type:  EventError,
				Error: err,
			})
			return
		}
		phases = append(phases, PhaseTiming{Name: "fetch", StartedAt: phaseStart, Duration: time.Since(phaseStart)})

		report := &AnalysisReport{
			ID:         uuid.New().String(),
			SourceType: SourceType(source.Type()),
			SourceID:   sourceData.ID,
			AnalyzedAt: startTime,
			Detections: make([]Detection, 0),
			Metrics:    sourceData.Metadata,
		}

		r.emit(ctx, events, StreamEvent{
			Type: EventProgress,
			Progress: &ProgressInfo{
				Phase:       "detecting",
				Current:     0,
				Total:       len(detectors),
				Message:     "Running detectors",
				ElapsedTime: time.Since(startTime),
			},
		})

		detectPhaseStart := time.Now()
		for i, detector := range detectors {
			select {
			case <-ctx.Done():
				r.emit(ctx, events, StreamEvent{
					Type:  EventError,
					Error: ctx.Err(),
				})
				return
			default:
			}

			detections, err := detector.Detect(ctx, sourceData)
			if err != nil {
				r.emit(ctx, events, StreamEvent{
					Type:  EventError,
					Error: err,
				})
				return
			}

			for j := range detections {
				report.Detections = append(report.Detections, detections[j])

				r.emit(ctx, events, StreamEvent{
					Type:      EventDetection,
					Detection: &detections[j],
				})
			}

			r.emit(ctx, events, StreamEvent{
				Type: EventProgress,
				Progress: &ProgressInfo{
					Phase:       "detecting",
					Current:     i + 1,
					Total:       len(detectors),
					Message:     "Running detectors",
					ElapsedTime: time.Since(startTime),
				},
			})
		}
		phases = append(phases, PhaseTiming{Name: "detect", StartedAt: detectPhaseStart, Duration: time.Since(detectPhaseStart)})

		completedAt := time.Now()
		report.Duration = completedAt.Sub(startTime)
		report.Timing = TimingInfo{
			StartedAt:   startTime,
			CompletedAt: completedAt,
			Duration:    report.Duration,
			Phases:      phases,
		}
		calculateReportStats(report)
		calculateSourceMetrics(report)

		r.logger.LogAnalysis(source.Type(), sourceData.ID,
			"phase", "stream_complete",
			"duration_ms", report.Duration.Milliseconds(),
			"total_detections", report.TotalDetections,
			"detected_count", report.DetectionCount,
		)

		r.emit(ctx, events, StreamEvent{
			Type:   EventComplete,
			Report: report,
		})
	}()

	return events
}

func (r *StreamingRunner) emit(ctx context.Context, ch chan<- StreamEvent, event StreamEvent) {
	select {
	case ch <- event:
	case <-ctx.Done():
	}
}

func CollectStream(events <-chan StreamEvent) (*AnalysisReport, error) {
	var lastReport *AnalysisReport
	var lastErr error

	for event := range events {
		switch event.Type {
		case EventComplete:
			lastReport = event.Report
		case EventError:
			lastErr = event.Error
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return lastReport, nil
}
