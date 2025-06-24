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

// Pool represents a pool of workers
type Pool struct {
	workers         []*Worker
	prover          zkproof.ProofEngine
	eventEmitter    events.Emitter
	queue           JobQueue
	resultStorage   ResultStorage
	callbackManager CallbackManager

	// Pool state
	running    int32 // atomic
	wg         sync.WaitGroup
	resultChan chan *Result

	// Configuration
	minWorkers  int
	maxWorkers  int
	scalePolicy ScalePolicy

	// Statistics
	stats      *PoolStats
	statsMutex sync.RWMutex

	// Scaling
	lastScaleTime time.Time
	scaleMutex    sync.Mutex
}

// ScalePolicy defines how the pool should scale
type ScalePolicy struct {
	ScaleUpThreshold   int           // Queue size to scale up
	ScaleDownThreshold int           // Idle time to scale down
	ScaleUpStep        int           // Workers to add when scaling up
	ScaleDownStep      int           // Workers to remove when scaling down
	ScaleInterval      time.Duration // Minimum time between scaling events
}

// PoolConfig holds pool configuration
type PoolConfig struct {
	MinWorkers   int
	MaxWorkers   int
	ScalePolicy  ScalePolicy
	QueueBuffer  int
	ResultBuffer int
}

// NewPool creates a new worker pool
func NewPool(
	prover zkproof.ProofEngine,
	eventEmitter events.Emitter,
	queue JobQueue,
	resultStorage ResultStorage,
	callbackManager CallbackManager,
	config PoolConfig,
) *Pool {
	return &Pool{
		prover:          prover,
		eventEmitter:    eventEmitter,
		queue:           queue,
		resultStorage:   resultStorage,
		callbackManager: callbackManager,
		minWorkers:      config.MinWorkers,
		maxWorkers:      config.MaxWorkers,
		scalePolicy:     config.ScalePolicy,
		resultChan:      make(chan *Result, config.ResultBuffer),
		stats: &PoolStats{
			TotalWorkers:  0,
			ActiveWorkers: 0,
			IdleWorkers:   0,
		},
	}
}

// Start starts the worker pool
func (p *Pool) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return fmt.Errorf("pool is already running")
	}

	logger.Info("Starting worker pool",
		"min_workers", p.minWorkers,
		"max_workers", p.maxWorkers)

	// Create initial workers
	for i := 0; i < p.minWorkers; i++ {
		if err := p.addWorker(ctx); err != nil {
			logger.Error("Failed to add initial worker", "error", err)
		}
	}

	// Start background goroutines
	p.wg.Add(3)
	go p.jobDispatcher(ctx)
	go p.resultProcessor(ctx)
	go p.scaler(ctx)

	p.eventEmitter.EmitPoolStarted(len(p.workers))
	logger.Info("Worker pool started", "workers", len(p.workers))

	return nil
}

// Stop stops the worker pool
func (p *Pool) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.running, 1, 0) {
		return fmt.Errorf("pool is not running")
	}

	logger.Info("Stopping worker pool...")

	// Stop all workers
	var wg sync.WaitGroup
	for _, worker := range p.workers {
		wg.Add(1)
		go func(w *Worker) {
			defer wg.Done()
			w.Stop()
		}(worker)
	}

	// Wait for workers to stop
	wg.Wait()

	// Wait for background goroutines
	p.wg.Wait()

	// Close result channel
	close(p.resultChan)

	p.eventEmitter.EmitPoolStopped(len(p.workers))
	logger.Info("Worker pool stopped")

	return nil
}

// SubmitJob submits a job to the pool
func (p *Pool) SubmitJob(ctx context.Context, job *Job) error {
	if atomic.LoadInt32(&p.running) == 0 {
		return fmt.Errorf("pool is not running")
	}

	// Set default values
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.Status == "" {
		job.Status = JobStatusPending
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}

	// Enqueue job
	if err := p.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	p.eventEmitter.EmitJobQueued(job.ID, job.Priority)
	logger.Info("Job submitted", "job_id", job.ID, "priority", job.Priority)

	return nil
}

