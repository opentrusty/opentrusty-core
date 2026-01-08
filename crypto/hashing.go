// Copyright 2026 The OpenTrusty Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
