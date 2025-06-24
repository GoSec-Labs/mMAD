package models

import (
	"encoding/json"
	"time"
)

// ProofType represents the type of ZK proof
type ProofType string

const (
	ProofTypeReserve    ProofType = "reserve"
	ProofTypeSolvency   ProofType = "solvency"
	ProofTypeBalance    ProofType = "balance"
	ProofTypeCompliance ProofType = "compliance"
)

// ProofStatus represents proof verification status
type ProofStatus string

const (
	ProofStatusPending   ProofStatus = "pending"
	ProofStatusGenerated ProofStatus = "generated"
	ProofStatusVerified  ProofStatus = "verified"
	ProofStatusFailed    ProofStatus = "failed"
	ProofStatusExpired   ProofStatus = "expired"
)

// ZKProof represents a zero-knowledge proof
type ZKProof struct {
	ID               string                 `json:"id" db:"id"`
	Type             ProofType              `json:"type" db:"type"`
	Status           ProofStatus            `json:"status" db:"status"`
	UserID           *string                `json:"user_id" db:"user_id"`
	AccountID        *string                `json:"account_id" db:"account_id"`
	ProofData        json.RawMessage        `json:"proof_data" db:"proof_data"`
	PublicInputs     json.RawMessage        `json:"public_inputs" db:"public_inputs"`
	PrivateInputs    json.RawMessage        `json:"-" db:"private_inputs"` // Never expose
	CircuitHash      string                 `json:"circuit_hash" db:"circuit_hash"`
	VerificationKey  string                 `json:"verification_key" db:"verification_key"`
	ProofHash        string                 `json:"proof_hash" db:"proof_hash"`
	MerkleRoot       string                 `json:"merkle_root" db:"merkle_root"`
	BlockNumber      *int64                 `json:"block_number" db:"block_number"`
	Timestamp        time.Time              `json:"timestamp" db:"timestamp"`
	ExpiresAt        *time.Time             `json:"expires_at" db:"expires_at"`
	GeneratedAt      *time.Time             `json:"generated_at" db:"generated_at"`
	VerifiedAt       *time.Time             `json:"verified_at" db:"verified_at"`
	FailedAt         *time.Time             `json:"failed_at" db:"failed_at"`
	FailureReason    string                 `json:"failure_reason" db:"failure_reason"`
	GenerationTime   *time.Duration         `json:"generation_time" db:"generation_time"`
	VerificationTime *time.Duration         `json:"verification_time" db:"verification_time"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`

	// Relations
	User    *User    `json:"user,omitempty"`
	Account *Account `json:"account,omitempty"`
}

// IsValid checks if proof is valid and not expired
func (p *ZKProof) IsValid() bool {
	if p.Status != ProofStatusVerified {
		return false
	}

	if p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}

// IsExpired checks if proof has expired
func (p *ZKProof) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// MarkGenerated marks proof as generated
func (p *ZKProof) MarkGenerated(proofData, publicInputs json.RawMessage, generationTime time.Duration) {
	p.Status = ProofStatusGenerated
	p.ProofData = proofData
	p.PublicInputs = publicInputs
	p.GenerationTime = &generationTime
	now := time.Now()
	p.GeneratedAt = &now
	p.UpdatedAt = now
}

// MarkVerified marks proof as verified
func (p *ZKProof) MarkVerified(verificationTime time.Duration) {
	p.Status = ProofStatusVerified
	p.VerificationTime = &verificationTime
	now := time.Now()
	p.VerifiedAt = &now
	p.UpdatedAt = now
}

// MarkFailed marks proof as failed
func (p *ZKProof) MarkFailed(reason string) {
	p.Status = ProofStatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.FailedAt = &now
	p.UpdatedAt = now
}
