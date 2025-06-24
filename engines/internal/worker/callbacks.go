package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// CallbackType represents different types of callbacks
type CallbackType string

const (
	CallbackTypeWebhook CallbackType = "webhook"
	CallbackTypeQueue   CallbackType = "queue"
	CallbackTypeCustom  CallbackType = "custom"
)

// Callback represents a result callback configuration
type Callback struct {
	ID       string                 `json:"id"`
	Type     CallbackType           `json:"type"`
	URL      string                 `json:"url,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Timeout  time.Duration          `json:"timeout"`
	Retries  int                    `json:"retries"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SimpleCallbackManager implements CallbackManager
type SimpleCallbackManager struct {
	callbacks map[string][]Callback
	mu        sync.RWMutex
	client    *http.Client
}

// NewSimpleCallbackManager creates a new callback manager
func NewSimpleCallbackManager() *SimpleCallbackManager {
	return &SimpleCallbackManager{
		callbacks: make(map[string][]Callback),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Register registers callbacks for a job
func (scm *SimpleCallbackManager) Register(jobID string, callbacks []string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	var callbackConfigs []Callback
	for _, callbackURL := range callbacks {
		callbackConfigs = append(callbackConfigs, Callback{
			ID:      fmt.Sprintf("%s_%d", jobID, len(callbackConfigs)),
			Type:    CallbackTypeWebhook,
			URL:     callbackURL,
			Timeout: 10 * time.Second,
			Retries: 3,
		})
	}

	scm.callbacks[jobID] = callbackConfigs
	return nil
}

// Execute executes all callbacks for a job result
func (scm *SimpleCallbackManager) Execute(ctx context.Context, result *Result) error {
	scm.mu.RLock()
	callbacks, exists := scm.callbacks[result.JobID]
	scm.mu.RUnlock()

	if !exists {
		return nil // No callbacks registered
	}

	var wg sync.WaitGroup
	for _, callback := range callbacks {
		wg.Add(1)
		go func(cb Callback) {
			defer wg.Done()
			scm.executeCallback(ctx, cb, result)
		}(callback)
	}

	wg.Wait()
	return nil
}

// Remove removes callbacks for a job
func (scm *SimpleCallbackManager) Remove(jobID string) error {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	delete(scm.callbacks, jobID)
	return nil
}

// executeCallback executes a single callback
func (scm *SimpleCallbackManager) executeCallback(ctx context.Context, callback Callback, result *Result) {
	for attempt := 0; attempt <= callback.Retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(delay)
		}

		if err := scm.sendWebhook(ctx, callback, result); err != nil {
			logger.Error("Callback failed",
				"callback_id", callback.ID,
				"job_id", result.JobID,
				"attempt", attempt+1,
				"error", err)
			continue
		}

		logger.Info("Callback executed successfully",
			"callback_id", callback.ID,
			"job_id", result.JobID,
			"attempt", attempt+1)
		return
	}

	logger.Error("Callback failed after all retries",
		"callback_id", callback.ID,
		"job_id", result.JobID,
		"max_retries", callback.Retries)
}

// sendWebhook sends a webhook callback
func (scm *SimpleCallbackManager) sendWebhook(ctx context.Context, callback Callback, result *Result) error {
	// Prepare payload
	payload := map[string]interface{}{
		"job_id":       result.JobID,
		"status":       result.Status,
		"worker_id":    result.WorkerID,
		"duration":     result.Duration.String(),
		"completed_at": result.CompletedAt,
		"metadata":     result.Metadata,
	}

	if result.Error != nil {
		payload["error"] = result.Error.Error()
	}

	if result.Proof != nil {
		payload["proof"] = map[string]interface{}{
			"id":           result.Proof.ID,
			"circuit_id":   result.Proof.CircuitID,
			"proof_type":   result.Proof.ProofType,
			"generated_at": result.Proof.GeneratedAt,
			"proof_size":   len(result.Proof.ProofData),
		}
	}

	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callback payload: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", callback.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create callback request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MMAD-Worker/1.0")
	for key, value := range callback.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := scm.client.Do(req)
	if err != nil {
		return fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("callback returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
