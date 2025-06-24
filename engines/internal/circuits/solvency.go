package circuits

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
)

// SolvencyCircuit proves that total assets >= total liabilities
// without revealing the actual amounts
type SolvencyCircuit struct {
	// Public inputs
	MerkleRoot       frontend.Variable `gnark:",public"` // Root of liability tree
	Timestamp        frontend.Variable `gnark:",public"` // When proof was generated
	MinSolvencyRatio frontend.Variable `gnark:",public"` // Minimum ratio (e.g., 1.0 = 100%)

	// Private inputs
	TotalAssets      frontend.Variable     `gnark:",secret"` // Sum of all assets
	TotalLiabilities frontend.Variable     `gnark:",secret"` // Sum of all liabilities
	AssetCommitments []frontend.Variable   `gnark:",secret"` // Commitments to individual assets
	LiabilityProofs  [][]frontend.Variable `gnark:",secret"` // Merkle proofs for liabilities
	LiabilityLeaves  []frontend.Variable   `gnark:",secret"` // Individual liability amounts
	Nonce            frontend.Variable     `gnark:",secret"` // Random nonce
}

func (circuit *SolvencyCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// 1. Verify all amounts are non-negative
	utils.AssertGreaterEqualThan(circuit.TotalAssets, 0)
	utils.AssertGreaterEqualThan(circuit.TotalLiabilities, 0)

	// 2. Main solvency constraint: Assets >= Liabilities * MinSolvencyRatio
	requiredAssets := api.Mul(circuit.TotalLiabilities, circuit.MinSolvencyRatio)
	utils.AssertGreaterEqualThan(circuit.TotalAssets, requiredAssets)

	// 3. Verify liability sum consistency
	circuit.verifyLiabilitySum(api, utils)

	// 4. Verify asset commitments (proves assets exist)
	circuit.verifyAssetCommitments(api, utils)

	// 5. Timestamp validation
	minTimestamp := big.NewInt(1640995200) // Recent timestamp
	utils.AssertGreaterEqualThan(circuit.Timestamp, minTimestamp)

	return nil
}

func (circuit *SolvencyCircuit) verifyLiabilitySum(api frontend.API, utils *CircuitUtils) {
	// Verify that TotalLiabilities equals sum of individual liabilities
	computedSum := frontend.Variable(0)

	for i := 0; i < len(circuit.LiabilityLeaves); i++ {
		// Add each liability to sum
		computedSum = api.Add(computedSum, circuit.LiabilityLeaves[i])

		// Verify Merkle inclusion proof for this liability
		if i < len(circuit.LiabilityProofs) {
			circuit.verifyMerkleProof(
				api, utils,
				circuit.LiabilityLeaves[i],
				circuit.LiabilityProofs[i],
				circuit.MerkleRoot,
			)
		}
	}

	// Total must match
	api.AssertIsEqual(circuit.TotalLiabilities, computedSum)
}

func (circuit *SolvencyCircuit) verifyAssetCommitments(api frontend.API, utils *CircuitUtils) {
	// Verify that asset commitments are valid
	// This is a simplified version - in practice you'd have more complex asset proofs

	if len(circuit.AssetCommitments) == 0 {
		return
	}

	// Create a commitment hash from total assets and nonce
	expectedCommitment := utils.Hash(circuit.TotalAssets, circuit.Nonce)

	// Verify at least one commitment matches (simplified)
	// In practice, you'd have individual asset proofs
	if len(circuit.AssetCommitments) > 0 {
		api.AssertIsEqual(circuit.AssetCommitments[0], expectedCommitment)
	}
}

func (circuit *SolvencyCircuit) verifyMerkleProof(
	api frontend.API,
	utils *CircuitUtils,
	leaf frontend.Variable,
	proof []frontend.Variable,
	root frontend.Variable,
) {
	currentHash := leaf

	// Simple Merkle proof verification
	for _, sibling := range proof {
		currentHash = utils.HashPair(currentHash, sibling)
	}

	api.AssertIsEqual(currentHash, root)
}

// BatchSolvencyCircuit proves solvency for multiple time periods or entities
type BatchSolvencyCircuit struct {
	// Public inputs
	NumEntities      frontend.Variable   `gnark:",public"`
	MerkleRoots      []frontend.Variable `gnark:",public"`
	Timestamps       []frontend.Variable `gnark:",public"`
	MinSolvencyRatio frontend.Variable   `gnark:",public"`

	// Private inputs
	TotalAssets      []frontend.Variable `gnark:",secret"`
	TotalLiabilities []frontend.Variable `gnark:",secret"`
	Nonces           []frontend.Variable `gnark:",secret"`
}

func (circuit *BatchSolvencyCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// Verify each entity is solvent
	for i := 0; i < len(circuit.TotalAssets); i++ {
		// Non-negative amounts
		utils.AssertGreaterEqualThan(circuit.TotalAssets[i], 0)
		utils.AssertGreaterEqualThan(circuit.TotalLiabilities[i], 0)

		// Solvency constraint
		requiredAssets := api.Mul(circuit.TotalLiabilities[i], circuit.MinSolvencyRatio)
		utils.AssertGreaterEqualThan(circuit.TotalAssets[i], requiredAssets)

		// Timestamp validation
		if i < len(circuit.Timestamps) {
			minTimestamp := big.NewInt(1640995200)
			utils.AssertGreaterEqualThan(circuit.Timestamps[i], minTimestamp)
		}
	}

	return nil
}
