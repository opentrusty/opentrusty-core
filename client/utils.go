// Copyright 2026 The OpenTrusty Authors
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// GenerateClientSecret generates a new cryptographically strong client secret
func GenerateClientSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// HashClientSecret hashes a client secret for storage
func HashClientSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// Validation errors
var (
	ErrInvalidRedirectURI = errors.New("invalid redirect_uri format")
	ErrInvalidClientURI   = errors.New("invalid client_uri format")
)
