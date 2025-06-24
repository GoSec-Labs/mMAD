package circuits

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/test"
)

// CircuitTester provides testing utilities for circuits
type CircuitTester struct {
	config   *CircuitConfig
	registry *CircuitRegistry
}

// TestCase represents a circuit test case
type TestCase struct {
	Name          string
	CircuitID     string
	PublicInputs  map[string]interface{}
	PrivateInputs map[string]interface{}
	ShouldPass    bool
	Description   string
}

// TestResult contains the result of a circuit test
type TestResult struct {
	TestCase    *TestCase
	Passed      bool
	Error       error
	ProofSize   int
	ProveTime   string
	VerifyTime  string
	Constraints int
}

// TestSuite contains multiple test cases
type TestSuite struct {
	Name      string
	CircuitID string
	TestCases []TestCase
}

// NewCircuitTester creates a new circuit tester
func NewCircuitTester(registry *CircuitRegistry) *CircuitTester {
	return &CircuitTester{
		config:   DefaultCircuitConfig(),
		registry: registry,
	}
}

// TestCircuit tests a single circuit with given inputs
func (ct *CircuitTester) TestCircuit(t *testing.T, testCase TestCase) *TestResult {
	result := &TestResult{
		TestCase: &testCase,
		Passed:   false,
	}

	// Get circuit
	circuit, err := ct.registry.GetCircuit(testCase.CircuitID)
	if err != nil {
		result.Error = fmt.Errorf("failed to get circuit: %w", err)
		return result
	}

	// Create witness
	witness, err := ct.createWitness(circuit, testCase.PublicInputs, testCase.PrivateInputs)
	if err != nil {
		result.Error = fmt.Errorf("failed to create witness: %w", err)
		return result
	}

	// Test the circuit
	err = test.IsSolved(circuit, witness, ecc.BN254.ScalarField())

	if testCase.ShouldPass {
		if err != nil {
			result.Error = fmt.Errorf("circuit test failed: %w", err)
			return result
		}
		result.Passed = true
	} else {
		if err == nil {
			result.Error = fmt.Errorf("circuit should have failed but passed")
			return result
		}
		result.Passed = true // Test passed because it correctly failed
	}

	return result
}

// RunTestSuite runs a complete test suite
func (ct *CircuitTester) RunTestSuite(t *testing.T, suite TestSuite) []TestResult {
	logger.Info("Running test suite", "suite", suite.Name, "tests", len(suite.TestCases))

	results := make([]TestResult, len(suite.TestCases))

	for i, testCase := range suite.TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			result := ct.TestCircuit(t, testCase)
			results[i] = *result

			if !result.Passed {
				t.Errorf("Test failed: %v", result.Error)
			}
		})
	}

	return results
}

// BenchmarkCircuit benchmarks circuit performance
func (ct *CircuitTester) BenchmarkCircuit(b *testing.B, circuitID string, publicInputs, privateInputs map[string]interface{}) {
	circuit, err := ct.registry.GetCircuit(circuitID)
	if err != nil {
		b.Fatalf("Failed to get circuit: %v", err)
	}

	witness, err := ct.createWitness(circuit, publicInputs, privateInputs)
	if err != nil {
		b.Fatalf("Failed to create witness: %v", err)
	}

	// Compile circuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, circuit)
	if err != nil {
		b.Fatalf("Failed to compile circuit: %v", err)
	}

	// Setup
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		b.Fatalf("Failed to setup: %v", err)
	}

	b.ResetTimer()

	// Benchmark proof generation
	b.Run("ProofGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := groth16.Prove(ccs, pk, witness)
			if err != nil {
				b.Fatalf("Failed to prove: %v", err)
			}
		}
	})

	// Generate a proof for verification benchmark
	proof, err := groth16.Prove(ccs, pk, witness)
	if err != nil {
		b.Fatalf("Failed to generate proof for verification benchmark: %v", err)
	}

	publicWitness, err := witness.Public()
	if err != nil {
		b.Fatalf("Failed to get public witness: %v", err)
	}

	// Benchmark verification
	b.Run("Verification", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := groth16.Verify(proof, vk, publicWitness)
			if err != nil {
				b.Fatalf("Failed to verify: %v", err)
			}
		}
	})
}

// createWitness creates a witness from input maps
func (ct *CircuitTester) createWitness(circuit frontend.Circuit, publicInputs, privateInputs map[string]interface{}) (frontend.Witness, error) {
	// This is a simplified implementation
	// In a real implementation, you'd use reflection to map inputs to circuit fields

	// For now, we'll create a basic witness structure
	assignment := make(map[string]interface{})

	// Add public inputs
	for k, v := range publicInputs {
		assignment[k] = v
	}

	// Add private inputs
	for k, v := range privateInputs {
		assignment[k] = v
	}

	// Create witness
	witness, err := frontend.NewWitness(assignment, ecc.BN254.ScalarField())
	if err != nil {
		return nil, fmt.Errorf("failed to create witness: %w", err)
	}

	return witness, nil
}

