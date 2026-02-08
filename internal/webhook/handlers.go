package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TryCadence/Cadence/internal/analysis"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git"
	"github.com/TryCadence/Cadence/internal/analysis/adapters/git/patterns"
	"github.com/TryCadence/Cadence/internal/analysis/detectors"
	"github.com/TryCadence/Cadence/internal/analysis/sources"
	"github.com/TryCadence/Cadence/internal/logging"
	gogit "github.com/go-git/go-git/v5"
	"github.com/gofiber/fiber/v2"
)

type WebhookHandlers struct {
	secret    string
	queue     *JobQueue
	processor *AnalysisProcessor
	cache     analysis.AnalysisCache
	metrics   analysis.AnalysisMetrics
	plugins   *analysis.PluginManager
}

type AnalysisProcessor struct {
	DetectorThresholds *patterns.Thresholds
	Logger             *logging.Logger
	Metrics            analysis.AnalysisMetrics
}

func (ap *AnalysisProcessor) log() *logging.Logger {
	if ap.Logger != nil {
		return ap.Logger
	}
	return logging.Default()
}

func (ap *AnalysisProcessor) metricsCollector() analysis.AnalysisMetrics {
	if ap.Metrics != nil {
		return ap.Metrics
	}
	return analysis.NullMetrics{}
}

func NewWebhookHandlers(secret string, queue *JobQueue, thresholds *patterns.Thresholds) *WebhookHandlers {
	return &WebhookHandlers{
		secret: secret,
		queue:  queue,
		processor: &AnalysisProcessor{
			DetectorThresholds: thresholds,
		},
		cache:   analysis.NullCache{},
		metrics: analysis.NullMetrics{},
		plugins: analysis.NewPluginManager(),
	}
}

// WithCache sets the analysis cache implementation.
func (wh *WebhookHandlers) WithCache(cache analysis.AnalysisCache) *WebhookHandlers {
	if cache != nil {
		wh.cache = cache
	}
	return wh
}

// WithMetrics sets the analysis metrics collector.
func (wh *WebhookHandlers) WithMetrics(metrics analysis.AnalysisMetrics) *WebhookHandlers {
	if metrics != nil {
		wh.metrics = metrics
		wh.processor.Metrics = metrics
	}
	return wh
}

// WithPlugins sets the plugin manager.
func (wh *WebhookHandlers) WithPlugins(plugins *analysis.PluginManager) *WebhookHandlers {
	if plugins != nil {
		wh.plugins = plugins
	}
	return wh
}

func (ap *AnalysisProcessor) Process(ctx context.Context, job *WebhookJob) error {
	ap.log().LogPhase(job.ID, "starting analysis", "event_type", job.EventType)

	job.Progress = "initializing"

	job.Result = &JobResult{
		JobID:      job.ID,
		RepoName:   job.RepoName,
		URL:        job.RepoURL,
		AnalyzedAt: time.Now(),
	}

	if job.EventType == "api_analysis_repo" {
		return ap.processGitAnalysis(ctx, job)
	} else if job.EventType == "api_analysis_website" {
		return ap.processWebAnalysis(ctx, job)
	}

	ap.log().LogPhase(job.ID, "analysis complete")
	job.Progress = "completed"
	return nil
}

