package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type JWTClaims struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
	jwt.RegisteredClaims
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

// HashPassword hashes a password using Argon2id with SECRET_KEY and returns a base64 encoded string
func HashPassword(password string) (string, error) {
	params := getDefaultParams()

	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash, err := hashPasswordWithSalt(password, salt)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	result := make([]byte, 0, len(salt)+len(hash))
	result = append(result, salt...)
	result = append(result, hash...)

	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}

// ValidatePassword validates a plain password against a hashed password
func ValidatePassword(plainPassword, hashedPassword string) error {
	params := getDefaultParams()

	data, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to decode hashed password: %w", err)
	}

	expectedLength := int(params.SaltLength + params.KeyLength)
	if len(data) != expectedLength {
		return fmt.Errorf("invalid hashed password format")
	}

	salt := data[:params.SaltLength]
	storedHash := data[params.SaltLength:]

	computedHash, err := hashPasswordWithSalt(plainPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if subtle.ConstantTimeCompare(storedHash, computedHash) != 1 {
		return fmt.Errorf("invalid password")
	}

	return nil
}

// hashPasswordWithSalt is a helper function that hashes a password with a given salt
func hashPasswordWithSalt(password string, salt []byte) ([]byte, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is not set")
	}

	params := getDefaultParams()

	// If no salt provided, generate a random one
	if salt == nil {
		salt = make([]byte, params.SaltLength)
		if _, err := rand.Read(salt); err != nil {
			return nil, fmt.Errorf("failed to generate salt: %w", err)
		}
	}

	// Hash the password + SECRET_KEY using Argon2id with the provided/generated salt
	hash := argon2.IDKey(
		[]byte(password+secretKey),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	return hash, nil
}

func EncryptString(str string) (string, error) {
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

	// Encrypt the string
	ciphertext := gcm.Seal(nil, nonce, []byte(str), nil)

	// Combine salt + nonce + ciphertext
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(result)

	return encoded, nil
}

func DecryptString(encryptedStr string) (string, error) {
	params := getDefaultParams()

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedStr)
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

	// Decrypt the string
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// getJWTSigningKey returns the signing key for JWT tokens
func getJWTSigningKey() ([]byte, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is not set")
	}
	return []byte(secretKey), nil
}

// CreateJWT creates a new JWT token for the given username with specified expiration duration
func CreateJWT(username string, expirationDuration time.Duration) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	signingKey, err := getJWTSigningKey()
	if err != nil {
		return "", fmt.Errorf("failed to get signing key: %w", err)
	}

	now := time.Now()
	expirationTime := now.Add(expirationDuration)

	claims := &JWTClaims{
		Username: username,
		Exp:      expirationTime.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims if valid
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	signingKey, err := getJWTSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get signing key: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Additional expiry check (jwt library also checks this, but we're being explicit)
	now := time.Now().Unix()
	if claims.Exp < now {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}