// GetStats returns pool statistics
func (p *Pool) GetStats() PoolStats {
	p.statsMutex.RLock()
	defer p.statsMutex.RUnlock()

	stats := *p.stats
	stats.TotalWorkers = len(p.workers)

	// Count active workers
	activeWorkers := 0
	for _, worker := range p.workers {
		if worker.IsRunning() {
			activeWorkers++
		}
	}
	stats.ActiveWorkers = activeWorkers
	stats.IdleWorkers = stats.TotalWorkers - activeWorkers

	// Get queue size
	if queueSize, err := p.queue.Size(context.Background()); err == nil {
		stats.QueueSize = queueSize
	}

	return stats
}

// GetWorkerStats returns statistics for all workers
func (p *Pool) GetWorkerStats() []WorkerStats {
	var stats []WorkerStats
	for _, worker := range p.workers {
		stats = append(stats, worker.GetStats())
	}
	return stats
}

// jobDispatcher dispatches jobs to available workers
func (p *Pool) jobDispatcher(ctx context.Context) {
	defer p.wg.Done()

	logger.Info("Job dispatcher started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Job dispatcher stopped")
			return
		default:
			// Get next job
			job, err := p.queue.Dequeue(ctx)
			if err != nil {
				if err == ctx.Err() {
					return
				}
				logger.Error("Failed to dequeue job", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Find available worker
			assigned := false
			for _, worker := range p.workers {
				if worker.ProcessJob(job) {
					assigned = true
					break
				}
			}

			if !assigned {
				// No workers available, re-queue job
				job.Status = JobStatusPending
				if err := p.queue.Enqueue(ctx, job); err != nil {
					logger.Error("Failed to re-queue job", "job_id", job.ID, "error", err)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

// resultProcessor processes job results
func (p *Pool) resultProcessor(ctx context.Context) {
	defer p.wg.Done()

	logger.Info("Result processor started")

	for {
		select {
		case result := <-p.resultChan:
			p.processResult(ctx, result)

		case <-ctx.Done():
			logger.Info("Result processor stopped")
			return
		}
	}
}

// processResult processes a single job result
func (p *Pool) processResult(ctx context.Context, result *Result) {
	// Update statistics
	p.updatePoolStats(result)

	// Store result
	if err := p.resultStorage.Store(ctx, result); err != nil {
		logger.Error("Failed to store result", "job_id", result.JobID, "error", err)
	}

	// Handle retry if needed
	if result.Status == JobStatusRetrying {
		p.handleRetry(ctx, result)
		return
	}

	// Execute callbacks
	if err := p.callbackManager.Execute(ctx, result); err != nil {
		logger.Error("Failed to execute callbacks", "job_id", result.JobID, "error", err)
	}

	// Update job status
	job, err := p.queue.GetJob(ctx, result.JobID)
	if err != nil {
		logger.Error("Failed to get job for status update", "job_id", result.JobID, "error", err)
		return
	}

	job.Status = result.Status
	now := time.Now()
	job.CompletedAt = &now
	if result.Error != nil {
		job.Error = result.Error.Error()
	}

	if err := p.queue.UpdateJob(ctx, job); err != nil {
		logger.Error("Failed to update job status", "job_id", result.JobID, "error", err)
	}
}

// handleRetry handles job retry logic
func (p *Pool) handleRetry(ctx context.Context, result *Result) {
	job, err := p.queue.GetJob(ctx, result.JobID)
	if err != nil {
		logger.Error("Failed to get job for retry", "job_id", result.JobID, "error", err)
		return
	}

	job.RetryCount++
	job.Status = JobStatusPending
	job.StartedAt = nil
	job.CompletedAt = nil

	// Apply retry delay
	if retryAfter, ok := result.Metadata["retry_after"].(time.Duration); ok {
		time.Sleep(retryAfter)
	}

	// Re-enqueue job
	if err := p.queue.Enqueue(ctx, job); err != nil {
		logger.Error("Failed to re-enqueue job for retry", "job_id", result.JobID, "error", err)
		return
	}

	logger.Info("Job re-queued for retry",
		"job_id", result.JobID,
		"retry_count", job.RetryCount,
		"max_retries", job.MaxRetries)
}

// scaler handles automatic scaling of the worker pool
func (p *Pool) scaler(ctx context.Context) {
	defer p.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	logger.Info("Pool scaler started")

	for {
		select {
		case <-ticker.C:
			p.checkAndScale(ctx)

		case <-ctx.Done():
			logger.Info("Pool scaler stopped")
			return
		}
	}
}

// checkAndScale checks if scaling is needed and performs it
func (p *Pool) checkAndScale(ctx context.Context) {
	p.scaleMutex.Lock()
	defer p.scaleMutex.Unlock()

	// Check if enough time has passed since last scaling
	if time.Since(p.lastScaleTime) < p.scalePolicy.ScaleInterval {
		return
	}

	queueSize, err := p.queue.Size(ctx)
	if err != nil {
		logger.Error("Failed to get queue size for scaling", "error", err)
		return
	}

	currentWorkers := len(p.workers)

	// Check if we should scale up
	if queueSize >= p.scalePolicy.ScaleUpThreshold && currentWorkers < p.maxWorkers {
		workersToAdd := p.scalePolicy.ScaleUpStep
		if currentWorkers+workersToAdd > p.maxWorkers {
			workersToAdd = p.maxWorkers - currentWorkers
		}

		for i := 0; i < workersToAdd; i++ {
			if err := p.addWorker(ctx); err != nil {
				logger.Error("Failed to add worker during scale up", "error", err)
				break
			}
		}

		p.lastScaleTime = time.Now()
		logger.Info("Scaled up worker pool",
			"added_workers", workersToAdd,
			"total_workers", len(p.workers),
			"queue_size", queueSize)

		p.eventEmitter.EmitPoolScaled("up", len(p.workers))
		return
	}

	// Check if we should scale down
	if queueSize <= p.scalePolicy.ScaleDownThreshold && currentWorkers > p.minWorkers {
		// Check if workers are idle
		idleWorkers := p.countIdleWorkers()
		if idleWorkers >= p.scalePolicy.ScaleDownStep {
			workersToRemove := p.scalePolicy.ScaleDownStep
			if currentWorkers-workersToRemove < p.minWorkers {
				workersToRemove = currentWorkers - p.minWorkers
			}

			for i := 0; i < workersToRemove; i++ {
				if err := p.removeWorker(); err != nil {
					logger.Error("Failed to remove worker during scale down", "error", err)
					break
				}
			}

			p.lastScaleTime = time.Now()
			logger.Info("Scaled down worker pool",
				"removed_workers", workersToRemove,
				"total_workers", len(p.workers),
				"queue_size", queueSize)

			p.eventEmitter.EmitPoolScaled("down", len(p.workers))
		}
	}
}

// addWorker adds a new worker to the pool
func (p *Pool) addWorker(ctx context.Context) error {
	workerID := len(p.workers) + 1
	worker := NewWorker(workerID, p.prover, p.eventEmitter)

	// Start worker
	worker.Start(ctx)

	// Connect result channel
	go func() {
		for result := range worker.GetResultChan() {
			select {
			case p.resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}()

	p.workers = append(p.workers, worker)

	logger.Info("Added worker to pool", "worker_id", workerID, "total_workers", len(p.workers))
	return nil
}

// removeWorker removes a worker from the pool
func (p *Pool) removeWorker() error {
	if len(p.workers) <= p.minWorkers {
		return fmt.Errorf("cannot remove worker: at minimum capacity")
	}

	// Find the last worker and stop it
	lastIndex := len(p.workers) - 1
	worker := p.workers[lastIndex]

	worker.Stop()
	p.workers = p.workers[:lastIndex]

	logger.Info("Removed worker from pool", "worker_id", worker.ID, "total_workers", len(p.workers))
	return nil
}

// countIdleWorkers counts workers that are currently idle
func (p *Pool) countIdleWorkers() int {
	idleCount := 0
	for _, worker := range p.workers {
		stats := worker.GetStats()
		// Consider worker idle if it hasn't processed a job in the last minute
		if time.Since(stats.LastJobAt) > time.Minute {
			idleCount++
		}
	}
	return idleCount
}

// updatePoolStats updates pool-level statistics
func (p *Pool) updatePoolStats(result *Result) {
	p.statsMutex.Lock()
	defer p.statsMutex.Unlock()

	if result.Error != nil {
		p.stats.FailedJobs++
	} else {
		p.stats.CompletedJobs++
	}

	// Update throughput (simplified calculation)
	totalJobs := p.stats.CompletedJobs + p.stats.FailedJobs
	if totalJobs > 0 {
		// This is a simplified throughput calculation
		// In practice, you'd want a more sophisticated time-window based calculation
		p.stats.ThroughputPerMin = float64(totalJobs) / time.Since(time.Now().Add(-time.Hour)).Minutes()
	}
}