func (ap *AnalysisProcessor) processGitAnalysis(ctx context.Context, job *WebhookJob) error {
	job.Progress = "cloning"
	ap.log().LogPhase(job.ID, "cloning repository", "repo_url", job.RepoURL)

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("cadence-analysis-%s", job.ID))
	defer os.RemoveAll(tmpDir)

	if err := cloneRepo(job.RepoURL, tmpDir); err != nil {
		ap.log().LogPhaseError(job.ID, "clone failed", err, "repo_url", job.RepoURL)
		ap.metricsCollector().RecordError("git", "clone")
		job.Progress = "clone-failed"
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	ap.log().LogPhase(job.ID, "clone completed, running analysis")
	job.Progress = "analyzing"

	source := sources.NewGitRepositorySource(tmpDir, job.Branch)
	det := detectors.NewGitDetector(ap.DetectorThresholds)
	runner := analysis.NewDefaultDetectionRunner()

	report, err := runner.Run(ctx, source, det)
	if err != nil {
		ap.log().LogPhaseError(job.ID, "analysis failed", err)
		ap.metricsCollector().RecordError("git", "analysis")
		job.Progress = "analysis-failed"
		return fmt.Errorf("analysis failed: %w", err)
	}

	job.Progress = "processing-results"
	ap.populateGitJobResult(job, report)

	ap.metricsCollector().RecordAnalysis("git", report.Duration)
	ap.metricsCollector().RecordDetections("git", report.TotalDetections, report.DetectionCount)

	ap.log().LogPhase(job.ID, "repository analysis complete",
		"total_commits", job.Result.TotalCommits,
		"suspicious_commits", job.Result.SuspiciousCommits,
	)
	job.Progress = "completed"
	return nil
}

func (ap *AnalysisProcessor) processWebAnalysis(ctx context.Context, job *WebhookJob) error {
	ap.log().LogPhase(job.ID, "starting website analysis", "url", job.RepoURL)
	job.Progress = "fetching-content"

	source := sources.NewWebsiteSource(job.RepoURL)
	det := detectors.NewWebDetector()
	runner := analysis.NewDefaultDetectionRunner()

	report, err := runner.Run(ctx, source, det)
	if err != nil {
		ap.log().LogPhaseError(job.ID, "website analysis failed", err, "url", job.RepoURL)
		ap.metricsCollector().RecordError("web", "analysis")
		job.Progress = "analysis-failed"
		return fmt.Errorf("analysis failed: %w", err)
	}

	job.Progress = "processing-results"

	if analysisErr, ok := report.Metrics["analysis_error"].(string); ok {
		ap.log().LogPhase(job.ID, "content analysis note", "note", analysisErr)
		job.Result.ConfidenceScore = 0
		job.Result.SuspicionRate = 0
		job.Result.Assessment = "Content too short for reliable analysis"
		job.Result.PatternCount = 0
	} else {
		ap.populateWebJobResult(job, report)
	}

	if wc, ok := report.Metrics["word_count"].(int); ok {
		job.Result.WordCount = wc
	}
	if cc, ok := report.Metrics["character_count"].(int); ok {
		job.Result.CharacterCount = cc
	}
	if hc, ok := report.Metrics["heading_count"].(int); ok {
		job.Result.HeadingCount = hc
	}
	if headings, ok := report.Metrics["headings"].([]string); ok {
		job.Result.Headings = headings
	}

	ap.metricsCollector().RecordAnalysis("web", report.Duration)
	ap.metricsCollector().RecordDetections("web", report.TotalDetections, report.DetectionCount)

	ap.log().LogPhase(job.ID, "website analysis complete",
		"pattern_count", job.Result.PatternCount,
		"confidence_score", job.Result.ConfidenceScore,
	)
	job.Progress = "completed"
	return nil
}

func (ap *AnalysisProcessor) populateGitJobResult(job *WebhookJob, report *analysis.AnalysisReport) {
	populateTimingAndMetrics(job.Result, report)

	if commitCount, ok := report.Metrics["commit_count"].(int); ok {
		job.Result.TotalCommits = commitCount
	}

	ap.log().LogPhase(job.ID, "populating git results", "total_commits", job.Result.TotalCommits)

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

			suspicion := Suspicion{
				CommitHash: hash,
				Message:    d.Description,
				Severity:   d.Severity,
				Reasons:    reasons,
				Score:      d.Score * 100,
			}
			job.Result.Suspicions = append(job.Result.Suspicions, suspicion)
		}
	}

	job.Result.SuspiciousCommits = len(job.Result.Suspicions)
	ap.log().LogPhase(job.ID, "suspicious commits found", "count", job.Result.SuspiciousCommits)

	if pairs, ok := report.Metrics["commit_pairs"].([]*git.CommitPair); ok {
		calculateMetrics(job.Result, pairs)
	}
}

