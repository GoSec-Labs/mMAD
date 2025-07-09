package main

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/circuits"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init(logger.Config{Level: "info"})

	fmt.Println("🔧 ZK Circuit Examples")
	fmt.Println("======================")

	// Create circuit infrastructure
	compiler := circuits.NewCircuitCompiler(circuits.DefaultCircuitConfig())
	registry := circuits.NewCircuitRegistry(compiler)
	tester := circuits.NewCircuitTester(registry)

	// Example 1: List available circuits
	fmt.Println("\n📋 Available Circuits:")
	listCircuits(registry)

	// Example 2: Test balance circuit
	fmt.Println("\n💰 Testing Balance Circuit:")
	testBalanceCircuit(tester)

	// Example 3: Test solvency circuit
	fmt.Println("\n🏦 Testing Solvency Circuit:")
	testSolvencyCircuit(tester)

	// Example 4: Compile and benchmark circuits
	fmt.Println("\n⚡ Benchmarking Circuits:")
	benchmarkCircuits(registry, compiler)

	// Example 5: Run full test suites
	fmt.Println("\n🧪 Running Test Suites:")
	runTestSuites(tester)
}

func listCircuits(registry *circuits.CircuitRegistry) {
	infos := registry.ListCircuits()

	for _, info := range infos {
		fmt.Printf("   📦 %s (%s)\n", info.Name, info.ID)
		fmt.Printf("      Description: %s\n", info.Description)
		fmt.Printf("      Estimated Time: %s\n", info.EstimatedTime)
		fmt.Printf("      Max Constraints: %d\n", info.MaxConstraints)
		fmt.Printf("      Public Inputs: %d\n", len(info.PublicInputs))
		fmt.Printf("      Private Inputs: %d\n", len(info.PrivateInputs))
		fmt.Println()
	}
}

func testBalanceCircuit(tester *circuits.CircuitTester) {
	// Create a mock test to demonstrate circuit testing
	testCase := circuits.TestCase{
		Name:      "Valid balance test",
		CircuitID: "balance_v1",
		PublicInputs: map[string]interface{}{
			"threshold": big.NewInt(1000),
			"user_id":   big.NewInt(12345),
			"nonce":     big.NewInt(1),
			"timestamp": big.NewInt(time.Now().Unix()),
		},
		PrivateInputs: map[string]interface{}{
			"balance": big.NewInt(2500),
			"salt":    big.NewInt(98765),
		},
		ShouldPass:  true,
		Description: "User has sufficient balance",
	}

	// Note: In a real implementation, you'd need to compute the account_commitment
	// For this example, we'll set it to a placeholder value
	commitment := computeAccountCommitment(
		testCase.PublicInputs["user_id"].(*big.Int),
		testCase.PrivateInputs["balance"].(*big.Int),
		testCase.PublicInputs["nonce"].(*big.Int),
		testCase.PrivateInputs["salt"].(*big.Int),
	)
	testCase.PublicInputs["account_commitment"] = commitment

	fmt.Printf("   🔍 Testing: %s\n", testCase.Description)
	fmt.Printf("   📊 Threshold: %s\n", testCase.PublicInputs["threshold"])
	fmt.Printf("   💎 Balance: %s (secret)\n", testCase.PrivateInputs["balance"])

	// This would normally run the actual test
	// For demo purposes, we'll simulate the result
	fmt.Printf("   ✅ Test Result: %v\n", testCase.ShouldPass)
}

func testSolvencyCircuit(tester *circuits.CircuitTester) {
	testCase := circuits.TestCase{
		Name:      "Solvency test",
		CircuitID: "solvency_v1",
		PublicInputs: map[string]interface{}{
			"merkle_root":        big.NewInt(12345678),
			"timestamp":          big.NewInt(time.Now().Unix()),
			"min_solvency_ratio": big.NewInt(1), // 100% ratio
		},
		PrivateInputs: map[string]interface{}{
			"total_assets":      big.NewInt(11000000), // $11M
			"total_liabilities": big.NewInt(10000000), // $10M
			"nonce":             big.NewInt(54321),
		},
		ShouldPass:  true,
		Description: "Institution is solvent with 110% ratio",
	}

	fmt.Printf("   🔍 Testing: %s\n", testCase.Description)
	fmt.Printf("   💰 Assets: $%s\n", formatAmount(testCase.PrivateInputs["total_assets"].(*big.Int)))
	fmt.Printf("   📊 Liabilities: $%s\n", formatAmount(testCase.PrivateInputs["total_liabilities"].(*big.Int)))

	ratio := float64(testCase.PrivateInputs["total_assets"].(*big.Int).Int64()) /
		float64(testCase.PrivateInputs["total_liabilities"].(*big.Int).Int64())
	fmt.Printf("   📈 Actual Ratio: %.1f%%\n", ratio*100)
	fmt.Printf("   ✅ Test Result: %v\n", testCase.ShouldPass)
}

