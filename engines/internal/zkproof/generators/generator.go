package generators

import (
	"context"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
)

// ProofGenerator defines the interface for generating ZK proofs
type ProofGenerator interface {
	// Generate creates a new zero-knowledge proof
	Generate(ctx context.Context, req *types.ProofRequest) (*types.ZKProof, error)

	// GenerateAsync creates a proof asynchronously
	GenerateAsync(ctx context.Context, req *types.ProofRequest) (string, error)

	// GetProgress returns the progress of an async proof generation
	GetProgress(ctx context.Context, proofID string) (*GenerationProgress, error)

	// Cancel cancels an ongoing proof generation
	Cancel(ctx context.Context, proofID string) error

	// SupportedTypes returns the proof types this generator supports
	SupportedTypes() []types.ProofType

	// EstimateTime estimates how long proof generation will take
	EstimateTime(req *types.ProofRequest) (time.Duration, error)
}

// ProofVerifier defines the interface for verifying ZK proofs
type ProofVerifier interface {
	// Verify verifies a zero-knowledge proof
	Verify(ctx context.Context, req *types.VerificationRequest) (*types.VerificationResult, error)

	// VerifyBatch verifies multiple proofs at once
	VerifyBatch(ctx context.Context, reqs []*types.VerificationRequest) ([]*types.VerificationResult, error)

	// ValidatePublicInputs validates that public inputs are correct
	ValidatePublicInputs(proofType types.ProofType, inputs map[string]interface{}) error
}

// CircuitManager defines the interface for managing ZK circuits
type CircuitManager interface {
	// GetCircuit retrieves a circuit by ID
	GetCircuit(circuitID string) (Circuit, error)

	// ListCircuits lists available circuits
	ListCircuits() ([]CircuitInfo, error)

	// CompileCircuit compiles a new circuit
	CompileCircuit(source string, options CompileOptions) (*CircuitInfo, error)

	// GetVerificationKey gets the verification key for a circuit
	GetVerificationKey(circuitID string) (*types.VerificationKey, error)
}

// GenerationProgress tracks proof generation progress
type GenerationProgress struct {
	ProofID      string                 `json:"proof_id"`
	Status       types.ProofStatus      `json:"status"`
	Progress     float64                `json:"progress"` // 0.0 to 1.0
	Stage        string                 `json:"stage"`    // Current stage
	EstimatedETA time.Duration          `json:"estimated_eta"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	StartedAt    time.Time              `json:"started_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Circuit represents a ZK circuit
type Circuit interface {
	// ID returns the circuit identifier
	ID() string

	// Hash returns the circuit hash
	Hash() string

	// Define defines the circuit constraints
	Define(api frontend.API, witness CircuitWitness) error

	// Compile compiles the circuit to constraint system
	Compile() (constraint.ConstraintSystem, error)

	// GenerateWitness creates a witness from inputs
	GenerateWitness(publicInputs, privateInputs map[string]interface{}) (CircuitWitness, error)

	// ValidateInputs validates circuit inputs
	ValidateInputs(publicInputs, privateInputs map[string]interface{}) error
}

// CircuitWitness represents circuit witness data
type CircuitWitness interface {
	// PublicInputs returns the public inputs
	PublicInputs() map[string]interface{}

	// PrivateInputs returns the private inputs (secret witness)
	PrivateInputs() map[string]interface{}

	// Serialize serializes the witness
	Serialize() ([]byte, error)

	// Deserialize deserializes the witness
	Deserialize(data []byte) error
}

// CircuitInfo contains information about a circuit
type CircuitInfo struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Version         string          `json:"version"`
	Hash            string          `json:"hash"`
	Type            types.ProofType `json:"type"`
	Description     string          `json:"description"`
	Constraints     int             `json:"constraints"`
	PublicInputs    []InputSpec     `json:"public_inputs"`
	PrivateInputs   []InputSpec     `json:"private_inputs"`
	CompilationTime time.Duration   `json:"compilation_time"`
	CompiledAt      time.Time       `json:"compiled_at"`
}

// InputSpec specifies an input parameter
type InputSpec struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Validation  string      `json:"validation,omitempty"`
}

// CompileOptions contains circuit compilation options
type CompileOptions struct {
	OptimizationLevel int                    `json:"optimization_level"`
	TargetCurve       string                 `json:"target_curve"`
	Metadata          map[string]interface{} `json:"metadata"`
}
