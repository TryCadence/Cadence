package webhook

import (
	"testing"
	"time"
)

func TestWebhookJob(t *testing.T) {
	t.Run("create webhook job", func(t *testing.T) {
		job := &WebhookJob{
			ID:        "test-123",
			EventType: "github_push",
			RepoURL:   "https://github.com/test/repo.git",
			RepoName:  "repo",
			Branch:    "main",
			Author:    "John Doe",
			Timestamp: time.Now(),
			Status:    StatusPending,
		}

		if job.ID != "test-123" {
			t.Errorf("ID = %s, want test-123", job.ID)
		}
		if job.EventType != "github_push" {
			t.Errorf("EventType = %s, want github_push", job.EventType)
		}
		if job.Status != StatusPending {
			t.Errorf("Status = %s, want %s", job.Status, StatusPending)
		}
	})

	t.Run("github push payload parsing", func(t *testing.T) {
		payload := &GithubPushPayload{
			Ref: "refs/heads/main",
		}
		payload.Repository.Name = "test-repo"
		payload.Repository.URL = "https://github.com/test/test-repo.git"
		payload.Pusher.Name = "Jane Doe"

		if payload.Ref != "refs/heads/main" {
			t.Errorf("Ref = %s, want refs/heads/main", payload.Ref)
		}
		if payload.Repository.Name != "test-repo" {
			t.Errorf("Repository.Name = %s, want test-repo", payload.Repository.Name)
		}
	})
}

func TestWebhookCommit(t *testing.T) {
	t.Run("webhook commit structure", func(t *testing.T) {
		now := time.Now()
		commit := WebhookCommit{
			Hash:      "abc123",
			Message:   "Test commit",
			Author:    "John Doe",
			Email:     "john@example.com",
			Timestamp: now,
			Added:     []string{"file1.txt"},
			Modified:  []string{"file2.txt"},
			Removed:   []string{"file3.txt"},
		}

		if commit.Hash != "abc123" {
			t.Errorf("Hash = %s, want abc123", commit.Hash)
		}
		if len(commit.Added) != 1 {
			t.Errorf("len(Added) = %d, want 1", len(commit.Added))
		}
		if len(commit.Modified) != 1 {
			t.Errorf("len(Modified) = %d, want 1", len(commit.Modified))
		}
		if len(commit.Removed) != 1 {
			t.Errorf("len(Removed) = %d, want 1", len(commit.Removed))
		}
	})
}

func TestJobResult(t *testing.T) {
	t.Run("create job result", func(t *testing.T) {
		result := &JobResult{
			JobID:             "job-123",
			RepoName:          "test-repo",
			TotalCommits:      5,
			SuspiciousCommits: 2,
			Suspicions:        make([]Suspicion, 0),
			AnalyzedAt:        time.Now(),
		}

		if result.JobID != "job-123" {
			t.Errorf("JobID = %s, want job-123", result.JobID)
		}
		if result.TotalCommits != 5 {
			t.Errorf("TotalCommits = %d, want 5", result.TotalCommits)
		}
		if result.SuspiciousCommits != 2 {
			t.Errorf("SuspiciousCommits = %d, want 2", result.SuspiciousCommits)
		}
	})

	t.Run("add suspicions to result", func(t *testing.T) {
		result := &JobResult{
			JobID:      "job-123",
			Suspicions: make([]Suspicion, 0),
		}

		result.Suspicions = append(result.Suspicions, Suspicion{
			CommitHash:  "abc123",
			CommitIndex: 0,
			Message:     "High velocity detected",
			Severity:    "high",
			Score:       0.85,
		})

		if len(result.Suspicions) != 1 {
			t.Errorf("len(Suspicions) = %d, want 1", len(result.Suspicions))
		}
		if result.Suspicions[0].Score != 0.85 {
			t.Errorf("Score = %f, want 0.85", result.Suspicions[0].Score)
		}
	})
}
