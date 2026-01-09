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
	"log/slog"

	"github.com/opentrusty/opentrusty-core/project"
	"github.com/opentrusty/opentrusty-core/role"
)

// UserRoleAssignment represents a role assigned to a user with scope.
//
// Purpose: Flattened representation of a user's role and its context.
// Domain: Authz
type UserRoleAssignment struct {
	RoleID   string  `json:"role_id"`
	RoleName string  `json:"role_name"`
	Scope    string  `json:"scope"`
	Context  *string `json:"context,omitempty"`
}

// ProjectInfo represents simplified project information for external systems
type ProjectInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UserInfoClaims represents the claims to be returned in the userinfo endpoint.
//
// Purpose: Structured results for OIDC UserInfo response.
// Domain: Authz
type UserInfoClaims struct {
	Roles    []string       `json:"roles"`
	Projects []*ProjectInfo `json:"projects"`
}

// Service provides authorization business logic.
//
// Purpose: Centralized engine for permission checks and role resolution.
// Domain: Authz
type Service struct {
	projectRepo    project.ProjectRepository
	roleRepo       role.RoleRepository
	assignmentRepo role.AssignmentRepository
}

// NewService creates a new authorization service.
//
// Purpose: Constructor for the authorization engine.
// Domain: Authz
// Audited: No
// Errors: None
func NewService(
	projectRepo project.ProjectRepository,
	roleRepo role.RoleRepository,
	assignmentRepo role.AssignmentRepository,
) *Service {
	return &Service{
		projectRepo:    projectRepo,
		roleRepo:       roleRepo,
		assignmentRepo: assignmentRepo,
	}
}

// GetUserRoles retrieves all unique role names for a user across all scopes.
//
// Purpose: Aggregation of platform and tenant roles for token issuance.
// Domain: Authz
// Audited: No
// Errors: System errors
func (s *Service) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	assignments, err := s.assignmentRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	roleMap := make(map[string]bool)
	for _, a := range assignments {
		r, err := s.roleRepo.GetByID(ctx, a.RoleID)
		if err != nil {
			continue
		}
		roleMap[r.Name] = true
	}

	roleNames := make([]string, 0, len(roleMap))
	for name := range roleMap {
		roleNames = append(roleNames, name)
	}

	return roleNames, nil
}

// GetUserRoleAssignments retrieves all role assignments for a user with details
func (s *Service) GetUserRoleAssignments(ctx context.Context, userID string) ([]UserRoleAssignment, error) {
	assignments, err := s.assignmentRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user assignments: %w", err)
	}

	var result []UserRoleAssignment
	for _, a := range assignments {
		r, err := s.roleRepo.GetByID(ctx, a.RoleID)
		if err != nil {
			result = append(result, UserRoleAssignment{
				RoleID:   a.RoleID,
				RoleName: "unknown",
				Scope:    string(a.Scope),
				Context:  a.ScopeContextID,
			})
			continue
		}
		result = append(result, UserRoleAssignment{
			RoleID:   a.RoleID,
			RoleName: r.Name,
			Scope:    string(a.Scope),
			Context:  a.ScopeContextID,
		})
	}

	return result, nil
}

// GetUserProjects retrieves all projects a user has access to
func (s *Service) GetUserProjects(ctx context.Context, userID string) ([]*project.Project, error) {
	return s.projectRepo.ListByUser(ctx, userID)
}

// BuildUserInfoClaims builds the authorization claims for a user
func (s *Service) BuildUserInfoClaims(ctx context.Context, userID string) (*UserInfoClaims, error) {
	roles, err := s.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	projects, err := s.GetUserProjects(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}

	projectInfos := make([]*ProjectInfo, 0, len(projects))
	for _, p := range projects {
		projectInfos = append(projectInfos, &ProjectInfo{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		})
	}

	return &UserInfoClaims{
		Roles:    roles,
		Projects: projectInfos,
	}, nil
}

// HasPermission checks if a user has a specific permission at a scope.
//
// Purpose: Core authorization check enforcing RBAC rules.
// Domain: Authz
// Security: Enforces scope context matching and platform administrator overrides.
// Audited: No
// Errors: System errors
func (s *Service) HasPermission(ctx context.Context, userID string, scope role.Scope, scopeContextID *string, permission string) (bool, error) {
	assignments, err := s.assignmentRepo.ListForUser(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "HasPermission: failed to get user assignments", "error", err)
		return false, fmt.Errorf("failed to get user assignments: %w", err)
	}

	for _, a := range assignments {
		matchesScope := false

		// Platform administrators have global authority across all scopes.
		if a.Scope == role.ScopePlatform {
			matchesScope = true
		} else if a.Scope == scope {
			// For context-specific scopes (tenant, client), the context IDs must match exactly.
			if scopeContextID != nil && a.ScopeContextID != nil && *a.ScopeContextID == *scopeContextID {
				matchesScope = true
			}
		}

		if !matchesScope {
			continue
		}

		r, err := s.roleRepo.GetByID(ctx, a.RoleID)
		if err != nil {
			slog.WarnContext(ctx, "HasPermission: failed to get role", "role_id", a.RoleID, "error", err)
			continue
		}

		if r.HasPermission(permission) {
			return true, nil
		} else {
			slog.InfoContext(ctx, "HasPermission: role does not have permission", "role", r.Name, "perm", permission)
		}
	}

	scID := ""
	if scopeContextID != nil {
		scID = *scopeContextID
	}
	slog.WarnContext(ctx, "HasPermission: DENIED", "user", userID, "scope", scope, "scopeID", scID, "perm", permission, "assignments_count", len(assignments))
	return false, nil
}

// HasPermissionAny checks if a user has a specific permission in ANY of their assigned scopes
func (s *Service) HasPermissionAny(ctx context.Context, userID string, permission string) (bool, error) {
	assignments, err := s.assignmentRepo.ListForUser(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user assignments: %w", err)
	}

	for _, a := range assignments {
		r, err := s.roleRepo.GetByID(ctx, a.RoleID)
		if err != nil {
			continue
		}

		if r.HasPermission(permission) {
			return true, nil
		}
	}

	return false, nil
}
