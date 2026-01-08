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
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opentrusty/opentrusty-core/policy"
	"github.com/opentrusty/opentrusty-core/role"
)

// RoleRepository implements role.RoleRepository and policy.RoleRepository
type RoleRepository struct {
	db *DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, ro *role.Role) error {
	tx, err := r.db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO rbac_roles (
			id, name, scope, description, created_at, updated_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, ro.ID, ro.Name, string(ro.Scope), ro.Description)
	if err != nil {
		return fmt.Errorf("failed to insert role: %w", err)
	}

	// Insert permissions
	for _, p := range ro.Permissions {
		var permID string
		err = tx.QueryRow(ctx, "SELECT id FROM rbac_permissions WHERE name = $1", p).Scan(&permID)
		if err != nil {
			if err == pgx.ErrNoRows {
				// Create permission if it doesn't exist?
				// For now, let's assume permissions are seeded or handled elsewhere.
				// Or we can just insert it.
				continue
			}
			return fmt.Errorf("failed to get permission ID for %s: %w", p, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO rbac_role_permissions (role_id, permission_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, ro.ID, permID)
		if err != nil {
			return fmt.Errorf("failed to insert role permission mapping: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id string) (*role.Role, error) {
	var ro role.Role
	var scopeStr string

	err := r.db.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.scope, COALESCE(r.description, ''),
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
		WHERE r.id = $1
		GROUP BY r.id, r.name, r.scope, r.description
	`, id).Scan(
		&ro.ID, &ro.Name, &scopeStr, &ro.Description, &ro.Permissions,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, policy.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	ro.Scope = role.Scope(scopeStr)
	return &ro, nil
}

// GetByName retrieves a role by name and scope
func (r *RoleRepository) GetByName(ctx context.Context, name string, scope role.Scope) (*role.Role, error) {
	var ro role.Role
	var scopeStr string

	err := r.db.pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.scope, COALESCE(r.description, ''),
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
		WHERE r.name = $1 AND r.scope = $2
		GROUP BY r.id, r.name, r.scope, r.description
	`, name, string(scope)).Scan(
		&ro.ID, &ro.Name, &scopeStr, &ro.Description, &ro.Permissions,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, policy.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	ro.Scope = role.Scope(scopeStr)
	return &ro, nil
}

// List retrieves all roles, optionally filtered by scope
func (r *RoleRepository) List(ctx context.Context, scope *role.Scope) ([]*role.Role, error) {
	query := `
		SELECT r.id, r.name, r.scope, COALESCE(r.description, ''),
		       COALESCE(array_agg(p.name) FILTER (WHERE p.name IS NOT NULL), '{}')
		FROM rbac_roles r
		LEFT JOIN rbac_role_permissions rp ON r.id = rp.role_id
		LEFT JOIN rbac_permissions p ON rp.permission_id = p.id
	`
	var args []interface{}
	if scope != nil {
		query += " WHERE r.scope = $1"
		args = append(args, string(*scope))
	}
	query += " GROUP BY r.id, r.name, r.scope, r.description ORDER BY r.name ASC"

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []*role.Role
	for rows.Next() {
		var ro role.Role
		var scopeStr string
		if err := rows.Scan(&ro.ID, &ro.Name, &scopeStr, &ro.Description, &ro.Permissions); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		ro.Scope = role.Scope(scopeStr)
		roles = append(roles, &ro)
	}

	return roles, nil
}

// Update updates role information
func (r *RoleRepository) Update(ctx context.Context, ro *role.Role) error {
	result, err := r.db.pool.Exec(ctx, `
		UPDATE rbac_roles SET description = $2, updated_at = NOW()
		WHERE id = $1
	`, ro.ID, ro.Description)

	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return policy.ErrRoleNotFound
	}

	return nil
}

// Delete deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.pool.Exec(ctx, `DELETE FROM rbac_roles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	if result.RowsAffected() == 0 {
		return policy.ErrRoleNotFound
	}
	return nil
}

// Support for policy.RoleRepository if needed can be added here or via a wrapper.
