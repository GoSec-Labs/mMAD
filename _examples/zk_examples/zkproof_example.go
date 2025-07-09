package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GoSec-Labs/mMAD/engines/internal/database/models"
	"github.com/GoSec-Labs/mMAD/engines/internal/database/repository"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/generators"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/merkle"
	"github.com/GoSec-Labs/mMAD/engines/internal/zkproof/types"
	"github.com/GoSec-Labs/mMAD/engines/pkg/logger"
	"github.com/GoSec-Labs/mMAD/engines/pkg/math"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	logger.Init(logger.Config{Level: "info"})

	// Create ZK proof engine
	engine := createZKProofEngine()

	// Start the engine
	if err := engine.Start(ctx); err != nil {
		log.Fatalf("Failed to start ZK proof engine: %v", err)
	}
	defer engine.Stop(ctx)

	// Example 1: Generate a balance proof
	fmt.Println("🔍 Example 1: Balance Proof Generation")
	balanceProof, err := generateBalanceProof(ctx, engine)
	if err != nil {
		log.Fatalf("Failed to generate balance proof: %v", err)
	}
	fmt.Printf("✅ Balance proof generated: %s\n", balanceProof.ID)

	// Example 2: Verify the balance proof
	fmt.Println("\n🔍 Example 2: Balance Proof Verification")
	verificationResult, err := verifyBalanceProof(ctx, engine, balanceProof)
	if err != nil {
		log.Fatalf("Failed to verify balance proof: %v", err)
	}
	fmt.Printf("✅ Proof verification: %v\n", verificationResult.Valid)

	// Example 3: Generate solvency proof
	fmt.Println("\n🔍 Example 3: Solvency Proof Generation")
	solvencyProof, err := generateSolvencyProof(ctx, engine)
	if err != nil {
		log.Fatalf("Failed to generate solvency proof: %v", err)
	}
	fmt.Printf("✅ Solvency proof generated: %s\n", solvencyProof.ID)

	// Example 4: Create Merkle tree and generate inclusion proofs
	fmt.Println("\n🔍 Example 4: Merkle Tree Operations")
	merkleExample()

	// Example 5: Async proof generation
	fmt.Println("\n🔍 Example 5: Async Proof Generation")
	asyncProofExample(ctx, engine)
}

func createZKProofEngine() *zkproof.Engine {
	// Mock dependencies for this example
	repo := &mockRepository{}
	circuitManager := &mockCircuitManager{}

	// Create generators
	balanceGen := generators.NewBalanceProofGenerator(circuitManager)
	solvencyGen := generators.NewSolvencyProofGenerator(circuitManager)

	// Create composite generator
	generator := &compositeGenerator{
		generators: map[types.ProofType]zkproof.ProofGenerator{
			types.ProofTypeBalance:  balanceGen,
			types.ProofTypeSolvency: solvencyGen,
		},
	}

	// Create verifier
	verifier := &mockVerifier{}

	// Configuration
	config := &zkproof.Config{
		MaxWorkers:     4,
		WorkerTimeout:  5 * time.Minute,
		DefaultTTL:     24 * time.Hour,
		MaxProofSize:   1024 * 1024, // 1MB
		MaxBatchSize:   10,
		EnableBatching: true,
		BatchTimeout:   30 * time.Second,
	}

	return zkproof.NewEngine(repo, circuitManager, generator, verifier, config)
}

func generateBalanceProof(ctx context.Context, engine *zkproof.Engine) (*types.ZKProof, error) {
	// Create balance proof request
	req := &types.ProofRequest{
		Type:      types.ProofTypeBalance,
		UserID:    "user123",
		AccountID: "acc456",
		PublicInputs: map[string]interface{}{
			"threshold": "1000.00", // Prove balance >= $1000
		},
		PrivateInputs: map[string]interface{}{
			"balance": "2500.75", // Actual balance (secret)
		},
		Options: types.ProofOptions{
			ExpiresIn: 24 * time.Hour,
		},
	}

	// Generate proof
	proof, err := engine.GenerateProof(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proof: %w", err)
	}

	fmt.Printf("   📋 Proof Type: %s\n", proof.Type)
	fmt.Printf("   🔑 Proof ID: %s\n", proof.ID)
	fmt.Printf("   ⏱️  Generation Time: %v\n", proof.GenerationTime)
	fmt.Printf("   📊 Status: %s\n", proof.Status)

	return proof, nil
}

