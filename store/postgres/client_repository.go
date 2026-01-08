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

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty-core/client"
)

// ClientRepository implements client.ClientRepository
type ClientRepository struct {
	db *DB
}

// NewClientRepository creates a new client repository
func NewClientRepository(db *DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// Create creates a new OAuth2 client
func (r *ClientRepository) Create(ctx context.Context, c *client.Client) error {
	redirectURIs, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	allowedScopes, err := json.Marshal(c.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantTypes, err := json.Marshal(c.GrantTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal grant types: %w", err)
	}

	responseTypes, err := json.Marshal(c.ResponseTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal response types: %w", err)
	}

	var ownerID sql.NullString
	if c.OwnerID != "" {
		ownerID = sql.NullString{String: c.OwnerID, Valid: true}
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	_, err = r.db.pool.Exec(ctx, `
		INSERT INTO oauth2_clients (
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	`,
		c.ID, c.ClientID, c.TenantID, c.ClientSecretHash, c.ClientName, c.ClientURI, c.LogoURI,
		redirectURIs, allowedScopes, grantTypes, responseTypes,
		c.TokenEndpointAuthMethod, c.AccessTokenLifetime, c.RefreshTokenLifetime, c.IDTokenLifetime,
		ownerID, c.IsTrusted, c.IsActive, c.CreatedAt, c.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

// GetByClientID retrieves a client by client_id and tenant_id
func (r *ClientRepository) GetByClientID(ctx context.Context, tenantID string, clientID string) (*client.Client, error) {
	var c client.Client
	var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
	var clientURI, logoURI, ownerID sql.NullString
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE client_id = $2 AND ($1 = '' OR tenant_id::text = $1) AND deleted_at IS NULL
	`, tenantID, clientID).Scan(
		&c.ID, &c.ClientID, &c.TenantID, &c.ClientSecretHash, &c.ClientName, &clientURI, &logoURI,
		&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
		&c.TokenEndpointAuthMethod, &c.AccessTokenLifetime, &c.RefreshTokenLifetime, &c.IDTokenLifetime,
		&ownerID, &c.IsTrusted, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, client.ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(redirectURIsJSON, &c.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
	}
	if err := json.Unmarshal(allowedScopesJSON, &c.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal(grantTypesJSON, &c.GrantTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grant types: %w", err)
	}
	if err := json.Unmarshal(responseTypesJSON, &c.ResponseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response types: %w", err)
	}

	if clientURI.Valid {
		c.ClientURI = clientURI.String
	}
	if logoURI.Valid {
		c.LogoURI = logoURI.String
	}
	if ownerID.Valid {
		c.OwnerID = ownerID.String
	}
	if deletedAt.Valid {
		c.DeletedAt = &deletedAt.Time
	}

	return &c, nil
}

// GetByID retrieves a client by tenant_id and internal ID
func (r *ClientRepository) GetByID(ctx context.Context, tenantID string, id string) (*client.Client, error) {
	var c client.Client
	var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
	var ownerID sql.NullString
	var deletedAt sql.NullTime

	err := r.db.pool.QueryRow(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE id = $2 AND tenant_id = $1 AND deleted_at IS NULL
	`, tenantID, id).Scan(
		&c.ID, &c.ClientID, &c.TenantID, &c.ClientSecretHash, &c.ClientName, &c.ClientURI, &c.LogoURI,
		&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
		&c.TokenEndpointAuthMethod, &c.AccessTokenLifetime, &c.RefreshTokenLifetime, &c.IDTokenLifetime,
		&ownerID, &c.IsTrusted, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, client.ErrClientNotFound
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(redirectURIsJSON, &c.RedirectURIs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redirect URIs: %w", err)
	}
	if err := json.Unmarshal(allowedScopesJSON, &c.AllowedScopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal allowed scopes: %w", err)
	}
	if err := json.Unmarshal(grantTypesJSON, &c.GrantTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grant types: %w", err)
	}
	if err := json.Unmarshal(responseTypesJSON, &c.ResponseTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response types: %w", err)
	}

	if ownerID.Valid {
		c.OwnerID = ownerID.String
	}
	if deletedAt.Valid {
		c.DeletedAt = &deletedAt.Time
	}

	return &c, nil
}

// Update updates client information
func (r *ClientRepository) Update(ctx context.Context, c *client.Client) error {
	redirectURIs, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return fmt.Errorf("failed to marshal redirect URIs: %w", err)
	}

	allowedScopes, err := json.Marshal(c.AllowedScopes)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed scopes: %w", err)
	}

	grantTypes, err := json.Marshal(c.GrantTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal grant types: %w", err)
	}

	responseTypes, err := json.Marshal(c.ResponseTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal response types: %w", err)
	}

	result, err := r.db.pool.Exec(ctx, `
		UPDATE oauth2_clients SET
			client_name = $2,
			client_uri = $3,
			logo_uri = $4,
			redirect_uris = $5,
			allowed_scopes = $6,
			grant_types = $7,
			response_types = $8,
			token_endpoint_auth_method = $9,
			access_token_lifetime = $10,
			refresh_token_lifetime = $11,
			id_token_lifetime = $12,
			is_trusted = $13,
			is_active = $14,
			updated_at = NOW()
		WHERE id = $1 AND tenant_id = $15 AND deleted_at IS NULL
	`,
		c.ID, c.ClientName, c.ClientURI, c.LogoURI,
		redirectURIs, allowedScopes, grantTypes, responseTypes,
		c.TokenEndpointAuthMethod, c.AccessTokenLifetime, c.RefreshTokenLifetime, c.IDTokenLifetime,
		c.IsTrusted, c.IsActive, c.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return client.ErrClientNotFound
	}

	return nil
}

// Delete soft-deletes a client by tenant_id and internal ID
func (r *ClientRepository) Delete(ctx context.Context, tenantID string, id string) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE oauth2_clients SET deleted_at = $3
		WHERE id = $2 AND tenant_id = $1 AND deleted_at IS NULL
	`, tenantID, id, time.Now())

	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return client.ErrClientNotFound
	}

	return nil
}

// ListByOwner retrieves all clients for an owner
func (r *ClientRepository) ListByOwner(ctx context.Context, ownerID string) ([]*client.Client, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE owner_id = $1 AND deleted_at IS NULL
	`, ownerID)

	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*client.Client
	for rows.Next() {
		var c client.Client
		var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
		var ownerID sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.ClientID, &c.TenantID, &c.ClientSecretHash, &c.ClientName, &c.ClientURI, &c.LogoURI,
			&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
			&c.TokenEndpointAuthMethod, &c.AccessTokenLifetime, &c.RefreshTokenLifetime, &c.IDTokenLifetime,
			&ownerID, &c.IsTrusted, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}

		if err := json.Unmarshal(redirectURIsJSON, &c.RedirectURIs); err != nil {
			continue
		}
		if err := json.Unmarshal(allowedScopesJSON, &c.AllowedScopes); err != nil {
			continue
		}
		if err := json.Unmarshal(grantTypesJSON, &c.GrantTypes); err != nil {
			continue
		}
		if err := json.Unmarshal(responseTypesJSON, &c.ResponseTypes); err != nil {
			continue
		}

		if ownerID.Valid {
			c.OwnerID = ownerID.String
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}

		clients = append(clients, &c)
	}

	return clients, nil
}