func (ap *AnalysisProcessor) populateWebJobResult(job *WebhookJob, report *analysis.AnalysisReport) {
	populateTimingAndMetrics(job.Result, report)

	job.Result.URL = report.SourceID

	job.Result.WebPatterns = make([]WebPattern, 0)
	job.Result.PassedPatterns = make([]WebPattern, 0)

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
				job.Result.WebPatterns = append(job.Result.WebPatterns, wp)
			} else {
				wp.Severity = 0
				wp.Passed = true
				job.Result.PassedPatterns = append(job.Result.PassedPatterns, wp)
			}
		}
	}

	job.Result.PatternCount = len(job.Result.WebPatterns)

	suspicionRate := report.SuspicionRate
	if slopRate, ok := report.Metrics["slop_suspicion_rate"].(float64); ok {
		suspicionRate = slopRate
	}

	job.Result.SuspicionRate = suspicionRate

	confidenceScore := int(suspicionRate * 100)
	if confidenceScore > 100 {
		confidenceScore = 100
	}
	job.Result.ConfidenceScore = confidenceScore
	job.Result.OverallSuspicion = float64(confidenceScore)
	job.Result.QualityScore = 1.0 - suspicionRate

	if suspicionRate >= 0.7 {
		job.Result.Assessment = "Likely AI-Generated"
	} else if suspicionRate >= 0.4 {
		job.Result.Assessment = "Suspicious Activity"
	} else {
		job.Result.Assessment = "Likely Human-Written"
	}
}

// populateTimingAndMetrics copies timing and cross-source metrics from the analysis report into the job result.
func populateTimingAndMetrics(result *JobResult, report *analysis.AnalysisReport) {
	result.StartedAt = report.Timing.StartedAt
	result.CompletedAt = report.Timing.CompletedAt
	result.DurationMs = report.Timing.Duration.Milliseconds()
	result.AnalyzedAt = report.Timing.StartedAt

	result.ItemsAnalyzed = report.SourceMetrics.ItemsAnalyzed
	result.ItemsFlagged = report.SourceMetrics.ItemsFlagged
	result.StrategiesUsed = report.SourceMetrics.StrategiesUsed
	result.StrategiesHit = report.SourceMetrics.StrategiesHit
	result.AverageScore = report.SourceMetrics.AverageScore
	result.CoverageRate = report.SourceMetrics.CoverageRate
	if report.SourceMetrics.UniqueAuthors > 0 {
		result.UniqueAuthors = report.SourceMetrics.UniqueAuthors
	}
}

// calculateMetrics computes repository metrics from commit pairs
func calculateMetrics(result *JobResult, commitPairs []*git.CommitPair) {
	if len(commitPairs) == 0 {
		return
	}

	// Calculate time span
	firstCommitTime := commitPairs[0].Current.Timestamp
	lastCommitTime := commitPairs[len(commitPairs)-1].Current.Timestamp
	timeSpan := lastCommitTime.Sub(firstCommitTime)

	// Format time span
	if timeSpan.Hours() < 24 {
		hours := int(timeSpan.Hours())
		if hours <= 1 {
			result.TimeSpan = "< 1 hour"
		} else {
			result.TimeSpan = fmt.Sprintf("%d hours", hours)
		}
	} else if timeSpan.Hours() < 24*30 {
		days := int(timeSpan.Hours() / 24)
		result.TimeSpan = fmt.Sprintf("%d days", days)
	} else {
		weeks := int(timeSpan.Hours() / (24 * 7))
		result.TimeSpan = fmt.Sprintf("%d weeks", weeks)
	}

	// Calculate total additions/deletions and unique authors
	totalAdditions := int64(0)
	totalDeletions := int64(0)
	authors := make(map[string]bool)

	for _, pair := range commitPairs {
		if pair.Stats != nil {
			totalAdditions += pair.Stats.Additions
			totalDeletions += pair.Stats.Deletions
		}
		authors[pair.Current.Author] = true
	}

	result.UniqueAuthors = len(authors)

	// Calculate velocity (additions per minute)
	if timeSpan.Minutes() > 0 {
		velocity := float64(totalAdditions) / timeSpan.Minutes()
		if velocity >= 1 {
			result.Velocity = fmt.Sprintf("%.0f additions/min", velocity)
		} else {
			result.Velocity = fmt.Sprintf("%.1f additions/min", velocity)
		}
	}

	// Calculate average commit size
	if len(commitPairs) > 0 {
		result.AverageCommitSize = int((totalAdditions + totalDeletions) / int64(len(commitPairs)))
	}

	// Calculate overall suspicion score (percentage of suspicious commits)
	if result.TotalCommits > 0 {
		result.OverallSuspicion = (float64(result.SuspiciousCommits) / float64(result.TotalCommits)) * 100
	}
}

