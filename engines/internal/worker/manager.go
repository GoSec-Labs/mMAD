package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/events"
	"github.com/GoSec-Labs/mMAD/engines/internal/worker/queue"
	"github.com/GoSec-Labs/mMAD/engines/internal/worker/storage"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// Manager manages the entire worker system
type Manager struct {
	pool            *Pool
	queue           JobQueue
	resultStorage   ResultStorage
	callbackManager CallbackManager
	scheduler       Scheduler
	eventEmitter    events.Emitter
	config          Config
}

// Config holds worker system configuration
type Config struct {
	Pool        PoolConfig
	Queue       QueueConfig
	Storage     StorageConfig
	Callbacks   CallbackConfig
	EventBuffer int
}

// QueueConfig holds queue configuration
type QueueConfig struct {
	Type        string // "memory" or "redis"
	RedisURL    string
	MaxSize     int
	Persistence bool
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type    string // "memory" or "database"
	MaxSize int
	TTL     time.Duration
}

// CallbackConfig holds callback configuration
type CallbackConfig struct {
	Timeout    time.Duration
	MaxRetries int
	Workers    int
}

// NewManager creates a new worker manager
func NewManager(
	prover zkproof.ProofEngine,
	eventEmitter events.Emitter,
	config Config,
) (*Manager, error) {

	// Create job queue
	var jobQueue JobQueue
	switch config.Queue.Type {
	case "memory":
		jobQueue = queue.NewMemoryQueue()
	case "redis":
		// Redis queue implementation would go here
		return nil, fmt.Errorf("redis queue not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported queue type: %s", config.Queue.Type)
	}

	// Create result storage
	var resultStorage ResultStorage
	switch config.Storage.Type {
	case "memory":
		resultStorage = storage.NewMemoryResultStorage(config.Storage.MaxSize)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Storage.Type)
	}

	// Create callback manager
	callbackManager := NewSimpleCallbackManager()

	// Create scheduler
	scheduler := NewSimpleScheduler(jobQueue)

	// Create worker pool
	pool := NewPool(
		prover,
		eventEmitter,
		jobQueue,
		resultStorage,
		callbackManager,
		config.Pool,
	)

	return &Manager{
		pool:            pool,
		queue:           jobQueue,
		resultStorage:   resultStorage,
		callbackManager: callbackManager,
		scheduler:       scheduler,
		eventEmitter:    eventEmitter,
		config:          config,
	}, nil
}

// Start starts the worker system
func (m *Manager) Start(ctx context.Context) error {
	logger.Info("Starting worker manager")

	// Start the worker pool
	if err := m.pool.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	m.eventEmitter.EmitSystemStarted("worker_manager")
	logger.Info("Worker manager started successfully")

	return nil
}

// Stop stops the worker system
func (m *Manager) Stop(ctx context.Context) error {
	logger.Info("Stopping worker manager")

	// Stop worker pool
	if err := m.pool.Stop(ctx); err != nil {
		logger.Error("Error stopping worker pool", "error", err)
	}

	// Close queue
	if err := m.queue.Close(); err != nil {
		logger.Error("Error closing queue", "error", err)
	}

	// Close result storage
	if err := m.resultStorage.Close(); err != nil {
		logger.Error("Error closing result storage", "error", err)
	}

	m.eventEmitter.EmitSystemStopped("worker_manager")
	logger.Info("Worker manager stopped")

	return nil
}

// SubmitJob submits a new job for processing
func (m *Manager) SubmitJob(ctx context.Context, job *Job) error {
	// Register callbacks if provided
	if len(job.Callbacks) > 0 {
		if err := m.callbackManager.Register(job.ID, job.Callbacks); err != nil {
			logger.Error("Failed to register callbacks", "job_id", job.ID, "error", err)
		}
	}

	// Schedule the job
	return m.scheduler.Schedule(ctx, job)
}

// GetJobStatus returns the status of a job
func (m *Manager) GetJobStatus(ctx context.Context, jobID string) (*Job, error) {
	return m.queue.GetJob(ctx, jobID)
}

// GetJobResult returns the result of a completed job
func (m *Manager) GetJobResult(ctx context.Context, jobID string) (*Result, error) {
	return m.resultStorage.Get(ctx, jobID)
}

// CancelJob cancels a pending or running job
func (m *Manager) CancelJob(ctx context.Context, jobID string) error {
	// Update job status
	job, err := m.queue.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	if job.Status == JobStatusRunning {
		return fmt.Errorf("cannot cancel running job")
	}

	job.Status = JobStatusCancelled
	now := time.Now()
	job.CompletedAt = &now

	if err := m.queue.UpdateJob(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Cancel from scheduler
	return m.scheduler.Cancel(ctx, jobID)
}

// ListJobs returns a list of jobs with the specified status
func (m *Manager) ListJobs(ctx context.Context, status JobStatus, limit int) ([]*Job, error) {
	return m.queue.ListJobs(ctx, status, limit)
}

// GetPoolStats returns worker pool statistics
func (m *Manager) GetPoolStats() PoolStats {
	return m.pool.GetStats()
}

// GetWorkerStats returns statistics for all workers
func (m *Manager) GetWorkerStats() []WorkerStats {
	return m.pool.GetWorkerStats()
}

// UpdateJobPriority updates the priority of a job
func (m *Manager) UpdateJobPriority(ctx context.Context, jobID string, priority JobPriority) error {
	return m.scheduler.UpdatePriority(ctx, jobID, priority)
}

// GetQueueSize returns the current queue size
func (m *Manager) GetQueueSize(ctx context.Context) (int, error) {
	return m.queue.Size(ctx)
}
