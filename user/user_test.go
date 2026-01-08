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

package user

import (
	"context"
	"testing"
	"time"

	"github.com/opentrusty/opentrusty-core/audit"
	"github.com/opentrusty/opentrusty-core/crypto"
)

// MockUserRepository implements UserRepository for testing
type MockUserRepository struct {
	users       map[string]*User
	credentials map[string]*Credentials
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:       make(map[string]*User),
		credentials: make(map[string]*Credentials),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) AddCredentials(ctx context.Context, credentials *Credentials) error {
	m.credentials[credentials.UserID] = credentials
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *MockUserRepository) GetByHash(ctx context.Context, hash string) (*User, error) {
	for _, u := range m.users {
		if u.EmailHash == hash {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *MockUserRepository) Update(ctx context.Context, user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateLockout(ctx context.Context, userID string, failedAttempts int, lockedUntil *time.Time) error {
	u, ok := m.users[userID]
	if !ok {
		return ErrUserNotFound
	}
	u.FailedLoginAttempts = failedAttempts
	u.LockedUntil = lockedUntil
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) GetCredentials(ctx context.Context, userID string) (*Credentials, error) {
	c, ok := m.credentials[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return c, nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	c, ok := m.credentials[userID]
	if !ok {
		return ErrUserNotFound
	}
	c.PasswordHash = passwordHash
	return nil
}

// MockAuditLogger implements audit.Logger for testing
type MockAuditLogger struct{}

func (m *MockAuditLogger) Log(ctx context.Context, event audit.Event) {}

func TestEmailNormalizationAndHashing(t *testing.T) {
	hmacKey := "test-key"
	email1 := "User@Example.Com "
	email2 := "user@example.com"

	hash1 := crypto.ComputeEmailHash(hmacKey, email1)
	hash2 := crypto.ComputeEmailHash(hmacKey, email2)

	if hash1 != hash2 {
		t.Errorf("expected hashes to match for normalized emails")
	}
}

func TestProvisionIdentity(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := NewPasswordHasher(65536, 1, 1, 16, 32)
	svc := NewService(repo, hasher, &MockAuditLogger{}, 5, time.Hour, "test-key")

	profile := Profile{
		GivenName:  "Test",
		FamilyName: "User",
	}

	u, err := svc.ProvisionIdentity(context.Background(), "test@example.com", profile)
	if err != nil {
		t.Fatalf("failed to provision identity: %v", err)
	}

	if u.EmailHash == "" {
		t.Error("expected email hash to be computed")
	}

	if u.Profile.Nickname != "test" {
		t.Errorf("expected nickname 'test', got %s", u.Profile.Nickname)
	}

	// Try provisioning same email again
	_, err = svc.ProvisionIdentity(context.Background(), "test@example.com", profile)
	if err == nil {
		t.Error("expected error for duplicate email")
	}
}

func TestAuthentication(t *testing.T) {
	repo := NewMockUserRepository()
	hasher := NewPasswordHasher(1024, 1, 1, 16, 32)
	svc := NewService(repo, hasher, &MockAuditLogger{}, 3, time.Hour, "test-key")

	email := "auth@example.com"
	password := "secure-password"

	u, _ := svc.ProvisionIdentity(context.Background(), email, Profile{})
	_ = svc.AddPassword(context.Background(), u.ID, password)

	// Test success
	authU, err := svc.Authenticate(context.Background(), email, password)
	if err != nil {
		t.Fatalf("authentication failed: %v", err)
	}
	if authU.ID != u.ID {
		t.Error("authenticated user ID mismatch")
	}

	// Test invalid password
	_, err = svc.Authenticate(context.Background(), email, "wrong-password")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}

	// Test account lockout
	_, _ = svc.Authenticate(context.Background(), email, "wrong-password")
	_, _ = svc.Authenticate(context.Background(), email, "wrong-password")
	_, err = svc.Authenticate(context.Background(), email, "wrong-password")

	if err != ErrAccountLocked {
		t.Errorf("expected ErrAccountLocked after max attempts, got %v", err)
	}
}
