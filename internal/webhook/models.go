package webhook

import "time"

const (
	// Job status constants
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

// WebhookJob represents an analysis job triggered by a webhook event
type WebhookJob struct {
	ID        string
	EventType string // "push", "pull_request", etc.
	RepoURL   string
	RepoName  string
	Branch    string
	Commits   []WebhookCommit
	Author    string
	Timestamp time.Time
	Status    string // StatusPending, StatusProcessing, StatusCompleted, StatusFailed
	Error     string
	Result    *JobResult
}

// WebhookCommit represents a commit from webhook payload
type WebhookCommit struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Timestamp time.Time
	Added     []string
	Modified  []string
	Removed   []string
}

// JobResult holds analysis results for a webhook job
type JobResult struct {
	JobID             string
	RepoName          string
	TotalCommits      int
	SuspiciousCommits int
	Suspicions        []Suspicion
	AnalyzedAt        time.Time
}

// Suspicion represents a suspicious commit detection
type Suspicion struct {
	CommitHash  string
	CommitIndex int
	Message     string
	Severity    string // "low", "medium", "high"
	Reasons     []string
	Score       float64
}

type GithubPushPayload struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		URL      string `json:"clone_url"`
	} `json:"repository"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Commits []struct {
		ID        string `json:"id"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string `json:"added"`
		Modified []string `json:"modified"`
		Removed  []string `json:"removed"`
	} `json:"commits"`
}
