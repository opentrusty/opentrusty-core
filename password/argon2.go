package password

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Hasher handles password hashing using Argon2id.
//
// Purpose: Primary mechanism for secure password storage and verification.
// Domain: Identity
// Invariants: Memory, Iterations, and Parallelism must be tuned for security.
type Hasher struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// NewHasher creates a new password hasher with Argon2id default settings.
//
// Purpose: Constructor for the password hashing utility.
// Domain: Identity
// Audited: No
// Errors: None
func NewHasher(memory, iterations uint32, parallelism uint8, saltLength, keyLength uint32) *Hasher {
	return &Hasher{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  saltLength,
		KeyLength:   keyLength,
	}
}

// Hash hashes a password using Argon2id.
//
// Purpose: Generates a cryptographically secure hash of a plaintext password.
// Domain: Identity
// Security: Uses Argon2id (memory-hard, side-channel resistant) with random salt.
// Audited: No
// Errors: System errors (e.g., random generation failure)
func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.Iterations,
		h.Memory,
		h.Parallelism,
		h.KeyLength,
	)

	return fmt.Sprintf(
		"=%d=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.Memory,
		h.Iterations,
		h.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// Verify verifies a password against a hash.
//
// Purpose: Validates an incoming password against a stored Argon2id hash.
// Domain: Identity
// Security: Uses constant-time comparison to prevent timing attacks.
// Audited: No
// Errors: Invalid hash format, decoding errors
func (h *Hasher) Verify(password, encodedHash string) (bool, error) {
	var version int
	var memory, iterations uint32
	var parallelism uint8
	var saltB64, hashB64 string

	_, err := fmt.Sscanf(encodedHash, "=%d=%d,t=%d,p=%d$%s$%s",
		&version, &memory, &iterations, &parallelism, &saltB64, &hashB64)
	if err != nil {
		return false, fmt.Errorf("invalid hash format: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expectedHash)),
	)

	if len(actualHash) != len(expectedHash) {
		return false, nil
	}

	var diff byte
	for i := range actualHash {
		diff |= actualHash[i] ^ expectedHash[i]
	}

	return diff == 0, nil
}
