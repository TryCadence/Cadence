package webhook

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TryCadence/Cadence/internal/logging"
	"github.com/google/uuid"
)

type JobQueue struct {
	jobs       chan *WebhookJob
	maxWorkers int
	workers    int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	processor  JobProcessor
	mu         sync.RWMutex
	jobStore   map[string]*WebhookJob
	logger     *logging.Logger
}

type JobProcessor interface {
	Process(ctx context.Context, job *WebhookJob) error
}

func NewJobQueue(maxWorkers int, processor JobProcessor) *JobQueue {
	ctx, cancel := context.WithCancel(context.Background())
	return &JobQueue{
		jobs:       make(chan *WebhookJob, 100),
		maxWorkers: maxWorkers,
		workers:    0,
		ctx:        ctx,
		cancel:     cancel,
		processor:  processor,
		jobStore:   make(map[string]*WebhookJob),
		logger:     logging.Default().With("component", "job_queue"),
	}
}

func (q *JobQueue) Start() error {
	for i := 0; i < q.maxWorkers; i++ {
		q.wg.Add(1)
		go q.worker()
		q.workers++
	}
	return nil
}

func (q *JobQueue) Stop() error {
	q.cancel()
	close(q.jobs)
	q.wg.Wait()
	return nil
}

func (q *JobQueue) Enqueue(job *WebhookJob) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	job.Status = StatusPending
	job.Timestamp = time.Now()

	q.mu.Lock()
	q.jobStore[job.ID] = job
	q.mu.Unlock()

	select {
	case q.jobs <- job:
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("job queue is shutting down")
	}
}

func (q *JobQueue) GetJob(jobID string) (*WebhookJob, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	job, exists := q.jobStore[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	return job, nil
}

func (q *JobQueue) ListJobs(limit int) []*WebhookJob {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]*WebhookJob, 0, len(q.jobStore))
	for _, job := range q.jobStore {
		jobs = append(jobs, job)
	}

	for i := 0; i < len(jobs)-1; i++ {
		for j := i + 1; j < len(jobs); j++ {
			if jobs[j].Timestamp.After(jobs[i].Timestamp) {
				jobs[i], jobs[j] = jobs[j], jobs[i]
			}
		}
	}

	if limit > 0 && len(jobs) > limit {
		jobs = jobs[:limit]
	}

	return jobs
}

func (q *JobQueue) worker() {
	defer q.wg.Done()

	for {
		select {
		case job := <-q.jobs:
			if job == nil {
				return
			}

			q.mu.Lock()
			job.Status = StatusProcessing
			q.mu.Unlock()

			q.logger.Info("processing job", "job_id", job.ID, "event_type", job.EventType)

			ctx, cancel := context.WithTimeout(q.ctx, 5*time.Minute)
			err := q.processor.Process(ctx, job)
			cancel()

			q.mu.Lock()
			if err != nil {
				q.logger.Error("job failed", "job_id", job.ID, "error", err)
				job.Status = StatusFailed
				job.Error = err.Error()
			} else {
				q.logger.Info("job completed", "job_id", job.ID)
				job.Status = StatusCompleted
			}
			q.mu.Unlock()

		case <-q.ctx.Done():
			return
		}
	}
}
