package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

// RandomBytes generates random bytes
func RandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// RandomHex generates a random hex string
func RandomHex(length int) (string, error) {
	bytes, err := RandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RandomBigInt generates a random big integer within a range
func RandomBigInt(max *big.Int) (*big.Int, error) {
	return rand.Int(rand.Reader, max)
}

// PadBytes pads bytes to a specific length
func PadBytes(data []byte, length int) []byte {
	if len(data) >= length {
		return data
	}

	padded := make([]byte, length)
	copy(padded[length-len(data):], data)
	return padded
}

// XOR performs XOR operation on two byte slices
func XOR(a, b []byte) ([]byte, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("byte slices must have same length")
	}

	result := make([]byte, len(a))
	for i := range a {
		result[i] = a[i] ^ b[i]
	}
	return result, nil
}

// IsValidPrivateKey checks if a big.Int is a valid private key for BN254
func IsValidPrivateKey(key *big.Int) bool {
	// BN254 field modulus
	modulus := new(big.Int)
	modulus.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	return key.Cmp(big.NewInt(0)) > 0 && key.Cmp(modulus) < 0
}

// ModInverse calculates modular inverse
func ModInverse(a, m *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, m)
}

// PowerMod calculates (base^exp) mod mod
func PowerMod(base, exp, mod *big.Int) *big.Int {
	return new(big.Int).Exp(base, exp, mod)
}