// cloneRepo is a helper function to clone a git repository
func cloneRepo(url, dest string) error {
	// Create a context with 2-minute timeout for clone operations
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Use go-git to clone
	_, err := gogit.PlainCloneContext(ctx, dest, false, &gogit.CloneOptions{
		URL:      url,
		Progress: nil,
	})
	return err
}

func (wh *WebhookHandlers) RegisterRoutes(app *fiber.App) {
	// Webhook endpoints
	app.Post("/webhooks/github", wh.HandleGithubWebhook)
	app.Post("/webhooks/gitlab", wh.HandleGitlabWebhook)

	// Public API endpoints for playground analysis
	app.Post("/api/analyze/repository", wh.AnalyzeRepository)
	app.Post("/api/analyze/website", wh.AnalyzeWebsite)

	// SSE streaming endpoints
	app.Post("/api/stream/repository", wh.StreamAnalyzeRepository)
	app.Post("/api/stream/website", wh.StreamAnalyzeWebsite)

	// Job status endpoints
	app.Get("/jobs/:id", wh.GetJobStatus)
	app.Get("/jobs", wh.ListJobs)
	app.Get("/api/results/:id", wh.GetJobResult)

	// Observability endpoints
	app.Get("/metrics", wh.MetricsEndpoint)
	app.Get("/api/metrics", wh.MetricsJSON)
	app.Get("/api/cache/stats", wh.CacheStats)
	app.Post("/api/cache/clear", wh.CacheClear)

	// Plugin endpoints
	app.Get("/api/plugins", wh.ListPlugins)

	app.Get("/health", wh.HealthCheck)
}

func (wh *WebhookHandlers) HandleGithubWebhook(c *fiber.Ctx) error {
	signature := c.Get("X-Hub-Signature-256")
	if signature == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing signature",
		})
	}

	body := c.Body()
	if err := wh.verifySignature(body, signature); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid signature",
		})
	}

	var payload GithubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	// Extract branch from ref (e.g., "refs/heads/main" -> "main")
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	job := &WebhookJob{
		EventType: "github_push",
		RepoURL:   payload.Repository.URL,
		RepoName:  payload.Repository.Name,
		Branch:    branch,
		Author:    payload.Pusher.Name,
		Commits:   make([]WebhookCommit, 0),
	}

	for i := range payload.Commits {
		commit := &payload.Commits[i]
		timestamp, _ := time.Parse(time.RFC3339, commit.Timestamp)
		job.Commits = append(job.Commits, WebhookCommit{
			Hash:      commit.ID,
			Message:   commit.Message,
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Timestamp: timestamp,
			Added:     commit.Added,
			Modified:  commit.Modified,
			Removed:   commit.Removed,
		})
	}

	if err := wh.queue.Enqueue(job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusAccepted).JSON(fiber.Map{
		"job_id": job.ID,
		"status": StatusPending,
	})
}

func (wh *WebhookHandlers) HandleGitlabWebhook(c *fiber.Ctx) error {
	token := c.Get("X-Gitlab-Token")
	if token == "" || token != wh.secret {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}

	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	return c.Status(http.StatusAccepted).JSON(fiber.Map{
		"status": StatusPending,
	})
}

