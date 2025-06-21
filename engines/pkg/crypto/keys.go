package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
)

// KeyManager handles key generation and derivation
type KeyManager struct {
	iterations int
	keyLength  int
}

func NewKeyManager() *KeyManager {
	return &KeyManager{
		iterations: 100000, // PBKDF2 iterations
		keyLength:  32,     // 256-bit keys
	}
}

// DeriveKey derives a key from password using PBKDF2
func (km *KeyManager) DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, km.iterations, km.keyLength, sha3.New256)
}

// GenerateSalt creates a random salt
func (km *KeyManager) GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GeneratePrivateKey generates a random private key for ZK proofs
func (km *KeyManager) GeneratePrivateKey() (*big.Int, error) {
	// Generate random 254-bit number (BN254 field size)
	max := new(big.Int)
	max.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	privateKey, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	return privateKey, nil
}

// PrivateKeyToHex converts private key to hex string
func (km *KeyManager) PrivateKeyToHex(privateKey *big.Int) string {
	return hex.EncodeToString(privateKey.Bytes())
}

// PrivateKeyFromHex creates private key from hex string
func (km *KeyManager) PrivateKeyFromHex(hexStr string) (*big.Int, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	return new(big.Int).SetBytes(bytes), nil
}

// GenerateCommitment creates a Pedersen commitment
func (km *KeyManager) GenerateCommitment(value, randomness *big.Int) (*Point, error) {
	curveManager := NewCurveManager()

	// Generate base points (in practice, these should be fixed trusted setup)
	g, err := curveManager.GenerateRandomPoint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate G point: %w", err)
	}

	h, err := curveManager.GenerateRandomPoint()
	if err != nil {
		return nil, fmt.Errorf("failed to generate H point: %w", err)
	}

	// Commitment = value * G + randomness * H
	valueG, err := curveManager.ScalarMultiply(g, value)
	if err != nil {
		return nil, fmt.Errorf("failed to multiply value with G: %w", err)
	}

	randomnessH, err := curveManager.ScalarMultiply(h, randomness)
	if err != nil {
		return nil, fmt.Errorf("failed to multiply randomness with H: %w", err)
	}

	commitment, err := curveManager.AddPoints(valueG, randomnessH)
	if err != nil {
		return nil, fmt.Errorf("failed to add points: %w", err)
	}

	return commitment, nil
}
