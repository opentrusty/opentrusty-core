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

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty-core/user"
)

// UserRepository implements user.UserRepository.
//
// Purpose: PostgreSQL implementation of user identity persistence.
// Domain: Identity (Infrastructure)
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository.
//
// Purpose: Constructor for the user persistence layer.
// Domain: Identity (Infrastructure)
// Audited: No
// Errors: None
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user identity.
//
// Purpose: Persists a new user record to the database.
// Domain: Identity (Infrastructure)
// Audited: No
// Errors: System errors
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	now := time.Now()
	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO users (
			id, email_hash, email_plain, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		u.ID, u.EmailHash, u.EmailPlain, u.EmailVerified,
		u.Profile.GivenName, u.Profile.FamilyName, u.Profile.FullName,
		u.Profile.Nickname, u.Profile.Picture, u.Profile.Locale, u.Profile.Timezone,
		now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	u.CreatedAt = now
	u.UpdatedAt = now

	return nil
}

// AddCredentials adds credentials for a user
func (r *UserRepository) AddCredentials(ctx context.Context, c *user.Credentials) error {
	now := time.Now()
	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO credentials (user_id, password_hash, updated_at)
		VALUES ($1, $2, $3)
	`, c.UserID, c.PasswordHash, now)
	if err != nil {
		return fmt.Errorf("failed to insert credentials: %w", err)
	}

	c.UpdatedAt = now

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, email_hash, email_plain, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(
		&u.ID, &u.EmailHash, &u.EmailPlain, &u.EmailVerified,
		&u.Profile.GivenName, &u.Profile.FamilyName, &u.Profile.FullName,
		&u.Profile.Nickname, &u.Profile.Picture, &u.Profile.Locale, &u.Profile.Timezone,
		&u.CreatedAt, &u.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}

	return &u, nil
}

// GetByHash retrieves a user by their global email hash
func (r *UserRepository) GetByHash(ctx context.Context, hash string) (*user.User, error) {
	var u user.User
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT id, email_hash, email_plain, email_verified,
			given_name, family_name, full_name, nickname, picture, locale, timezone,
			created_at, updated_at, deleted_at
		FROM users
		WHERE email_hash = $1 AND deleted_at IS NULL
	`, hash).Scan(
		&u.ID, &u.EmailHash, &u.EmailPlain, &u.EmailVerified,
		&u.Profile.GivenName, &u.Profile.FamilyName, &u.Profile.FullName,
		&u.Profile.Nickname, &u.Profile.Picture, &u.Profile.Locale, &u.Profile.Timezone,
		&u.CreatedAt, &u.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by hash: %w", err)
	}

	if deletedAt.Valid {
		u.DeletedAt = &deletedAt.Time
	}

	return &u, nil
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE users SET
			email_plain = $2,
			email_verified = $3,
			given_name = $4,
			family_name = $5,
			full_name = $6,
			nickname = $7,
			picture = $8,
			locale = $9,
			timezone = $10,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`,
		u.ID, u.EmailPlain, u.EmailVerified,
		u.Profile.GivenName, u.Profile.FamilyName, u.Profile.FullName,
		u.Profile.Nickname, u.Profile.Picture, u.Profile.Locale, u.Profile.Timezone,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// UpdateLockout updates user lockout status
func (r *UserRepository) UpdateLockout(ctx context.Context, userID string, failedAttempts int, lockedUntil *time.Time) error {
	_, err := r.db.pool.Exec(ctx, `
		UPDATE users
		SET failed_login_attempts = $1, locked_until = $2, updated_at = NOW()
		WHERE id = $3
	`, failedAttempts, lockedUntil, userID)
	if err != nil {
		return fmt.Errorf("failed to update user lockout status: %w", err)
	}
	return nil
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE users SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}

// GetCredentials retrieves user credentials
func (r *UserRepository) GetCredentials(ctx context.Context, userID string) (*user.Credentials, error) {
	var c user.Credentials
	err := r.db.pool.QueryRow(ctx, `
		SELECT user_id, password_hash, updated_at
		FROM credentials
		WHERE user_id = $1
	`, userID).Scan(&c.UserID, &c.PasswordHash, &c.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	return &c, nil
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE credentials SET password_hash = $2, updated_at = NOW()
		WHERE user_id = $1
	`, userID, passwordHash)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrUserNotFound
	}

	return nil
}
