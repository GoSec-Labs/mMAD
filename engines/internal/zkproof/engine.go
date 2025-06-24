package zkproof

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/database/repository"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

// Engine is the main ZK proof engine
type Engine struct {
	// Dependencies
	repo           repository.Manager
	circuitManager CircuitManager
	generator      ProofGenerator
	verifier       ProofVerifier

	// Configuration
	config *Config

	// State management
	mu           sync.RWMutex
	activeProofs map[string]*GenerationProgress

	// Workers
	workerPool *WorkerPool

	// Metrics
	metrics *Metrics
}

// Config contains engine configuration
type Config struct {
	// Worker configuration
	MaxWorkers    int           `json:"max_workers"`
	WorkerTimeout time.Duration `json:"worker_timeout"`

	// Proof configuration
	DefaultTTL   time.Duration `json:"default_ttl"`
	MaxProofSize int64         `json:"max_proof_size"`
	MaxBatchSize int           `json:"max_batch_size"`

	// Circuit configuration
	CircuitCacheSize int           `json:"circuit_cache_size"`
	CircuitCacheTTL  time.Duration `json:"circuit_cache_ttl"`

	// Performance tuning
	EnableBatching    bool          `json:"enable_batching"`
	BatchTimeout      time.Duration `json:"batch_timeout"`
	EnableCompression bool          `json:"enable_compression"`
}

// NewEngine creates a new ZK proof engine
func NewEngine(
	repo repository.Manager,
	circuitManager CircuitManager,
	generator ProofGenerator,
	verifier ProofVerifier,
	config *Config,
) *Engine {
	return &Engine{
		repo:           repo,
		circuitManager: circuitManager,
		generator:      generator,
		verifier:       verifier,
		config:         config,
		activeProofs:   make(map[string]*GenerationProgress),
		workerPool:     NewWorkerPool(config.MaxWorkers),
		metrics:        NewMetrics(),
	}
}

// Start starts the proof engine
func (e *Engine) Start(ctx context.Context) error {
	logger.Info("Starting ZK proof engine",
		"max_workers", e.config.MaxWorkers,
		"default_ttl", e.config.DefaultTTL)

	// Start worker pool
	if err := e.workerPool.Start(ctx); err != nil {
		return fmt.Errorf("failed to start worker pool: %w", err)
	}

	// Start background tasks
	go e.runCleanupWorker(ctx)
	go e.runMetricsWorker(ctx)

	return nil
}

// Stop stops the proof engine
func (e *Engine) Stop(ctx context.Context) error {
	logger.Info("Stopping ZK proof engine")

	// Stop worker pool
	e.workerPool.Stop()

	// Cancel active proofs
	e.mu.Lock()
	for proofID := range e.activeProofs {
		e.generator.Cancel(ctx, proofID)
	}
	e.mu.Unlock()

	return nil
}

