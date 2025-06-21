package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// EncryptionManager handles AES encryption/decryption
type EncryptionManager struct {
	keySize int // 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
}

func NewEncryptionManager(keySize int) *EncryptionManager {
	return &EncryptionManager{
		keySize: keySize,
	}
}

// GenerateKey creates a random encryption key
func (em *EncryptionManager) GenerateKey() ([]byte, error) {
	key := make([]byte, em.keySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts data using AES-GCM
func (em *EncryptionManager) Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != em.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", em.keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-GCM
func (em *EncryptionManager) Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != em.keySize {
		return nil, fmt.Errorf("invalid key size: expected %d, got %d", em.keySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns hex-encoded result
func (em *EncryptionManager) EncryptString(plaintext string, key []byte) (string, error) {
	encrypted, err := em.Encrypt([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

// DecryptString decrypts a hex-encoded string
func (em *EncryptionManager) DecryptString(hexCiphertext string, key []byte) (string, error) {
	ciphertext, err := hex.DecodeString(hexCiphertext)
	if err != nil {
		return "", fmt.Errorf("invalid hex ciphertext: %w", err)
	}

	decrypted, err := em.Decrypt(ciphertext, key)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
