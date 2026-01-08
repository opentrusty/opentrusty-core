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
	"fmt"

	"github.com/opentrusty/opentrusty-core/policy"
	"github.com/opentrusty/opentrusty-core/role"
)

// AssignmentRepository implements role.AssignmentRepository
type AssignmentRepository struct {
	db *DB
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(db *DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Grant assigns a role to a user
func (r *AssignmentRepository) Grant(ctx context.Context, a *role.Assignment) error {
	var grantedBy interface{} = a.GrantedBy
	if a.GrantedBy == "" {
		grantedBy = nil
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO rbac_assignments (
			id, user_id, role_id, scope, scope_context_id, granted_at, granted_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, role_id, scope, scope_context_id) DO NOTHING
	`, a.ID, a.UserID, a.RoleID, string(a.Scope), a.ScopeContextID, a.GrantedAt, grantedBy)

	if err != nil {
		return fmt.Errorf("failed to grant role: %w", err)
	}
	return nil
}

// Revoke removes a role assignment
func (r *AssignmentRepository) Revoke(ctx context.Context, userID, roleID string, scope role.Scope, scopeContextID *string) error {
	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			DELETE FROM rbac_assignments
			WHERE user_id = $1 AND role_id = $2 AND scope = $3 AND scope_context_id IS NULL
		`
		args = []interface{}{userID, roleID, string(scope)}
	} else {
		query = `
			DELETE FROM rbac_assignments
			WHERE user_id = $1 AND role_id = $2 AND scope = $3 AND scope_context_id = $4
		`
		args = []interface{}{userID, roleID, string(scope), *scopeContextID}
	}

	_, err := r.db.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to revoke role: %w", err)
	}
	return nil
}

// ListForUser retrieves all assignments for a user
func (r *AssignmentRepository) ListForUser(ctx context.Context, userID string) ([]*role.Assignment, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, user_id, role_id, scope, scope_context_id, granted_at, granted_by
		FROM rbac_assignments
		WHERE user_id = $1
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*role.Assignment
	for rows.Next() {
		var a role.Assignment
		var scopeStr string
		var grantedBy *string
		if err := rows.Scan(&a.ID, &a.UserID, &a.RoleID, &scopeStr, &a.ScopeContextID, &a.GrantedAt, &grantedBy); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		if grantedBy != nil {
			a.GrantedBy = *grantedBy
		}
		a.Scope = role.Scope(scopeStr)
		assignments = append(assignments, &a)
	}
	return assignments, nil
}

// ListByRole retrieves all users assigned a specific role at a scope
func (r *AssignmentRepository) ListByRole(ctx context.Context, roleID string, scope role.Scope, scopeContextID *string) ([]string, error) {
	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			SELECT user_id FROM rbac_assignments
			WHERE role_id = $1 AND scope = $2 AND scope_context_id IS NULL
		`
		args = []interface{}{roleID, string(scope)}
	} else {
		query = `
			SELECT user_id FROM rbac_assignments
			WHERE role_id = $1 AND scope = $2 AND scope_context_id = $3
		`
		args = []interface{}{roleID, string(scope), *scopeContextID}
	}

	rows, err := r.db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by role: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// CheckExists checks if a specific assignment exists
func (r *AssignmentRepository) CheckExists(ctx context.Context, roleID string, scope role.Scope, scopeContextID *string) (bool, error) {
	var query string
	var args []interface{}

	if scopeContextID == nil {
		query = `
			SELECT EXISTS (
				SELECT 1 FROM rbac_assignments
				WHERE role_id = $1 AND scope = $2 AND scope_context_id IS NULL
			)
		`
		args = []interface{}{roleID, string(scope)}
	} else {
		query = `
			SELECT EXISTS (
				SELECT 1 FROM rbac_assignments
				WHERE role_id = $1 AND scope = $2 AND scope_context_id = $3
			)
		`
		args = []interface{}{roleID, string(scope), *scopeContextID}
	}

	var exists bool
	err := r.db.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check assignment existence: %w", err)
	}
	return exists, nil
}

// DeleteByContextID removes all assignments for a specific scope and context
func (r *AssignmentRepository) DeleteByContextID(ctx context.Context, scope role.Scope, contextID string) error {
	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM rbac_assignments
		WHERE scope = $1 AND scope_context_id = $2
	`, string(scope), contextID)

	if err != nil {
		return fmt.Errorf("failed to delete assignments by context: %w", err)
	}
	return nil
}

// PolicyAssignmentRepository implements policy.AssignmentRepository
type PolicyAssignmentRepository struct {
	r *AssignmentRepository
}

func NewPolicyAssignmentRepository(db *DB) *PolicyAssignmentRepository {
	return &PolicyAssignmentRepository{r: NewAssignmentRepository(db)}
}

func (pr *PolicyAssignmentRepository) Grant(ctx context.Context, a *policy.Assignment) error {
	return pr.r.Grant(ctx, &role.Assignment{
		ID:             a.ID,
		UserID:         a.UserID,
		RoleID:         a.RoleID,
		Scope:          role.Scope(a.Scope),
		ScopeContextID: a.ScopeContextID,
		GrantedAt:      a.GrantedAt,
		GrantedBy:      a.GrantedBy,
	})
}

func (pr *PolicyAssignmentRepository) Revoke(ctx context.Context, userID, roleID string, scope policy.Scope, scopeContextID *string) error {
	return pr.r.Revoke(ctx, userID, roleID, role.Scope(scope), scopeContextID)
}

func (pr *PolicyAssignmentRepository) ListForUser(ctx context.Context, userID string) ([]*policy.Assignment, error) {
	assignments, err := pr.r.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*policy.Assignment, len(assignments))
	for i, a := range assignments {
		result[i] = &policy.Assignment{
			ID:             a.ID,
			UserID:         a.UserID,
			RoleID:         a.RoleID,
			Scope:          policy.Scope(a.Scope),
			ScopeContextID: a.ScopeContextID,
			GrantedAt:      a.GrantedAt,
			GrantedBy:      a.GrantedBy,
		}
	}
	return result, nil
}

func (pr *PolicyAssignmentRepository) ListByRole(ctx context.Context, roleID string, scope policy.Scope, scopeContextID *string) ([]string, error) {
	return pr.r.ListByRole(ctx, roleID, role.Scope(scope), scopeContextID)
}

func (pr *PolicyAssignmentRepository) CheckExists(ctx context.Context, roleID string, scope policy.Scope, scopeContextID *string) (bool, error) {
	return pr.r.CheckExists(ctx, roleID, role.Scope(scope), scopeContextID)
}

func (pr *PolicyAssignmentRepository) DeleteByContextID(ctx context.Context, scope policy.Scope, contextID string) error {
	return pr.r.DeleteByContextID(ctx, role.Scope(scope), contextID)
}
