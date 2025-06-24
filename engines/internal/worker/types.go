package worker

import (
	"context"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// JobStatus represents job execution status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
	JobStatusRetrying  JobStatus = "retrying"
)

// JobPriority represents job priority levels
type JobPriority int

const (
	PriorityLow      JobPriority = 1
	PriorityNormal   JobPriority = 5
	PriorityHigh     JobPriority = 8
	PriorityCritical JobPriority = 10
)

// Job represents a proof generation job
type Job struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    JobPriority            `json:"priority"`
	Status      JobStatus              `json:"status"`
	ProofReq    *types.ProofRequest    `json:"proof_request"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Metadata    map[string]interface{} `json:"metadata"`
	Callbacks   []string               `json:"callbacks"`
	WorkerID    int                    `json:"worker_id,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// Result represents job execution result
type Result struct {
	JobID       string                 `json:"job_id"`
	Status      JobStatus              `json:"status"`
	Proof       *types.ZKProof           `json:"proof,omitempty"`
	Error       error                  `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	WorkerID    int                    `json:"worker_id"`
	CompletedAt time.Time              `json:"completed_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// WorkerStats represents worker performance statistics
type WorkerStats struct {
	WorkerID        int           `json:"worker_id"`
	JobsProcessed   int64         `json:"jobs_processed"`
	JobsSuccessful  int64         `json:"jobs_successful"`
	JobsFailed      int64         `json:"jobs_failed"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	LastJobAt       time.Time     `json:"last_job_at"`
	IsActive        bool          `json:"is_active"`
}

// PoolStats represents worker pool statistics
type PoolStats struct {
	TotalWorkers     int           `json:"total_workers"`
	ActiveWorkers    int           `json:"active_workers"`
	IdleWorkers      int           `json:"idle_workers"`
	QueueSize        int           `json:"queue_size"`
	ProcessingJobs   int           `json:"processing_jobs"`
	CompletedJobs    int64         `json:"completed_jobs"`
	FailedJobs       int64         `json:"failed_jobs"`
	AverageWaitTime  time.Duration `json:"average_wait_time"`
	ThroughputPerMin float64       `json:"throughput_per_min"`
}

// Interfaces

// JobQueue defines the job queue interface
type JobQueue interface {
	Enqueue(ctx context.Context, job *Job) error
	Dequeue(ctx context.Context) (*Job, error)
	GetJob(ctx context.Context, jobID string) (*Job, error)
	UpdateJob(ctx context.Context, job *Job) error
	DeleteJob(ctx context.Context, jobID string) error
	ListJobs(ctx context.Context, status JobStatus, limit int) ([]*Job, error)
	Size(ctx context.Context) (int, error)
	Close() error
}

// ResultStorage defines the result storage interface
type ResultStorage interface {
	Store(ctx context.Context, result *Result) error
	Get(ctx context.Context, jobID string) (*Result, error)
	List(ctx context.Context, limit int, offset int) ([]*Result, error)
	Delete(ctx context.Context, jobID string) error
	Close() error
}

// CallbackManager defines the callback interface
type CallbackManager interface {
	Register(jobID string, callbacks []string) error
	Execute(ctx context.Context, result *Result) error
	Remove(jobID string) error
}

// Scheduler defines the job scheduling interface
type Scheduler interface {
	Schedule(ctx context.Context, job *Job) error
	GetNext(ctx context.Context) (*Job, error)
	UpdatePriority(ctx context.Context, jobID string, priority JobPriority) error
	Cancel(ctx context.Context, jobID string) error
}
