package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/events"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// Worker represents a single worker that processes jobs
type Worker struct {
	ID           int
	prover       zkproof.ProofEngine
	eventEmitter events.Emitter
	stats        *WorkerStats
	running      int32 // atomic
	jobChan      chan *Job
	quitChan     chan struct{}
	resultChan   chan *Result
	wg           sync.WaitGroup
}

// NewWorker creates a new worker
func NewWorker(id int, prover zkproof.ProofEngine, eventEmitter events.Emitter) *Worker {
	return &Worker{
		ID:           id,
		prover:       prover,
		eventEmitter: eventEmitter,
		stats: &WorkerStats{
			WorkerID: id,
			IsActive: false,
		},
		jobChan:    make(chan *Job, 1),
		quitChan:   make(chan struct{}),
		resultChan: make(chan *Result, 10),
	}
}

// Start begins the worker's job processing loop
func (w *Worker) Start(ctx context.Context) {
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return // Already running
	}

	w.stats.IsActive = true
	w.wg.Add(1)

	go w.run(ctx)

	logger.Info("Worker started", "worker_id", w.ID)
	w.eventEmitter.EmitWorkerStarted(w.ID)
}

// Stop stops the worker
func (w *Worker) Stop() {
	if !atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		return // Already stopped
	}

	close(w.quitChan)
	w.wg.Wait()

	w.stats.IsActive = false

	logger.Info("Worker stopped", "worker_id", w.ID)
	w.eventEmitter.EmitWorkerStopped(w.ID)
}

// ProcessJob sends a job to the worker for processing
func (w *Worker) ProcessJob(job *Job) bool {
	select {
	case w.jobChan <- job:
		return true
	case <-time.After(100 * time.Millisecond):
		return false // Worker busy
	}
}

// GetStats returns worker statistics
func (w *Worker) GetStats() WorkerStats {
	return *w.stats
}

// GetResultChan returns the result channel
func (w *Worker) GetResultChan() <-chan *Result {
	return w.resultChan
}

// IsRunning returns true if worker is running
func (w *Worker) IsRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

// run is the main worker loop
func (w *Worker) run(ctx context.Context) {
	defer w.wg.Done()

	logger.Info("Worker running", "worker_id", w.ID)

	for {
		select {
		case job := <-w.jobChan:
			w.processJob(ctx, job)

		case <-w.quitChan:
			logger.Info("Worker quit signal received", "worker_id", w.ID)
			return

		case <-ctx.Done():
			logger.Info(fmt.Sprintf("Worker context cancelled: worker_id=%d", w.ID))
			return
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(ctx context.Context, job *Job) {
	start := time.Now()
	job.WorkerID = w.ID

	logger.Info(fmt.Sprintf("Processing job: worker_id=%d job_id=%s priority=%v", w.ID, job.ID, job.Priority))

	w.eventEmitter.EmitJobStarted(job.ID, w.ID)

	// Create job context with timeout
	jobCtx := ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		jobCtx, cancel = context.WithTimeout(ctx, job.Timeout)
		defer cancel()
	}

	// Generate proof
	result := w.generateProof(jobCtx, job)
	result.Duration = time.Since(start)
	result.WorkerID = w.ID
	result.CompletedAt = time.Now()

	// Update statistics
	w.updateStats(result)

	// Send result
	select {
	case w.resultChan <- result:
	case <-ctx.Done():
		return
	}

	// Emit events
	if result.Error != nil {
		w.eventEmitter.EmitJobFailed(job.ID, w.ID, result.Error.Error())
		logger.Error(fmt.Sprintf("Job failed: worker_id=%d job_id=%s error=%v duration=%s", w.ID, job.ID, result.Error, result.Duration))
	} else {
		w.eventEmitter.EmitJobCompleted(job.ID, w.ID, result.Duration)
		logger.Info(fmt.Sprintf("Job completed: worker_id=%d job_id=%s duration=%s", w.ID, job.ID, result.Duration))
	}
}

// generateProof generates a ZK proof for the job
func (w *Worker) generateProof(ctx context.Context, job *Job) *Result {
	result := &Result{
		JobID:    job.ID,
		Status:   JobStatusCompleted,
		Metadata: make(map[string]interface{}),
	}

	// Generate proof using the proof engine
	proof, err := w.prover.GenerateProof(ctx, job.ProofReq)
	if err != nil {
		result.Status = JobStatusFailed
		result.Error = fmt.Errorf("proof generation failed: %w", err)

		// Check if we should retry
		if job.RetryCount < job.MaxRetries && w.shouldRetry(err) {
			result.Status = JobStatusRetrying
			result.Metadata["should_retry"] = true
			result.Metadata["retry_after"] = w.getRetryDelay(job.RetryCount)
		}

		return result
	}

	result.Proof = proof
	result.Metadata["proof_size"] = len(proof.ProofData)
	// result.Metadata["circuit_id"] = job.ProofReq.CircuitID // Removed: field does not exist
	// result.Metadata["proof_type"] = job.ProofReq.ProofType // Removed: field does not exist

	return result
}

// shouldRetry determines if a job should be retried based on the error
func (w *Worker) shouldRetry(err error) bool {
	// Define retry logic based on error type
	errStr := err.Error()

	// Don't retry validation errors
	if contains(errStr, "invalid input") || contains(errStr, "validation failed") {
		return false
	}

	// Retry timeout and temporary errors
	if contains(errStr, "timeout") || contains(errStr, "temporary") {
		return true
	}

	// Default: don't retry
	return false
}

// getRetryDelay calculates delay before retry
func (w *Worker) getRetryDelay(retryCount int) time.Duration {
	// Exponential backoff: 2^retryCount seconds
	delay := time.Duration(1<<uint(retryCount)) * time.Second
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	return delay
}

// updateStats updates worker statistics
func (w *Worker) updateStats(result *Result) {
	w.stats.JobsProcessed++
	w.stats.TotalDuration += result.Duration
	w.stats.LastJobAt = result.CompletedAt

	if result.Error != nil {
		w.stats.JobsFailed++
	} else {
		w.stats.JobsSuccessful++
	}

	// Calculate average duration
	if w.stats.JobsProcessed > 0 {
		w.stats.AverageDuration = w.stats.TotalDuration / time.Duration(w.stats.JobsProcessed)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
