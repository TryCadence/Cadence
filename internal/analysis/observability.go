package analysis

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// AnalysisMetrics defines the interface for recording server-level observability data.
type AnalysisMetrics interface {
	// RecordAnalysis records a completed analysis with its type and duration.
	RecordAnalysis(sourceType string, duration time.Duration)

	// RecordDetections records the number of detections found in an analysis.
	RecordDetections(sourceType string, count int, flaggedCount int)

	// RecordError records an analysis error.
	RecordError(sourceType string, phase string)

	// RecordStrategyExecution records a strategy run with its result.
	RecordStrategyExecution(strategy string, detected bool, duration time.Duration)

	// RecordCacheHit records a cache hit.
	RecordCacheHit(sourceType string)

	// RecordCacheMiss records a cache miss.
	RecordCacheMiss(sourceType string)

	// Snapshot returns a point-in-time copy of all metrics.
	Snapshot() *MetricsSnapshot

	// PrometheusFormat returns metrics as Prometheus text exposition format.
	PrometheusFormat() string

	// Reset clears all metrics.
	Reset()
}

// MetricsSnapshot is a serializable point-in-time view of all collected metrics.
type MetricsSnapshot struct {
	CollectedAt     time.Time                     `json:"collectedAt"`
	UptimeSeconds   float64                       `json:"uptimeSeconds"`
	TotalAnalyses   int64                         `json:"totalAnalyses"`
	TotalErrors     int64                         `json:"totalErrors"`
	TotalDetections int64                         `json:"totalDetections"`
	TotalFlagged    int64                         `json:"totalFlagged"`
	BySource        map[string]*SourceMetricsData `json:"bySource"`
	ByStrategy      map[string]*StrategyMetrics   `json:"byStrategy"`
	ErrorsByPhase   map[string]int64              `json:"errorsByPhase"`
	CacheHits       int64                         `json:"cacheHits"`
	CacheMisses     int64                         `json:"cacheMisses"`
	AvgDurationMs   float64                       `json:"avgDurationMs"`
}

// SourceMetricsData holds per-source-type metrics.
type SourceMetricsData struct {
	Analyses        int64   `json:"analyses"`
	Errors          int64   `json:"errors"`
	Detections      int64   `json:"detections"`
	Flagged         int64   `json:"flagged"`
	TotalDurationMs int64   `json:"totalDurationMs"`
	AvgDurationMs   float64 `json:"avgDurationMs"`
	CacheHits       int64   `json:"cacheHits"`
	CacheMisses     int64   `json:"cacheMisses"`
}

// StrategyMetrics holds per-strategy execution metrics.
type StrategyMetrics struct {
	Executions      int64   `json:"executions"`
	Detections      int64   `json:"detections"`
	TotalDurationMs int64   `json:"totalDurationMs"`
	AvgDurationMs   float64 `json:"avgDurationMs"`
}

// InMemoryMetrics is a thread-safe in-memory implementation of AnalysisMetrics.
type InMemoryMetrics struct {
	mu            sync.RWMutex
	startTime     time.Time
	totalAnalyses atomic.Int64
	totalErrors   atomic.Int64
	totalDetect   atomic.Int64
	totalFlagged  atomic.Int64
	cacheHits     atomic.Int64
	cacheMisses   atomic.Int64

	// Guarded by mu
	sources    map[string]*sourceCounter
	strategies map[string]*strategyCounter
	errors     map[string]*atomic.Int64 // phase -> count
}

type sourceCounter struct {
	analyses    atomic.Int64
	errors      atomic.Int64
	detections  atomic.Int64
	flagged     atomic.Int64
	totalDurMs  atomic.Int64
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64
}

type strategyCounter struct {
	executions atomic.Int64
	detections atomic.Int64
	totalDurMs atomic.Int64
}

// NewInMemoryMetrics creates a new in-memory metrics collector.
func NewInMemoryMetrics() *InMemoryMetrics {
	return &InMemoryMetrics{
		startTime:  time.Now(),
		sources:    make(map[string]*sourceCounter),
		strategies: make(map[string]*strategyCounter),
		errors:     make(map[string]*atomic.Int64),
	}
}

