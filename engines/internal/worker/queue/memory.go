package queue

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/worker"
)

// PriorityQueue implements a priority queue for jobs
type PriorityQueue struct {
	jobs []*worker.Job
	mu   sync.RWMutex
}

func (pq *PriorityQueue) Len() int { return len(pq.jobs) }

func (pq *PriorityQueue) Less(i, j int) bool {
	// Higher priority first, then by creation time
	if pq.jobs[i].Priority == pq.jobs[j].Priority {
		return pq.jobs[i].CreatedAt.Before(pq.jobs[j].CreatedAt)
	}
	return pq.jobs[i].Priority > pq.jobs[j].Priority
}

func (pq *PriorityQueue) Swap(i, j int) {
	pq.jobs[i], pq.jobs[j] = pq.jobs[j], pq.jobs[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	pq.jobs = append(pq.jobs, x.(*worker.Job))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.jobs
	n := len(old)
	job := old[n-1]
	pq.jobs = old[0 : n-1]
	return job
}

// MemoryQueue implements JobQueue using in-memory storage
type MemoryQueue struct {
	queue   *PriorityQueue
	jobs    map[string]*worker.Job
	mu      sync.RWMutex
	closed  bool
	waiting []chan *worker.Job
}

// NewMemoryQueue creates a new memory-based job queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		queue: &PriorityQueue{},
		jobs:  make(map[string]*worker.Job),
	}
}

// Enqueue adds a job to the queue
func (mq *MemoryQueue) Enqueue(ctx context.Context, job *worker.Job) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return fmt.Errorf("queue is closed")
	}

	// Store job
	mq.jobs[job.ID] = job

	// Add to priority queue
	heap.Push(mq.queue, job)

	// Notify waiting workers
	if len(mq.waiting) > 0 {
		waiter := mq.waiting[0]
		mq.waiting = mq.waiting[1:]
		go func() {
			select {
			case waiter <- job:
			case <-ctx.Done():
			}
		}()
	}

	return nil
}

// Dequeue removes and returns the highest priority job
func (mq *MemoryQueue) Dequeue(ctx context.Context) (*worker.Job, error) {
	mq.mu.Lock()

	if mq.closed {
		mq.mu.Unlock()
		return nil, fmt.Errorf("queue is closed")
	}

	// Check if jobs available
	if mq.queue.Len() == 0 {
		// Create a channel to wait for jobs
		waiter := make(chan *worker.Job, 1)
		mq.waiting = append(mq.waiting, waiter)
		mq.mu.Unlock()

		// Wait for job or context cancellation
		select {
		case job := <-waiter:
			return job, nil
		case <-ctx.Done():
			// Remove from waiting list
			mq.mu.Lock()
			for i, w := range mq.waiting {
				if w == waiter {
					mq.waiting = append(mq.waiting[:i], mq.waiting[i+1:]...)
					break
				}
			}
			mq.mu.Unlock()
			return nil, ctx.Err()
		}
	}

	// Get highest priority job
	job := heap.Pop(mq.queue).(*worker.Job)
	job.Status = worker.JobStatusRunning
	now := time.Now()
	job.StartedAt = &now

	mq.mu.Unlock()
	return job, nil
}

// GetJob retrieves a job by ID
func (mq *MemoryQueue) GetJob(ctx context.Context, jobID string) (*worker.Job, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	job, exists := mq.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy
	jobCopy := *job
	return &jobCopy, nil
}

// UpdateJob updates an existing job
func (mq *MemoryQueue) UpdateJob(ctx context.Context, job *worker.Job) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return fmt.Errorf("queue is closed")
	}

	mq.jobs[job.ID] = job
	return nil
}

// DeleteJob removes a job from the queue
func (mq *MemoryQueue) DeleteJob(ctx context.Context, jobID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.jobs, jobID)
	return nil
}

// ListJobs returns jobs with the specified status
func (mq *MemoryQueue) ListJobs(ctx context.Context, status worker.JobStatus, limit int) ([]*worker.Job, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	var jobs []*worker.Job
	count := 0

	for _, job := range mq.jobs {
		if job.Status == status && count < limit {
			jobCopy := *job
			jobs = append(jobs, &jobCopy)
			count++
		}
	}

	return jobs, nil
}

// Size returns the number of jobs in the queue
func (mq *MemoryQueue) Size(ctx context.Context) (int, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	return mq.queue.Len(), nil
}

// Close closes the queue
func (mq *MemoryQueue) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.closed = true

	// Close all waiting channels
	for _, waiter := range mq.waiting {
		close(waiter)
	}
	mq.waiting = nil

	return nil
}