// ListByTenant retrieves all clients for a tenant
func (r *ClientRepository) ListByTenant(ctx context.Context, tenantID string) ([]*client.Client, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT 
			id, client_id, tenant_id, client_secret_hash, client_name, client_uri, logo_uri,
			redirect_uris, allowed_scopes, grant_types, response_types,
			token_endpoint_auth_method, access_token_lifetime, refresh_token_lifetime, id_token_lifetime,
			owner_id, is_trusted, is_active, created_at, updated_at, deleted_at
		FROM oauth2_clients
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID)

	if err != nil {
		fmt.Printf("DEBUG: ListByTenant failed for tenant %s: %v\n", tenantID, err)
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []*client.Client
	for rows.Next() {
		var c client.Client
		var redirectURIsJSON, allowedScopesJSON, grantTypesJSON, responseTypesJSON []byte
		var ownerID sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&c.ID, &c.ClientID, &c.TenantID, &c.ClientSecretHash, &c.ClientName, &c.ClientURI, &c.LogoURI,
			&redirectURIsJSON, &allowedScopesJSON, &grantTypesJSON, &responseTypesJSON,
			&c.TokenEndpointAuthMethod, &c.AccessTokenLifetime, &c.RefreshTokenLifetime, &c.IDTokenLifetime,
			&ownerID, &c.IsTrusted, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}

		if err := json.Unmarshal(redirectURIsJSON, &c.RedirectURIs); err != nil {
			continue
		}
		if err := json.Unmarshal(allowedScopesJSON, &c.AllowedScopes); err != nil {
			continue
		}
		if err := json.Unmarshal(grantTypesJSON, &c.GrantTypes); err != nil {
			continue
		}
		if err := json.Unmarshal(responseTypesJSON, &c.ResponseTypes); err != nil {
			continue
		}

		if ownerID.Valid {
			c.OwnerID = ownerID.String
		}
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}

		clients = append(clients, &c)
	}

	return clients, nil
}

// DeleteByTenantID soft-deletes all clients belonging to a tenant
func (r *ClientRepository) DeleteByTenantID(ctx context.Context, tenantID string) error {
	_, err := r.db.pool.Exec(ctx, `
		UPDATE oauth2_clients SET deleted_at = NOW()
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete clients by tenant: %w", err)
	}
	return nil
}
