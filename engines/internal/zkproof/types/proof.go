package types

import (
	"time"
)

// ProofType defines the type of zero-knowledge proof
type ProofType string

const (
	ProofTypeBalance   ProofType = "balance"   // Prove balance > threshold
	ProofTypeSolvency  ProofType = "solvency"  // Prove assets >= liabilities
	ProofTypeReserve   ProofType = "reserve"   // Prove reserve backing
	ProofTypeInclusion ProofType = "inclusion" // Prove account in merkle tree
	ProofTypeRange     ProofType = "range"     // Prove value in range
)

// ProofStatus defines the current status of a proof
type ProofStatus string

const (
	ProofStatusPending    ProofStatus = "pending"
	ProofStatusGenerating ProofStatus = "generating"
	ProofStatusGenerated  ProofStatus = "generated"
	ProofStatusVerified   ProofStatus = "verified"
	ProofStatusFailed     ProofStatus = "failed"
	ProofStatusExpired    ProofStatus = "expired"
)

// ZKProof represents a zero-knowledge proof
type ZKProof struct {
	ID      string      `json:"id"`
	Type    ProofType   `json:"type"`
	Status  ProofStatus `json:"status"`
	Version string      `json:"version"`

	// Circuit information
	CircuitID   string `json:"circuit_id"`
	CircuitHash string `json:"circuit_hash"`

	// Public inputs (visible to verifiers)
	PublicInputs map[string]interface{} `json:"public_inputs"`

	// Private witness (secret inputs)
	PrivateWitness map[string]interface{} `json:"-"` // Never serialize

	// Generated proof data
	Proof *ProofData `json:"proof,omitempty"`

	// Verification key
	VerificationKey *VerificationKey `json:"verification_key,omitempty"`

	// Metadata
	GeneratedAt      *time.Time    `json:"generated_at,omitempty"`
	VerifiedAt       *time.Time    `json:"verified_at,omitempty"`
	ExpiresAt        *time.Time    `json:"expires_at,omitempty"`
	GenerationTime   time.Duration `json:"generation_time,omitempty"`
	VerificationTime time.Duration `json:"verification_time,omitempty"`

	// Context
	UserID      string  `json:"user_id,omitempty"`
	AccountID   string  `json:"account_id,omitempty"`
	BlockNumber *uint64 `json:"block_number,omitempty"`
	MerkleRoot  string  `json:"merkle_root,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ProofData contains the actual cryptographic proof
type ProofData struct {
	// Groth16 proof components
	A Point  `json:"a"` // [A]_1
	B Point2 `json:"b"` // [B]_2
	C Point  `json:"c"` // [C]_1

	// Additional proof metadata
	Hash string `json:"hash"` // Hash of the proof
	Size int    `json:"size"` // Proof size in bytes
}

// VerificationKey contains the circuit verification key
type VerificationKey struct {
	Alpha Point   `json:"alpha"` // [α]_1
	Beta  Point2  `json:"beta"`  // [β]_2
	Gamma Point2  `json:"gamma"` // [γ]_2
	Delta Point2  `json:"delta"` // [δ]_2
	IC    []Point `json:"ic"`    // [γ^-1 * (β * u_i(x) + α * v_i(x) + w_i(x))]_1
	Hash  string  `json:"hash"`  // VK hash
}

// Point represents a point on elliptic curve (G1)
type Point struct {
	X string `json:"x"`
	Y string `json:"y"`
}

// Point2 represents a point on elliptic curve (G2)
type Point2 struct {
	X [2]string `json:"x"` // [x0, x1]
	Y [2]string `json:"y"` // [y0, y1]
}

// ProofRequest represents a request to generate a proof
type ProofRequest struct {
	Type          ProofType              `json:"type"`
	UserID        string                 `json:"user_id,omitempty"`
	AccountID     string                 `json:"account_id,omitempty"`
	PublicInputs  map[string]interface{} `json:"public_inputs"`
	PrivateInputs map[string]interface{} `json:"private_inputs"`
	Options       ProofOptions           `json:"options,omitempty"`
}

// ProofOptions contains options for proof generation
type ProofOptions struct {
	Timeout   time.Duration          `json:"timeout,omitempty"`
	Priority  int                    `json:"priority,omitempty"`
	BatchWith []string               `json:"batch_with,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	ExpiresIn time.Duration          `json:"expires_in,omitempty"`
}

// VerificationRequest represents a request to verify a proof
type VerificationRequest struct {
	ProofID         string                 `json:"proof_id,omitempty"`
	Proof           *ProofData             `json:"proof"`
	PublicInputs    map[string]interface{} `json:"public_inputs"`
	VerificationKey *VerificationKey       `json:"verification_key"`
	CircuitHash     string                 `json:"circuit_hash"`
}

// VerificationResult contains the result of proof verification
type VerificationResult struct {
	Valid            bool                   `json:"valid"`
	ProofID          string                 `json:"proof_id,omitempty"`
	VerifiedAt       time.Time              `json:"verified_at"`
	VerificationTime time.Duration          `json:"verification_time"`
	Error            string                 `json:"error,omitempty"`
	Details          map[string]interface{} `json:"details,omitempty"`
}

// IsValid checks if the proof is valid and not expired
func (p *ZKProof) IsValid() bool {
	if p.Status != ProofStatusGenerated && p.Status != ProofStatusVerified {
		return false
	}

	if p.Proof == nil || p.VerificationKey == nil {
		return false
	}

	if p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt) {
		return false
	}

	return true
}

// IsExpired checks if the proof has expired
func (p *ZKProof) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

// GetPublicInput safely retrieves a public input
func (p *ZKProof) GetPublicInput(key string) (interface{}, bool) {
	val, exists := p.PublicInputs[key]
	return val, exists
}

// SetPublicInput safely sets a public input
func (p *ZKProof) SetPublicInput(key string, value interface{}) {
	if p.PublicInputs == nil {
		p.PublicInputs = make(map[string]interface{})
	}
	p.PublicInputs[key] = value
}

// MarkAsGenerated marks the proof as successfully generated
func (p *ZKProof) MarkAsGenerated(proof *ProofData, vk *VerificationKey, duration time.Duration) {
	now := time.Now()
	p.Status = ProofStatusGenerated
	p.Proof = proof
	p.VerificationKey = vk
	p.GeneratedAt = &now
	p.GenerationTime = duration
	p.UpdatedAt = now
}

// MarkAsVerified marks the proof as successfully verified
func (p *ZKProof) MarkAsVerified(duration time.Duration) {
	now := time.Now()
	p.Status = ProofStatusVerified
	p.VerifiedAt = &now
	p.VerificationTime = duration
	p.UpdatedAt = now
}

// MarkAsFailed marks the proof as failed
func (p *ZKProof) MarkAsFailed(reason string) {
	p.Status = ProofStatusFailed
	p.UpdatedAt = time.Now()
	if p.PublicInputs == nil {
		p.PublicInputs = make(map[string]interface{})
	}
	p.PublicInputs["failure_reason"] = reason
}
