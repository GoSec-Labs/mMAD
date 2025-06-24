package generators

import (
	"context"
	"fmt"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
)

// SolvencyProofGenerator generates proofs that total_assets >= total_liabilities
type SolvencyProofGenerator struct {
	circuitManager CircuitManager
}

// NewSolvencyProofGenerator creates a new solvency proof generator
func NewSolvencyProofGenerator(cm CircuitManager) *SolvencyProofGenerator {
	return &SolvencyProofGenerator{
		circuitManager: cm,
	}
}

// Generate generates a solvency proof
func (g *SolvencyProofGenerator) Generate(ctx context.Context, req *types.ProofRequest) (*types.ZKProof, error) {
	// Validate inputs
	merkleRoot, ok := req.PublicInputs["merkle_root"]
	if !ok {
		return nil, fmt.Errorf("merkle_root is required")
	}

	timestamp, ok := req.PublicInputs["timestamp"]
	if !ok {
		return nil, fmt.Errorf("timestamp is required")
	}

	// Private inputs
	totalAssets, ok := req.PrivateInputs["total_assets"]
	if !ok {
		return nil, fmt.Errorf("total_assets is required")
	}

	totalLiabilities, ok := req.PrivateInputs["total_liabilities"]
	if !ok {
		return nil, fmt.Errorf("total_liabilities is required")
	}

	assetProofs, ok := req.PrivateInputs["asset_proofs"]
	if !ok {
		return nil, fmt.Errorf("asset_proofs is required")
	}

	liabilityProofs, ok := req.PrivateInputs["liability_proofs"]
	if !ok {
		return nil, fmt.Errorf("liability_proofs is required")
	}

	// Convert values
	assetsDecimal, err := g.toDecimal(totalAssets)
	if err != nil {
		return nil, fmt.Errorf("invalid total_assets: %w", err)
	}

	liabilitiesDecimal, err := g.toDecimal(totalLiabilities)
	if err != nil {
		return nil, fmt.Errorf("invalid total_liabilities: %w", err)
	}

	// Validate solvency before generating proof
	if assetsDecimal.LessThan(liabilitiesDecimal) {
		return nil, fmt.Errorf("institution is insolvent: assets=%s < liabilities=%s",
			assetsDecimal.String(), liabilitiesDecimal.String())
	}

	// Get circuit
	circuit, err := g.circuitManager.GetCircuit("solvency_v1")
	if err != nil {
		return nil, fmt.Errorf("failed to get circuit: %w", err)
	}

	// Prepare witness
	witness, err := circuit.GenerateWitness(
		map[string]interface{}{
			"merkle_root": merkleRoot,
			"timestamp":   timestamp,
		},
		map[string]interface{}{
			"total_assets":      assetsDecimal.BigInt(),
			"total_liabilities": liabilitiesDecimal.BigInt(),
			"asset_proofs":      assetProofs,
			"liability_proofs":  liabilityProofs,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate witness: %w", err)
	}

	// Generate proof
	proof, vk, err := g.generateGroth16Proof(circuit, witness)
	if err != nil {
		return nil, fmt.Errorf("failed to generate groth16 proof: %w", err)
	}

	// Create ZK proof object
	zkProof := &types.ZKProof{
		ID:              generateProofID(),
		Type:            types.ProofTypeSolvency,
		Status:          types.ProofStatusGenerated,
		CircuitID:       circuit.ID(),
		CircuitHash:     circuit.Hash(),
		PublicInputs:    req.PublicInputs,
		Proof:           proof,
		VerificationKey: vk,
		MerkleRoot:      merkleRoot.(string),
	}

	return zkProof, nil
}

// SupportedTypes returns supported proof types
func (g *SolvencyProofGenerator) SupportedTypes() []types.ProofType {
	return []types.ProofType{types.ProofTypeSolvency}
}

// EstimateTime estimates proof generation time
func (g *SolvencyProofGenerator) EstimateTime(req *types.ProofRequest) (time.Duration, error) {
	// Solvency proofs are more complex - estimate 30-90 seconds
	return 60 * time.Second, nil
}

func (g *SolvencyProofGenerator) toDecimal(value interface{}) (*math.Decimal, error) {
	switch v := value.(type) {
	case string:
		return math.NewDecimalFromString(v)
	case *math.Decimal:
		return v, nil
	case float64:
		return math.NewDecimalFromFloat(v), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

func (g *SolvencyProofGenerator) generateGroth16Proof(circuit Circuit, witness CircuitWitness) (*types.ProofData, *types.VerificationKey, error) {
	// Mock implementation - in reality this would use gnark
	proof := &types.ProofData{
		A: types.Point{X: "0xsolvency_a_x", Y: "0xsolvency_a_y"},
		B: types.Point2{
			X: [2]string{"0xsolv_b_x0", "0xsolv_b_x1"},
			Y: [2]string{"0xsolv_b_y0", "0xsolv_b_y1"},
		},
		C:    types.Point{X: "0xsolvency_c_x", Y: "0xsolvency_c_y"},
		Hash: "0xsolvency_proof_hash",
		Size: 512,
	}

	vk := &types.VerificationKey{
		Alpha: types.Point{X: "0xsolv_alpha_x", Y: "0xsolv_alpha_y"},
		Beta: types.Point2{
			X: [2]string{"0xsolv_beta_x0", "0xsolv_beta_x1"},
			Y: [2]string{"0xsolv_beta_y0", "0xsolv_beta_y1"},
		},
		Gamma: types.Point2{
			X: [2]string{"0xsolv_gamma_x0", "0xsolv_gamma_x1"},
			Y: [2]string{"0xsolv_gamma_y0", "0xsolv_gamma_y1"},
		},
		Delta: types.Point2{
			X: [2]string{"0xsolv_delta_x0", "0xsolv_delta_x1"},
			Y: [2]string{"0xsolv_delta_y0", "0xsolv_delta_y1"},
		},
		IC: []types.Point{
			{X: "0xsolv_ic0_x", Y: "0xsolv_ic0_y"},
			{X: "0xsolv_ic1_x", Y: "0xsolv_ic1_y"},
			{X: "0xsolv_ic2_x", Y: "0xsolv_ic2_y"},
		},
		Hash: "0xsolvency_vk_hash",
	}

	return proof, vk, nil
}
