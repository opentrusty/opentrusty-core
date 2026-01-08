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
	"time"

	"github.com/opentrusty/opentrusty-core/tenant"
)

// MembershipRepository implements tenant.MembershipRepository
type MembershipRepository struct {
	db *DB
}

// NewMembershipRepository creates a new membership repository
func NewMembershipRepository(db *DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// AddMember inserts a new membership record
func (r *MembershipRepository) AddMember(ctx context.Context, m *tenant.Membership) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	_, err := r.db.pool.Exec(ctx, `
		INSERT INTO tenant_members (id, tenant_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, user_id) DO NOTHING
	`, m.ID, m.TenantID, m.UserID, m.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

// RemoveMember removes a specific membership record
func (r *MembershipRepository) RemoveMember(ctx context.Context, tenantID, userID string) error {
	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM tenant_members
		WHERE tenant_id = $1 AND user_id = $2
	`, tenantID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

// ListMembers retrieves all memberships for a tenant
func (r *MembershipRepository) ListMembers(ctx context.Context, tenantID string) ([]*tenant.Membership, error) {
	rows, err := r.db.pool.Query(ctx, `
		SELECT id, tenant_id, user_id, created_at
		FROM tenant_members
		WHERE tenant_id = $1
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	var result []*tenant.Membership
	for rows.Next() {
		m := &tenant.Membership{}
		if err := rows.Scan(&m.ID, &m.TenantID, &m.UserID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}
		result = append(result, m)
	}
	return result, nil
}

// CheckMembership checks if a user is a member of a tenant
func (r *MembershipRepository) CheckMembership(ctx context.Context, tenantID, userID string) (bool, error) {
	var exists bool
	err := r.db.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tenant_members
			WHERE tenant_id = $1 AND user_id = $2
		)
	`, tenantID, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	return exists, nil
}

// DeleteByTenantID removes all memberships for a tenant
func (r *MembershipRepository) DeleteByTenantID(ctx context.Context, tenantID string) error {
	_, err := r.db.pool.Exec(ctx, `
		DELETE FROM tenant_members
		WHERE tenant_id = $1
	`, tenantID)

	if err != nil {
		return fmt.Errorf("failed to delete memberships by tenant: %w", err)
	}
	return nil
}
