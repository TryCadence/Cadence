package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/codemeapixel/cadence/internal/detector"
	"github.com/gofiber/fiber/v2"
)

// WebhookHandlers handles incoming webhook requests
type WebhookHandlers struct {
	secret    string
	queue     *JobQueue
	processor *AnalysisProcessor
}

// AnalysisProcessor implements JobProcessor for running Cadence analysis
type AnalysisProcessor struct {
	DetectorThresholds *detector.Thresholds
}

// NewWebhookHandlers creates a new webhook handler
func NewWebhookHandlers(secret string, queue *JobQueue, thresholds *detector.Thresholds) *WebhookHandlers {
	return &WebhookHandlers{
		secret: secret,
		queue:  queue,
		processor: &AnalysisProcessor{
			DetectorThresholds: thresholds,
		},
	}
}

// Process implements the JobProcessor interface for analysis
func (ap *AnalysisProcessor) Process(ctx context.Context, job *WebhookJob) error {
	// In a real implementation, this would:
	// 1. Clone the repository
	// 2. Run analysis on the commits
	// 3. Store results
	// 4. Post results back to GitHub/GitLab if configured

	// For now, we'll create a mock analysis result
	job.Result = &JobResult{
		JobID:             job.ID,
		RepoName:          job.RepoName,
		TotalCommits:      len(job.Commits),
		SuspiciousCommits: 0,
		Suspicions:        make([]Suspicion, 0),
		AnalyzedAt:        time.Now(),
	}

	// TODO: Implement actual analysis against the repository
	// This would require:
	// - Cloning the repo temporarily
	// - Running git.OpenRepository()
	// - Creating an analyzer instance
	// - Running detection
	// - Storing/reporting results

	return nil
}

// RegisterRoutes registers webhook routes on the Fiber app
func (wh *WebhookHandlers) RegisterRoutes(app *fiber.App) {
	app.Post("/webhooks/github", wh.HandleGithubWebhook)
	app.Post("/webhooks/gitlab", wh.HandleGitlabWebhook)
	app.Get("/jobs/:id", wh.GetJobStatus)
	app.Get("/jobs", wh.ListJobs)
	app.Get("/health", wh.HealthCheck)
}

// HandleGithubWebhook handles incoming GitHub push webhooks
func (wh *WebhookHandlers) HandleGithubWebhook(c *fiber.Ctx) error {
	// Verify webhook signature
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

	// Parse payload
	var payload GithubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	// Extract branch name from ref (e.g., "refs/heads/main" -> "main")
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// Convert to WebhookJob
	job := &WebhookJob{
		EventType: "github_push",
		RepoURL:   payload.Repository.URL,
		RepoName:  payload.Repository.Name,
		Branch:    branch,
		Author:    payload.Pusher.Name,
		Commits:   make([]WebhookCommit, 0),
	}

	// Convert commits
	for _, commit := range payload.Commits {
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

	// Enqueue the job
	if err := wh.queue.Enqueue(job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusAccepted).JSON(fiber.Map{
		"job_id": job.ID,
		"status": "pending",
	})
}

// HandleGitlabWebhook handles incoming GitLab push webhooks
func (wh *WebhookHandlers) HandleGitlabWebhook(c *fiber.Ctx) error {
	// Verify webhook token
	token := c.Get("X-Gitlab-Token")
	if token == "" || token != wh.secret {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid or missing token",
		})
	}

	// Parse GitLab payload (simplified version)
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid payload",
		})
	}

	// For now, return a placeholder response
	// Full implementation would parse GitLab-specific webhook format
	return c.Status(http.StatusAccepted).JSON(fiber.Map{
		"status": "pending",
	})
}

// GetJobStatus retrieves the status of a specific job
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

// ListJobs lists recent webhook jobs
func (wh *WebhookHandlers) ListJobs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	jobs := wh.queue.ListJobs(limit)

	jobsData := make([]fiber.Map, 0)
	for _, job := range jobs {
		jobsData = append(jobsData, fiber.Map{
			"id":        job.ID,
			"status":    job.Status,
			"repo":      job.RepoName,
			"branch":    job.Branch,
			"timestamp": job.Timestamp,
			"author":    job.Author,
		})
	}

	return c.JSON(fiber.Map{
		"total": len(jobsData),
		"jobs":  jobsData,
	})
}

// HealthCheck returns server health status
func (wh *WebhookHandlers) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "ok",
		"time":   time.Now(),
	})
}

// verifySignature verifies the GitHub webhook signature
func (wh *WebhookHandlers) verifySignature(body []byte, signature string) error {
	// GitHub sends signature as "sha256=<hash>"
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

// DefaultAnalysisProcessor is a simple processor that just marks jobs as completed
func NewDefaultProcessor() JobProcessor {
	return &AnalysisProcessor{
		DetectorThresholds: &detector.Thresholds{
			SuspiciousAdditions: 500,
			MaxAdditionsPerMin:  100,
		},
	}
}
