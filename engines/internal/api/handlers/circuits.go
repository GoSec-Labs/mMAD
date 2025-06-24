package handlers

import (
	"net/http"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/api/models"
	"github.com/gin-gonic/gin"
)

// ListCircuits returns available circuits
func (h *Handler) ListCircuits(c *gin.Context) {
	circuits := h.circuits.ListCircuits()

	circuitInfos := make([]models.CircuitInfo, len(circuits))
	for i, circuit := range circuits {
		// Check if circuit is compiled
		_, err := h.circuits.GetCompiledCircuit(c.Request.Context(), circuit.ID)
		isCompiled := err == nil

		circuitInfos[i] = models.CircuitInfo{
			ID:            circuit.ID,
			Name:          circuit.Name,
			Description:   circuit.Description,
			Version:       circuit.Version,
			ProofType:     circuit.ProofType,
			Constraints:   circuit.MaxConstraints,
			IsCompiled:    isCompiled,
			EstimatedTime: circuit.EstimatedTime,
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"circuits": circuitInfos,
			"total":    len(circuitInfos),
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetCircuit returns specific circuit information
func (h *Handler) GetCircuit(c *gin.Context) {
	circuitID := c.Param("id")
	if circuitID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_CIRCUIT_ID",
				Message: "Circuit ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	// Get circuit info
	info := h.circuits.GetCircuitInfo(circuitID)

	// Check if compiled
	compiled, err := h.circuits.GetCompiledCircuit(c.Request.Context(), circuitID)
	isCompiled := err == nil

	circuitInfo := models.CircuitInfo{
		ID:            info.ID,
		Name:          info.Name,
		Description:   info.Description,
		Version:       info.Version,
		ProofType:     info.ProofType,
		Constraints:   info.MaxConstraints,
		IsCompiled:    isCompiled,
		EstimatedTime: info.EstimatedTime,
	}

	if isCompiled {
		circuitInfo.Constraints = compiled.NumConstraints
		circuitInfo.CompiledAt = compiled.CompiledAt
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success:   true,
		Data:      circuitInfo,
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// CompileCircuit compiles a circuit
func (h *Handler) CompileCircuit(c *gin.Context) {
	circuitID := c.Param("id")
	if circuitID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_CIRCUIT_ID",
				Message: "Circuit ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	var req models.CompileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use circuit ID from URL if no body provided
		req.CircuitID = circuitID
	}

	start := time.Now()

	// Compile circuit
	compiled, err := h.circuits.CompileCircuit(c.Request.Context(), req.CircuitID, req.Force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "COMPILATION_FAILED",
				Message: "Failed to compile circuit",
				Details: err.Error(),
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	duration := time.Since(start)

	// Emit event
	info := h.circuits.GetCircuitInfo(req.CircuitID)
	emitter := h.events.GetEmitter("api")
	emitter.EmitCircuitCompiled(req.CircuitID, info.Name, compiled.NumConstraints, duration)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"circuit_id":   req.CircuitID,
			"compiled":     true,
			"constraints":  compiled.NumConstraints,
			"compile_time": duration,
			"compiled_at":  compiled.CompiledAt,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}

// GetCircuitInfo returns detailed circuit information
func (h *Handler) GetCircuitInfo(c *gin.Context) {
	circuitID := c.Param("id")
	if circuitID == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: &models.APIError{
				Code:    "MISSING_CIRCUIT_ID",
				Message: "Circuit ID is required",
			},
			RequestID: getRequestID(c),
			Timestamp: time.Now(),
		})
		return
	}

	info := h.circuits.GetCircuitInfo(circuitID)

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"id":              info.ID,
			"name":            info.Name,
			"description":     info.Description,
			"version":         info.Version,
			"proof_type":      info.ProofType,
			"public_inputs":   info.PublicInputs,
			"private_inputs":  info.PrivateInputs,
			"max_constraints": info.MaxConstraints,
			"required_memory": info.RequiredMemory,
			"estimated_time":  info.EstimatedTime,
		},
		RequestID: getRequestID(c),
		Timestamp: time.Now(),
	})
}
