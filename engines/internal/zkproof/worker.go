package zkproof

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// WorkerPool manages a pool of workers for proof generation
type WorkerPool struct {
	workers  int
	jobQueue chan Job
	workerWG sync.WaitGroup
	quit     chan bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// Job represents a proof generation job
type Job interface {
	Execute(ctx context.Context) error
	ID() string
	Type() string
	Priority() int
}

// ProofJob represents a proof generation job
type ProofJob struct {
	ID       string
	Request  *types.ProofRequest
	Engine   *Engine
	Priority int

	// Progress tracking
	progress chan *GenerationProgress
	result   chan *types.ZKProof
	err      chan error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers:  workers,
		jobQueue: make(chan Job, workers*2), // Buffer for jobs
		quit:     make(chan bool),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start(ctx context.Context) error {
	wp.ctx, wp.cancel = context.WithCancel(ctx)

	logger.Info("Starting worker pool", "workers", wp.workers)

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.workerWG.Add(1)
		go wp.worker(i)
	}

	return nil
}

// Stop stops the worker pool
func (wp *WorkerPool) Stop() {
	logger.Info("Stopping worker pool")

	close(wp.quit)
	wp.cancel()
	wp.workerWG.Wait()

	logger.Info("Worker pool stopped")
}

// Submit submits a job to the worker pool
func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// worker runs the worker loop
func (wp *WorkerPool) worker(id int) {
	defer wp.workerWG.Done()

	logger.Debug("Worker started", "worker_id", id)

	for {
		select {
		case job := <-wp.jobQueue:
			wp.executeJob(id, job)
		case <-wp.quit:
			logger.Debug("Worker stopping", "worker_id", id)
			return
		case <-wp.ctx.Done():
			logger.Debug("Worker context cancelled", "worker_id", id)
			return
		}
	}
}

func (wp *WorkerPool) executeJob(workerID int, job Job) {
	start := time.Now()

	logger.Info("Executing job",
		"worker_id", workerID,
		"job_id", job.ID(),
		"job_type", job.Type())

	// Execute the job with timeout
	ctx, cancel := context.WithTimeout(wp.ctx, 5*time.Minute)
	defer cancel()

	err := job.Execute(ctx)
	duration := time.Since(start)

	if err != nil {
		logger.Error("Job execution failed",
			"worker_id", workerID,
			"job_id", job.ID(),
			"duration", duration,
			"error", err)
	} else {
		logger.Info("Job completed successfully",
			"worker_id", workerID,
			"job_id", job.ID(),
			"duration", duration)
	}
}

// ProofJob methods

func (j *ProofJob) Execute(ctx context.Context) error {
	// Update progress
	j.updateProgress(types.ProofStatusGenerating, 0.1, "starting generation")

	// Generate the proof
	proof, err := j.Engine.GenerateProof(ctx, j.Request)
	if err != nil {
		j.updateProgress(types.ProofStatusFailed, 1.0, "generation failed")
		return err
	}

	j.updateProgress(types.ProofStatusGenerated, 1.0, "generation completed")

	// Send result
	select {
	case j.result <- proof:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (j *ProofJob) ID() string {
	return j.ID
}

func (j *ProofJob) Type() string {
	return string(j.Request.Type)
}

func (j *ProofJob) Priority() int {
	return j.Priority
}

func (j *ProofJob) updateProgress(status types.ProofStatus, progress float64, stage string) {
	progressUpdate := &GenerationProgress{
		ProofID:   j.ID,
		Status:    status,
		Progress:  progress,
		Stage:     stage,
		UpdatedAt: time.Now(),
	}

	select {
	case j.progress <- progressUpdate:
	default:
		// Channel full, skip update
	}
}
