package analysis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// mockSource is a test AnalysisSource implementation.
type mockSource struct {
	sourceType  string
	validateErr error
	fetchErr    error
	data        *SourceData
}

func (m *mockSource) Type() string                     { return m.sourceType }
func (m *mockSource) Validate(_ context.Context) error { return m.validateErr }
func (m *mockSource) Fetch(_ context.Context) (*SourceData, error) {
	if m.fetchErr != nil {
		return nil, m.fetchErr
	}
	return m.data, nil
}

// mockDetector is a test Detector implementation.
type mockDetector struct {
	detections []Detection
	err        error
}

func (m *mockDetector) Detect(_ context.Context, _ *SourceData) ([]Detection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.detections, nil
}

func TestStreamingRunner_HappyPath(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	detector := &mockDetector{
		detections: []Detection{
			{Strategy: "strat1", Detected: true, Severity: "high", Score: 0.9, Confidence: 0.8, Category: "pattern"},
			{Strategy: "strat2", Detected: false, Severity: "none", Score: 0, Confidence: 0.5, Category: "linguistic"},
		},
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source, detector)

	var progressCount int
	var detectionCount int
	var report *AnalysisReport
	var gotError error

	for event := range events {
		switch event.Type {
		case EventProgress:
			progressCount++
			if event.Progress == nil {
				t.Error("Progress event has nil Progress")
			}
		case EventDetection:
			detectionCount++
			if event.Detection == nil {
				t.Error("Detection event has nil Detection")
			}
		case EventComplete:
			report = event.Report
		case EventError:
			gotError = event.Error
		}
	}

	if gotError != nil {
		t.Fatalf("unexpected error: %v", gotError)
	}

	if progressCount < 3 {
		t.Errorf("expected at least 3 progress events, got %d", progressCount)
	}

	if detectionCount != 2 {
		t.Errorf("expected 2 detection events, got %d", detectionCount)
	}

	if report == nil {
		t.Fatal("expected a complete event with report")
	}

	if report.SourceID != "test-id" {
		t.Errorf("report.SourceID = %q, want %q", report.SourceID, "test-id")
	}

	if len(report.Detections) != 2 {
		t.Errorf("report has %d detections, want 2", len(report.Detections))
	}
}

func TestStreamingRunner_ValidationError(t *testing.T) {
	source := &mockSource{
		sourceType:  "test",
		validateErr: fmt.Errorf("bad source"),
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source)

	var gotError error
	for event := range events {
		if event.Type == EventError {
			gotError = event.Error
		}
	}

	if gotError == nil {
		t.Fatal("expected error event for validation failure")
	}
}

func TestStreamingRunner_FetchError(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		fetchErr:   fmt.Errorf("fetch failed"),
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source)

	var gotError error
	for event := range events {
		if event.Type == EventError {
			gotError = event.Error
		}
	}

	if gotError == nil {
		t.Fatal("expected error event for fetch failure")
	}
}

func TestStreamingRunner_DetectorError(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	detector := &mockDetector{
		err: fmt.Errorf("detector failed"),
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source, detector)

	var gotError error
	for event := range events {
		if event.Type == EventError {
			gotError = event.Error
		}
	}

	if gotError == nil {
		t.Fatal("expected error event for detector failure")
	}
}

func TestStreamingRunner_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	// Slow detector that gives us time to cancel
	slowDetector := &mockDetector{
		detections: []Detection{
			{Strategy: "slow", Detected: true, Severity: "high", Score: 0.9, Confidence: 0.8},
		},
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(ctx, source, slowDetector, slowDetector, slowDetector)

	// Cancel after first event
	firstEvent := <-events
	if firstEvent.Type == "" {
		t.Fatal("expected first event")
	}
	cancel()

	// Drain remaining events — should terminate
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-events:
			if !ok {
				return // channel closed, test passes
			}
		case <-timeout:
			t.Fatal("stream did not close after context cancellation")
		}
	}
}

func TestCollectStream_HappyPath(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	detector := &mockDetector{
		detections: []Detection{
			{Strategy: "s1", Detected: true, Severity: "low", Score: 0.3, Confidence: 0.7},
		},
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source, detector)

	report, err := CollectStream(events)
	if err != nil {
		t.Fatalf("CollectStream error: %v", err)
	}

	if report == nil {
		t.Fatal("CollectStream returned nil report")
	}

	if len(report.Detections) != 1 {
		t.Errorf("report has %d detections, want 1", len(report.Detections))
	}
}

func TestCollectStream_Error(t *testing.T) {
	source := &mockSource{
		sourceType:  "test",
		validateErr: fmt.Errorf("fail"),
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source)

	report, err := CollectStream(events)
	if err == nil {
		t.Fatal("CollectStream should return error")
	}

	if report != nil {
		t.Error("CollectStream should return nil report on error")
	}
}

func TestStreamingRunner_MultipleDetectors(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	d1 := &mockDetector{detections: []Detection{
		{Strategy: "a", Detected: true, Severity: "high", Confidence: 0.9},
	}}
	d2 := &mockDetector{detections: []Detection{
		{Strategy: "b", Detected: true, Severity: "low", Confidence: 0.4},
		{Strategy: "c", Detected: false, Severity: "none", Confidence: 0.5},
	}}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source, d1, d2)

	report, err := CollectStream(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Detections) != 3 {
		t.Errorf("expected 3 detections, got %d", len(report.Detections))
	}
}

func TestStreamingRunner_EmptyDetectors(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source)

	report, err := CollectStream(events)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report == nil {
		t.Fatal("expected report even with no detectors")
	}

	if len(report.Detections) != 0 {
		t.Errorf("expected 0 detections, got %d", len(report.Detections))
	}
}

// panicDetector triggers a panic inside Detect.
type panicDetector struct{}

func (p *panicDetector) Detect(_ context.Context, _ *SourceData) ([]Detection, error) {
	panic("deliberate panic in detector")
}

func TestStreamingRunner_DetectorPanicRecovery(t *testing.T) {
	source := &mockSource{
		sourceType: "test",
		data: &SourceData{
			ID:       "test-id",
			Type:     "test",
			Metadata: map[string]interface{}{},
		},
	}

	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), source, &panicDetector{})

	_, err := CollectStream(events)
	if err == nil {
		t.Fatal("expected an error from panicking detector, got nil")
	}

	if !contains(err.Error(), "internal analysis error") {
		t.Errorf("expected error containing 'internal analysis error', got: %s", err.Error())
	}
}

// panicSource panics during Fetch.
type panicSource struct{}

func (p *panicSource) Type() string                     { return "test" }
func (p *panicSource) Validate(_ context.Context) error { return nil }
func (p *panicSource) Fetch(_ context.Context) (*SourceData, error) {
	panic("deliberate panic in source fetch")
}

func TestStreamingRunner_FetchPanicRecovery(t *testing.T) {
	runner := NewStreamingRunner()
	events := runner.RunStream(context.Background(), &panicSource{}, &mockDetector{})

	_, err := CollectStream(events)
	if err == nil {
		t.Fatal("expected an error from panicking source, got nil")
	}

	if !contains(err.Error(), "internal analysis error") {
		t.Errorf("expected error containing 'internal analysis error', got: %s", err.Error())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
