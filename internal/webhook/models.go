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
	Progress  string // Current step being processed (e.g., "cloning", "analyzing", "detecting")
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

type JobResult struct {
	JobID             string      `json:"job_id"`
	RepoName          string      `json:"repo_name,omitempty"`
	URL               string      `json:"url,omitempty"`
	TotalCommits      int         `json:"total_commits,omitempty"`
	SuspiciousCommits int         `json:"suspicious_commits,omitempty"`
	Suspicions        []Suspicion `json:"suspicions,omitempty"`
	AnalyzedAt        time.Time   `json:"analyzed_at"`
	// Timing fields
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	DurationMs  int64     `json:"duration_ms"`
	// Git-specific fields
	Velocity          string  `json:"velocity,omitempty"`
	TimeSpan          string  `json:"time_span,omitempty"`
	UniqueAuthors     int     `json:"unique_authors,omitempty"`
	AverageCommitSize int     `json:"average_commit_size,omitempty"`
	OverallSuspicion  float64 `json:"overall_suspicion,omitempty"`
	// Web-specific fields
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
}

type WebPattern struct {
	Type        string   `json:"type"`
	Severity    float64  `json:"severity"`
	Description string   `json:"description"`
	Examples    []string `json:"examples,omitempty"`
	Passed      bool     `json:"passed"`
}

type Suspicion struct {
	CommitHash  string   `json:"commit_hash"`
	CommitIndex int      `json:"commit_index"`
	Message     string   `json:"message"`
	Severity    string   `json:"severity"`
	Reasons     []string `json:"reasons"`
	Score       float64  `json:"score"`
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