func (wh *WebhookHandlers) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")

	job, err := wh.queue.GetJob(jobID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "job not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":        job.ID,
		"status":    job.Status,
		"repo":      job.RepoName,
		"branch":    job.Branch,
		"timestamp": job.Timestamp,
		"error":     job.Error,
		"result":    job.Result,
	})
}

func (wh *WebhookHandlers) ListJobs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	jobs := wh.queue.ListJobs(limit)

	jobList := make([]fiber.Map, 0)
	for _, job := range jobs {
		jobList = append(jobList, fiber.Map{
			"id":        job.ID,
			"status":    job.Status,
			"repo":      job.RepoName,
			"branch":    job.Branch,
			"timestamp": job.Timestamp,
			"author":    job.Author,
		})
	}

	return c.JSON(fiber.Map{
		"total": len(jobList),
		"jobs":  jobList,
	})
}

func (wh *WebhookHandlers) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now(),
	})
}

func (wh *WebhookHandlers) verifySignature(body []byte, signature string) error {
	parts := strings.Split(signature, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid signature format")
	}

	expectedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("invalid signature encoding")
	}

	h := hmac.New(sha256.New, []byte(wh.secret))
	if _, err := h.Write(body); err != nil {
		return fmt.Errorf("failed to compute signature: %w", err)
	}

	actualHash := h.Sum(nil)

	if !hmac.Equal(actualHash, expectedHash) {
		return fmt.Errorf("signature mismatch")
	}

	return nil
}

func NewDefaultProcessor() JobProcessor {
	return &AnalysisProcessor{
		DetectorThresholds: &patterns.Thresholds{
			SuspiciousAdditions: 500,
			MaxAdditionsPerMin:  100,
		},
	}
}

type AnalyzeRepositoryRequest struct {
	RepositoryURL string `json:"repository_url"`
	Branch        string `json:"branch,omitempty"`
}

type AnalyzeWebsiteRequest struct {
	URL string `json:"url"`
}

type AnalysisResponse struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

