package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func getDefaultParams() Argon2Params {
	return Argon2Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,         // 3 iterations
		Parallelism: 2,         // 2 parallel threads
		SaltLength:  16,        // 16 bytes salt
		KeyLength:   32,        // 32 bytes key length (AES-256)
	}
}

func deriveKey(salt []byte) ([]byte, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is not set")
	}

	params := getDefaultParams()

	key := argon2.IDKey(
		[]byte(secretKey),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return key, nil
}

func EncryptPassword(password string) (string, error) {
	params := getDefaultParams()

	// Generate random salt for key derivation
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key using Argon2id
	key, err := deriveKey(salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the password
	ciphertext := gcm.Seal(nil, nonce, []byte(password), nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}

func DecryptPassword(encryptedPassword string) (string, error) {
	params := getDefaultParams()

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length (salt + nonce + at least some ciphertext)
	minLength := int(params.SaltLength) + 12 + 16 // salt + nonce + GCM tag
	if len(data) < minLength {
		return "", fmt.Errorf("encrypted data too short")
	}

	// Extract salt
	salt := data[:params.SaltLength]

	// Derive decryption key using Argon2id
	key, err := deriveKey(salt)
	if err != nil {
		return "", fmt.Errorf("failed to derive key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(data) < int(params.SaltLength)+nonceSize {
		return "", fmt.Errorf("encrypted data too short for nonce")
	}

	nonce := data[params.SaltLength : params.SaltLength+uint32(nonceSize)]
	ciphertext := data[params.SaltLength+uint32(nonceSize):]

	// Decrypt the password
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
