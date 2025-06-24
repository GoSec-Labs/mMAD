package circuits

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
)

// CircuitVariable represents a circuit variable that can be frontend.Variable or a constant
type CircuitVariable interface{}

// Field represents the field we're working in (BN254 scalar field)
type Field = emulated.BN254Fp

// Point represents an elliptic curve point
type Point struct {
	X frontend.Variable
	Y frontend.Variable
}

// Hash represents a hash value in the circuit
type Hash frontend.Variable

// Balance represents a balance value (as field element)
type Balance frontend.Variable

// Timestamp represents a timestamp
type Timestamp frontend.Variable

// CircuitConfig contains configuration for circuits
type CircuitConfig struct {
	// Hash function to use (MiMC by default)
	HashFunction string

	// Maximum tree depth for Merkle proofs
	MaxTreeDepth int

	// Maximum number of accounts in batch operations
	MaxBatchSize int

	// Precision for decimal operations (number of decimal places)
	DecimalPrecision int

	// Field modulus
	FieldModulus string
}

// DefaultCircuitConfig returns default circuit configuration
func DefaultCircuitConfig() *CircuitConfig {
	return &CircuitConfig{
		HashFunction:     "mimc",
		MaxTreeDepth:     32,
		MaxBatchSize:     1000,
		DecimalPrecision: 8, // 8 decimal places
		FieldModulus:     "21888242871839275222246405745257275088548364400416034343698204186575808495617",
	}
}