func verifyBalanceProof(ctx context.Context, engine *zkproof.Engine, proof *types.ZKProof) (*types.VerificationResult, error) {
	// Create verification request
	req := &types.VerificationRequest{
		ProofID:         proof.ID,
		Proof:           proof.Proof,
		PublicInputs:    proof.PublicInputs,
		VerificationKey: proof.VerificationKey,
		CircuitHash:     proof.CircuitHash,
	}

	// Verify proof
	result, err := engine.VerifyProof(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify proof: %w", err)
	}

	fmt.Printf("   ✅ Valid: %v\n", result.Valid)
	fmt.Printf("   ⏱️  Verification Time: %v\n", result.VerificationTime)

	return result, nil
}

func generateSolvencyProof(ctx context.Context, engine *zkproof.Engine) (*types.ZKProof, error) {
	// Create accounts for merkle tree
	accounts := []merkle.Account{
		{ID: "acc1", Balance: math.NewDecimalFromString("1000.50"), Currency: "USD"},
		{ID: "acc2", Balance: math.NewDecimalFromString("2500.75"), Currency: "USD"},
		{ID: "acc3", Balance: math.NewDecimalFromString("750.25"), Currency: "USD"},
		{ID: "acc4", Balance: math.NewDecimalFromString("3200.00"), Currency: "USD"},
	}

	// Create merkle tree
	tree, err := merkle.NewMerkleTree(accounts)
	if err != nil {
		return nil, fmt.Errorf("failed to create merkle tree: %w", err)
	}

	// Calculate totals
	totalLiabilities, _ := tree.GetTotalBalance("USD")
	totalAssets := totalLiabilities.Multiply(math.NewDecimalFromString("1.1")) // 110% reserve ratio

	// Create solvency proof request
	req := &types.ProofRequest{
		Type: types.ProofTypeSolvency,
		PublicInputs: map[string]interface{}{
			"merkle_root": tree.GetRoot(),
			"timestamp":   time.Now().Unix(),
		},
		PrivateInputs: map[string]interface{}{
			"total_assets":      totalAssets.String(),
			"total_liabilities": totalLiabilities.String(),
			"asset_proofs":      []string{"proof1", "proof2"}, // Mock proofs
			"liability_proofs":  []string{"proof3", "proof4"}, // Mock proofs
		},
	}

	// Generate proof
	proof, err := engine.GenerateProof(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate solvency proof: %w", err)
	}

	fmt.Printf("   📋 Proof Type: %s\n", proof.Type)
	fmt.Printf("   🔑 Proof ID: %s\n", proof.ID)
	fmt.Printf("   🌳 Merkle Root: %s\n", proof.MerkleRoot)
	fmt.Printf("   💰 Reserve Ratio: 110%%\n")
	fmt.Printf("   ⏱️  Generation Time: %v\n", proof.GenerationTime)

	return proof, nil
}

