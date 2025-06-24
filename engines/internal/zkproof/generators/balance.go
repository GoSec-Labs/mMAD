package generators

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// BalanceProofGenerator generates proofs that balance >= threshold without revealing balance
type BalanceProofGenerator struct {
	circuitManager CircuitManager
}

// NewBalanceProofGenerator creates a new balance proof generator
func NewBalanceProofGenerator(cm CircuitManager) *BalanceProofGenerator {
	return &BalanceProofGenerator{
		circuitManager: cm,
	}
}

// Generate generates a balance proof
func (g *BalanceProofGenerator) Generate(ctx context.Context, req *types.ProofRequest) (*types.ZKProof, error) {
	// Validate inputs
	threshold, ok := req.PublicInputs["threshold"]
	if !ok {
		return nil, fmt.Errorf("threshold is required")
	}

	balance, ok := req.PrivateInputs["balance"]
	if !ok {
		return nil, fmt.Errorf("balance is required")
	}

	// Convert to big.Int for circuit
	thresholdBig, err := g.toBigInt(threshold)
	if err != nil {
		return nil, fmt.Errorf("invalid threshold: %w", err)
	}

	balanceBig, err := g.toBigInt(balance)
	if err != nil {
		return nil, fmt.Errorf("invalid balance: %w", err)
	}

	// Get circuit
	circuit, err := g.circuitManager.GetCircuit("balance_v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	// Prepare witness
	witness, err := circuit.GenerateWitness(
		map[string]interface{}{
			"threshold": thresholdBig,
		},
		map[string]interface{}{
			"balance": balanceBig,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate witness: %w", err)
	}

	// Generate proof using the circuit
	// Generate proof using the circuit
	proof, vk, err := g.generateGroth16Proof(circuit, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate groth16 proof: %w", err)
	}

	// Create ZK proof object
	zkProof := &types.ZKProof{
		ID:              generateProofID(),
		Type:            types.ProofTypeBalance,
		Status:          types.ProofStatusGenerated,
		CircuitID:       circuit.ID(),
		CircuitHash:     circuit.Hash(),
		PublicInputs:    req.PublicInputs,
		Proof:           proof,
		VerificationKey: vk,
		UserID:          req.UserID,
		AccountID:       req.AccountID,
	}

	return zkProof, nil
}

// SupportedTypes returns supported proof types
func (g *BalanceProofGenerator) SupportedTypes() []types.ProofType {
	return []types.ProofType{types.ProofTypeBalance}
}

// EstimateTime estimates proof generation time
func (g *BalanceProofGenerator) EstimateTime(req *types.ProofRequest) (time.Duration, error) {
	// Balance proofs are relatively fast - estimate 5-15 seconds
	return 10 * time.Second, nil
}

// Helper methods

func (g *BalanceProofGenerator) toBigInt(value interface{}) (*big.Int, error) {
	switch v := value.(type) {
	case string:
		decimal, err := math.NewDecimalFromString(v)
		if err != nil {
			return nil, err
		}
		return decimal.BigInt(), nil
	case *math.Decimal:
		return v.BigInt(), nil
	case int64:
		return big.NewInt(v), nil
	case float64:
		decimal := math.NewDecimalFromFloat(v)
		return decimal.BigInt(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

func (g *BalanceProofGenerator) generateGroth16Proof(circuit Circuit, witness CircuitWitness) (*types.ProofData, *types.VerificationKey, error) {
	// This is where we'd integrate with gnark or other ZK libraries
	// For now, we'll create a mock implementation

	// In real implementation:
	// 1. Compile circuit to constraint system
	// 2. Setup proving/verification keys (if not cached)
	// 3. Generate witness
	// 4. Create Groth16 proof

	// Mock proof data
	proof := &types.ProofData{
		A: types.Point{
			X: "0x1234567890abcdef",
			Y: "0xfedcba0987654321",
		},
		B: types.Point2{
			X: [2]string{"0xaaaa", "0xbbbb"},
			Y: [2]string{"0xcccc", "0xdddd"},
		},
		C: types.Point{
			X: "0x1111111111111111",
			Y: "0x2222222222222222",
		},
		Hash: "0xproof_hash_here",
		Size: 256,
	}

	vk := &types.VerificationKey{
		Alpha: types.Point{
			X: "0xalpha_x",
			Y: "0xalpha_y",
		},
		Beta: types.Point2{
			X: [2]string{"0xbeta_x0", "0xbeta_x1"},
			Y: [2]string{"0xbeta_y0", "0xbeta_y1"},
		},
		Gamma: types.Point2{
			X: [2]string{"0xgamma_x0", "0xgamma_x1"},
			Y: [2]string{"0xgamma_y0", "0xgamma_y1"},
		},
		Delta: types.Point2{
			X: [2]string{"0xdelta_x0", "0xdelta_x1"},
			Y: [2]string{"0xdelta_y0", "0xdelta_y1"},
		},
		IC: []types.Point{
			{X: "0xic0_x", Y: "0xic0_y"},
			{X: "0xic1_x", Y: "0xic1_y"},
		},
		Hash: "0xvk_hash_here",
	}

	return proof, vk, nil
}
