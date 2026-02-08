package analysis

import (
	"strings"
	"testing"
	"time"
)

func TestInMemoryMetrics_RecordAnalysis(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordAnalysis("git", 2*time.Second)
	m.RecordAnalysis("git", 3*time.Second)
	m.RecordAnalysis("web", 1*time.Second)

	snap := m.Snapshot()
	if snap.TotalAnalyses != 3 {
		t.Fatalf("got %d total analyses, want 3", snap.TotalAnalyses)
	}

	gitData := snap.BySource["git"]
	if gitData == nil {
		t.Fatal("expected git source data")
	}
	if gitData.Analyses != 2 {
		t.Fatalf("got %d git analyses, want 2", gitData.Analyses)
	}
	if gitData.AvgDurationMs != 2500 {
		t.Fatalf("got avg duration %f, want 2500", gitData.AvgDurationMs)
	}

	webData := snap.BySource["web"]
	if webData == nil {
		t.Fatal("expected web source data")
	}
	if webData.Analyses != 1 {
		t.Fatalf("got %d web analyses, want 1", webData.Analyses)
	}
}

func TestInMemoryMetrics_RecordDetections(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordDetections("git", 10, 3)
	m.RecordDetections("web", 5, 2)

	snap := m.Snapshot()
	if snap.TotalDetections != 15 {
		t.Fatalf("got %d total detections, want 15", snap.TotalDetections)
	}
	if snap.TotalFlagged != 5 {
		t.Fatalf("got %d total flagged, want 5", snap.TotalFlagged)
	}
}

func TestInMemoryMetrics_RecordError(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordError("git", "clone")
	m.RecordError("git", "analysis")
	m.RecordError("web", "analysis")

	snap := m.Snapshot()
	if snap.TotalErrors != 3 {
		t.Fatalf("got %d total errors, want 3", snap.TotalErrors)
	}

	if snap.ErrorsByPhase["clone"] != 1 {
		t.Fatalf("got %d clone errors, want 1", snap.ErrorsByPhase["clone"])
	}
	if snap.ErrorsByPhase["analysis"] != 2 {
		t.Fatalf("got %d analysis errors, want 2", snap.ErrorsByPhase["analysis"])
	}
}

func TestInMemoryMetrics_RecordStrategyExecution(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordStrategyExecution("velocity_analysis", true, 100*time.Millisecond)
	m.RecordStrategyExecution("velocity_analysis", false, 50*time.Millisecond)
	m.RecordStrategyExecution("size_analysis", true, 200*time.Millisecond)

	snap := m.Snapshot()
	velStrat := snap.ByStrategy["velocity_analysis"]
	if velStrat == nil {
		t.Fatal("expected velocity_analysis strategy data")
	}
	if velStrat.Executions != 2 {
		t.Fatalf("got %d executions, want 2", velStrat.Executions)
	}
	if velStrat.Detections != 1 {
		t.Fatalf("got %d detections, want 1", velStrat.Detections)
	}
}

func TestInMemoryMetrics_CacheHitsMisses(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordCacheHit("git")
	m.RecordCacheHit("git")
	m.RecordCacheMiss("git")
	m.RecordCacheMiss("web")

	snap := m.Snapshot()
	if snap.CacheHits != 2 {
		t.Fatalf("got %d cache hits, want 2", snap.CacheHits)
	}
	if snap.CacheMisses != 2 {
		t.Fatalf("got %d cache misses, want 2", snap.CacheMisses)
	}

	gitData := snap.BySource["git"]
	if gitData.CacheHits != 2 {
		t.Fatalf("got %d git cache hits, want 2", gitData.CacheHits)
	}
}

func TestInMemoryMetrics_Reset(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordAnalysis("git", time.Second)
	m.RecordError("git", "clone")
	m.RecordDetections("git", 5, 2)

	m.Reset()

	snap := m.Snapshot()
	if snap.TotalAnalyses != 0 {
		t.Fatalf("got %d analyses after reset, want 0", snap.TotalAnalyses)
	}
	if snap.TotalErrors != 0 {
		t.Fatalf("got %d errors after reset, want 0", snap.TotalErrors)
	}
}

func TestInMemoryMetrics_Snapshot_UptimeSeconds(t *testing.T) {
	m := NewInMemoryMetrics()
	time.Sleep(10 * time.Millisecond)

	snap := m.Snapshot()
	if snap.UptimeSeconds < 0.01 {
		t.Fatalf("uptime too low: %f", snap.UptimeSeconds)
	}
}

func TestInMemoryMetrics_PrometheusFormat(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordAnalysis("git", 2*time.Second)
	m.RecordDetections("git", 10, 3)
	m.RecordError("git", "clone")
	m.RecordStrategyExecution("velocity_analysis", true, 100*time.Millisecond)
	m.RecordCacheHit("git")

	output := m.PrometheusFormat()

	// Check for key Prometheus metrics
	expected := []string{
		"cadence_analyses_total",
		"cadence_errors_total",
		"cadence_detections_total",
		"cadence_flagged_total",
		"cadence_uptime_seconds",
		"cadence_cache_hits_total",
		"cadence_cache_misses_total",
		"cadence_source_analyses_total{source=\"git\"}",
		"cadence_strategy_executions_total{strategy=\"velocity_analysis\"}",
		"cadence_errors_by_phase{phase=\"clone\"}",
		"# HELP",
		"# TYPE",
	}

	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("prometheus output missing %q", s)
		}
	}
}

func TestInMemoryMetrics_AvgDurationMs(t *testing.T) {
	m := NewInMemoryMetrics()

	m.RecordAnalysis("git", 1*time.Second)
	m.RecordAnalysis("web", 3*time.Second)

	snap := m.Snapshot()
	// Total duration = 4000ms / 2 analyses = 2000ms avg
	if snap.AvgDurationMs != 2000 {
		t.Fatalf("got avg duration %f, want 2000", snap.AvgDurationMs)
	}
}

func TestNullMetrics(t *testing.T) {
	m := NullMetrics{}

	// All methods should be no-ops
	m.RecordAnalysis("git", time.Second)
	m.RecordDetections("git", 5, 2)
	m.RecordError("git", "test")
	m.RecordStrategyExecution("test", true, time.Second)
	m.RecordCacheHit("git")
	m.RecordCacheMiss("git")
	m.Reset()

	snap := m.Snapshot()
	if snap == nil {
		t.Fatal("NullMetrics.Snapshot should return non-nil")
	}
	if snap.TotalAnalyses != 0 {
		t.Fatal("NullMetrics should have zero metrics")
	}

	if m.PrometheusFormat() != "" {
		t.Fatal("NullMetrics.PrometheusFormat should return empty string")
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := sortedKeys(m)
	if len(keys) != 3 {
		t.Fatalf("got %d keys, want 3", len(keys))
	}
	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("keys not sorted: %v", keys)
	}
}

func TestInMemoryMetrics_ConcurrentAccess(t *testing.T) {
	m := NewInMemoryMetrics()

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.RecordAnalysis("git", time.Millisecond)
				m.RecordDetections("git", 1, 0)
				m.RecordError("git", "test")
				m.RecordStrategyExecution("test", true, time.Millisecond)
				m.RecordCacheHit("git")
				m.RecordCacheMiss("web")
				_ = m.Snapshot()
				_ = m.PrometheusFormat()
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	snap := m.Snapshot()
	if snap.TotalAnalyses != 1000 {
		t.Fatalf("got %d analyses, want 1000", snap.TotalAnalyses)
	}
}
