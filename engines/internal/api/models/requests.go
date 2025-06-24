package models

import (
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// ProofRequest represents a proof generation request
type ProofRequest struct {
	CircuitID     string                 `json:"circuit_id" binding:"required"`
	ProofType     types.ProofType        `json:"proof_type" binding:"required"`
	PublicInputs  map[string]interface{} `json:"public_inputs" binding:"required"`
	PrivateInputs map[string]interface{} `json:"private_inputs" binding:"required"`
	UserID        string                 `json:"user_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	Priority      int                    `json:"priority,omitempty"` // 1-10, 10 = highest
	Timeout       time.Duration          `json:"timeout,omitempty"`
}

// VerifyRequest represents a proof verification request
type VerifyRequest struct {
	ProofData    string                 `json:"proof_data" binding:"required"`
	PublicInputs map[string]interface{} `json:"public_inputs" binding:"required"`
	CircuitID    string                 `json:"circuit_id" binding:"required"`
	VerifyingKey string                 `json:"verifying_key,omitempty"`
}

// BatchProofRequest represents multiple proof requests
type BatchProofRequest struct {
	Requests []ProofRequest `json:"requests" binding:"required,min=1,max=100"`
	BatchID  string         `json:"batch_id,omitempty"`
	Parallel bool           `json:"parallel,omitempty"`
	MaxWait  time.Duration  `json:"max_wait,omitempty"`
}

// BatchVerifyRequest represents multiple verification requests
type BatchVerifyRequest struct {
	Requests []VerifyRequest `json:"requests" binding:"required,min=1,max=100"`
	BatchID  string          `json:"batch_id,omitempty"`
}

// CompileRequest represents a circuit compilation request
type CompileRequest struct {
	CircuitID string `json:"circuit_id" binding:"required"`
	Force     bool   `json:"force,omitempty"` // Force recompilation
}
