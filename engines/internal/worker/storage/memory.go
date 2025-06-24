package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/GoSec-Labs/mMAD/engines/internal/worker"
)

// MemoryResultStorage implements ResultStorage using in-memory storage
type MemoryResultStorage struct {
	results map[string]*worker.Result
	mu      sync.RWMutex
	maxSize int
}

// NewMemoryResultStorage creates a new memory-based result storage
func NewMemoryResultStorage(maxSize int) *MemoryResultStorage {
	return &MemoryResultStorage{
		results: make(map[string]*worker.Result),
		maxSize: maxSize,
	}
}

// Store stores a job result
func (mrs *MemoryResultStorage) Store(ctx context.Context, result *worker.Result) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	// Check if we need to evict old results
	if len(mrs.results) >= mrs.maxSize {
		mrs.evictOldest()
	}

	// Store result (make a copy to avoid race conditions)
	resultCopy := *result
	mrs.results[result.JobID] = &resultCopy

	return nil
}

// Get retrieves a result by job ID
func (mrs *MemoryResultStorage) Get(ctx context.Context, jobID string) (*worker.Result, error) {
	mrs.mu.RLock()
	defer mrs.mu.RUnlock()

	result, exists := mrs.results[jobID]
	if !exists {
		return nil, fmt.Errorf("result not found for job: %s", jobID)
	}

	// Return a copy
	resultCopy := *result
	return &resultCopy, nil
}

// List returns a list of results with pagination
func (mrs *MemoryResultStorage) List(ctx context.Context, limit int, offset int) ([]*worker.Result, error) {
	mrs.mu.RLock()
	defer mrs.mu.RUnlock()

	// Convert map to slice
	var results []*worker.Result
	for _, result := range mrs.results {
		resultCopy := *result
		results = append(results, &resultCopy)
	}

	// Sort by completion time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CompletedAt.After(results[j].CompletedAt)
	})

	// Apply pagination
	start := offset
	if start >= len(results) {
		return []*worker.Result{}, nil
	}

	end := start + limit
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// Delete removes a result
func (mrs *MemoryResultStorage) Delete(ctx context.Context, jobID string) error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	delete(mrs.results, jobID)
	return nil
}

// Close closes the storage
func (mrs *MemoryResultStorage) Close() error {
	mrs.mu.Lock()
	defer mrs.mu.Unlock()

	mrs.results = make(map[string]*worker.Result)
	return nil
}

// evictOldest removes the oldest result to make space
func (mrs *MemoryResultStorage) evictOldest() {
	var oldestJobID string
	var oldestTime = mrs.results[oldestJobID].CompletedAt

	for jobID, result := range mrs.results {
		if oldestJobID == "" || result.CompletedAt.Before(oldestTime) {
			oldestJobID = jobID
			oldestTime = result.CompletedAt
		}
	}

	if oldestJobID != "" {
		delete(mrs.results, oldestJobID)
	}
}
