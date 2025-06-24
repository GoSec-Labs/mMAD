package circuits

import (
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/cmp"
)

// CircuitUtils provides utility functions for circuits
type CircuitUtils struct {
	api    frontend.API
	config *CircuitConfig
}

// NewCircuitUtils creates new circuit utilities
func NewCircuitUtils(api frontend.API, config *CircuitConfig) *CircuitUtils {
	return &CircuitUtils{
		api:    api,
		config: config,
	}
}

// Hash computes MiMC hash of given inputs
func (u *CircuitUtils) Hash(inputs ...frontend.Variable) frontend.Variable {
	hasher, err := mimc.NewMiMC(u.api)
	if err != nil {
		panic(err) // This should never happen in a well-formed circuit
	}

	hasher.Write(inputs...)
	return hasher.Sum()
}

// HashPair computes hash of two values
func (u *CircuitUtils) HashPair(left, right frontend.Variable) frontend.Variable {
	return u.Hash(left, right)
}

// AssertLessEqualThan asserts that a <= b
func (u *CircuitUtils) AssertLessEqualThan(a, b frontend.Variable) {
	u.api.AssertIsEqual(
		cmp.IsLessOrEqual(u.api, a, b),
		1,
	)
}

// AssertGreaterEqualThan asserts that a >= b
func (u *CircuitUtils) AssertGreaterEqualThan(a, b frontend.Variable) {
	u.api.AssertIsEqual(
		cmp.IsLessOrEqual(u.api, b, a),
		1,
	)
}

// AssertInRange asserts that min <= value <= max
func (u *CircuitUtils) AssertInRange(value, min, max frontend.Variable) {
	u.AssertGreaterEqualThan(value, min)
	u.AssertLessEqualThan(value, max)
}

// IsZero returns 1 if value is zero, 0 otherwise
func (u *CircuitUtils) IsZero(value frontend.Variable) frontend.Variable {
	return u.api.IsZero(value)
}

// Select returns a if selector is 1, b if selector is 0
func (u *CircuitUtils) Select(selector, a, b frontend.Variable) frontend.Variable {
	return u.api.Select(selector, a, b)
}

// ConditionalSelect returns a if condition is true, b otherwise
func (u *CircuitUtils) ConditionalSelect(condition, a, b frontend.Variable) frontend.Variable {
	return u.api.Select(condition, a, b)
}

// ScaleToFieldElement converts a decimal value to field element
// Multiplies by 10^precision to handle decimals
func (u *CircuitUtils) ScaleToFieldElement(value string) *big.Int {
	// Parse decimal string and scale by precision
	decimal := new(big.Float)
	decimal.SetString(value)

	// Scale by 10^precision
	precision := new(big.Float).SetInt64(int64(u.config.DecimalPrecision))
	scale := new(big.Float).Exp(big.NewFloat(10), precision, nil)

	scaled := new(big.Float).Mul(decimal, scale)

	// Convert to integer
	result := new(big.Int)
	scaled.Int(result)

	return result
}

// BatchHash computes hash of multiple values efficiently
func (u *CircuitUtils) BatchHash(values []frontend.Variable) frontend.Variable {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return u.Hash(values[0])
	}

	// Hash pairs recursively for efficiency
	if len(values) == 2 {
		return u.HashPair(values[0], values[1])
	}

	// For more values, hash in tree structure
	mid := len(values) / 2
	leftHash := u.BatchHash(values[:mid])
	rightHash := u.BatchHash(values[mid:])

	return u.HashPair(leftHash, rightHash)
}

// VerifySignature verifies an EdDSA signature (placeholder for now)
func (u *CircuitUtils) VerifySignature(message, signature, publicKey frontend.Variable) frontend.Variable {
	// In a real implementation, this would use EdDSA verification
	// For now, we'll just return a placeholder
	return u.Hash(message, signature, publicKey)
}