// GetBuiltinTestSuites returns predefined test suites for built-in circuits
func (ct *CircuitTester) GetBuiltinTestSuites() []TestSuite {
	return []TestSuite{
		ct.getBalanceTestSuite(),
		ct.getSolvencyTestSuite(),
		ct.getMerkleTestSuite(),
	}
}

func (ct *CircuitTester) getBalanceTestSuite() TestSuite {
	return TestSuite{
		Name:      "Balance Circuit Tests",
		CircuitID: "balance_v1",
		TestCases: []TestCase{
			{
				Name:      "Valid balance above threshold",
				CircuitID: "balance_v1",
				PublicInputs: map[string]interface{}{
					"threshold":          big.NewInt(1000),
					"user_id":            big.NewInt(12345),
					"nonce":              big.NewInt(1),
					"timestamp":          big.NewInt(1640995200),
					"account_commitment": big.NewInt(0), // Would be computed
				},
				PrivateInputs: map[string]interface{}{
					"balance": big.NewInt(2500),
					"salt":    big.NewInt(98765),
				},
				ShouldPass:  true,
				Description: "Balance (2500) is above threshold (1000)",
			},
			{
				Name:      "Invalid balance below threshold",
				CircuitID: "balance_v1",
				PublicInputs: map[string]interface{}{
					"threshold":          big.NewInt(1000),
					"user_id":            big.NewInt(12345),
					"nonce":              big.NewInt(1),
					"timestamp":          big.NewInt(1640995200),
					"account_commitment": big.NewInt(0),
				},
				PrivateInputs: map[string]interface{}{
					"balance": big.NewInt(500),
					"salt":    big.NewInt(98765),
				},
				ShouldPass:  false,
				Description: "Balance (500) is below threshold (1000) - should fail",
			},
			{
				Name:      "Edge case: balance equals threshold",
				CircuitID: "balance_v1",
				PublicInputs: map[string]interface{}{
					"threshold":          big.NewInt(1000),
					"user_id":            big.NewInt(12345),
					"nonce":              big.NewInt(1),
					"timestamp":          big.NewInt(1640995200),
					"account_commitment": big.NewInt(0),
				},
				PrivateInputs: map[string]interface{}{
					"balance": big.NewInt(1000),
					"salt":    big.NewInt(98765),
				},
				ShouldPass:  true,
				Description: "Balance exactly equals threshold",
			},
		},
	}
}

func (ct *CircuitTester) getSolvencyTestSuite() TestSuite {
	return TestSuite{
		Name:      "Solvency Circuit Tests",
		CircuitID: "solvency_v1",
		TestCases: []TestCase{
			{
				Name:      "Solvent institution",
				CircuitID: "solvency_v1",
				PublicInputs: map[string]interface{}{
					"merkle_root":        big.NewInt(12345678),
					"timestamp":          big.NewInt(1640995200),
					"min_solvency_ratio": big.NewInt(1), // 100%
				},
				PrivateInputs: map[string]interface{}{
					"total_assets":      big.NewInt(11000),
					"total_liabilities": big.NewInt(10000),
					"asset_commitments": []interface{}{big.NewInt(111)},
					"liability_proofs":  []interface{}{[]interface{}{big.NewInt(222)}},
					"liability_leaves":  []interface{}{big.NewInt(10000)},
					"nonce":             big.NewInt(54321),
				},
				ShouldPass:  true,
				Description: "Assets (11000) > Liabilities (10000) * Ratio (1)",
			},
			{
				Name:      "Insolvent institution",
				CircuitID: "solvency_v1",
				PublicInputs: map[string]interface{}{
					"merkle_root":        big.NewInt(12345678),
					"timestamp":          big.NewInt(1640995200),
					"min_solvency_ratio": big.NewInt(1),
				},
				PrivateInputs: map[string]interface{}{
					"total_assets":      big.NewInt(9000),
					"total_liabilities": big.NewInt(10000),
					"asset_commitments": []interface{}{big.NewInt(111)},
					"liability_proofs":  []interface{}{[]interface{}{big.NewInt(222)}},
					"liability_leaves":  []interface{}{big.NewInt(10000)},
					"nonce":             big.NewInt(54321),
				},
				ShouldPass:  false,
				Description: "Assets (9000) < Liabilities (10000) - should fail",
			},
		},
	}
}

func (ct *CircuitTester) getMerkleTestSuite() TestSuite {
	return TestSuite{
		Name:      "Merkle Inclusion Tests",
		CircuitID: "merkle_inclusion_v1",
		TestCases: []TestCase{
			{
				Name:      "Valid inclusion proof",
				CircuitID: "merkle_inclusion_v1",
				PublicInputs: map[string]interface{}{
					"root":       big.NewInt(987654321),
					"leaf_index": big.NewInt(0),
				},
				PrivateInputs: map[string]interface{}{
					"leaf_value": big.NewInt(42),
					"path":       []interface{}{big.NewInt(111), big.NewInt(222)},
					"directions": []interface{}{big.NewInt(0), big.NewInt(1)},
				},
				ShouldPass:  true,
				Description: "Valid Merkle inclusion proof",
			},
		},
	}
}
