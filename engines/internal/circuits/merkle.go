package circuits

import (
	"github.com/consensys/gnark/frontend"
)

// MerkleInclusionCircuit proves that a value is included in a Merkle tree
type MerkleInclusionCircuit struct {
	// Public inputs
	Root      frontend.Variable `gnark:",public"` // Merkle tree root
	LeafIndex frontend.Variable `gnark:",public"` // Index of the leaf (optional)

	// Private inputs
	LeafValue  frontend.Variable   `gnark:",secret"` // The actual leaf value
	Path       []frontend.Variable `gnark:",secret"` // Sibling hashes along path to root
	Directions []frontend.Variable `gnark:",secret"` // 0 = left, 1 = right
}

func (circuit *MerkleInclusionCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// Start with leaf hash
	currentHash := utils.Hash(circuit.LeafValue)

	// Traverse up the tree
	for i := 0; i < len(circuit.Path); i++ {
		sibling := circuit.Path[i]
		direction := circuit.Directions[i]

		// If direction is 0, current node is left child
		// If direction is 1, current node is right child
		leftChild := utils.Select(direction, sibling, currentHash)
		rightChild := utils.Select(direction, currentHash, sibling)

		// Hash the pair
		currentHash = utils.HashPair(leftChild, rightChild)
	}

	// Final hash must equal root
	api.AssertIsEqual(currentHash, circuit.Root)

	return nil
}

// MerkleNonInclusionCircuit proves that a value is NOT in a Merkle tree
type MerkleNonInclusionCircuit struct {
	// Public inputs
	Root         frontend.Variable `gnark:",public"`
	ClaimedValue frontend.Variable `gnark:",public"` // Value claimed to not be in tree

	// Private inputs
	WitnessPath []frontend.Variable `gnark:",secret"` // Path showing different value at position
	ActualValue frontend.Variable   `gnark:",secret"` // What's actually at that position
	Directions  []frontend.Variable `gnark:",secret"`
}

func (circuit *MerkleNonInclusionCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// Verify that ActualValue != ClaimedValue
	difference := api.Sub(circuit.ActualValue, circuit.ClaimedValue)
	api.AssertIsDifferent(difference, 0)

	// Prove that ActualValue is in the tree at the claimed position
	currentHash := utils.Hash(circuit.ActualValue)

	for i := 0; i < len(circuit.WitnessPath); i++ {
		sibling := circuit.WitnessPath[i]
		direction := circuit.Directions[i]

		leftChild := utils.Select(direction, sibling, currentHash)
		rightChild := utils.Select(direction, currentHash, sibling)

		currentHash = utils.HashPair(leftChild, rightChild)
	}

	api.AssertIsEqual(currentHash, circuit.Root)

	return nil
}

// SparseMerkleInclusionCircuit for sparse Merkle trees (more efficient for sparse data)
type SparseMerkleInclusionCircuit struct {
	// Public inputs
	Root frontend.Variable `gnark:",public"`
	Key  frontend.Variable `gnark:",public"` // Key being proven

	// Private inputs
	Value               frontend.Variable   `gnark:",secret"` // Value at key (0 for non-inclusion)
	Path                []frontend.Variable `gnark:",secret"`
	Siblings            []frontend.Variable `gnark:",secret"`
	IsNonInclusionProof frontend.Variable   `gnark:",secret"` // 1 if proving non-inclusion
}

func (circuit *SparseMerkleInclusionCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())
	config := DefaultCircuitConfig()

	// For non-inclusion, value must be 0
	nonInclusionConstraint := api.Mul(circuit.IsNonInclusionProof, circuit.Value)
	api.AssertIsEqual(nonInclusionConstraint, 0)

	// Start with leaf hash
	leafHash := utils.Hash(circuit.Key, circuit.Value)
	currentHash := leafHash

	// Traverse up the tree
	for i := 0; i < len(circuit.Path) && i < config.MaxTreeDepth; i++ {
		pathBit := circuit.Path[i] // 0 or 1
		sibling := circuit.Siblings[i]

		// Ensure path bit is binary
		api.AssertIsBoolean(pathBit)

		// Select left and right based on path bit
		leftChild := utils.Select(pathBit, sibling, currentHash)
		rightChild := utils.Select(pathBit, currentHash, sibling)

		currentHash = utils.HashPair(leftChild, rightChild)
	}

	// Must equal root
	api.AssertIsEqual(currentHash, circuit.Root)

	return nil
}