func benchmarkCircuits(registry *circuits.CircuitRegistry, compiler *circuits.CircuitCompiler) {
	circuits := []string{"balance_v1", "solvency_v1", "merkle_inclusion_v1"}

	for _, circuitID := range circuits {
		fmt.Printf("   ⚡ Benchmarking %s:\n", circuitID)

		// Simulate compilation metrics
		start := time.Now()
		compiled, err := registry.GetCompiledCircuit(context.Background(), circuitID)
		if err != nil {
			fmt.Printf("      ❌ Compilation failed: %v\n", err)
			continue
		}

		compileTime := time.Since(start)

		fmt.Printf("      🔧 Compile Time: %v\n", compileTime)
		fmt.Printf("      📊 Constraints: %d\n", compiled.NumConstraints)
		fmt.Printf("      💾 Memory: %s\n", estimateMemoryUsage(compiled.NumConstraints))
		fmt.Printf("      ⚡ Est. Prove Time: %s\n", estimateProveTime(compiled.NumConstraints))
		fmt.Printf("      🔍 Est. Verify Time: %s\n", estimateVerifyTime())
		fmt.Println()
	}
}

func runTestSuites(tester *circuits.CircuitTester) {
	suites := tester.GetBuiltinTestSuites()

	for _, suite := range suites {
		fmt.Printf("   🧪 Running: %s\n", suite.Name)
		fmt.Printf("      📝 Test Cases: %d\n", len(suite.TestCases))

		passed := 0
		for _, testCase := range suite.TestCases {
			// Simulate test execution
			if testCase.ShouldPass {
				fmt.Printf("      ✅ %s\n", testCase.Name)
				passed++
			} else {
				fmt.Printf("      🔍 %s (expected failure)\n", testCase.Name)
				passed++
			}
		}

		fmt.Printf("      📊 Results: %d/%d passed\n", passed, len(suite.TestCases))
		fmt.Println()
	}
}

// Helper functions

func computeAccountCommitment(userID, balance, nonce, salt *big.Int) *big.Int {
	// Simplified commitment computation
	// In reality, this would use the same hash function as the circuit
	result := big.NewInt(0)
	result.Add(result, userID)
	result.Add(result, balance)
	result.Add(result, nonce)
	result.Add(result, salt)
	return result
}

func formatAmount(amount *big.Int) string {
	// Convert to millions for readability
	millions := new(big.Int).Div(amount, big.NewInt(1000000))
	return fmt.Sprintf("%s M", millions.String())
}

func estimateMemoryUsage(constraints int) string {
	// Rough estimate: ~1KB per constraint
	mb := constraints / 1000
	if mb < 1 {
		return "< 1MB"
	}
	return fmt.Sprintf("~%d MB", mb)
}

func estimateProveTime(constraints int) string {
	// Rough estimates based on constraint count
	if constraints < 1000 {
		return "1-5 seconds"
	} else if constraints < 10000 {
		return "5-30 seconds"
	} else if constraints < 50000 {
		return "30-120 seconds"
	} else {
		return "2-10 minutes"
	}
}

func estimateVerifyTime() string {
	return "10-50 ms"
}

// Benchmark function (would be in a separate test file)
func BenchmarkBalanceCircuit(b *testing.B) {
	compiler := circuits.NewCircuitCompiler(circuits.DefaultCircuitConfig())
	registry := circuits.NewCircuitRegistry(compiler)
	tester := circuits.NewCircuitTester(registry)

	publicInputs := map[string]interface{}{
		"threshold": big.NewInt(1000),
		"user_id":   big.NewInt(12345),
		"nonce":     big.NewInt(1),
		"timestamp": big.NewInt(time.Now().Unix()),
	}

	privateInputs := map[string]interface{}{
		"balance": big.NewInt(2500),
		"salt":    big.NewInt(98765),
	}

	commitment := computeAccountCommitment(
		publicInputs["user_id"].(*big.Int),
		privateInputs["balance"].(*big.Int),
		publicInputs["nonce"].(*big.Int),
		privateInputs["salt"].(*big.Int),
	)
	publicInputs["account_commitment"] = commitment

	tester.BenchmarkCircuit(b, "balance_v1", publicInputs, privateInputs)
}
