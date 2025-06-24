package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/models"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateProof generates a new ZK proof
func (h *Handler) GenerateProof(c *gin.Context) {
	var req models.ProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Generate proof ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Emit event
	emitter := h.events.GetEmitter("api")
	emitter.EmitProofRequested(req.RequestID, req.ProofType, req.UserID, getRequestID(c))

	// Start proof generation
	start := time.Now()

	// Create proof request for engine
	proofReq := &types.ProofRequest{
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
	proof, err := h.prover.GenerateProof(c.Request.Context(), proofReq)
	if err != nil {
		emitter.EmitProofFailed(req.RequestID, req.ProofType, err, req.UserID, getRequestID(c))

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PROOF_GENERATION_FAILED",
				Message: "Failed to generate proof",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	duration := time.Since(start)
	emitter.EmitProofGenerated(req.RequestID, req.ProofType, duration, len(proof.ProofData), req.UserID, getRequestID(c))

	response := models.ProofResponse{
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

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// VerifyProof verifies a ZK proof
func (h *Handler) VerifyProof(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request format",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	start := time.Now()

	// Create verification request
	verifyReq := &types.VerifyRequest{
		ProofData:    req.ProofData,
		PublicInputs: req.PublicInputs,
		CircuitID:    req.CircuitID,
		VerifyingKey: req.VerifyingKey,
	}

	// Verify proof
	result, err := h.prover.VerifyProof(c.Request.Context(), verifyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "VERIFICATION_FAILED",
				Message: "Failed to verify proof",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	duration := time.Since(start)

	// Emit event
	emitter := h.events.GetEmitter("api")
	emitter.EmitProofVerified("", types.ProofType(""), duration, "", getRequestID(c))

	response := models.VerifyResponse{
		Valid:      result.Valid,
		VerifiedAt: time.Now(),
		Duration:   duration,
		CircuitID:  req.CircuitID,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetProof retrieves proof information
func (h *Handler) GetProof(c *gin.Context) {
	proofID := c.Param("id")
	if proofID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PROOF_ID",
				Message: "Proof ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Get proof from storage/cache
	proof, err := h.prover.GetProof(c.Request.Context(), proofID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PROOF_NOT_FOUND",
				Message: fmt.Sprintf("Proof with ID %s not found", proofID),
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	response := models.ProofResponse{
		ProofID:      proof.ID,
		Status:       proof.Status,
		ProofData:    proof.ProofData,
		ProofType:    proof.ProofType,
		CircuitID:    proof.CircuitID,
		GeneratedAt:  proof.GeneratedAt,
		ProofSize:    len(proof.ProofData),
		VerifyingKey: proof.VerifyingKey,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      response,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetProofStatus returns proof generation status
func (h *Handler) GetProofStatus(c *gin.Context) {
	proofID := c.Param("id")
	if proofID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PROOF_ID",
				Message: "Proof ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Get status from proof engine
	status, err := h.prover.GetProofStatus(c.Request.Context(), proofID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "PROOF_NOT_FOUND",
				Message: fmt.Sprintf("Proof with ID %s not found", proofID),
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"proof_id": proofID,
			"status":   status,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// DeleteProof deletes a proof
func (h *Handler) DeleteProof(c *gin.Context) {
	proofID := c.Param("id")
	if proofID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_PROOF_ID",
				Message: "Proof ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Delete proof
	err := h.prover.DeleteProof(c.Request.Context(), proofID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete proof",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"proof_id": proofID,
			"deleted":  true,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}
