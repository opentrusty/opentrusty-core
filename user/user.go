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

package user

import (
	"context"
	"errors"
	"time"
)

// Domain errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
	ErrAccountLocked      = errors.New("account is locked")
)

// Platform Authorization Principles:
// 1. No tenant represents the platform
// 2. Platform authorization is expressed only via scoped roles
// 3. Tenant context must never be elevated to platform context
//
// Anti-Patterns (FORBIDDEN):
// - Magic tenant IDs
// - Empty tenantID implying platform privileges

// User represents a user identity in the system.
//
// Purpose: Core identity entity representing a digital actor.
// Domain: Identity
// Invariants: ID must be a UUIDv7. EmailHash must be a valid HMAC-SHA256 of the normalized email.
type User struct {
	ID         string
	EmailHash  string  // Global Identity Key (HMAC-SHA256)
	EmailPlain *string // Nullable PII Metadata

	EmailVerified       bool
	Profile             Profile
	FailedLoginAttempts int
	LockedUntil         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

// Profile represents user profile information.
//
// Purpose: PII metadata associated with a user identity.
// Domain: Identity
type Profile struct {
	GivenName  string
	FamilyName string
	FullName   string
	Nickname   string
	Picture    string
	Locale     string
	Timezone   string
}

// Credentials represents user authentication credentials
type Credentials struct {
	UserID       string
	PasswordHash string
	UpdatedAt    time.Time
}

// UserRepository defines the interface for user persistence.
//
// Purpose: Abstraction for managing user identity storage.
// Domain: Identity
type UserRepository interface {
	// Create creates a new user identity
	Create(ctx context.Context, user *User) error

	// AddCredentials adds credentials for a user
	AddCredentials(ctx context.Context, credentials *Credentials) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*User, error)

	// GetByHash retrieves a user by their global email hash
	GetByHash(ctx context.Context, hash string) (*User, error)

	// Update updates user information
	Update(ctx context.Context, user *User) error

	// UpdateLockout updates user lockout status
	UpdateLockout(ctx context.Context, userID string, failedAttempts int, lockedUntil *time.Time) error

	// Delete soft-deletes a user
	Delete(ctx context.Context, id string) error

	// GetCredentials retrieves user credentials
	GetCredentials(ctx context.Context, userID string) (*Credentials, error)

	// UpdatePassword updates user password
	UpdatePassword(ctx context.Context, userID string, passwordHash string) error
}
