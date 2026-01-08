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

	"github.com/opentrusty/opentrusty-core/id"
	"github.com/opentrusty/opentrusty-core/role"
	"github.com/opentrusty/opentrusty-core/tenant"
)

// TenantRoleRepository implements tenant.RoleRepository
type TenantRoleRepository struct {
	db *DB
}

// NewTenantRoleRepository creates a new tenant role repository
func NewTenantRoleRepository(db *DB) *TenantRoleRepository {
	return &TenantRoleRepository{db: db}
}

// MapTenantRole maps internal tenant role names to seeded RBAC role IDs
func MapTenantRole(roleName string) string {
	switch roleName {
	case role.RoleTenantOwner:
		return role.RoleIDTenantOwner
	case role.RoleTenantAdmin:
		return role.RoleIDTenantAdmin
	case role.RoleTenantMember:
		return role.RoleIDMember
	default:
		return role.RoleIDMember
	}
}

// AssignRole assigns a role to a user in a tenant
func (r *TenantRoleRepository) AssignRole(ctx context.Context, tenantID, userID, roleName, grantedBy string) error {
	roleID := MapTenantRole(roleName)
	assignmentID := id.NewUUIDv7()

	var grantedByUUID sql.NullString
	if grantedBy != "" {
		grantedByUUID = sql.NullString{String: grantedBy, Valid: true}
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO rbac_assignments (id, user_id, role_id, scope, scope_context_id, granted_at, granted_by)
		VALUES ($1, $2, $3, 'tenant', $4, NOW(), $5)
		ON CONFLICT (user_id, role_id, scope, scope_context_id) DO NOTHING
	`, assignmentID, userID, roleID, tenantID, grantedByUUID)

	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// RevokeRole revokes a role from a user in a tenant
func (r *TenantRoleRepository) RevokeRole(ctx context.Context, tenantID, userID, roleName string) error {
	roleID := MapTenantRole(roleName)
	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM rbac_assignments
		WHERE user_id = $1 AND role_id = $2 AND scope = 'tenant' AND scope_context_id = $3
	`, userID, roleID, tenantID)

	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	return nil
}

// GetUserRoles retrieves all roles a user has in a tenant
func (r *TenantRoleRepository) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*tenant.TenantUserRole, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT a.id, a.scope_context_id, a.user_id, r.name, u.email_plain, u.full_name, u.nickname, u.picture, a.granted_at, a.granted_by
		FROM rbac_assignments a
		JOIN rbac_roles r ON a.role_id = r.id
		JOIN users u ON a.user_id = u.id
		WHERE a.user_id = $1 AND a.scope = 'tenant' AND a.scope_context_id = $2
	`, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()

	var roles []*tenant.TenantUserRole
	for rows.Next() {
		var role tenant.TenantUserRole
		var grantedBy sql.NullString
		var nickname, picture sql.NullString
		if err := rows.Scan(&role.ID, &role.TenantID, &role.UserID, &role.Role, &role.EmailPlain, &role.FullName, &nickname, &picture, &role.GrantedAt, &grantedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		if nickname.Valid {
			role.Nickname = &nickname.String
		}
		if picture.Valid {
			role.Picture = &picture.String
		}
		if grantedBy.Valid {
			role.GrantedBy = grantedBy.String
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// GetTenantUsers retrieves all users with roles in a tenant
func (r *TenantRoleRepository) GetTenantUsers(ctx context.Context, tenantID string) ([]*tenant.TenantUserRole, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT a.id, a.scope_context_id, a.user_id, r.name, u.email_plain, u.full_name, u.nickname, u.picture, a.granted_at, a.granted_by
		FROM rbac_assignments a
		JOIN rbac_roles r ON a.role_id = r.id
		JOIN users u ON a.user_id = u.id
		WHERE a.scope = 'tenant' AND a.scope_context_id = $1
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant users: %w", err)
	}
	defer rows.Close()

	var roles []*tenant.TenantUserRole
	for rows.Next() {
		var role tenant.TenantUserRole
		var grantedBy sql.NullString
		var nickname, picture sql.NullString
		if err := rows.Scan(&role.ID, &role.TenantID, &role.UserID, &role.Role, &role.EmailPlain, &role.FullName, &nickname, &picture, &role.GrantedAt, &grantedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		if nickname.Valid {
			role.Nickname = &nickname.String
		}
		if picture.Valid {
			role.Picture = &picture.String
		}
		if grantedBy.Valid {
			role.GrantedBy = grantedBy.String
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

// DeleteByTenantID removes all role assignments for a specific tenant
func (r *TenantRoleRepository) DeleteByTenantID(ctx context.Context, tenantID string) error {
	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM rbac_assignments
		WHERE scope = 'tenant' AND scope_context_id = $1
	`, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete tenant roles: %w", err)
	}

	return nil
}
