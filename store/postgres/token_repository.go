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

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty-core/client"
)

// AccessTokenRepository implements client.AccessTokenRepository
type AccessTokenRepository struct {
	db *DB
}

// NewAccessTokenRepository creates a new access token repository
func NewAccessTokenRepository(db *DB) *AccessTokenRepository {
	return &AccessTokenRepository{db: db}
}

// Create creates a new access token
func (r *AccessTokenRepository) Create(t *client.AccessToken) error {
	ctx := context.Background()

	var revokedAt sql.NullTime
	if t.RevokedAt != nil {
		revokedAt = sql.NullTime{Time: *t.RevokedAt, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO access_tokens (
			id, tenant_id, token_hash, client_id, user_id, 
			scope, token_type, expires_at, revoked_at, is_revoked, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		t.ID, t.TenantID, t.TokenHash, t.ClientID, t.UserID,
		t.Scope, t.TokenType, t.ExpiresAt, revokedAt, t.IsRevoked, t.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create access token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves an access token
func (r *AccessTokenRepository) GetByTokenHash(tokenHash string) (*client.AccessToken, error) {
	ctx := context.Background()

	var t client.AccessToken
	var revokedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, tenant_id, token_hash, client_id, user_id, 
			scope, token_type, expires_at, revoked_at, is_revoked, created_at
		FROM access_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&t.ID, &t.TenantID, &t.TokenHash, &t.ClientID, &t.UserID,
		&t.Scope, &t.TokenType, &t.ExpiresAt, &revokedAt, &t.IsRevoked, &t.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, client.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}

	return &t, nil
}

// Revoke revokes an access token
func (r *AccessTokenRepository) Revoke(tokenHash string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE access_tokens SET is_revoked = true, revoked_at = NOW()
		WHERE token_hash = $1
	`, tokenHash)

	if err != nil {
		return fmt.Errorf("failed to revoke access token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return client.ErrTokenNotFound
	}

	return nil
}

// DeleteExpired deletes all expired access tokens
func (r *AccessTokenRepository) DeleteExpired() error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `DELETE FROM access_tokens WHERE expires_at < NOW()`)

	if err != nil {
		return fmt.Errorf("failed to delete expired access tokens: %w", err)
	}

	return nil
}

// RefreshTokenRepository implements client.RefreshTokenRepository
type RefreshTokenRepository struct {
	db *DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(t *client.RefreshToken) error {
	ctx := context.Background()

	var revokedAt sql.NullTime
	if t.RevokedAt != nil {
		revokedAt = sql.NullTime{Time: *t.RevokedAt, Valid: true}
	}

	var accessTokenID sql.NullString
	if t.AccessTokenID != "" {
		accessTokenID = sql.NullString{String: t.AccessTokenID, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (
			id, tenant_id, token_hash, access_token_id, client_id, user_id, 
			scope, expires_at, revoked_at, is_revoked, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		t.ID, t.TenantID, t.TokenHash, accessTokenID, t.ClientID, t.UserID,
		t.Scope, t.ExpiresAt, revokedAt, t.IsRevoked, t.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a refresh token
func (r *RefreshTokenRepository) GetByTokenHash(tokenHash string) (*client.RefreshToken, error) {
	ctx := context.Background()

	var t client.RefreshToken
	var revokedAt sql.NullTime
	var accessTokenID sql.NullString

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, tenant_id, token_hash, access_token_id, client_id, user_id, 
			scope, expires_at, revoked_at, is_revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(
		&t.ID, &t.TenantID, &t.TokenHash, &accessTokenID, &t.ClientID, &t.UserID,
		&t.Scope, &t.ExpiresAt, &revokedAt, &t.IsRevoked, &t.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, client.ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if revokedAt.Valid {
		t.RevokedAt = &revokedAt.Time
	}
	if accessTokenID.Valid {
		t.AccessTokenID = accessTokenID.String
	}

	return &t, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(tokenHash string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE refresh_tokens SET is_revoked = true, revoked_at = NOW()
		WHERE token_hash = $1
	`, tokenHash)

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return client.ErrTokenNotFound
	}

	return nil
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired() error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < NOW()`)

	if err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return nil
}
