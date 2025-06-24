package circuits

import (
    "github.com/consensys/gnark/frontend"
)

// AggregateBalanceCircuit proves aggregate properties across multiple accounts
type AggregateBalanceCircuit struct {
    // Public inputs
    NumAccounts     frontend.Variable `gnark:",public"`
    MinTotalBalance frontend.Variable `gnark:",public"` // Minimum sum of all balances
    MerkleRoot      frontend.Variable `gnark:",public"` // Root of account commitment tree
    
    // Private inputs
    Balances        []frontend.Variable `gnark:",secret"` // Individual balances
    UserIDs         []frontend.Variable `gnark:",secret"` // Account identifiers
    Salts           []frontend.Variable `gnark:",secret"` // Random salts
    MerkleProofs    [][]frontend.Variable `gnark:",secret"` // Inclusion proofs
}

func (circuit *AggregateBalanceCircuit) Define(api frontend.API) error {
    utils := NewCircuitUtils(api, DefaultCircuitConfig())
    
    // Validate input lengths
    api.AssertIsEqual(len(circuit.Balances), len(circuit.UserIDs))
    api.AssertIsEqual(len(circuit.Balances), len(circuit.Salts))
    
    // Compute total balance
    totalBalance := frontend.Variable(0)
    
    for i := 0; i < len(circuit.Balances); i++ {
        // Each balance must be non-negative
        utils.AssertGreaterEqualThan(circuit.Balances[i], 0)
        
        // Add to total
        totalBalance = api.Add(totalBalance, circuit.Balances[i])
        
        // Create commitment for this account
        commitment := utils.Hash(circuit.UserIDs[i], circuit.Balances[i], circuit.Salts[i])
        
        // Verify Merkle inclusion
        if i < len(circuit.MerkleProofs) {
            circuit.verifyMerkleInclusion(api, utils, commitment, circuit.MerkleProofs[i], circuit.MerkleRoot)
        }
    }
    
    // Total balance must meet minimum
    utils.AssertGreaterEqualThan(totalBalance, circuit.MinTotalBalance)
    
    return nil
}

func (circuit *AggregateBalanceCircuit) verifyMerkleInclusion(
    api frontend.API,
    utils *CircuitUtils,
    leaf frontend.Variable,
    proof []frontend.Variable,
    root frontend.Variable,
) {
    currentHash := leaf
    
    for _, sibling := range proof {
        currentHash = utils.HashPair(currentHash, sibling)
    }
    
    api.AssertIsEqual(currentHash, root)
}

// PortfolioCircuit proves properties about a portfolio of assets
type PortfolioCircuit struct {
    // Public inputs
    NumAssets        frontend.Variable `gnark:",public"`
    MinTotalValue    frontend.Variable `gnark:",public"`
    MaxRiskScore     frontend.Variable `gnark:",public"`
    PriceCommitment  frontend.Variable `gnark:",public"` // Commitment to price data
    
    // Private inputs
    AssetAmounts     []frontend.Variable `gnark:",secret"` // Amount of each asset
    AssetPrices      []frontend.Variable `gnark:",secret"` // Price of each asset
    RiskWeights      []frontend.Variable `gnark:",secret"` // Risk weight of each asset
    PriceNonce       frontend.Variable   `gnark:",secret"` // Nonce for price commitment
}

func (circuit *PortfolioCircuit) Define(api frontend.API) error {
    utils := NewCircuitUtils(api, DefaultCircuitConfig())
    
    // Validate price commitment
    priceHash := utils.BatchHash(circuit.AssetPrices)
    expectedCommitment := utils.Hash(priceHash, circuit.PriceNonce)
    api.AssertIsEqual(circuit.PriceCommitment, expectedCommitment)
    
    totalValue := frontend.Variable(0)
    totalRisk := frontend.Variable(0)
    
    for i := 0; i < len(circuit.AssetAmounts); i++ {
        // All amounts must be non-negative
        utils.AssertGreaterEqualThan(circuit.AssetAmounts[i], 0)
        utils.AssertGreaterEqualThan(circuit.AssetPrices[i], 0)
        
        // Calculate value of this asset
        assetValue := api.Mul(circuit.AssetAmounts[i], circuit.AssetPrices[i])
        totalValue = api.Add(totalValue, assetValue)
        
        // Calculate risk contribution
        if i < len(circuit.RiskWeights) {
            riskContribution := api.Mul(assetValue, circuit.RiskWeights[i])
            totalRisk = api.Add(totalRisk, riskContribution)
        }
    }
    
    // Portfolio constraints
    utils.AssertGreaterEqualThan(totalValue, circuit.MinTotalValue)
    utils.AssertLessEqualThan(totalRisk, circuit.MaxRiskScore)
    
    return nil
}

// TimeSeriesCircuit proves properties about time series data
type TimeSeriesCircuit struct {
    // Public inputs
    NumDataPoints    frontend.Variable `gnark:",public"`
    MinValue         frontend.Variable `gnark:",public"`
    MaxValue         frontend.Variable `gnark:",public"`
    CommitmentRoot   frontend.Variable `gnark:",public"`
    
    // Private inputs
    Values           []frontend.Variable `gnark:",secret"`
    Timestamps       []frontend.Variable `gnark:",secret"`
    CommitmentProofs [][]frontend.Variable `gnark:",secret"`
}

func (circuit *TimeSeriesCircuit) Define(api frontend.API) error {
    utils := NewCircuitUtils(api, DefaultCircuitConfig())
    
    for i := 0; i < len(circuit.Values); i++ {
        // Value bounds
        utils.AssertInRange(circuit.Values[i], circuit.MinValue, circuit.MaxValue)
        
        // Timestamp ordering (each timestamp > previous)
        if i > 0 {
            utils.AssertGreaterEqualThan(circuit.Timestamps[i], circuit.Timestamps[i-1])
        }
        
        // Verify commitment
        dataPoint := utils.Hash(circuit.Values[i], circuit.Timestamps[i])
        if i < len(circuit.CommitmentProofs) {
            circuit.verifyCommitmentInclusion(api, utils, dataPoint, circuit.CommitmentProofs[i], circuit.CommitmentRoot)
        }
    }
    
    return nil
}

func (circuit *TimeSeriesCircuit) verifyCommitmentInclusion(
	api frontend.API,
	utils *CircuitUtils,
	dataPoint frontend.Variable,
	proof []frontend.Variable,
	root frontend.Variable,
 ) {
	currentHash := dataPoint
	
	for _, sibling := range proof {
		currentHash = utils.HashPair(currentHash, sibling)
	}
	
	api.AssertIsEqual(currentHash, root)
 }