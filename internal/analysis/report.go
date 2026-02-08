package analysis

import "time"

type SourceType string

const (
	SourceTypeGit SourceType = "git"
	SourceTypeWeb SourceType = "web"
)

type Detection struct {
	Strategy    string
	Detected    bool
	Severity    string
	Score       float64
	Confidence  float64 // Strategy's base confidence weight (0.0-1.0)
	Category    string
	Description string
	Examples    []string
}

// TimingInfo holds structured timing data for an analysis run.
type TimingInfo struct {
	StartedAt   time.Time     `json:"startedAt"`
	CompletedAt time.Time     `json:"completedAt"`
	Duration    time.Duration `json:"duration"`
	// Phase-level timing breakdown
	Phases []PhaseTiming `json:"phases,omitempty"`
}

// PhaseTiming records the duration of a single analysis phase.
type PhaseTiming struct {
	Name      string        `json:"name"`
	StartedAt time.Time     `json:"startedAt"`
	Duration  time.Duration `json:"duration"`
}

// SourceMetrics holds cross-source summary metrics computed from any analysis.
type SourceMetrics struct {
	// Common across all sources
	ItemsAnalyzed  int     `json:"itemsAnalyzed"` // commits for git, pages for web
	ItemsFlagged   int     `json:"itemsFlagged"`  // suspicious items
	UniqueAuthors  int     `json:"uniqueAuthors,omitempty"`
	AverageScore   float64 `json:"averageScore"`   // mean detection score
	CoverageRate   float64 `json:"coverageRate"`   // fraction of items that triggered at least one strategy
	StrategiesUsed int     `json:"strategiesUsed"` // how many strategies ran
	StrategiesHit  int     `json:"strategiesHit"`  // how many strategies fired

	// Source-specific extensions (kept as structured map for flexibility)
	Extra map[string]interface{} `json:"extra,omitempty"`
}

type AnalysisReport struct {
	ID                  string
	SourceType          SourceType
	SourceID            string
	AnalyzedAt          time.Time
	Duration            time.Duration
	Timing              TimingInfo
	SourceMetrics       SourceMetrics
	Detections          []Detection
	OverallScore        float64
	Assessment          string
	SuspicionRate       float64
	TotalDetections     int
	DetectionCount      int
	PassedDetections    int
	HighSeverityCount   int
	MediumSeverityCount int
	LowSeverityCount    int
	Metrics             map[string]interface{}
	Error               string
}

func (r *AnalysisReport) GetDetectionsBySeverity(severity string) []Detection {
	var filtered []Detection
	for _, d := range r.Detections {
		if d.Severity == severity {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

// calculateSourceMetrics populates cross-source summary metrics from detections and metadata.
func calculateSourceMetrics(report *AnalysisReport) {
	sm := &report.SourceMetrics

	// Items analyzed — use source-specific counts from metadata when available
	if count, ok := report.Metrics["commit_count"].(int); ok {
		sm.ItemsAnalyzed = count
	} else if count, ok := report.Metrics["word_count"].(int); ok {
		sm.ItemsAnalyzed = count // for web, use word count as item proxy
	}

	sm.ItemsFlagged = report.DetectionCount

	// Unique authors from metadata
	if authors, ok := report.Metrics["unique_authors"].(int); ok {
		sm.UniqueAuthors = authors
	}

	// Strategy counts
	strategySeen := make(map[string]bool)
	strategyHit := make(map[string]bool)
	var totalScore float64
	var scoredCount int

	for _, d := range report.Detections {
		strategySeen[d.Strategy] = true
		if d.Detected {
			strategyHit[d.Strategy] = true
			totalScore += d.Score
			scoredCount++
		}
	}

	sm.StrategiesUsed = len(strategySeen)
	sm.StrategiesHit = len(strategyHit)

	if scoredCount > 0 {
		sm.AverageScore = totalScore / float64(scoredCount)
	}

	if sm.ItemsAnalyzed > 0 {
		sm.CoverageRate = float64(sm.ItemsFlagged) / float64(sm.ItemsAnalyzed)
		if sm.CoverageRate > 1.0 {
			sm.CoverageRate = 1.0
		}
	}

	// Preserve source-specific extras
	if sm.Extra == nil {
		sm.Extra = make(map[string]interface{})
	}

	// Git-specific extras
	if ts, ok := report.Metrics["time_span"].(string); ok {
		sm.Extra["timeSpan"] = ts
	}
	if v, ok := report.Metrics["velocity"].(string); ok {
		sm.Extra["velocity"] = v
	}

	// Web-specific extras
	if wc, ok := report.Metrics["word_count"].(int); ok {
		sm.Extra["wordCount"] = wc
	}
	if cc, ok := report.Metrics["character_count"].(int); ok {
		sm.Extra["characterCount"] = cc
	}
	if hc, ok := report.Metrics["heading_count"].(int); ok {
		sm.Extra["headingCount"] = hc
	}
}
