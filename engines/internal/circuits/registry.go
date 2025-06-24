package circuits

import (
	"context"
	"fmt"
	"sync"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/signature/eddsa"
)

// CircuitRegistry manages available circuits
type CircuitRegistry struct {
	circuits map[string]CircuitFactory
	compiler *CircuitCompiler
	mu       sync.RWMutex
}

// CircuitFactory creates circuit instances
type CircuitFactory func() frontend.Circuit

// CircuitInfo contains metadata about a circuit
type CircuitInfo struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Version        string          `json:"version"`
	ProofType      types.ProofType `json:"proof_type"`
	PublicInputs   []InputSpec     `json:"public_inputs"`
	PrivateInputs  []InputSpec     `json:"private_inputs"`
	EstimatedTime  string          `json:"estimated_time"`
	MaxConstraints int             `json:"max_constraints"`
	RequiredMemory string          `json:"required_memory"`
}

// InputSpec describes an input parameter
type InputSpec struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	MinValue    *string `json:"min_value,omitempty"`
	MaxValue    *string `json:"max_value,omitempty"`
}

// NewCircuitRegistry creates a new circuit registry
func NewCircuitRegistry(compiler *CircuitCompiler) *CircuitRegistry {
	registry := &CircuitRegistry{
		circuits: make(map[string]CircuitFactory),
		compiler: compiler,
	}

	// Register built-in circuits
	registry.registerBuiltinCircuits()

	return registry
}

// RegisterCircuit registers a new circuit
func (r *CircuitRegistry) RegisterCircuit(id string, factory CircuitFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.circuits[id]; exists {
		return fmt.Errorf("circuit already registered: %s", id)
	}

	r.circuits[id] = factory
	return nil
}

// GetCircuit creates a circuit instance
func (r *CircuitRegistry) GetCircuit(id string) (frontend.Circuit, error) {
	r.mu.RLock()
	factory, exists := r.circuits[id]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("circuit not found: %s", id)
	}

	return factory(), nil
}

// GetCompiledCircuit gets or compiles a circuit
func (r *CircuitRegistry) GetCompiledCircuit(ctx context.Context, id string) (*CompiledCircuit, error) {
	// Try to get from cache first
	if compiled, err := r.compiler.GetCircuit(id); err == nil {
		return compiled, nil
	}

	// Create circuit instance
	circuit, err := r.GetCircuit(id)
	if err != nil {
		return nil, err
	}

	// Compile it
	return r.compiler.CompileCircuit(ctx, id, circuit)
}

// ListCircuits returns information about all registered circuits
func (r *CircuitRegistry) ListCircuits() []CircuitInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]CircuitInfo, 0, len(r.circuits))

	for id := range r.circuits {
		info := r.getCircuitInfo(id)
		infos = append(infos, info)
	}

	return infos
}

// GetCircuitInfo returns information about a specific circuit
func (r *CircuitRegistry) GetCircuitInfo(id string) (CircuitInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, exists := r.circuits[id]; !exists {
		return CircuitInfo{}, fmt.Errorf("circuit not found: %s", id)
	}

	return r.getCircuitInfo(id), nil
}

// Private methods

func (r *CircuitRegistry) registerBuiltinCircuits() {
	// Balance proof circuits
	r.circuits["balance_v1"] = func() frontend.Circuit {
		return &BalanceCircuit{}
	}

	r.circuits["balance_range_v1"] = func() frontend.Circuit {
		return &BalanceRangeCircuit{}
	}

	r.circuits["multi_balance_v1"] = func() frontend.Circuit {
		return &MultiBalanceCircuit{
			Thresholds:   make([]frontend.Variable, 10),
			UserIDs:      make([]frontend.Variable, 10),
			Balances:     make([]frontend.Variable, 10),
			Salts:        make([]frontend.Variable, 10),
			MerkleProofs: make([][]frontend.Variable, 10),
		}
	}

	// Solvency proof circuits
	r.circuits["solvency_v1"] = func() frontend.Circuit {
		return &SolvencyCircuit{
			AssetCommitments: make([]frontend.Variable, 5),
			LiabilityProofs:  make([][]frontend.Variable, 100),
			LiabilityLeaves:  make([]frontend.Variable, 100),
		}
	}

	r.circuits["batch_solvency_v1"] = func() frontend.Circuit {
		return &BatchSolvencyCircuit{
			MerkleRoots:      make([]frontend.Variable, 10),
			Timestamps:       make([]frontend.Variable, 10),
			TotalAssets:      make([]frontend.Variable, 10),
			TotalLiabilities: make([]frontend.Variable, 10),
			Nonces:           make([]frontend.Variable, 10),
		}
	}

	// Merkle proof circuits
	r.circuits["merkle_inclusion_v1"] = func() frontend.Circuit {
		return &MerkleInclusionCircuit{
			Path:       make([]frontend.Variable, 32),
			Directions: make([]frontend.Variable, 32),
		}
	}

	r.circuits["merkle_non_inclusion_v1"] = func() frontend.Circuit {
		return &MerkleNonInclusionCircuit{
			WitnessPath: make([]frontend.Variable, 32),
			Directions:  make([]frontend.Variable, 32),
		}
	}

	r.circuits["sparse_merkle_v1"] = func() frontend.Circuit {
		return &SparseMerkleInclusionCircuit{
			Path:     make([]frontend.Variable, 32),
			Siblings: make([]frontend.Variable, 32),
		}
	}

	// Aggregate circuits
	r.circuits["aggregate_balance_v1"] = func() frontend.Circuit {
		return &AggregateBalanceCircuit{
			Balances:     make([]frontend.Variable, 100),
			UserIDs:      make([]frontend.Variable, 100),
			Salts:        make([]frontend.Variable, 100),
			MerkleProofs: make([][]frontend.Variable, 100),
		}
	}

	r.circuits["portfolio_v1"] = func() frontend.Circuit {
		return &PortfolioCircuit{
			AssetAmounts: make([]frontend.Variable, 50),
			AssetPrices:  make([]frontend.Variable, 50),
			RiskWeights:  make([]frontend.Variable, 50),
		}
	}

	// Signature circuits
	r.circuits["signature_v1"] = func() frontend.Circuit {
		return &SignatureCircuit{}
	}

	r.circuits["multi_signature_v1"] = func() frontend.Circuit {
		return &MultiSignatureCircuit{
			PublicKeys: make([]eddsa.PublicKey, 10),
			Messages:   make([]frontend.Variable, 10),
			Signatures: make([]eddsa.Signature, 10),
		}
	}

	// Arithmetic circuits
	r.circuits["private_arithmetic_v1"] = func() frontend.Circuit {
		return &PrivateArithmeticCircuit{}
	}

	r.circuits["comparison_v1"] = func() frontend.Circuit {
		return &ComparisonCircuit{}
	}

	r.circuits["range_proof_v1"] = func() frontend.Circuit {
		return &RangeProofCircuit{
			Bits: make([]frontend.Variable, 64),
		}
	}
}

