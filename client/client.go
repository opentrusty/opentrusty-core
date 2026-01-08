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

package client

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Domain errors (Internal)
var (
	ErrClientNotFound           = errors.New("client not found")
	ErrClientAlreadyExists      = errors.New("client already exists")
	ErrDomainInvalidRedirectURI = errors.New("invalid redirect URI")
	ErrDomainInvalidScope       = errors.New("invalid scope")
	ErrDomainInvalidGrantType   = errors.New("invalid grant type")
	ErrCodeExpired              = errors.New("authorization code expired")
	ErrCodeAlreadyUsed          = errors.New("authorization code already used")
	ErrCodeNotFound             = errors.New("authorization code not found")
	ErrDomainInvalidClient      = errors.New("invalid client credentials")
	ErrTokenExpired             = errors.New("token expired")
	ErrTokenRevoked             = errors.New("token revoked")
	ErrTokenNotFound            = errors.New("token not found")
)

// OIDC Standard Scope Constants
const (
	ScopeOpenID        = "openid"
	ScopeProfile       = "profile"
	ScopeEmail         = "email"
	ScopeAddress       = "address"
	ScopePhone         = "phone"
	ScopeOfflineAccess = "offline_access"
)

// OIDCScopes defines the valid OIDC standard scopes (RFC compliant)
// Scopes control claim RELEASE, not authorization.
var OIDCScopes = map[string]bool{
	ScopeOpenID:        true, // Required for OIDC
	ScopeProfile:       true, // name, given_name, family_name, picture, locale
	ScopeEmail:         true, // email, email_verified
	ScopeAddress:       true, // address claim
	ScopePhone:         true, // phone_number, phone_number_verified
	ScopeOfflineAccess: true, // refresh token permission
}

// ValidateOIDCScopes validates that scopes are OIDC-compliant.
// Rules:
// - Scope list cannot be empty
// - 'openid' scope is mandatory
// - All scopes must be in OIDCScopes
func ValidateOIDCScopes(scopes []string) error {
	if len(scopes) == 0 {
		return fmt.Errorf("%w: scope list cannot be empty", ErrDomainInvalidScope)
	}

	hasOpenID := false
	for _, s := range scopes {
		if s == ScopeOpenID {
			hasOpenID = true
		}
		if !OIDCScopes[s] {
			return fmt.Errorf("%w: unknown scope '%s'", ErrDomainInvalidScope, s)
		}
	}

	if !hasOpenID {
		return fmt.Errorf("%w: 'openid' scope is required", ErrDomainInvalidScope)
	}

	return nil
}

// Client represents an OAuth2 client application.
//
// Purpose: Entity representing a third-party application or service using OIDC/OAuth2.
// Domain: OAuth2
// Invariants: ClientID must be unique. RedirectURIs must be valid.
type Client struct {
	ID                      string     `json:"id"`
	ClientID                string     `json:"client_id"`
	TenantID                string     `json:"tenant_id"`
	ClientSecretHash        string     `json:"-"`
	ClientName              string     `json:"client_name"`
	ClientURI               string     `json:"client_uri,omitempty"`
	LogoURI                 string     `json:"logo_uri,omitempty"`
	RedirectURIs            []string   `json:"redirect_uris"`
	AllowedScopes           []string   `json:"allowed_scopes"`
	GrantTypes              []string   `json:"grant_types"`
	ResponseTypes           []string   `json:"response_types"`
	TokenEndpointAuthMethod string     `json:"token_endpoint_auth_method"`
	AccessTokenLifetime     int        `json:"access_token_lifetime"`
	RefreshTokenLifetime    int        `json:"refresh_token_lifetime"`
	IDTokenLifetime         int        `json:"id_token_lifetime"`
	OwnerID                 string     `json:"owner_id,omitempty"`
	IsTrusted               bool       `json:"is_trusted"`
	IsActive                bool       `json:"is_active"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty"`
}