func merkleExample() {
	// Create sample accounts
	accounts := []merkle.Account{
		{ID: "alice", Balance: math.NewDecimalFromString("1000.50"), Currency: "USD", Nonce: 1},
		{ID: "bob", Balance: math.NewDecimalFromString("2500.75"), Currency: "USD", Nonce: 2},
		{ID: "charlie", Balance: math.NewDecimalFromString("750.25"), Currency: "USD", Nonce: 3},
		{ID: "david", Balance: math.NewDecimalFromString("3200.00"), Currency: "USD", Nonce: 4},
	}

	// Create merkle tree
	tree, err := merkle.NewMerkleTree(accounts)
	if err != nil {
		log.Fatalf("Failed to create merkle tree: %v", err)
	}

	fmt.Printf("   🌳 Merkle Root: %s\n", tree.GetRoot())

	// Generate inclusion proof for Alice
	proof, err := tree.GetProof(0) // Alice is at index 0
	if err != nil {
		log.Fatalf("Failed to generate proof: %v", err)
	}

	fmt.Printf("   🔍 Inclusion Proof for Alice:\n")
	fmt.Printf("      - Leaf Index: %d\n", proof.LeafIndex)
	fmt.Printf("      - Leaf Hash: %s\n", proof.LeafHash[:16]+"...")
	fmt.Printf("      - Path Length: %d\n", len(proof.Path))

	// Verify the proof
	isValid := merkle.VerifyProof(proof)
	fmt.Printf("   ✅ Proof Valid: %v\n", isValid)

	// Calculate total balance
	total, err := tree.GetTotalBalance("USD")
	if err != nil {
		log.Fatalf("Failed to calculate total: %v", err)
	}
	fmt.Printf("   💰 Total Balance: $%s\n", total.String())
}

func asyncProofExample(ctx context.Context, engine *zkproof.Engine) {
	// Generate proof asynchronously
	req := &types.ProofRequest{
		Type: types.ProofTypeBalance,
		PublicInputs: map[string]interface{}{
			"threshold": "5000.00",
		},
		PrivateInputs: map[string]interface{}{
			"balance": "7500.50",
		},
	}

	proofID, err := engine.GenerateProofAsync(ctx, req)
	if err != nil {
		log.Fatalf("Failed to start async proof generation: %v", err)
	}

	fmt.Printf("   🚀 Started async proof generation: %s\n", proofID)

	// Poll for progress
	for {
		progress, err := engine.GetProgress(ctx, proofID)
		if err != nil {
			log.Fatalf("Failed to get progress: %v", err)
		}

		fmt.Printf("   📊 Progress: %.1f%% - %s\n", progress.Progress*100, progress.Stage)

		if progress.Status == types.ProofStatusGenerated {
			fmt.Printf("   ✅ Async proof completed!\n")
			break
		} else if progress.Status == types.ProofStatusFailed {
			fmt.Printf("   ❌ Async proof failed: %s\n", progress.Error)
			break
		}

		time.Sleep(1 * time.Second)
	}
}

// Mock implementations for the example

type mockRepository struct{}

func (r *mockRepository) Proofs() repository.ProofRepository                        { return &mockProofRepo{} }
func (r *mockRepository) Users() repository.UserRepository                          { return nil }
func (r *mockRepository) Accounts() repository.AccountRepository                    { return nil }
func (r *mockRepository) Transactions() repository.TransactionRepository            { return nil }
func (r *mockRepository) Reserves() repository.ReserveRepository                    { return nil }
func (r *mockRepository) Compliance() repository.ComplianceRepository               { return nil }
func (r *mockRepository) BeginTx(ctx context.Context) (repository.TxManager, error) { return nil, nil }
func (r *mockRepository) WithTx(ctx context.Context, fn func(repository.TxManager) error) error {
	return nil
}

type mockProofRepo struct{}

func (r *mockProofRepo) Create(ctx context.Context, proof *models.ZKProof) error { return nil }
func (r *mockProofRepo) GetByID(ctx context.Context, id string) (*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) GetByHash(ctx context.Context, hash string) (*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) GetByUserID(ctx context.Context, userID string, filters repository.ProofFilters) ([]*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) GetByAccountID(ctx context.Context, accountID string, filters repository.ProofFilters) ([]*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) Update(ctx context.Context, proof *models.ZKProof) error { return nil }
func (r *mockProofRepo) UpdateStatus(ctx context.Context, id string, status models.ProofStatus) error {
	return nil
}
func (r *mockProofRepo) Delete(ctx context.Context, id string) error { return nil }
func (r *mockProofRepo) List(ctx context.Context, filters repository.ProofFilters) ([]*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) Count(ctx context.Context, filters repository.ProofFilters) (int64, error) {
	return 0, nil
}
func (r *mockProofRepo) GetLatestByType(ctx context.Context, proofType models.ProofType) (*models.ZKProof, error) {
	return nil, nil
}
func (r *mockProofRepo) CleanupExpired(ctx context.Context) (int64, error) { return 0, nil }

