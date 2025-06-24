package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/models"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BatchGenerateProofs generates multiple proofs
func (h *Handler) BatchGenerateProofs(c *gin.Context) {
	var req models.BatchProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid batch request format",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Generate batch ID
	if req.BatchID == "" {
		req.BatchID = "batch_" + uuid.New().String()
	}

	logger.Info("Starting batch proof generation",
		"batch_id", req.BatchID,
		"count", len(req.Requests),
		"parallel", req.Parallel)

	start := time.Now()
	results := make([]models.ProofResponse, len(req.Requests))

	if req.Parallel {
		// Process in parallel
		results = h.processBatchParallel(c.Request.Context(), req.Requests)
	} else {
		// Process sequentially
		results = h.processBatchSequential(c.Request.Context(), req.Requests)
	}

	duration := time.Since(start)

	// Count successes and failures
	completed := 0
	failed := 0
	for _, result := range results {
		if result.Status == "generated" {
			completed++
		} else {
			failed++
		}
	}

	response := models.BatchResponse{
		BatchID:     req.BatchID,
		Status:      "completed",
		Total:       len(req.Requests),
		Completed:   completed,
		Failed:      failed,
		Results:     results,
		StartedAt:   start,
		CompletedAt: time.Now(),
		Duration:    duration,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// BatchVerifyProofs verifies multiple proofs
func (h *Handler) BatchVerifyProofs(c *gin.Context) {
	var req models.BatchVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid batch verify request format",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	if req.BatchID == "" {
		req.BatchID = "verify_" + uuid.New().String()
	}

	start := time.Now()
	results := make([]models.VerifyResponse, len(req.Requests))

	// Process verifications in parallel (they're typically fast)
	var wg sync.WaitGroup
	for i, verifyReq := range req.Requests {
		wg.Add(1)
		go func(idx int, vReq models.VerifyRequest) {
			defer wg.Done()

			verifyStart := time.Now()

			// Create verification request
			verifyRequest := &types.VerifyRequest{
				ProofData:    vReq.ProofData,
				PublicInputs: vReq.PublicInputs,
				CircuitID:    vReq.CircuitID,
				VerifyingKey: vReq.VerifyingKey,
			}

			// Verify proof
			result, err := h.prover.VerifyProof(c.Request.Context(), verifyRequest)

			verifyDuration := time.Since(verifyStart)

			if err != nil {
				results[idx] = models.VerifyResponse{
					Valid:      false,
					VerifiedAt: time.Now(),
					Duration:   verifyDuration,
					CircuitID:  vReq.CircuitID,
				}
			} else {
				results[idx] = models.VerifyResponse{
					Valid:      result.Valid,
					VerifiedAt: time.Now(),
					Duration:   verifyDuration,
					CircuitID:  vReq.CircuitID,
				}
			}
		}(i, verifyReq)
	}

	wg.Wait()
	duration := time.Since(start)

	// Count results
	valid := 0
	invalid := 0
	for _, result := range results {
		if result.Valid {
			valid++
		} else {
			invalid++
		}
	}

	response := models.BatchResponse{
		BatchID:     req.BatchID,
		Status:      "completed",
		Total:       len(req.Requests),
		Completed:   valid,
		Failed:      invalid,
		Results:     results,
		StartedAt:   start,
		CompletedAt: time.Now(),
		Duration:    duration,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetBatchStatus returns batch operation status
func (h *Handler) GetBatchStatus(c *gin.Context) {
	batchID := c.Param("id")
	if batchID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_BATCH_ID",
				Message: "Batch ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// In a real implementation, you'd store batch status
	// For now, return a mock response
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"batch_id": batchID,
			"status":   "completed",
			"message":  "Batch operation completed",
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// Helper functions

func (h *Handler) processBatchParallel(ctx context.Context, requests []models.ProofRequest) []models.ProofResponse {
	results := make([]models.ProofResponse, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(idx int, proofReq models.ProofRequest) {
			defer wg.Done()

			if proofReq.RequestID == "" {
				proofReq.RequestID = uuid.New().String()
			}

			start := time.Now()

			// Create proof request
			request := &types.ProofRequest{
				ID:            proofReq.RequestID,
				CircuitID:     proofReq.CircuitID,
				ProofType:     proofReq.ProofType,
				PublicInputs:  proofReq.PublicInputs,
				PrivateInputs: proofReq.PrivateInputs,
				UserID:        proofReq.UserID,
				Priority:      proofReq.Priority,
				Timeout:       proofReq.Timeout,
			}

			// Generate proof
			proof, err := h.prover.GenerateProof(ctx, request)
			duration := time.Since(start)

			if err != nil {
				results[idx] = models.ProofResponse{
					ProofID:   proofReq.RequestID,
					Status:    "failed",
					ProofType: proofReq.ProofType,
					CircuitID: proofReq.CircuitID,
					Duration:  duration,
				}
			} else {
				results[idx] = models.ProofResponse{
					ProofID:      proof.ID,
					Status:       "generated",
					ProofData:    proof.ProofData,
					ProofType:    proof.ProofType,
					CircuitID:    proof.CircuitID,
					GeneratedAt:  proof.GeneratedAt,
					Duration:     duration,
					ProofSize:    len(proof.ProofData),
					VerifyingKey: proof.VerifyingKey,
				}
			}
		}(i, req)
	}

	wg.Wait()
	return results
}

func (h *Handler) processBatchSequential(ctx context.Context, requests []models.ProofRequest) []models.ProofResponse {
	results := make([]models.ProofResponse, len(requests))

	for i, req := range requests {
		if req.RequestID == "" {
			req.RequestID = uuid.New().String()
		}

		start := time.Now()

		// Create proof request
		request := &types.ProofRequest{
			ID:            req.RequestID,
			CircuitID:     req.CircuitID,
			ProofType:     req.ProofType,
			PublicInputs:  req.PublicInputs,
			PrivateInputs: req.PrivateInputs,
			UserID:        req.UserID,
			Priority:      req.Priority,
			Timeout:       req.Timeout,
		}

		// Generate proof
		proof, err := h.prover.GenerateProof(ctx, request)
		duration := time.Since(start)

		if err != nil {
			results[i] = models.ProofResponse{
				ProofID:   req.RequestID,
				Status:    "failed",
				ProofType: req.ProofType,
				CircuitID: req.CircuitID,
				Duration:  duration,
			}
		} else {
			results[i] = models.ProofResponse{
				ProofID:      proof.ID,
				Status:       "generated",
				ProofData:    proof.ProofData,
				ProofType:    proof.ProofType,
				CircuitID:    proof.CircuitID,
				GeneratedAt:  proof.GeneratedAt,
				Duration:     duration,
				ProofSize:    len(proof.ProofData),
				VerifyingKey: proof.VerifyingKey,
			}
		}
	}

	return results
}
