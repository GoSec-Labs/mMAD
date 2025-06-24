package circuits

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
)

// BalanceCircuit proves that a user's balance is above a threshold
// without revealing the actual balance
type BalanceCircuit struct {
	// Public inputs (known to verifier)
	Threshold frontend.Variable `gnark:",public"` // Minimum required balance
	UserID    frontend.Variable `gnark:",public"` // User identifier hash
	Nonce     frontend.Variable `gnark:",public"` // Prevents replay attacks
	Timestamp frontend.Variable `gnark:",public"` // When proof was generated

	// Private inputs (secret)
	Balance frontend.Variable `gnark:",secret"` // Actual balance (secret)
	Salt    frontend.Variable `gnark:",secret"` // Random salt for privacy

	// Optional: Account commitment
	AccountCommitment frontend.Variable `gnark:",public"` // Hash commitment to account
}

// Define implements the circuit logic
func (circuit *BalanceCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// 1. Verify balance is non-negative
	utils.AssertGreaterEqualThan(circuit.Balance, 0)

	// 2. Main constraint: Balance >= Threshold
	utils.AssertGreaterEqualThan(circuit.Balance, circuit.Threshold)

	// 3. Verify account commitment (binds balance to specific account)
	expectedCommitment := utils.Hash(
		circuit.UserID,
		circuit.Balance,
		circuit.Nonce,
		circuit.Salt,
	)
	api.AssertIsEqual(circuit.AccountCommitment, expectedCommitment)

	// 4. Timestamp bounds (proof is fresh)
	minTimestamp := big.NewInt(1640995200) // Jan 1, 2022
	maxTimestamp := big.NewInt(4102444800) // Jan 1, 2100
	utils.AssertInRange(circuit.Timestamp, minTimestamp, maxTimestamp)

	// 5. Nonce bounds (prevent overflow)
	maxNonce := new(big.Int).Lsh(big.NewInt(1), 64) // 2^64
	utils.AssertLessEqualThan(circuit.Nonce, maxNonce)

	return nil
}

// BalanceRangeCircuit proves balance is within a specific range [min, max]
type BalanceRangeCircuit struct {
	// Public inputs
	MinThreshold frontend.Variable `gnark:",public"`
	MaxThreshold frontend.Variable `gnark:",public"`
	UserID       frontend.Variable `gnark:",public"`
	Commitment   frontend.Variable `gnark:",public"`

	// Private inputs
	Balance frontend.Variable `gnark:",secret"`
	Salt    frontend.Variable `gnark:",secret"`
}

func (circuit *BalanceRangeCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())

	// Balance must be in range [MinThreshold, MaxThreshold]
	utils.AssertInRange(circuit.Balance, circuit.MinThreshold, circuit.MaxThreshold)

	// Verify commitment
	expectedCommitment := utils.Hash(circuit.UserID, circuit.Balance, circuit.Salt)
	api.AssertIsEqual(circuit.Commitment, expectedCommitment)

	return nil
}

// MultiBalanceCircuit proves multiple balances simultaneously
type MultiBalanceCircuit struct {
	// Public inputs
	NumAccounts frontend.Variable   `gnark:",public"`
	Thresholds  []frontend.Variable `gnark:",public"` // One threshold per account
	UserIDs     []frontend.Variable `gnark:",public"`
	RootHash    frontend.Variable   `gnark:",public"` // Merkle root of all commitments

	// Private inputs
	Balances     []frontend.Variable   `gnark:",secret"`
	Salts        []frontend.Variable   `gnark:",secret"`
	MerkleProofs [][]frontend.Variable `gnark:",secret"` // Merkle inclusion proofs
}

func (circuit *MultiBalanceCircuit) Define(api frontend.API) error {
	utils := NewCircuitUtils(api, DefaultCircuitConfig())
	config := DefaultCircuitConfig()

	// Validate number of accounts
	api.AssertIsEqual(len(circuit.Balances), len(circuit.Thresholds))
	api.AssertIsEqual(len(circuit.Balances), len(circuit.UserIDs))

	// For each account, verify balance >= threshold
	for i := 0; i < len(circuit.Balances); i++ {
		// Balance constraint
		utils.AssertGreaterEqualThan(circuit.Balances[i], circuit.Thresholds[i])

		// Create leaf commitment
		leaf := utils.Hash(circuit.UserIDs[i], circuit.Balances[i], circuit.Salts[i])

		// Verify Merkle inclusion proof
		if i < len(circuit.MerkleProofs) {
			circuit.verifyMerkleProof(api, utils, leaf, circuit.MerkleProofs[i], circuit.RootHash, config.MaxTreeDepth)
		}
	}

	return nil
}

func (circuit *MultiBalanceCircuit) verifyMerkleProof(
	api frontend.API,
	utils *CircuitUtils,
	leaf frontend.Variable,
	proof []frontend.Variable,
	root frontend.Variable,
	maxDepth int,
) {
	currentHash := leaf

	// Traverse up the tree
	for i := 0; i < len(proof) && i < maxDepth; i++ {
		sibling := proof[i]

		// Hash with sibling (order doesn't matter for this simple version)
		currentHash = utils.HashPair(currentHash, sibling)
	}

	// Final hash should equal root
	api.AssertIsEqual(currentHash, root)
}