// ValidateRedirectURI checks if the redirect URI is allowed for this client
func (c *Client) ValidateRedirectURI(redirectURI string) bool {
	for _, uri := range c.RedirectURIs {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

// ValidateScope checks if the requested scope is allowed for this client
func (c *Client) ValidateScope(requestedScope string) bool {
	if requestedScope == "" {
		return true
	}

	// Split space-separated scopes
	requestedScopes := strings.Fields(requestedScope)

	// Check if all requested scopes are allowed
	for _, reqScope := range requestedScopes {
		allowed := false
		for _, allowedScope := range c.AllowedScopes {
			if allowedScope == reqScope || allowedScope == "*" {
				allowed = true
				break
			}
		}
		if !allowed {
			return false
		}
	}

	return true
}

// AuthorizationCode represents a short-lived authorization code.
//
// Purpose: One-time use token for exchanging with an access token.
// Domain: OAuth2
// Invariants: Code must be a cryptographically secure token. Must expire within 10 minutes.
type AuthorizationCode struct {
	ID                  string
	Code                string
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	State               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
	UsedAt              *time.Time
	IsUsed              bool
	CreatedAt           time.Time
}

// IsExpired checks if the authorization code has expired
func (a *AuthorizationCode) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// AccessToken represents an OAuth2 access token.
//
// Purpose: Credential for accessing protected resources.
// Domain: OAuth2
// Invariants: TokenHash must be unique. Lifetime is configurable.
type AccessToken struct {
	ID        string
	TenantID  string
	TokenHash string
	ClientID  string
	UserID    string
	Scope     string
	TokenType string
	ExpiresAt time.Time
	RevokedAt *time.Time
	IsRevoked bool
	CreatedAt time.Time
}

// IsExpired checks if the access token has expired
func (a *AccessToken) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

// RefreshToken represents an OAuth2 refresh token.
//
// Purpose: Long-lived credential to obtain new access tokens.
// Domain: OAuth2
// Invariants: Associated with a specific client and user.
type RefreshToken struct {
	ID            string
	TenantID      string
	TokenHash     string
	AccessTokenID string
	ClientID      string
	UserID        string
	Scope         string
	ExpiresAt     time.Time
	RevokedAt     *time.Time
	IsRevoked     bool
	CreatedAt     time.Time
}

// IsExpired checks if the refresh token has expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// ClientRepository defines the interface for OAuth2 client persistence.
//
// Purpose: Abstraction for managing persistent storage of client metadata.
// Domain: OAuth2
type ClientRepository interface {
	// Create creates a new OAuth2 client
	Create(ctx context.Context, client *Client) error

	// GetByClientID retrieves a client by tenant_id and client_id
	GetByClientID(ctx context.Context, tenantID string, clientID string) (*Client, error)

	// GetByID retrieves a client by tenant_id and internal ID
	GetByID(ctx context.Context, tenantID string, id string) (*Client, error)

	// Update updates client information
	Update(ctx context.Context, client *Client) error

	// Delete soft-deletes a client by tenant_id and internal ID
	Delete(ctx context.Context, tenantID string, id string) error

	// ListByOwner retrieves all clients for an owner
	ListByOwner(ctx context.Context, ownerID string) ([]*Client, error)

	// ListByTenant retrieves all clients for a tenant
	ListByTenant(ctx context.Context, tenantID string) ([]*Client, error)

	// DeleteByTenantID soft-deletes all clients belonging to a tenant
	DeleteByTenantID(ctx context.Context, tenantID string) error
}

// AuthorizationCodeRepository defines the interface for authorization code persistence.
//
// Purpose: Abstraction for managing short-lived authorization codes.
// Domain: OAuth2
type AuthorizationCodeRepository interface {
	// Create creates a new authorization code
	Create(code *AuthorizationCode) error

	// GetByCode retrieves an authorization code
	GetByCode(code string) (*AuthorizationCode, error)

	// MarkAsUsed marks the code as used
	MarkAsUsed(code string) error

	// Delete deletes an authorization code
	Delete(code string) error

	// DeleteExpired deletes all expired authorization codes
	DeleteExpired() error
}

// AccessTokenRepository defines the interface for access token persistence
type AccessTokenRepository interface {
	// Create creates a new access token
	Create(token *AccessToken) error

	// GetByTokenHash retrieves an access token
	GetByTokenHash(tokenHash string) (*AccessToken, error)

	// Revoke revokes an access token
	Revoke(tokenHash string) error

	// DeleteExpired deletes all expired access tokens
	DeleteExpired() error
}

// RefreshTokenRepository defines the interface for refresh token persistence
type RefreshTokenRepository interface {
	// Create creates a new refresh token
	Create(token *RefreshToken) error

	// GetByTokenHash retrieves a refresh token
	GetByTokenHash(tokenHash string) (*RefreshToken, error)

	// Revoke revokes a refresh token
	Revoke(tokenHash string) error

	// DeleteExpired deletes all expired refresh tokens
	DeleteExpired() error
}
