package webhook

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/analysis/detectors"
	"github.com/TryCadence/Cadence/internal/analysis/sources"
	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SSE event types sent to clients.
const (
	SSEEventProgress  = "progress"
	SSEEventDetection = "detection"
	SSEEventResult    = "result"
	SSEEventError     = "error"
)

// SSEProgressEvent is sent during analysis phases.
type SSEProgressEvent struct {
	Phase     string  `json:"phase"`
	Message   string  `json:"message"`
	Current   int     `json:"current,omitempty"`
	Total     int     `json:"total,omitempty"`
	ElapsedMs int64   `json:"elapsed_ms,omitempty"`
	Percent   float64 `json:"percent,omitempty"`
}

// SSEDetectionEvent is sent when a detection fires during streaming analysis.
type SSEDetectionEvent struct {
	Strategy    string   `json:"strategy"`
	Detected    bool     `json:"detected"`
	Severity    string   `json:"severity"`
	Score       float64  `json:"score"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
}

// StreamAnalyzeRepository handles POST /api/stream/repository
// It runs analysis using the StreamingRunner and sends SSE events in real-time.
func (wh *WebhookHandlers) StreamAnalyzeRepository(c *fiber.Ctx) error {
	var req AnalyzeRepositoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.RepositoryURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "repository_url is required",
		})
	}

	log := logging.Default().With("component", "stream_handler")
	jobID := uuid.New().String()
	branch := req.Branch
	repoURL := req.RepositoryURL
	thresholds := wh.processor.DetectorThresholds

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	c.Set("X-Job-ID", jobID)

	// Clear the per-connection write deadline so fasthttp doesn't kill the SSE stream.
	// The stream has its own 5-minute context timeout.
	c.Context().Conn().SetWriteDeadline(time.Time{})

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Recover from panics so the connection closes cleanly with an error event
		// instead of an abrupt TCP RST that the browser sees as "network error".
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic in stream writer", "job_id", jobID, "panic", fmt.Sprintf("%v", r))
				writeSSE(w, SSEEventError, fiber.Map{
					"message": fmt.Sprintf("Internal server error: %v", r),
				})
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		log.Info("SSE stream started", "job_id", jobID, "type", "repository", "url", repoURL)

		// Send initial progress
		writeSSE(w, SSEEventProgress, SSEProgressEvent{
			Phase:   "queued",
			Message: "Analysis job accepted",
		})

		// Clone repository — run in goroutine with keepalive heartbeats
		// so the SSE connection doesn't appear idle to browsers/proxies.
		writeSSE(w, SSEEventProgress, SSEProgressEvent{
			Phase:   "cloning",
			Message: fmt.Sprintf("Cloning repository %s", repoURL),
		})

		tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("cadence-stream-%s", jobID))
		defer os.RemoveAll(tmpDir)

		cloneErr := make(chan error, 1)
		cloneStart := time.Now()
		go func() {
			cloneErr <- cloneRepo(repoURL, tmpDir)
		}()

		// Send keepalive heartbeats every 10s while clone is in progress
		heartbeat := time.NewTicker(10 * time.Second)
		defer heartbeat.Stop()

		var err error
	cloneLoop:
		for {
			select {
			case err = <-cloneErr:
				break cloneLoop
			case <-heartbeat.C:
				elapsed := time.Since(cloneStart)
				writeSSE(w, SSEEventProgress, SSEProgressEvent{
					Phase:     "cloning",
					Message:   fmt.Sprintf("Still cloning repository... (%ds elapsed)", int(elapsed.Seconds())),
					ElapsedMs: elapsed.Milliseconds(),
					Percent:   15,
				})
			case <-ctx.Done():
				err = ctx.Err()
				break cloneLoop
			}
		}

		if err != nil {
			log.Error("clone failed", "error", err, "job_id", jobID)
			writeSSE(w, SSEEventError, fiber.Map{
				"message": fmt.Sprintf("Failed to clone repository: %s", err.Error()),
			})
			return
		}

		writeSSE(w, SSEEventProgress, SSEProgressEvent{
			Phase:     "analyzing",
			Message:   "Repository cloned, starting analysis",
			ElapsedMs: time.Since(cloneStart).Milliseconds(),
		})

		source := sources.NewGitRepositorySource(tmpDir, branch)
		det := detectors.NewGitDetector(thresholds)
		runner := analysis.NewStreamingRunner()

		events := runner.RunStream(ctx, source, det)
		streamEventsToSSEWithMetrics(w, events, log, jobID, "api_analysis_repo", wh.metrics)

		log.Info("SSE stream ended", "job_id", jobID, "type", "repository")
	})

	return nil
}

// StreamAnalyzeWebsite handles POST /api/stream/website
// Same as StreamAnalyzeRepository but for web content.
func (wh *WebhookHandlers) StreamAnalyzeWebsite(c *fiber.Ctx) error {
	var req AnalyzeWebsiteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.URL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "url is required",
		})
	}

	log := logging.Default().With("component", "stream_handler")
	jobID := uuid.New().String()
	targetURL := req.URL

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")
	c.Set("X-Job-ID", jobID)

	// Clear the per-connection write deadline so fasthttp doesn't kill the SSE stream.
	c.Context().Conn().SetWriteDeadline(time.Time{})

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Recover from panics so the connection closes cleanly with an error event
		// instead of an abrupt TCP RST that the browser sees as "network error".
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic in stream writer", "job_id", jobID, "panic", fmt.Sprintf("%v", r))
				writeSSE(w, SSEEventError, fiber.Map{
					"message": fmt.Sprintf("Internal server error: %v", r),
				})
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		log.Info("SSE stream started", "job_id", jobID, "type", "website", "url", targetURL)

		writeSSE(w, SSEEventProgress, SSEProgressEvent{
			Phase:   "fetching",
			Message: fmt.Sprintf("Fetching content from %s", targetURL),
		})

		source := sources.NewWebsiteSource(targetURL)
		det := detectors.NewWebDetector()
		runner := analysis.NewStreamingRunner()

		events := runner.RunStream(ctx, source, det)
		streamEventsToSSEWithMetrics(w, events, log, jobID, "api_analysis_website", wh.metrics)

		log.Info("SSE stream ended", "job_id", jobID, "type", "website")
	})

	return nil
}

// streamEventsToSSE reads from the StreamingRunner channel and writes SSE events to the response writer.
// It sends heartbeat comments when no events arrive for 15 seconds, keeping the chunked
// connection alive through proxies and browsers.
func streamEventsToSSE(w *bufio.Writer, events <-chan analysis.StreamEvent, log *logging.Logger, jobID string, eventType string) {
	streamEventsToSSEWithMetrics(w, events, log, jobID, eventType, analysis.NullMetrics{})
}

// streamEventsToSSEWithMetrics is the same as streamEventsToSSE but also records analysis metrics.
func streamEventsToSSEWithMetrics(w *bufio.Writer, events <-chan analysis.StreamEvent, log *logging.Logger, jobID string, eventType string, metrics analysis.AnalysisMetrics) {
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return // channel closed — stream complete
			}
			heartbeat.Reset(15 * time.Second)

			switch event.Type {
			case analysis.EventProgress:
				if event.Progress != nil {
					if !writeSSE(w, SSEEventProgress, SSEProgressEvent{
						Phase:     event.Progress.Phase,
						Message:   event.Progress.Message,
						Current:   event.Progress.Current,
						Total:     event.Progress.Total,
						ElapsedMs: event.Progress.ElapsedTime.Milliseconds(),
						Percent:   progressPercent(event.Progress),
					}) {
						log.Warn("SSE write failed (client disconnected?)", "job_id", jobID)
						return
					}
				}

			case analysis.EventDetection:
				if event.Detection != nil {
					if !writeSSE(w, SSEEventDetection, SSEDetectionEvent{
						Strategy:    event.Detection.Strategy,
						Detected:    event.Detection.Detected,
						Severity:    event.Detection.Severity,
						Score:       event.Detection.Score,
						Category:    event.Detection.Category,
						Description: event.Detection.Description,
						Examples:    event.Detection.Examples,
					}) {
						log.Warn("SSE write failed (client disconnected?)", "job_id", jobID)
						return
					}
				}

			case analysis.EventComplete:
				if event.Report != nil {
					log.Info("stream analysis complete",
						"job_id", jobID,
						"detections", event.Report.DetectionCount,
						"duration_ms", event.Report.Duration.Milliseconds(),
					)

					// Record metrics
					sourceType := string(event.Report.SourceType)
					metrics.RecordAnalysis(sourceType, event.Report.Duration)
					metrics.RecordDetections(sourceType, event.Report.TotalDetections, event.Report.DetectionCount)

					// Build the final result in the same format as the non-streaming endpoint
					result := buildJobResult(event.Report, eventType)
					writeSSE(w, SSEEventResult, result)
				}

			case analysis.EventError:
				if event.Error != nil {
					log.Error("stream analysis error", "job_id", jobID, "error", event.Error)
					metrics.RecordError(eventType, "stream")
					writeSSE(w, SSEEventError, fiber.Map{
						"message": event.Error.Error(),
					})
				}
			}

		case <-heartbeat.C:
			// Send an SSE comment to keep the chunked connection alive.
			// Browsers and proxies may drop idle chunked streams.
			if _, err := fmt.Fprint(w, ":heartbeat\n\n"); err != nil {
				log.Warn("heartbeat write failed (client disconnected?)", "job_id", jobID)
				return
			}
			if err := w.Flush(); err != nil {
				log.Warn("heartbeat flush failed (client disconnected?)", "job_id", jobID)
				return
			}
		}
	}
}

// buildJobResult converts an AnalysisReport into a JobResultResponse for the final SSE event.
func buildJobResult(report *analysis.AnalysisReport, eventType string) *JobResultResponse {
	resp := &JobResultResponse{
		Status:     StatusCompleted,
		AnalyzedAt: report.Timing.StartedAt,
	}

	if eventType == "api_analysis_repo" {
		// Populate repository fields
		if commitCount, ok := report.Metrics["commit_count"].(int); ok {
			resp.TotalCommits = commitCount
		}

		for _, d := range report.Detections {
			if d.Category == "git-analysis" && d.Detected {
				hash := ""
				var reasons []string
				if len(d.Examples) > 0 {
					hash = d.Examples[0]
					if len(d.Examples) > 1 {
						reasons = d.Examples[1:]
					}
				}
				resp.Suspicions = append(resp.Suspicions, Suspicion{
					CommitHash: hash,
					Message:    d.Description,
					Severity:   d.Severity,
					Reasons:    reasons,
					Score:      d.Score * 100,
				})
			}
		}
		resp.SuspiciousCommits = len(resp.Suspicions)

		if pairs, ok := report.Metrics["commit_pairs"].([]*git.CommitPair); ok {
			jr := mapToJobResult(resp)
			calculateMetrics(jr, pairs)
			resp.Velocity = jr.Velocity
			resp.TimeSpan = jr.TimeSpan
			resp.UniqueAuthors = jr.UniqueAuthors
			resp.AverageCommitSize = jr.AverageCommitSize
			resp.OverallSuspicion = jr.OverallSuspicion
		} else if resp.TotalCommits > 0 {
			resp.OverallSuspicion = (float64(resp.SuspiciousCommits) / float64(resp.TotalCommits)) * 100
		}
	} else if eventType == "api_analysis_website" {
		// Populate website fields
		resp.URL = report.SourceID
		resp.WebPatterns = make([]WebPattern, 0)
		resp.PassedPatterns = make([]WebPattern, 0)

		for _, d := range report.Detections {
			if d.Category == "web-pattern" {
				wp := WebPattern{
					Type:        d.Strategy,
					Description: d.Description,
					Examples:    d.Examples,
				}
				if d.Detected {
					wp.Severity = d.Score
					wp.Passed = false
					resp.WebPatterns = append(resp.WebPatterns, wp)
				} else {
					wp.Severity = 0
					wp.Passed = true
					resp.PassedPatterns = append(resp.PassedPatterns, wp)
				}
			}
		}

		resp.PatternCount = len(resp.WebPatterns)

		suspicionRate := report.SuspicionRate
		if slopRate, ok := report.Metrics["slop_suspicion_rate"].(float64); ok {
			suspicionRate = slopRate
		}
		resp.SuspicionRate = suspicionRate

		confidenceScore := int(suspicionRate * 100)
		if confidenceScore > 100 {
			confidenceScore = 100
		}
		resp.ConfidenceScore = confidenceScore
		resp.OverallSuspicion = float64(confidenceScore)
		resp.QualityScore = 1.0 - suspicionRate

		if suspicionRate >= 0.7 {
			resp.Assessment = "Likely AI-Generated"
		} else if suspicionRate >= 0.4 {
			resp.Assessment = "Suspicious Activity"
		} else {
			resp.Assessment = "Likely Human-Written"
		}

		if wc, ok := report.Metrics["word_count"].(int); ok {
			resp.WordCount = wc
		}
		if cc, ok := report.Metrics["character_count"].(int); ok {
			resp.CharacterCount = cc
		}
		if hc, ok := report.Metrics["heading_count"].(int); ok {
			resp.HeadingCount = hc
		}
		if headings, ok := report.Metrics["headings"].([]string); ok {
			resp.Headings = headings
		}
	}

	// Cross-source metrics
	resp.ItemsAnalyzed = report.SourceMetrics.ItemsAnalyzed
	resp.ItemsFlagged = report.SourceMetrics.ItemsFlagged
	resp.StrategiesUsed = report.SourceMetrics.StrategiesUsed
	resp.StrategiesHit = report.SourceMetrics.StrategiesHit
	resp.AverageScore = report.SourceMetrics.AverageScore
	resp.CoverageRate = report.SourceMetrics.CoverageRate

	return resp
}

// mapToJobResult creates a temporary JobResult from a JobResultResponse for metric calculation.
func mapToJobResult(resp *JobResultResponse) *JobResult {
	return &JobResult{
		TotalCommits:      resp.TotalCommits,
		SuspiciousCommits: resp.SuspiciousCommits,
		Velocity:          resp.Velocity,
		TimeSpan:          resp.TimeSpan,
		UniqueAuthors:     resp.UniqueAuthors,
		AverageCommitSize: resp.AverageCommitSize,
		OverallSuspicion:  resp.OverallSuspicion,
	}
}

// progressPercent computes a rough progress percentage.
func progressPercent(p *analysis.ProgressInfo) float64 {
	if p.Total > 0 && p.Current > 0 {
		return float64(p.Current) / float64(p.Total) * 100
	}
	switch p.Phase {
	case "validating":
		return 5
	case "fetching", "cloning":
		return 15
	case "detecting":
		return 50
	default:
		return 0
	}
}

// writeSSE writes a single Server-Sent Event to the writer and flushes.
// Returns false if the write or flush failed (broken connection).
func writeSSE(w *bufio.Writer, event string, data interface{}) bool {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return false
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonBytes)); err != nil {
		return false
	}
	return w.Flush() == nil
}
