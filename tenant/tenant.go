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

package tenant

import (
	"context"
	"errors"
	"time"
)

// Domain errors
var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrInvalidTenantName   = errors.New("invalid tenant name")
)

// TenantUserRole represents a user's role assignment in a tenant
type TenantUserRole struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	UserID     string    `json:"user_id"`
	Role       string    `json:"role"`
	EmailPlain string    `json:"email_plain"`
	FullName   string    `json:"full_name"`
	Nickname   *string   `json:"nickname"`
	Picture    *string   `json:"picture"`
	GrantedAt  time.Time `json:"granted_at"`
	GrantedBy  string    `json:"granted_by"`
}

// Tenant represents an isolated environment or customer account.
//
// Purpose: Root container for data isolation in multi-tenant architecture.
// Domain: Tenant
// Invariants: ID must be unique. Status must be Active or Inactive.
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultTenantID is the ID of the default tenant
const DefaultTenantID = "default"

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type TenantMetrics struct {
	TotalUsers    int `json:"total_users"`
	TotalClients  int `json:"total_clients"`
	AuditCount24h int `json:"audit_count_24h"`
}

// Membership represents a user's membership in a tenant.
//
// Purpose: Linkage between a global identity and a specific tenant.
// Domain: Tenant
type Membership struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository defines the interface for tenant persistence.
//
// Purpose: Abstraction for managing tenant lifecycle storage.
// Domain: Tenant
type Repository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	GetByName(ctx context.Context, name string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
}

// RoleRepository defines the interface for tenant role persistence.
//
// Purpose: Management of role assignments within a tenant.
// Domain: Authz
type RoleRepository interface {
	AssignRole(ctx context.Context, tenantID, userID, roleName, grantedBy string) error
	RevokeRole(ctx context.Context, tenantID, userID, roleName string) error
	GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error)
	GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error)
	DeleteByTenantID(ctx context.Context, tenantID string) error
}

// MembershipRepository defines the interface for tenant membership persistence.
//
// Purpose: Management of tenant membership lifecycle.
// Domain: Tenant
type MembershipRepository interface {
	AddMember(ctx context.Context, membership *Membership) error
	RemoveMember(ctx context.Context, tenantID, userID string) error
	ListMembers(ctx context.Context, tenantID string) ([]*Membership, error)
	CheckMembership(ctx context.Context, tenantID, userID string) (bool, error)
	DeleteByTenantID(ctx context.Context, tenantID string) error
}