type JobResultResponse struct {
	JobID    string `json:"job_id"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Progress string `json:"progress,omitempty"`
	// Repository fields
	RepoName          string      `json:"repo_name,omitempty"`
	TotalCommits      int         `json:"total_commits,omitempty"`
	SuspiciousCommits int         `json:"suspicious_commits,omitempty"`
	Suspicions        []Suspicion `json:"suspicions,omitempty"`
	Velocity          string      `json:"velocity,omitempty"`
	TimeSpan          string      `json:"time_span,omitempty"`
	UniqueAuthors     int         `json:"unique_authors,omitempty"`
	AverageCommitSize int         `json:"average_commit_size,omitempty"`
	OverallSuspicion  float64     `json:"overall_suspicion,omitempty"`
	// Website fields
	URL             string       `json:"url,omitempty"`
	WordCount       int          `json:"word_count,omitempty"`
	CharacterCount  int          `json:"character_count,omitempty"`
	HeadingCount    int          `json:"heading_count,omitempty"`
	Headings        []string     `json:"headings,omitempty"`
	QualityScore    float64      `json:"quality_score,omitempty"`
	ConfidenceScore int          `json:"confidence_score,omitempty"`
	SuspicionRate   float64      `json:"suspicion_rate,omitempty"`
	PatternCount    int          `json:"pattern_count,omitempty"`
	Assessment      string       `json:"assessment,omitempty"`
	WebPatterns     []WebPattern `json:"web_patterns,omitempty"`
	PassedPatterns  []WebPattern `json:"passed_patterns,omitempty"`
	// Cross-source metrics
	ItemsAnalyzed  int     `json:"items_analyzed,omitempty"`
	ItemsFlagged   int     `json:"items_flagged,omitempty"`
	StrategiesUsed int     `json:"strategies_used,omitempty"`
	StrategiesHit  int     `json:"strategies_hit,omitempty"`
	AverageScore   float64 `json:"average_score,omitempty"`
	CoverageRate   float64 `json:"coverage_rate,omitempty"`
	// Common
	AnalyzedAt time.Time `json:"analyzed_at,omitempty"`
}

func (wh *WebhookHandlers) AnalyzeRepository(c *fiber.Ctx) error {
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

	job := &WebhookJob{
		EventType: "api_analysis_repo",
		RepoURL:   req.RepositoryURL,
		Branch:    req.Branch,
		Timestamp: time.Now(),
		Commits:   make([]WebhookCommit, 0),
	}

	if err := wh.queue.Enqueue(job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to queue analysis job",
		})
	}

	return c.Status(http.StatusAccepted).JSON(AnalysisResponse{
		JobID:  job.ID,
		Status: StatusPending,
	})
}

func (wh *WebhookHandlers) AnalyzeWebsite(c *fiber.Ctx) error {
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

	job := &WebhookJob{
		EventType: "api_analysis_website",
		RepoURL:   req.URL,
		Timestamp: time.Now(),
		Commits:   make([]WebhookCommit, 0),
	}

	if err := wh.queue.Enqueue(job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to queue analysis job",
		})
	}

	return c.Status(http.StatusAccepted).JSON(AnalysisResponse{
		JobID:  job.ID,
		Status: StatusPending,
	})
}

func (wh *WebhookHandlers) GetJobResult(c *fiber.Ctx) error {
	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "job id is required",
		})
	}

	job, err := wh.queue.GetJob(jobID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "job not found",
		})
	}

	response := JobResultResponse{
		JobID:    job.ID,
		Status:   job.Status,
		Error:    job.Error,
		Progress: job.Progress,
	}

	if job.Result != nil {
		// Repository fields
		response.RepoName = job.Result.RepoName
		response.TotalCommits = job.Result.TotalCommits
		response.SuspiciousCommits = job.Result.SuspiciousCommits
		response.Suspicions = job.Result.Suspicions
		response.Velocity = job.Result.Velocity
		response.TimeSpan = job.Result.TimeSpan
		response.UniqueAuthors = job.Result.UniqueAuthors
		response.AverageCommitSize = job.Result.AverageCommitSize
		response.OverallSuspicion = job.Result.OverallSuspicion
		// Website fields
		response.URL = job.Result.URL
		response.WordCount = job.Result.WordCount
		response.CharacterCount = job.Result.CharacterCount
		response.HeadingCount = job.Result.HeadingCount
		response.Headings = job.Result.Headings
		response.QualityScore = job.Result.QualityScore
		response.ConfidenceScore = job.Result.ConfidenceScore
		response.SuspicionRate = job.Result.SuspicionRate
		response.PatternCount = job.Result.PatternCount
		response.Assessment = job.Result.Assessment
		response.WebPatterns = job.Result.WebPatterns
		response.PassedPatterns = job.Result.PassedPatterns
		// Common
		response.AnalyzedAt = job.Result.AnalyzedAt
	}

	return c.JSON(response)
}

// MetricsEndpoint serves Prometheus text format metrics at GET /metrics.
func (wh *WebhookHandlers) MetricsEndpoint(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	return c.SendString(wh.metrics.PrometheusFormat())
}

// MetricsJSON serves metrics as JSON at GET /api/metrics.
func (wh *WebhookHandlers) MetricsJSON(c *fiber.Ctx) error {
	return c.JSON(wh.metrics.Snapshot())
}

// CacheStats returns cache statistics at GET /api/cache/stats.
func (wh *WebhookHandlers) CacheStats(c *fiber.Ctx) error {
	return c.JSON(wh.cache.Stats())
}

// CacheClear empties the analysis cache at POST /api/cache/clear.
func (wh *WebhookHandlers) CacheClear(c *fiber.Ctx) error {
	wh.cache.Clear()
	return c.JSON(fiber.Map{"status": "cleared"})
}

// ListPlugins returns registered plugin metadata at GET /api/plugins.
func (wh *WebhookHandlers) ListPlugins(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"plugins": wh.plugins.List(),
		"count":   wh.plugins.Count(),
	})
}