func (r *CircuitRegistry) getCircuitInfo(id string) CircuitInfo {
	// This would typically be stored in a database or config file
	// For now, we'll return hardcoded info based on circuit ID

	switch id {
	case "balance_v1":
		return CircuitInfo{
			ID:          id,
			Name:        "Balance Proof v1",
			Description: "Proves that an account balance meets a minimum threshold without revealing the actual balance",
			Version:     "1.0.0",
			ProofType:   types.ProofTypeBalance,
			PublicInputs: []InputSpec{
				{Name: "threshold", Type: "field", Description: "Minimum required balance", Required: true},
				{Name: "user_id", Type: "field", Description: "User identifier hash", Required: true},
				{Name: "nonce", Type: "field", Description: "Replay protection nonce", Required: true},
				{Name: "timestamp", Type: "field", Description: "Proof generation timestamp", Required: true},
				{Name: "account_commitment", Type: "field", Description: "Account commitment hash", Required: true},
			},
			PrivateInputs: []InputSpec{
				{Name: "balance", Type: "field", Description: "Actual account balance", Required: true, MinValue: stringPtr("0")},
				{Name: "salt", Type: "field", Description: "Random salt for privacy", Required: true},
			},
			EstimatedTime:  "5-15 seconds",
			MaxConstraints: 1000,
			RequiredMemory: "256MB",
		}

	case "solvency_v1":
		return CircuitInfo{
			ID:          id,
			Name:        "Solvency Proof v1",
			Description: "Proves that total assets exceed total liabilities by a minimum ratio",
			Version:     "1.0.0",
			ProofType:   types.ProofTypeSolvency,
			PublicInputs: []InputSpec{
				{Name: "merkle_root", Type: "field", Description: "Root of liability Merkle tree", Required: true},
				{Name: "timestamp", Type: "field", Description: "Proof generation timestamp", Required: true},
				{Name: "min_solvency_ratio", Type: "field", Description: "Minimum required solvency ratio", Required: true, MinValue: stringPtr("1")},
			},
			PrivateInputs: []InputSpec{
				{Name: "total_assets", Type: "field", Description: "Sum of all assets", Required: true, MinValue: stringPtr("0")},
				{Name: "total_liabilities", Type: "field", Description: "Sum of all liabilities", Required: true, MinValue: stringPtr("0")},
				{Name: "asset_commitments", Type: "field[]", Description: "Asset commitment proofs", Required: true},
				{Name: "liability_proofs", Type: "field[][]", Description: "Merkle inclusion proofs for liabilities", Required: true},
				{Name: "liability_leaves", Type: "field[]", Description: "Individual liability amounts", Required: true},
				{Name: "nonce", Type: "field", Description: "Random nonce", Required: true},
			},
			EstimatedTime:  "30-90 seconds",
			MaxConstraints: 50000,
			RequiredMemory: "1GB",
		}

	case "merkle_inclusion_v1":
		return CircuitInfo{
			ID:          id,
			Name:        "Merkle Inclusion Proof v1",
			Description: "Proves that a value is included in a Merkle tree",
			Version:     "1.0.0",
			ProofType:   types.ProofTypeInclusion,
			PublicInputs: []InputSpec{
				{Name: "root", Type: "field", Description: "Merkle tree root hash", Required: true},
				{Name: "leaf_index", Type: "field", Description: "Index of the leaf", Required: false},
			},
			PrivateInputs: []InputSpec{
				{Name: "leaf_value", Type: "field", Description: "Value of the leaf", Required: true},
				{Name: "path", Type: "field[]", Description: "Sibling hashes on path to root", Required: true},
				{Name: "directions", Type: "field[]", Description: "Direction bits (0=left, 1=right)", Required: true},
			},
			EstimatedTime:  "1-5 seconds",
			MaxConstraints: 5000,
			RequiredMemory: "128MB",
		}

	default:
		return CircuitInfo{
			ID:          id,
			Name:        "Unknown Circuit",
			Description: "Circuit information not available",
			Version:     "unknown",
		}
	}
}

func stringPtr(s string) *string {
	return &s
}
