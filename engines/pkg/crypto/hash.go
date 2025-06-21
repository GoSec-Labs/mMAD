package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

type HashFunction string

const (
	SHA256Hash    HashFunction = "sha256"
	Keccak256Hash HashFunction = "keccak256"
	PoseidonHash  HashFunction = "poseidon"
)

type Hasher interface {
	Hash(data []byte) []byte
	HashString(data string) string
	HashBigInt(data *big.Int) *big.Int
}

// SHA256Hasher implements SHA256 hashing
type SHA256Hasher struct{}

func NewSHA256Hasher() *SHA256Hasher {
	return &SHA256Hasher{}
}

func (h *SHA256Hasher) Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (h *SHA256Hasher) HashString(data string) string {
	hash := h.Hash([]byte(data))
	return hex.EncodeToString(hash)
}

func (h *SHA256Hasher) HashBigInt(data *big.Int) *big.Int {
	hash := h.Hash(data.Bytes())
	return new(big.Int).SetBytes(hash)
}

type Keccak256Hasher struct{}

func NewKeccak256Hasher() *Keccak256Hasher {
	return &Keccak256Hasher{}
}

func (h *Keccak256Hasher) Hash(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

func (h *Keccak256Hasher) HashString(data string) string {
	hash := h.Hash([]byte(data))
	return hex.EncodeToString(hash)
}

func (h *Keccak256Hasher) HashBigInt(data *big.Int) *big.Int {
	hash := h.Hash(data.Bytes())
	return new(big.Int).SetBytes(hash)
}


func SHA256(data []byte) []byte {
	hasher := NewSHA256Hasher()
	return hasher.Hash(data)
}

func Keccak256(data []byte) []byte {
	hasher := NewKeccak256Hasher()
	return hasher.Hash(data)
}

func HashToHex(data []byte, hashFunc HashFunction) (string, error) {
	switch hashFunc {
	case SHA256Hash:
		return NewSHA256Hasher().HashString(string(data)), nil
	case Keccak256Hash:
		return NewKeccak256Hasher().HashString(string(data)), nil
	default:
		return "", fmt.Errorf("unsupported hash function: %s", hashFunc)
	}
}
