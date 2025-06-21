package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Signature represents a digital signature
type Signature struct {
	R *big.Int
	S *big.Int
	V uint8 // Recovery ID for Ethereum
}

// KeyPair represents an ECDSA key pair
type KeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// SignatureManager handles digital signatures
type SignatureManager struct {
	curve elliptic.Curve
}

func NewSignatureManager() *SignatureManager {
	return &SignatureManager{
		curve: elliptic.P256(), // Can switch to secp256k1 for Ethereum
	}
}

// GenerateKeyPair creates a new ECDSA key pair
func (sm *SignatureManager) GenerateKeyPair() (*KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(sm.curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}, nil
}

// Sign creates a digital signature
func (sm *SignatureManager) Sign(data []byte, privateKey *ecdsa.PrivateKey) (*Signature, error) {
	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	return &Signature{
		R: r,
		S: s,
		V: 0, // Recovery ID calculation needed for Ethereum
	}, nil
}

// Verify checks a digital signature
func (sm *SignatureManager) Verify(data []byte, signature *Signature, publicKey *ecdsa.PublicKey) bool {
	hash := sha256.Sum256(data)
	return ecdsa.Verify(publicKey, hash[:], signature.R, signature.S)
}

// SignatureToHex converts signature to hex string
func (sig *Signature) ToHex() string {
	rHex := hex.EncodeToString(sig.R.Bytes())
	sHex := hex.EncodeToString(sig.S.Bytes())
	return fmt.Sprintf("%s%s%02x", rHex, sHex, sig.V)
}

// SignatureFromHex creates signature from hex string
func SignatureFromHex(hexStr string) (*Signature, error) {
	if len(hexStr) < 130 { // 64 + 64 + 2 chars
		return nil, fmt.Errorf("invalid signature hex length")
	}

	rBytes, err := hex.DecodeString(hexStr[:64])
	if err != nil {
		return nil, fmt.Errorf("invalid R component: %w", err)
	}

	sBytes, err := hex.DecodeString(hexStr[64:128])
	if err != nil {
		return nil, fmt.Errorf("invalid S component: %w", err)
	}

	vByte, err := hex.DecodeString(hexStr[128:130])
	if err != nil {
		return nil, fmt.Errorf("invalid V component: %w", err)
	}

	return &Signature{
		R: new(big.Int).SetBytes(rBytes),
		S: new(big.Int).SetBytes(sBytes),
		V: vByte[0],
	}, nil
}
