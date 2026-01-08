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

package authz

import (
	"context"
	"fmt"
	"testing"

	"github.com/opentrusty/opentrusty-core/project"
	"github.com/opentrusty/opentrusty-core/role"
)

// Mock repos
type mockProjectRepo struct {
	project.ProjectRepository
}

func (m *mockProjectRepo) ListByUser(ctx context.Context, userID string) ([]*project.Project, error) {
	return []*project.Project{{ID: "p1", Name: "Project 1"}}, nil
}

type mockRoleRepo struct {
	role.RoleRepository
	roles map[string]*role.Role
}

func (m *mockRoleRepo) GetByID(ctx context.Context, id string) (*role.Role, error) {
	r, ok := m.roles[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return r, nil
}

type mockAssignmentRepo struct {
	role.AssignmentRepository
	assignments []*role.Assignment
}

func (m *mockAssignmentRepo) ListForUser(ctx context.Context, userID string) ([]*role.Assignment, error) {
	var res []*role.Assignment
	for _, a := range m.assignments {
		if a.UserID == userID {
			res = append(res, a)
		}
	}
	return res, nil
}

func TestHasPermission(t *testing.T) {
	adminRole := &role.Role{
		ID:          "role-admin",
		Name:        "admin",
		Scope:       role.ScopePlatform,
		Permissions: []string{"*"},
	}
	tenantRole := &role.Role{
		ID:          "role-tenant",
		Name:        "editor",
		Scope:       role.ScopeTenant,
		Permissions: []string{"edit:stuff"},
	}

	roleRepo := &mockRoleRepo{
		roles: map[string]*role.Role{
			adminRole.ID:  adminRole,
			tenantRole.ID: tenantRole,
		},
	}

	assignmentRepo := &mockAssignmentRepo{
		assignments: []*role.Assignment{
			{UserID: "user-admin", RoleID: adminRole.ID, Scope: role.ScopePlatform},
			{UserID: "user-tenant", RoleID: tenantRole.ID, Scope: role.ScopeTenant, ScopeContextID: stringPtr("t1")},
		},
	}

	svc := NewService(&mockProjectRepo{}, roleRepo, assignmentRepo)

	tests := []struct {
		name       string
		userID     string
		scope      role.Scope
		contextID  *string
		permission string
		want       bool
	}{
		{
			name:       "platform admin has any permission",
			userID:     "user-admin",
			scope:      role.ScopePlatform,
			permission: "any:action",
			want:       true,
		},
		{
			name:       "platform admin has tenant permission (override)",
			userID:     "user-admin",
			scope:      role.ScopeTenant,
			contextID:  stringPtr("any-tenant"),
			permission: "tenant:action",
			want:       true,
		},
		{
			name:       "tenant editor has specific permission in context",
			userID:     "user-tenant",
			scope:      role.ScopeTenant,
			contextID:  stringPtr("t1"),
			permission: "edit:stuff",
			want:       true,
		},
		{
			name:       "tenant editor lacks permission in wrong context",
			userID:     "user-tenant",
			scope:      role.ScopeTenant,
			contextID:  stringPtr("t2"),
			permission: "edit:stuff",
			want:       false,
		},
		{
			name:       "tenant editor lacks wrong permission",
			userID:     "user-tenant",
			scope:      role.ScopeTenant,
			contextID:  stringPtr("t1"),
			permission: "delete:stuff",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.HasPermission(context.Background(), tt.userID, tt.scope, tt.contextID, tt.permission)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
