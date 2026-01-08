package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ComputeEmailHash computes a HMAC-SHA256 hash for an email using the provided key.
//
// Purpose: Generates a stable, opaque primary identifier for users to prevent email exposure in secondary indices.
// Domain: Identity
// Invariants: Normalizes email to lowercase and trims whitespace before hashing.
// Audited: No
// Errors: None
func ComputeEmailHash(key string, emailPlain string) string {
	normalized := strings.TrimSpace(strings.ToLower(emailPlain))

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(normalized))

	return hex.EncodeToString(h.Sum(nil))
}
