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

// AuthorizationCodeRepository implements client.AuthorizationCodeRepository
type AuthorizationCodeRepository struct {
	db *DB
}

// NewAuthorizationCodeRepository creates a new authorization code repository
func NewAuthorizationCodeRepository(db *DB) *AuthorizationCodeRepository {
	return &AuthorizationCodeRepository{db: db}
}

// Create creates a new authorization code
func (r *AuthorizationCodeRepository) Create(c *client.AuthorizationCode) error {
	ctx := context.Background()

	var usedAt sql.NullTime
	if c.UsedAt != nil {
		usedAt = sql.NullTime{Time: *c.UsedAt, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO authorization_codes (
			id, code, client_id, user_id, 
			redirect_uri, scope, state, nonce,
			code_challenge, code_challenge_method,
			expires_at, used_at, is_used, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`,
		c.ID, c.Code, c.ClientID, c.UserID,
		c.RedirectURI, c.Scope, c.State, c.Nonce,
		c.CodeChallenge, c.CodeChallengeMethod,
		c.ExpiresAt, usedAt, c.IsUsed, c.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create authorization code: %w", err)
	}

	return nil
}

// GetByCode retrieves an authorization code
func (r *AuthorizationCodeRepository) GetByCode(codeStr string) (*client.AuthorizationCode, error) {
	ctx := context.Background()

	var c client.AuthorizationCode
	var usedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, code, client_id, user_id, 
			redirect_uri, scope, state, nonce,
			code_challenge, code_challenge_method,
			expires_at, used_at, is_used, created_at
		FROM authorization_codes
		WHERE code = $1
	`, codeStr).Scan(
		&c.ID, &c.Code, &c.ClientID, &c.UserID,
		&c.RedirectURI, &c.Scope, &c.State, &c.Nonce,
		&c.CodeChallenge, &c.CodeChallengeMethod,
		&c.ExpiresAt, &usedAt, &c.IsUsed, &c.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, client.ErrCodeNotFound
		}
		return nil, fmt.Errorf("failed to get authorization code: %w", err)
	}

	if usedAt.Valid {
		c.UsedAt = &usedAt.Time
	}

	return &c, nil
}

// MarkAsUsed marks the code as used
func (r *AuthorizationCodeRepository) MarkAsUsed(code string) error {
	ctx := context.Background()

	result, err := r.db.pool.Exec(ctx, `
		UPDATE authorization_codes SET is_used = true, used_at = NOW()
		WHERE code = $1
	`, code)

	if err != nil {
		return fmt.Errorf("failed to mark code as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return client.ErrCodeNotFound
	}

	return nil
}

// Delete deletes an authorization code
func (r *AuthorizationCodeRepository) Delete(code string) error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM authorization_codes WHERE code = $1
	`, code)

	if err != nil {
		return fmt.Errorf("failed to delete code: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired authorization codes
func (r *AuthorizationCodeRepository) DeleteExpired() error {
	ctx := context.Background()

	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM authorization_codes WHERE expires_at < NOW()
	`)

	if err != nil {
		return fmt.Errorf("failed to delete expired codes: %w", err)
	}

	return nil
}