func (m *InMemoryMetrics) getSource(sourceType string) *sourceCounter {
	m.mu.RLock()
	sc, ok := m.sources[sourceType]
	m.mu.RUnlock()
	if ok {
		return sc
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	// Double-check after acquiring write lock
	if sc, ok = m.sources[sourceType]; ok {
		return sc
	}
	sc = &sourceCounter{}
	m.sources[sourceType] = sc
	return sc
}

func (m *InMemoryMetrics) getStrategy(name string) *strategyCounter {
	m.mu.RLock()
	sc, ok := m.strategies[name]
	m.mu.RUnlock()
	if ok {
		return sc
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if sc, ok = m.strategies[name]; ok {
		return sc
	}
	sc = &strategyCounter{}
	m.strategies[name] = sc
	return sc
}

func (m *InMemoryMetrics) getPhaseError(phase string) *atomic.Int64 {
	m.mu.RLock()
	counter, ok := m.errors[phase]
	m.mu.RUnlock()
	if ok {
		return counter
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if counter, ok = m.errors[phase]; ok {
		return counter
	}
	counter = &atomic.Int64{}
	m.errors[phase] = counter
	return counter
}

// RecordAnalysis records a completed analysis.
func (m *InMemoryMetrics) RecordAnalysis(sourceType string, duration time.Duration) {
	m.totalAnalyses.Add(1)
	sc := m.getSource(sourceType)
	sc.analyses.Add(1)
	sc.totalDurMs.Add(duration.Milliseconds())
}

// RecordDetections records detection counts.
func (m *InMemoryMetrics) RecordDetections(sourceType string, count int, flaggedCount int) {
	m.totalDetect.Add(int64(count))
	m.totalFlagged.Add(int64(flaggedCount))
	sc := m.getSource(sourceType)
	sc.detections.Add(int64(count))
	sc.flagged.Add(int64(flaggedCount))
}

// RecordError records an analysis error.
func (m *InMemoryMetrics) RecordError(sourceType string, phase string) {
	m.totalErrors.Add(1)
	sc := m.getSource(sourceType)
	sc.errors.Add(1)
	m.getPhaseError(phase).Add(1)
}

// RecordStrategyExecution records a single strategy execution.
func (m *InMemoryMetrics) RecordStrategyExecution(strategy string, detected bool, duration time.Duration) {
	sc := m.getStrategy(strategy)
	sc.executions.Add(1)
	sc.totalDurMs.Add(duration.Milliseconds())
	if detected {
		sc.detections.Add(1)
	}
}

// RecordCacheHit records a cache hit for the source type.
func (m *InMemoryMetrics) RecordCacheHit(sourceType string) {
	m.cacheHits.Add(1)
	m.getSource(sourceType).cacheHits.Add(1)
}

// RecordCacheMiss records a cache miss for the source type.
func (m *InMemoryMetrics) RecordCacheMiss(sourceType string) {
	m.cacheMisses.Add(1)
	m.getSource(sourceType).cacheMisses.Add(1)
}

// Snapshot returns a point-in-time copy of all metrics.
func (m *InMemoryMetrics) Snapshot() *MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	now := time.Now()
	totalAnalyses := m.totalAnalyses.Load()

	snap := &MetricsSnapshot{
		CollectedAt:     now,
		UptimeSeconds:   now.Sub(m.startTime).Seconds(),
		TotalAnalyses:   totalAnalyses,
		TotalErrors:     m.totalErrors.Load(),
		TotalDetections: m.totalDetect.Load(),
		TotalFlagged:    m.totalFlagged.Load(),
		CacheHits:       m.cacheHits.Load(),
		CacheMisses:     m.cacheMisses.Load(),
		BySource:        make(map[string]*SourceMetricsData),
		ByStrategy:      make(map[string]*StrategyMetrics),
		ErrorsByPhase:   make(map[string]int64),
	}

	// Per-source breakdown
	var totalDurMs int64
	for name, sc := range m.sources {
		analyses := sc.analyses.Load()
		durMs := sc.totalDurMs.Load()
		totalDurMs += durMs
		var avgMs float64
		if analyses > 0 {
			avgMs = float64(durMs) / float64(analyses)
		}
		snap.BySource[name] = &SourceMetricsData{
			Analyses:        analyses,
			Errors:          sc.errors.Load(),
			Detections:      sc.detections.Load(),
			Flagged:         sc.flagged.Load(),
			TotalDurationMs: durMs,
			AvgDurationMs:   avgMs,
			CacheHits:       sc.cacheHits.Load(),
			CacheMisses:     sc.cacheMisses.Load(),
		}
	}

	if totalAnalyses > 0 {
		snap.AvgDurationMs = float64(totalDurMs) / float64(totalAnalyses)
	}

	// Per-strategy breakdown
	for name, sc := range m.strategies {
		execs := sc.executions.Load()
		durMs := sc.totalDurMs.Load()
		var avgMs float64
		if execs > 0 {
			avgMs = float64(durMs) / float64(execs)
		}
		snap.ByStrategy[name] = &StrategyMetrics{
			Executions:      execs,
			Detections:      sc.detections.Load(),
			TotalDurationMs: durMs,
			AvgDurationMs:   avgMs,
		}
	}

	// Error breakdown by phase
	for phase, counter := range m.errors {
		snap.ErrorsByPhase[phase] = counter.Load()
	}

	return snap
}

// PrometheusFormat returns all metrics in Prometheus text exposition format.
func (m *InMemoryMetrics) PrometheusFormat() string {
	snap := m.Snapshot()
	var b strings.Builder

	// Global counters
	writePromMetric(&b, "cadence_analyses_total", "counter", "Total number of analyses performed",
		fmt.Sprintf("%d", snap.TotalAnalyses))
	writePromMetric(&b, "cadence_errors_total", "counter", "Total number of analysis errors",
		fmt.Sprintf("%d", snap.TotalErrors))
	writePromMetric(&b, "cadence_detections_total", "counter", "Total number of detections",
		fmt.Sprintf("%d", snap.TotalDetections))
	writePromMetric(&b, "cadence_flagged_total", "counter", "Total flagged items",
		fmt.Sprintf("%d", snap.TotalFlagged))
	writePromMetric(&b, "cadence_uptime_seconds", "gauge", "Server uptime in seconds",
		fmt.Sprintf("%.1f", snap.UptimeSeconds))
	writePromMetric(&b, "cadence_analysis_avg_duration_ms", "gauge", "Average analysis duration in milliseconds",
		fmt.Sprintf("%.2f", snap.AvgDurationMs))

	// Cache counters
	writePromMetric(&b, "cadence_cache_hits_total", "counter", "Total cache hits",
		fmt.Sprintf("%d", snap.CacheHits))
	writePromMetric(&b, "cadence_cache_misses_total", "counter", "Total cache misses",
		fmt.Sprintf("%d", snap.CacheMisses))

	// Per-source metrics
	sourceNames := sortedKeys(snap.BySource)
	for _, name := range sourceNames {
		sd := snap.BySource[name]
		b.WriteString(fmt.Sprintf("cadence_source_analyses_total{source=\"%s\"} %d\n", name, sd.Analyses))
		b.WriteString(fmt.Sprintf("cadence_source_errors_total{source=\"%s\"} %d\n", name, sd.Errors))
		b.WriteString(fmt.Sprintf("cadence_source_detections_total{source=\"%s\"} %d\n", name, sd.Detections))
		b.WriteString(fmt.Sprintf("cadence_source_avg_duration_ms{source=\"%s\"} %.2f\n", name, sd.AvgDurationMs))
	}

	// Per-strategy metrics (top strategies only to avoid huge output)
	stratNames := sortedKeys(snap.ByStrategy)
	for _, name := range stratNames {
		sm := snap.ByStrategy[name]
		b.WriteString(fmt.Sprintf("cadence_strategy_executions_total{strategy=\"%s\"} %d\n", name, sm.Executions))
		b.WriteString(fmt.Sprintf("cadence_strategy_detections_total{strategy=\"%s\"} %d\n", name, sm.Detections))
		b.WriteString(fmt.Sprintf("cadence_strategy_avg_duration_ms{strategy=\"%s\"} %.2f\n", name, sm.AvgDurationMs))
	}

	// Errors by phase
	phaseNames := sortedKeys(snap.ErrorsByPhase)
	for _, phase := range phaseNames {
		b.WriteString(fmt.Sprintf("cadence_errors_by_phase{phase=\"%s\"} %d\n", phase, snap.ErrorsByPhase[phase]))
	}

	return b.String()
}

// Reset clears all metrics.
func (m *InMemoryMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.startTime = time.Now()
	m.totalAnalyses.Store(0)
	m.totalErrors.Store(0)
	m.totalDetect.Store(0)
	m.totalFlagged.Store(0)
	m.cacheHits.Store(0)
	m.cacheMisses.Store(0)
	m.sources = make(map[string]*sourceCounter)
	m.strategies = make(map[string]*strategyCounter)
	m.errors = make(map[string]*atomic.Int64)
}

// NullMetrics is a no-op AnalysisMetrics implementation for when metrics are disabled.
type NullMetrics struct{}

func (NullMetrics) RecordAnalysis(string, time.Duration)                {}
func (NullMetrics) RecordDetections(string, int, int)                   {}
func (NullMetrics) RecordError(string, string)                          {}
func (NullMetrics) RecordStrategyExecution(string, bool, time.Duration) {}
func (NullMetrics) RecordCacheHit(string)                               {}
func (NullMetrics) RecordCacheMiss(string)                              {}
func (NullMetrics) Snapshot() *MetricsSnapshot                          { return &MetricsSnapshot{} }
func (NullMetrics) PrometheusFormat() string                            { return "" }
func (NullMetrics) Reset()                                              {}

// Helper to write a single Prometheus metric with TYPE and HELP.
func writePromMetric(b *strings.Builder, name, metricType, help, value string) {
	b.WriteString(fmt.Sprintf("# HELP %s %s\n", name, help))
	b.WriteString(fmt.Sprintf("# TYPE %s %s\n", name, metricType))
	b.WriteString(fmt.Sprintf("%s %s\n", name, value))
}

// sortedKeys returns sorted keys from a map with string keys.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
