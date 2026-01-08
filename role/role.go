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

package role

import (
	"context"
	"time"

	"github.com/opentrusty/opentrusty-core/policy"
)

// -----------------------------------------------------------------------------
// Role Name Constants
// These are the canonical names for roles stored in the database.
// -----------------------------------------------------------------------------

const (
	// RolePlatformAdmin is the platform-wide administrator role.
	// Scope: Platform
	// Permissions: * (wildcard - all permissions)
	RolePlatformAdmin = "platform_admin"

	// RoleTenantOwner is the tenant owner role with full tenant control.
	// Scope: Tenant
	RoleTenantOwner = "tenant_owner"

	// RoleTenantAdmin is the tenant administrator role.
	// Scope: Tenant
	RoleTenantAdmin = "tenant_admin"

	// RoleTenantMember is a basic tenant membership role.
	// Scope: Tenant
	RoleTenantMember = "tenant_member"
)

// RoleID constants (Seeded UUIDs from initial migration)
const (
	RoleIDPlatformAdmin = "00000000-0000-0000-0000-000000000001"
	RoleIDTenantOwner   = "00000000-0000-0000-0000-000000000002"
	RoleIDTenantAdmin   = "00000000-0000-0000-0000-000000000003"
	RoleIDMember        = "00000000-0000-0000-0000-000000000004"
)

// -----------------------------------------------------------------------------
// Actor Type Constants
// These identify the type of actor making a request.
// -----------------------------------------------------------------------------

type ActorType string

const (
	// ActorUser represents a human user.
	ActorUser ActorType = "user"

	// ActorClient represents an OAuth2 client acting on behalf of a user.
	ActorClient ActorType = "client"

	// ActorSystem represents internal system operations (e.g., bootstrap, scheduled jobs).
	ActorSystem ActorType = "system"
)

// -----------------------------------------------------------------------------
// Role Permission Mappings
// These define the default permissions for each role.
// Used for seeding and validation.
// -----------------------------------------------------------------------------

// PlatformAdminPermissions defines permissions for the platform_admin role.
var PlatformAdminPermissions = []string{
	"*", // Wildcard: all permissions
}

// TenantOwnerPermissions defines permissions for the tenant_owner role.
var TenantOwnerPermissions = []string{
	policy.PermTenantManageUsers,
	policy.PermTenantManageClients,
	policy.PermTenantManageSettings,
	policy.PermTenantViewUsers,
	policy.PermTenantView,
	policy.PermTenantViewAudit,
	policy.PermUserReadProfile,
	policy.PermUserWriteProfile,
	policy.PermUserChangePassword,
	policy.PermUserManageSessions,
}

// TenantAdminPermissions defines permissions for the tenant_admin role.
var TenantAdminPermissions = []string{
	policy.PermTenantManageUsers,
	policy.PermTenantManageClients,
	policy.PermTenantViewUsers,
	policy.PermTenantView,
	policy.PermTenantViewAudit,
	policy.PermUserReadProfile,
	policy.PermUserWriteProfile,
	policy.PermUserChangePassword,
	policy.PermUserManageSessions,
}

// TenantMemberPermissions defines permissions for the tenant_member role.
var TenantMemberPermissions = []string{
	policy.PermTenantView,
	policy.PermUserReadProfile,
	policy.PermUserWriteProfile,
	policy.PermUserChangePassword,
}

// -----------------------------------------------------------------------------
// Models & Interfaces
// -----------------------------------------------------------------------------

// Scope defines the level at which a role is assigned.
//
// Purpose: Classification of authorization boundaries.
// Domain: Authz
type Scope string

const (
	ScopePlatform Scope = "platform"
	ScopeTenant   Scope = "tenant"
	ScopeClient   Scope = "client"
)

// Role represents a scoped role with associated permission names.
//
// Purpose: Container for a set of permissions with a defined scope.
// Domain: Authz
// Invariants: Name must be unique within scope. Scope must be valid.
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Scope       Scope    `json:"scope"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p == "*" || p == permission {
			return true
		}
	}
	return false
}

// Assignment represents a role granted to a user at a specific scope.
//
// Purpose: Association between an identity and a role context.
// Domain: Authz
// Invariants: UserID and RoleID must exist. ScopeContextID mandatory for tenant/client scopes.
type Assignment struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	RoleID         string    `json:"role_id"`
	Scope          Scope     `json:"scope"`
	ScopeContextID *string   `json:"scope_context_id,omitempty"` // NULL for platform, tenant_id for tenant, etc.
	GrantedAt      time.Time `json:"granted_at"`
	GrantedBy      string    `json:"granted_by"`
}

// RoleRepository defines the interface for role persistence.
//
// Purpose: Abstraction for managing role definition storage.
// Domain: Authz
type RoleRepository interface {
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByName(ctx context.Context, name string, scope Scope) (*Role, error)
	List(ctx context.Context, scope *Scope) ([]*Role, error)
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id string) error
}

// AssignmentRepository defines the interface for RBAC assignments.
//
// Purpose: Abstraction for managing user role associations.
// Domain: Authz
type AssignmentRepository interface {
	ListForUser(ctx context.Context, userID string) ([]*Assignment, error)
	Grant(ctx context.Context, assignment *Assignment) error
	Revoke(ctx context.Context, userID, roleID string, scope Scope, scopeContextID *string) error
	ListByRole(ctx context.Context, roleID string, scope Scope, scopeContextID *string) ([]string, error)
	CheckExists(ctx context.Context, roleID string, scope Scope, scopeContextID *string) (bool, error)
	DeleteByContextID(ctx context.Context, scope Scope, contextID string) error
}
