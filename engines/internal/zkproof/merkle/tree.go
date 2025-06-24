package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// MerkleTree represents a Merkle tree for account balances
type MerkleTree struct {
	leaves []Node
	nodes  [][]Node
	root   *Node
}

// Node represents a node in the Merkle tree
type Node struct {
	Hash   string `json:"hash"`
	Value  string `json:"value,omitempty"` // For leaf nodes
	Left   *Node  `json:"left,omitempty"`  // For internal nodes
	Right  *Node  `json:"right,omitempty"` // For internal nodes
	Index  int    `json:"index"`
	IsLeaf bool   `json:"is_leaf"`
}

// Account represents an account for Merkle tree inclusion
type Account struct {
	ID       string        `json:"id"`
	Balance  *math.Decimal `json:"balance"`
	Currency string        `json:"currency"`
	Nonce    uint64        `json:"nonce"`
}

// MerkleProof represents a proof of inclusion in the Merkle tree
type MerkleProof struct {
	LeafIndex  int      `json:"leaf_index"`
	LeafHash   string   `json:"leaf_hash"`
	LeafValue  string   `json:"leaf_value"`
	Path       []string `json:"path"`       // Hashes of sibling nodes
	Directions []bool   `json:"directions"` // true = right, false = left
	Root       string   `json:"root"`
}

// NewMerkleTree creates a new Merkle tree from accounts
func NewMerkleTree(accounts []Account) (*MerkleTree, error) {
	if len(accounts) == 0 {
		return nil, fmt.Errorf("cannot create tree with zero accounts")
	}

	// Sort accounts by ID for consistency
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].ID < accounts[j].ID
	})

	tree := &MerkleTree{}

	// Create leaf nodes
	tree.leaves = make([]Node, len(accounts))
	for i, account := range accounts {
		leafValue := fmt.Sprintf("%s:%s:%s:%d",
			account.ID, account.Balance.String(), account.Currency, account.Nonce)
		leafHash := hashValue(leafValue)

		tree.leaves[i] = Node{
			Hash:   leafHash,
			Value:  leafValue,
			Index:  i,
			IsLeaf: true,
		}
	}

	// Build tree layers
	if err := tree.buildTree(); err != nil {
		return nil, fmt.Errorf("failed to build tree: %w", err)
	}

	return tree, nil
}

// GetRoot returns the root hash of the tree
func (t *MerkleTree) GetRoot() string {
	if t.root == nil {
		return ""
	}
	return t.root.Hash
}

// GetProof generates a proof of inclusion for the account at given index
func (t *MerkleTree) GetProof(leafIndex int) (*MerkleProof, error) {
	if leafIndex < 0 || leafIndex >= len(t.leaves) {
		return nil, fmt.Errorf("invalid leaf index: %d", leafIndex)
	}

	leaf := &t.leaves[leafIndex]
	proof := &MerkleProof{
		LeafIndex:  leafIndex,
		LeafHash:   leaf.Hash,
		LeafValue:  leaf.Value,
		Root:       t.GetRoot(),
		Path:       []string{},
		Directions: []bool{},
	}

	// Traverse from leaf to root, collecting sibling hashes
	currentIndex := leafIndex
	for level := 0; level < len(t.nodes); level++ {
		// Determine sibling index
		var siblingIndex int
		var isRight bool

		if currentIndex%2 == 0 {
			// Current node is left child
			siblingIndex = currentIndex + 1
			isRight = true
		} else {
			// Current node is right child
			siblingIndex = currentIndex - 1
			isRight = false
		}

		// Add sibling hash to proof if it exists
		if siblingIndex < len(t.nodes[level]) {
			proof.Path = append(proof.Path, t.nodes[level][siblingIndex].Hash)
			proof.Directions = append(proof.Directions, isRight)
		}

		// Move to parent index
		currentIndex = currentIndex / 2
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof
func VerifyProof(proof *MerkleProof) bool {
	if proof == nil {
		return false
	}

	// Start with leaf hash
	currentHash := proof.LeafHash

	// Apply each step in the proof path
	for i, siblingHash := range proof.Path {
		if i >= len(proof.Directions) {
			return false
		}

		if proof.Directions[i] {
			// Sibling is on the right
			currentHash = hashPair(currentHash, siblingHash)
		} else {
			// Sibling is on the left
			currentHash = hashPair(siblingHash, currentHash)
		}
	}

	// Final hash should match the root
	return currentHash == proof.Root
}

// GetTotalBalance calculates the total balance of all accounts in the tree
func (t *MerkleTree) GetTotalBalance(currency string) (*math.Decimal, error) {
	total := math.NewDecimalFromInt(0)

	for _, leaf := range t.leaves {
		// Parse account from leaf value
		account, err := parseAccountFromLeaf(leaf.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse account: %w", err)
		}

		if account.Currency == currency {
			total = total.Add(account.Balance)
		}
	}

	return total, nil
}

// UpdateAccount updates an account in the tree and rebuilds affected nodes
func (t *MerkleTree) UpdateAccount(index int, account Account) error {
	if index < 0 || index >= len(t.leaves) {
		return fmt.Errorf("invalid leaf index: %d", index)
	}

	// Update leaf
	leafValue := fmt.Sprintf("%s:%s:%s:%d",
		account.ID, account.Balance.String(), account.Currency, account.Nonce)
	leafHash := hashValue(leafValue)

	t.leaves[index] = Node{
		Hash:   leafHash,
		Value:  leafValue,
		Index:  index,
		IsLeaf: true,
	}

	// Rebuild tree
	return t.buildTree()
}

// Private methods

func (t *MerkleTree) buildTree() error {
	if len(t.leaves) == 0 {
		return fmt.Errorf("no leaves to build tree")
	}

	// Start with leaves as the first level
	currentLevel := make([]Node, len(t.leaves))
	copy(currentLevel, t.leaves)

	t.nodes = [][]Node{currentLevel}

	// Build internal nodes level by level
	for len(currentLevel) > 1 {
		nextLevel := []Node{}

		// Pair up nodes and create parents
		for i := 0; i < len(currentLevel); i += 2 {
			left := &currentLevel[i]
			var right *Node

			if i+1 < len(currentLevel) {
				right = &currentLevel[i+1]
			} else {
				// Odd number of nodes - duplicate the last one
				right = left
			}

			// Create parent node
			parentHash := hashPair(left.Hash, right.Hash)
			parent := Node{
				Hash:   parentHash,
				Left:   left,
				Right:  right,
				Index:  len(nextLevel),
				IsLeaf: false,
			}

			nextLevel = append(nextLevel, parent)
		}

		t.nodes = append(t.nodes, nextLevel)
		currentLevel = nextLevel
	}

	// Root is the single node in the last level
	if len(currentLevel) == 1 {
		t.root = &currentLevel[0]
	}

	return nil
}

func hashValue(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func hashPair(left, right string) string {
	combined := left + right
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func parseAccountFromLeaf(leafValue string) (*Account, error) {
	// Parse "id:balance:currency:nonce" format
	// This is a simplified parser - in production you'd want more robust parsing
	parts := strings.Split(leafValue, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid leaf format")
	}

	balance, err := math.NewDecimalFromString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid balance: %w", err)
	}

	nonce, err := strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce: %w", err)
	}

	return &Account{
		ID:       parts[0],
		Balance:  balance,
		Currency: parts[2],
		Nonce:    nonce,
	}, nil
}
