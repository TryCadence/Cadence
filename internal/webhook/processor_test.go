package webhook

import (
	"context"
	"testing"
	"time"
)

func TestJobQueue_Enqueue(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)

	t.Run("enqueue job", func(t *testing.T) {
		job := &WebhookJob{
			EventType: "github_push",
			RepoName:  "test-repo",
			Branch:    "main",
		}

		err := queue.Enqueue(job)
		if err != nil {
			t.Fatalf("Enqueue() unexpected error = %v", err)
		}

		if job.ID == "" {
			t.Error("Job ID should be generated")
		}
		if job.Status != StatusPending {
			t.Errorf("Job status = %s, want %s", job.Status, StatusPending)
		}
	})
}

func TestJobQueue_GetJob(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)

	job := &WebhookJob{
		EventType: "github_push",
		RepoName:  "test-repo",
	}

	err := queue.Enqueue(job)
	if err != nil {
		t.Fatalf("Enqueue() failed: %v", err)
	}

	jobID := job.ID

	t.Run("get existing job", func(t *testing.T) {
		retrieved, err := queue.GetJob(jobID)
		if err != nil {
			t.Fatalf("GetJob() unexpected error = %v", err)
		}
		if retrieved.ID != jobID {
			t.Errorf("Retrieved job ID = %s, want %s", retrieved.ID, jobID)
		}
	})

	t.Run("get non-existent job", func(t *testing.T) {
		_, err := queue.GetJob("non-existent")
		if err == nil {
			t.Error("GetJob() expected error for non-existent job")
		}
	})
}

func TestJobQueue_ListJobs(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)

	t.Run("list empty jobs", func(t *testing.T) {
		jobs := queue.ListJobs(10)
		if len(jobs) != 0 {
			t.Errorf("len(jobs) = %d, want 0", len(jobs))
		}
	})

	t.Run("list multiple jobs", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			job := &WebhookJob{
				EventType: "github_push",
				RepoName:  "test-repo",
			}
			queue.Enqueue(job)
		}

		jobs := queue.ListJobs(10)
		if len(jobs) != 3 {
			t.Errorf("len(jobs) = %d, want 3", len(jobs))
		}
	})

	t.Run("list jobs with limit", func(t *testing.T) {
		processor2 := NewDefaultProcessor()
		queue2 := NewJobQueue(2, processor2)

		for i := 0; i < 5; i++ {
			job := &WebhookJob{
				EventType: "github_push",
				RepoName:  "test-repo",
			}
			queue2.Enqueue(job)
			time.Sleep(10 * time.Millisecond)
		}

		jobs := queue2.ListJobs(3)
		if len(jobs) > 3 {
			t.Errorf("len(jobs) = %d, want <= 3", len(jobs))
		}
	})
}

func TestJobQueue_StartStop(t *testing.T) {
	processor := NewDefaultProcessor()
	queue := NewJobQueue(2, processor)

	t.Run("start and stop queue", func(t *testing.T) {
		err := queue.Start()
		if err != nil {
			t.Fatalf("Start() unexpected error = %v", err)
		}

		err = queue.Stop()
		if err != nil {
			t.Fatalf("Stop() unexpected error = %v", err)
		}
	})

	t.Run("enqueue after stop returns error", func(t *testing.T) {
		processor2 := NewDefaultProcessor()
		queue2 := NewJobQueue(2, processor2)
		queue2.Start()
		queue2.Stop()

		job := &WebhookJob{
			EventType: "github_push",
			RepoName:  "test-repo",
		}

		_ = queue2.Enqueue(job)
	})
}

func TestAnalysisProcessor_Process(t *testing.T) {
	processor := &AnalysisProcessor{}

	t.Run("process job", func(t *testing.T) {
		job := &WebhookJob{
			ID:       "job-123",
			RepoName: "test-repo",
			Commits:  make([]WebhookCommit, 0),
		}

		job.Commits = append(job.Commits, WebhookCommit{
			Hash:    "abc123",
			Message: "Test commit",
		})

		ctx := context.Background()
		err := processor.Process(ctx, job)
		if err != nil {
			t.Fatalf("Process() unexpected error = %v", err)
		}

		if job.Result == nil {
			t.Error("Job result should not be nil")
		}
		if job.Result.JobID != "job-123" {
			t.Errorf("Result.JobID = %s, want job-123", job.Result.JobID)
		}
		if job.Result.TotalCommits != 1 {
			t.Errorf("Result.TotalCommits = %d, want 1", job.Result.TotalCommits)
		}
	})
}
