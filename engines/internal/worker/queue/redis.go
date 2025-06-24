package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/worker"
	"github.com/go-redis/redis/v8"
)

// RedisQueue implements JobQueue using Redis
type RedisQueue struct {
	client    *redis.Client
	queueKey  string
	jobsKey   string
	resultKey string
	timeout   time.Duration
}

// NewRedisQueue creates a new Redis-based job queue
func NewRedisQueue(client *redis.Client, prefix string) *RedisQueue {
	return &RedisQueue{
		client:    client,
		queueKey:  prefix + ":queue",
		jobsKey:   prefix + ":jobs:",
		resultKey: prefix + ":results:",
		timeout:   30 * time.Second,
	}
}

// Enqueue adds a job to the Redis queue
func (rq *RedisQueue) Enqueue(ctx context.Context, job *worker.Job) error {
	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	pipe := rq.client.TxPipeline()

	// Store job data
	pipe.Set(ctx, rq.jobsKey+job.ID, jobData, 24*time.Hour)

	// Add to priority queue (using score as priority * -1 for reverse order)
	score := float64(-job.Priority)
	pipe.ZAdd(ctx, rq.queueKey, &redis.Z{
		Score:  score,
		Member: job.ID,
	})

	_, err = pipe.Exec(ctx)
	return err
}

// Dequeue removes and returns the highest priority job
func (rq *RedisQueue) Dequeue(ctx context.Context) (*worker.Job, error) {
	for {
		// Use BZPOPMIN for blocking pop with timeout
		result, err := rq.client.BZPopMin(ctx, rq.timeout, rq.queueKey).Result()
		if err != nil {
			if err == redis.Nil {
				// Timeout - check if context cancelled
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					continue // Keep trying
				}
			}
			return nil, err
		}

		jobID := result.Member.(string)

		// Get job data
		jobData, err := rq.client.Get(ctx, rq.jobsKey+jobID).Result()
		if err != nil {
			if err == redis.Nil {
				// Job was deleted, continue to next
				continue
			}
			return nil, err
		}

		// Deserialize job
		var job worker.Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			return nil, fmt.Errorf("failed to unmarshal job: %w", err)
		}

		// Update job status
		job.Status = worker.JobStatusRunning
		now := time.Now()
		job.StartedAt = &now

		// Update job in Redis
		if err := rq.UpdateJob(ctx, &job); err != nil {
			return nil, err
		}

		return &job, nil
	}
}

// GetJob retrieves a job by ID
func (rq *RedisQueue) GetJob(ctx context.Context, jobID string) (*worker.Job, error) {
	jobData, err := rq.client.Get(ctx, rq.jobsKey+jobID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, err
	}

	var job worker.Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// UpdateJob updates an existing job
func (rq *RedisQueue) UpdateJob(ctx context.Context, job *worker.Job) error {
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	return rq.client.Set(ctx, rq.jobsKey+job.ID, jobData, 24*time.Hour).Err()
}

// DeleteJob removes a job from the queue
func (rq *RedisQueue) DeleteJob(ctx context.Context, jobID string) error {
	pipe := rq.client.TxPipeline()

	// Remove from queue
	pipe.ZRem(ctx, rq.queueKey, jobID)

	// Remove job data
	pipe.Del(ctx, rq.jobsKey+jobID)

	_, err := pipe.Exec(ctx)
	return err
}

// ListJobs returns jobs with the specified status
func (rq *RedisQueue) ListJobs(ctx context.Context, status worker.JobStatus, limit int) ([]*worker.Job, error) {
	// This is a simplified implementation
	// In practice, you'd want to maintain separate indexes for different statuses

	var cursor uint64
	var jobs []*worker.Job
	count := 0

	for count < limit {
		keys, newCursor, err := rq.client.Scan(ctx, cursor, rq.jobsKey+"*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			if count >= limit {
				break
			}

			jobData, err := rq.client.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var job worker.Job
			if err := json.Unmarshal([]byte(jobData), &job); err != nil {
				continue
			}

			if job.Status == status {
				jobs = append(jobs, &job)
				count++
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return jobs, nil
}

// Size returns the number of jobs in the queue
func (rq *RedisQueue) Size(ctx context.Context) (int, error) {
	count, err := rq.client.ZCard(ctx, rq.queueKey).Result()
	return int(count), err
}

// Close closes the Redis connection
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}
