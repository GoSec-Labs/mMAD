package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// SimpleScheduler implements a simple priority-based scheduler
type SimpleScheduler struct {
	queue         JobQueue
	mu            sync.RWMutex
	scheduledJobs map[string]*ScheduledJob
}

// ScheduledJob represents a job with scheduling information
type ScheduledJob struct {
	Job         *Job
	ScheduledAt time.Time
	Delay       time.Duration
}

// NewSimpleScheduler creates a new simple scheduler
func NewSimpleScheduler(queue JobQueue) *SimpleScheduler {
	return &SimpleScheduler{
		queue:         queue,
		scheduledJobs: make(map[string]*ScheduledJob),
	}
}

// Schedule schedules a job for execution
func (ss *SimpleScheduler) Schedule(ctx context.Context, job *Job) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Check if job should be delayed (for retries)
	if retryDelay, exists := job.Metadata["retry_after"]; exists {
		if delay, ok := retryDelay.(time.Duration); ok {
			// Schedule for later execution
			scheduledJob := &ScheduledJob{
				Job:         job,
				ScheduledAt: time.Now().Add(delay),
				Delay:       delay,
			}

			ss.scheduledJobs[job.ID] = scheduledJob

			// Start a goroutine to enqueue the job later
			go ss.scheduleDelayed(ctx, scheduledJob)

			logger.Info("Job scheduled for delayed execution",
				"job_id", job.ID,
				"delay", delay,
				"scheduled_at", scheduledJob.ScheduledAt)

			return nil
		}
	}

	// Immediate scheduling
	return ss.queue.Enqueue(ctx, job)
}

// GetNext gets the next job to be executed
func (ss *SimpleScheduler) GetNext(ctx context.Context) (*Job, error) {
	return ss.queue.Dequeue(ctx)
}

// UpdatePriority updates the priority of a scheduled job
func (ss *SimpleScheduler) UpdatePriority(ctx context.Context, jobID string, priority JobPriority) error {
	// Get job from queue
	job, err := ss.queue.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	// Update priority
	job.Priority = priority

	// Update in queue
	return ss.queue.UpdateJob(ctx, job)
}

// Cancel cancels a scheduled job
func (ss *SimpleScheduler) Cancel(ctx context.Context, jobID string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	// Remove from scheduled jobs
	delete(ss.scheduledJobs, jobID)

	// Remove from queue
	return ss.queue.DeleteJob(ctx, jobID)
}

// scheduleDelayed handles delayed job scheduling
func (ss *SimpleScheduler) scheduleDelayed(ctx context.Context, scheduledJob *ScheduledJob) {
	// Wait for the scheduled time
	timer := time.NewTimer(time.Until(scheduledJob.ScheduledAt))
	defer timer.Stop()

	select {
	case <-timer.C:
		// Time to enqueue the job
		if err := ss.queue.Enqueue(ctx, scheduledJob.Job); err != nil {
			logger.Error("Failed to enqueue delayed job",
				"job_id", scheduledJob.Job.ID,
				"error", err)
		} else {
			logger.Info("Delayed job enqueued",
				"job_id", scheduledJob.Job.ID,
				"scheduled_at", scheduledJob.ScheduledAt)
		}

		// Remove from scheduled jobs
		ss.mu.Lock()
		delete(ss.scheduledJobs, scheduledJob.Job.ID)
		ss.mu.Unlock()

	case <-ctx.Done():
		logger.Info("Delayed job scheduling cancelled",
			"job_id", scheduledJob.Job.ID)
		return
	}
}

// GetScheduledJobs returns all scheduled jobs
func (ss *SimpleScheduler) GetScheduledJobs() map[string]*ScheduledJob {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	result := make(map[string]*ScheduledJob)
	for id, job := range ss.scheduledJobs {
		result[id] = job
	}
	return result
}
