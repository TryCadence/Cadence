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

	"github.com/TryCadence/Cadence/internal/detector"
	"github.com/gofiber/fiber/v2"
)

type WebhookHandlers struct {
	secret    string
	queue     *JobQueue
	processor *AnalysisProcessor
}

type AnalysisProcessor struct {
	DetectorThresholds *detector.Thresholds
}

func NewWebhookHandlers(secret string, queue *JobQueue, thresholds *detector.Thresholds) *WebhookHandlers {
	return &WebhookHandlers{
		secret: secret,
		queue:  queue,
		processor: &AnalysisProcessor{
			DetectorThresholds: thresholds,
		},
	}
}

func (ap *AnalysisProcessor) Process(ctx context.Context, job *WebhookJob) error {
	job.Result = &JobResult{
		JobID:             job.ID,
		RepoName:          job.RepoName,
		TotalCommits:      len(job.Commits),
		SuspiciousCommits: 0,
		Suspicions:        make([]Suspicion, 0),
		AnalyzedAt:        time.Now(),
	}
	return nil
}

func (wh *WebhookHandlers) RegisterRoutes(app *fiber.App) {
	app.Post("/webhooks/github", wh.HandleGithubWebhook)
	app.Post("/webhooks/gitlab", wh.HandleGitlabWebhook)
	app.Get("/jobs/:id", wh.GetJobStatus)
	app.Get("/jobs", wh.ListJobs)
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

// GitHub sends signature as "sha256=<hash>"
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
		DetectorThresholds: &detector.Thresholds{
			SuspiciousAdditions: 500,
			MaxAdditionsPerMin:  100,
		},
	}
}