type mockCircuitManager struct{}

func (m *mockCircuitManager) GetCircuit(circuitID string) (zkproof.Circuit, error) {
	return &mockCircuit{id: circuitID}, nil
}
func (m *mockCircuitManager) ListCircuits() ([]zkproof.CircuitInfo, error) { return nil, nil }
func (m *mockCircuitManager) CompileCircuit(source string, options zkproof.CompileOptions) (*zkproof.CircuitInfo, error) {
	return nil, nil
}
func (m *mockCircuitManager) GetVerificationKey(circuitID string) (*types.VerificationKey, error) {
	return nil, nil
}

type mockCircuit struct{ id string }

func (c *mockCircuit) ID() string                                                    { return c.id }
func (c *mockCircuit) Hash() string                                                  { return "mock_hash" }
func (c *mockCircuit) Define(api frontend.API, witness zkproof.CircuitWitness) error { return nil }
func (c *mockCircuit) Compile() (constraint.ConstraintSystem, error)                 { return nil, nil }
func (c *mockCircuit) GenerateWitness(publicInputs, privateInputs map[string]interface{}) (zkproof.CircuitWitness, error) {
	return &mockWitness{}, nil
}
func (c *mockCircuit) ValidateInputs(publicInputs, privateInputs map[string]interface{}) error {
	return nil
}

type mockWitness struct{}

func (w *mockWitness) PublicInputs() map[string]interface{}  { return nil }
func (w *mockWitness) PrivateInputs() map[string]interface{} { return nil }
func (w *mockWitness) Serialize() ([]byte, error)            { return nil, nil }
func (w *mockWitness) Deserialize(data []byte) error         { return nil }

type compositeGenerator struct {
	generators map[types.ProofType]zkproof.ProofGenerator
}

func (g *compositeGenerator) Generate(ctx context.Context, req *types.ProofRequest) (*types.ZKProof, error) {
	gen, exists := g.generators[req.Type]
	if !exists {
		return nil, fmt.Errorf("unsupported proof type: %s", req.Type)
	}
	return gen.Generate(ctx, req)
}

func (g *compositeGenerator) GenerateAsync(ctx context.Context, req *types.ProofRequest) (string, error) {
	return "", fmt.Errorf("async generation not implemented in mock")
}

func (g *compositeGenerator) GetProgress(ctx context.Context, proofID string) (*zkproof.GenerationProgress, error) {
	return nil, fmt.Errorf("progress tracking not implemented in mock")
}

func (g *compositeGenerator) Cancel(ctx context.Context, proofID string) error {
	return nil
}

func (g *compositeGenerator) SupportedTypes() []types.ProofType {
	var types []types.ProofType
	for t := range g.generators {
		types = append(types, t)
	}
	return types
}

func (g *compositeGenerator) EstimateTime(req *types.ProofRequest) (time.Duration, error) {
	return 30 * time.Second, nil
}

type mockVerifier struct{}

func (v *mockVerifier) Verify(ctx context.Context, req *types.VerificationRequest) (*types.VerificationResult, error) {
	return &types.VerificationResult{
		Valid:            true,
		ProofID:          req.ProofID,
		VerifiedAt:       time.Now(),
		VerificationTime: 50 * time.Millisecond,
	}, nil
}

func (v *mockVerifier) VerifyBatch(ctx context.Context, reqs []*types.VerificationRequest) ([]*types.VerificationResult, error) {
	return nil, nil
}

func (v *mockVerifier) ValidatePublicInputs(proofType types.ProofType, inputs map[string]interface{}) error {
	return nil
}
