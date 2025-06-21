package crypto

import (
	"encoding/hex"
	"fmt"
)

// MerkleNode represents a node in the Merkle tree
type MerkleNode struct {
	Hash   []byte
	Left   *MerkleNode
	Right  *MerkleNode
	Parent *MerkleNode
}

// MerkleTree represents a Merkle tree
type MerkleTree struct {
	Root     *MerkleNode
	Leaves   []*MerkleNode
	hashFunc Hasher
}

// NewMerkleTree creates a new Merkle tree
func NewMerkleTree(data [][]byte, hashFunc Hasher) (*MerkleTree, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}

	tree := &MerkleTree{
		hashFunc: hashFunc,
	}

	// Create leaf nodes
	var leaves []*MerkleNode
	for _, item := range data {
		hash := hashFunc.Hash(item)
		leaf := &MerkleNode{Hash: hash}
		leaves = append(leaves, leaf)
	}

	tree.Leaves = leaves
	tree.Root = tree.buildTree(leaves)

	return tree, nil
}

// buildTree recursively builds the Merkle tree
func (mt *MerkleTree) buildTree(nodes []*MerkleNode) *MerkleNode {
	if len(nodes) == 1 {
		return nodes[0]
	}

	var parentNodes []*MerkleNode

	for i := 0; i < len(nodes); i += 2 {
		left := nodes[i]
		var right *MerkleNode

		if i+1 < len(nodes) {
			right = nodes[i+1]
		} else {
			// Duplicate the last node if odd number of nodes
			right = nodes[i]
		}

		// Calculate parent hash
		combined := append(left.Hash, right.Hash...)
		parentHash := mt.hashFunc.Hash(combined)

		parent := &MerkleNode{
			Hash:  parentHash,
			Left:  left,
			Right: right,
		}

		left.Parent = parent
		right.Parent = parent

		parentNodes = append(parentNodes, parent)
	}

	return mt.buildTree(parentNodes)
}

// GetProof generates a Merkle proof for a given data item
func (mt *MerkleTree) GetProof(data []byte) ([][]byte, error) {
	targetHash := mt.hashFunc.Hash(data)

	// Find the leaf node
	var targetLeaf *MerkleNode
	for _, leaf := range mt.Leaves {
		if string(leaf.Hash) == string(targetHash) {
			targetLeaf = leaf
			break
		}
	}

	if targetLeaf == nil {
		return nil, fmt.Errorf("data not found in tree")
	}

	var proof [][]byte
	current := targetLeaf

	for current.Parent != nil {
		parent := current.Parent
		if parent.Left == current {
			// Current is left child, include right sibling
			proof = append(proof, parent.Right.Hash)
		} else {
			// Current is right child, include left sibling
			proof = append(proof, parent.Left.Hash)
		}
		current = parent
	}

	return proof, nil
}

// VerifyProof verifies a Merkle proof
func (mt *MerkleTree) VerifyProof(data []byte, proof [][]byte, rootHash []byte) bool {
	hash := mt.hashFunc.Hash(data)

	for _, siblingHash := range proof {
		// Try both orders to handle left/right positioning
		combined1 := append(hash, siblingHash...)
		combined2 := append(siblingHash, hash...)

		hash1 := mt.hashFunc.Hash(combined1)
		hash2 := mt.hashFunc.Hash(combined2)

		// Use the hash that would lead to the root
		if string(hash1) == string(rootHash) || len(proof) > 1 {
			hash = hash1
		} else {
			hash = hash2
		}
	}

	return string(hash) == string(rootHash)
}

// GetRootHash returns the root hash
func (mt *MerkleTree) GetRootHash() []byte {
	if mt.Root == nil {
		return nil
	}
	return mt.Root.Hash
}

func (mt *MerkleTree) getProofRecursive(node *MerkleNode, targetHash []byte, proof [][]byte) [][]byte {
	if node == nil {
		return nil
	}

	// Check if this is a leaf node with target hash
	if node.Left == nil && node.Right == nil {
		if hex.EncodeToString(node.Hash) == hex.EncodeToString(targetHash) {
			return proof
		}
		return nil
	}

	// Search left subtree
	if leftProof := mt.getProofRecursive(node.Left, targetHash, append(proof, node.Right.Hash)); leftProof != nil {
		return leftProof
	}

	// Search right subtree
	if rightProof := mt.getProofRecursive(node.Right, targetHash, append(proof, node.Left.Hash)); rightProof != nil {
		return rightProof
	}

	return nil
}