// GenerateProof generates a new zero-knowledge proof
func (e *Engine) GenerateProof(ctx context.Context, req *types.ProofRequest) (*types.ZKProof, error) {
	start := time.Now()

	// Validate request
	if err := e.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create proof record
	proof := &types.ZKProof{
		ID:             generateID(),
		Type:           req.Type,
		Status:         types.ProofStatusPending,
		PublicInputs:   req.PublicInputs,
		PrivateWitness: req.PrivateInputs,
		UserID:         req.UserID,
		AccountID:      req.AccountID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Set expiration
	if req.Options.ExpiresIn > 0 {
		expiresAt := time.Now().Add(req.Options.ExpiresIn)
		proof.ExpiresAt = &expiresAt
	} else {
		expiresAt := time.Now().Add(e.config.DefaultTTL)
		proof.ExpiresAt = &expiresAt
	}

	// Store proof record
	if err := e.repo.Proofs().Create(ctx, proof.ToModel()); err != nil {
		return nil, fmt.Errorf("failed to store proof: %w", err)
	}

	// Update metrics
	e.metrics.ProofRequests.WithLabelValues(string(req.Type)).Inc()

	// Generate proof
	proof.Status = types.ProofStatusGenerating
	e.repo.Proofs().UpdateStatus(ctx, proof.ID, models.ProofStatusGenerating)

	// Track progress
	progress := &GenerationProgress{
		ProofID:   proof.ID,
		Status:    types.ProofStatusGenerating,
		Progress:  0.0,
		Stage:     "initializing",
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	e.mu.Lock()
	e.activeProofs[proof.ID] = progress
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		delete(e.activeProofs, proof.ID)
		e.mu.Unlock()
	}()

	// Generate the actual proof
	generatedProof, err := e.generator.Generate(ctx, req)
	if err != nil {
		proof.MarkAsFailed(err.Error())
		e.repo.Proofs().UpdateStatus(ctx, proof.ID, models.ProofStatusFailed)
		e.metrics.ProofFailures.WithLabelValues(string(req.Type)).Inc()
		return nil, fmt.Errorf("proof generation failed: %w", err)
	}

	// Update proof with generated data
	proof.Proof = generatedProof.Proof
	proof.VerificationKey = generatedProof.VerificationKey
	proof.CircuitHash = generatedProof.CircuitHash
	proof.MarkAsGenerated(generatedProof.Proof, generatedProof.VerificationKey, time.Since(start))

	// Update in database
	if err := e.repo.Proofs().Update(ctx, proof.ToModel()); err != nil {
		logger.Error("Failed to update proof after generation", "error", err, "proof_id", proof.ID)
	}

	// Update metrics
	e.metrics.ProofGenerationDuration.WithLabelValues(string(req.Type)).Observe(time.Since(start).Seconds())
	e.metrics.ProofGenerated.WithLabelValues(string(req.Type)).Inc()

	logger.Info("Proof generated successfully",
		"proof_id", proof.ID,
		"type", req.Type,
		"duration", time.Since(start))

	return proof, nil
}

// GenerateProofAsync generates a proof asynchronously
func (e *Engine) GenerateProofAsync(ctx context.Context, req *types.ProofRequest) (string, error) {
	// Submit to worker pool
	proofID := generateID()

	job := &ProofJob{
		ID:      proofID,
		Request: req,
		Engine:  e,
	}

	if err := e.workerPool.Submit(job); err != nil {
		return "", fmt.Errorf("failed to submit proof job: %w", err)
	}

	return proofID, nil
}

// VerifyProof verifies a zero-knowledge proof
func (e *Engine) VerifyProof(ctx context.Context, req *types.VerificationRequest) (*types.VerificationResult, error) {
	start := time.Now()

	// Validate request
	if req.Proof == nil {
		return nil, fmt.Errorf("proof data is required")
	}

	if req.VerificationKey == nil {
		return nil, fmt.Errorf("verification key is required")
	}

	// Verify the proof
	result, err := e.verifier.Verify(ctx, req)
	if err != nil {
		e.metrics.VerificationFailures.Inc()
		return nil, fmt.Errorf("verification failed: %w", err)
	}

	// Update metrics
	e.metrics.VerificationDuration.Observe(time.Since(start).Seconds())
	if result.Valid {
		e.metrics.ValidProofs.Inc()
	} else {
		e.metrics.InvalidProofs.Inc()
	}

	// Update proof record if proof ID provided
	if req.ProofID != "" {
		if result.Valid {
			e.repo.Proofs().UpdateStatus(ctx, req.ProofID, models.ProofStatusVerified)
		}
	}

	return result, nil
}

// GetProof retrieves a proof by ID
func (e *Engine) GetProof(ctx context.Context, proofID string) (*types.ZKProof, error) {
	proofModel, err := e.repo.Proofs().GetByID(ctx, proofID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof: %w", err)
	}

	return types.ProofFromModel(proofModel), nil
}

// GetProgress returns the generation progress of a proof
func (e *Engine) GetProgress(ctx context.Context, proofID string) (*GenerationProgress, error) {
	e.mu.RLock()
	progress, exists := e.activeProofs[proofID]
	e.mu.RUnlock()

	if !exists {
		// Check if proof exists in database
		proof, err := e.GetProof(ctx, proofID)
		if err != nil {
			return nil, fmt.Errorf("proof not found: %s", proofID)
		}

		// Return synthetic progress based on proof status
		return &GenerationProgress{
			ProofID:   proofID,
			Status:    proof.Status,
			Progress:  e.statusToProgress(proof.Status),
			Stage:     string(proof.Status),
			StartedAt: proof.CreatedAt,
			UpdatedAt: proof.UpdatedAt,
		}, nil
	}

	return progress, nil
}

// ListProofs lists proofs with filters
func (e *Engine) ListProofs(ctx context.Context, filters repository.ProofFilters) ([]*types.ZKProof, error) {
	proofModels, err := e.repo.Proofs().List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list proofs: %w", err)
	}

	proofs := make([]*types.ZKProof, len(proofModels))
	for i, model := range proofModels {
		proofs[i] = types.ProofFromModel(model)
	}

	return proofs, nil
}

// Helper methods

func (e *Engine) validateRequest(req *types.ProofRequest) error {
	if req.Type == "" {
		return fmt.Errorf("proof type is required")
	}

	if req.PublicInputs == nil {
		return fmt.Errorf("public inputs are required")
	}

	if req.PrivateInputs == nil {
		return fmt.Errorf("private inputs are required")
	}

	// Validate with verifier
	return e.verifier.ValidatePublicInputs(req.Type, req.PublicInputs)
}

func (e *Engine) statusToProgress(status types.ProofStatus) float64 {
	switch status {
	case types.ProofStatusPending:
		return 0.0
	case types.ProofStatusGenerating:
		return 0.5
	case types.ProofStatusGenerated, types.ProofStatusVerified:
		return 1.0
	case types.ProofStatusFailed, types.ProofStatusExpired:
		return 0.0
	default:
		return 0.0
	}
}

func (e *Engine) runCleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if deleted, err := e.repo.Proofs().CleanupExpired(ctx); err != nil {
				logger.Error("Failed to cleanup expired proofs", "error", err)
			} else if deleted > 0 {
				logger.Info("Cleaned up expired proofs", "count", deleted)
			}
		}
	}
}

func (e *Engine) runMetricsWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.updateMetrics(ctx)
		}
	}
}

func (e *Engine) updateMetrics(ctx context.Context) {
	// Update active proofs count
	e.mu.RLock()
	activeCount := len(e.activeProofs)
	e.mu.RUnlock()

	e.metrics.ActiveProofs.Set(float64(activeCount))

	// Update other metrics from database
	// TODO: Add more detailed metrics collection
}

func generateID() string {
	// Generate a unique ID for the proof
	// You can use UUID, nanoid, or custom implementation
	return fmt.Sprintf("proof_%d", time.Now().UnixNano())
}
