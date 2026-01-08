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

package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/opentrusty/opentrusty-core/audit"
	"github.com/opentrusty/opentrusty-core/client"
	"github.com/opentrusty/opentrusty-core/id"
	"github.com/opentrusty/opentrusty-core/policy"
	"github.com/opentrusty/opentrusty-core/role"
	"github.com/opentrusty/opentrusty-core/user"
)

// Service provides tenant management business logic
type Service struct {
	repo            Repository
	roleRepo        RoleRepository
	authzRepo       policy.AssignmentRepository
	identityService *user.Service
	clientRepo      client.ClientRepository
	membershipRepo  MembershipRepository
	auditLogger     audit.Logger
}

// NewService creates a new tenant service
func NewService(
	repo Repository,
	roleRepo RoleRepository,
	authzRepo policy.AssignmentRepository,
	identityService *user.Service,
	clientRepo client.ClientRepository,
	membershipRepo MembershipRepository,
	auditLogger audit.Logger,
) *Service {
	return &Service{
		repo:            repo,
		roleRepo:        roleRepo,
		authzRepo:       authzRepo,
		identityService: identityService,
		clientRepo:      clientRepo,
		membershipRepo:  membershipRepo,
		auditLogger:     auditLogger,
	}
}

// CreateTenant creates a new tenant and provisions an initial tenant_owner.
// If ownerPassword is empty, a one-time bootstrap secret should be generated (handled by caller or here).
func (s *Service) CreateTenant(ctx context.Context, name string, ownerEmail string, ownerPassword string, creatorUserID string) (*Tenant, error) {
	// 1. Validate name
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidTenantName
	}
	if len(name) < 3 || len(name) > 100 {
		return nil, ErrInvalidTenantName
	}

	// 2. Check for duplicate name
	existing, err := s.repo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, ErrTenantAlreadyExists
	}

	// 3. Provision owner identity (optional)
	var owner *user.User
	if ownerEmail != "" {
		owner, err = s.identityService.GetByEmail(ctx, ownerEmail)
		if err != nil {
			if errors.Is(err, user.ErrUserNotFound) {
				// Provision new identity
				owner, err = s.identityService.ProvisionIdentity(ctx, ownerEmail, user.Profile{
					GivenName:  "Tenant",
					FamilyName: "Owner",
				})
				if err != nil {
					return nil, fmt.Errorf("failed to provision tenant owner identity: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to check owner identity: %w", err)
			}
		}

		// Always set/update password if provided in this bootstrap flow
		if ownerPassword != "" {
			if err := s.identityService.SetPassword(ctx, owner.ID, ownerPassword); err != nil {
				return nil, fmt.Errorf("failed to set tenant owner password: %w", err)
			}
		}
	}

	// 4. Generate UUID v7 (RFC 9562) for tenant
	tenantID := id.NewUUIDv7()

	now := time.Now()
	tenant := &Tenant{
		ID:        tenantID,
		Name:      name,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 5. Create tenant
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// 6. Assign Permission: tenant_owner role (if owner exists)
	if owner != nil {
		if err := s.AssignRole(ctx, tenantID, owner.ID, role.RoleTenantOwner, creatorUserID); err != nil {
			return nil, fmt.Errorf("failed to assign tenant owner role: %w", err)
		}
	}

	auditMetadata := map[string]any{
		audit.AttrTenantID:   tenantID,
		audit.AttrTenantName: tenant.Name,
	}
	if owner != nil {
		auditMetadata["owner_id"] = owner.ID
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeTenantCreated,
		ActorID:    creatorUserID,
		Resource:   audit.ResourceTenant,
		TargetName: tenant.Name,
		TargetID:   tenantID,
		Metadata:   auditMetadata,
	})

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(ctx context.Context, id string) (*Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

// GetTenantByName retrieves a tenant by name
func (s *Service) GetTenantByName(ctx context.Context, name string) (*Tenant, error) {
	return s.repo.GetByName(ctx, name)
}

// ListTenants retrieves tenants with pagination.
func (s *Service) ListTenants(ctx context.Context, limit, offset int) ([]*Tenant, error) {
	return s.repo.List(ctx, limit, offset)
}

// UpdateTenant updates a tenant
func (s *Service) UpdateTenant(ctx context.Context, tenantID string, name string, actorID string) (*Tenant, error) {
	t, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	oldName := t.Name
	if name != "" {
		t.Name = name
	}

	// Handle status transitions if needed
	// Valid status transitions: active <-> suspended
	// Platform admin can suspend tenants, suspended tenants can be reactivated
	// This can be extended in the future to support more status values
	// For now, status changes are not supported via UpdateTenant, use dedicated status endpoints if needed

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Audit log
	metadata := map[string]any{
		audit.AttrTenantID:   tenantID,
		audit.AttrTenantName: t.Name,
	}

	if oldName != t.Name {
		metadata["changes"] = map[string]string{
			"name_from": oldName,
			"name_to":   t.Name,
		}
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeTenantUpdated,
		ActorID:    actorID,
		Resource:   audit.ResourceTenant,
		TargetName: t.Name,
		TargetID:   t.ID,
		Metadata:   metadata,
	})
	return t, nil
}

// DeleteTenant deletes a tenant and performs cascading soft-deletion of associated data
func (s *Service) DeleteTenant(ctx context.Context, tenantID string, actorID string) error {
	// 1. Fetch tenant first to get name for audit
	t, err := s.repo.GetByID(ctx, tenantID)
	tenantName := "Unknown"
	if err == nil && t != nil {
		tenantName = t.Name
	}

	// 2. Perform cascading soft-deletion
	// Note: In a production system, these should ideally be in a transaction.
	// However, since we are doing soft-deletes (UPDATE), partial failure is recoverable.

	// 1. Delete memberships
	if s.membershipRepo != nil {
		if err := s.membershipRepo.DeleteByTenantID(ctx, tenantID); err != nil {
			return fmt.Errorf("failed to cascade membership deletion: %w", err)
		}
	}

	// 2. Delete clients
	if s.clientRepo != nil {
		if err := s.clientRepo.DeleteByTenantID(ctx, tenantID); err != nil {
			return fmt.Errorf("failed to cascade client deletion: %w", err)
		}
	}

	// 3. Delete role assignments (Tenant internal table)
	if s.roleRepo != nil {
		if err := s.roleRepo.DeleteByTenantID(ctx, tenantID); err != nil {
			return fmt.Errorf("failed to cascade tenant role deletion: %w", err)
		}
	}

	// 4. Delete RBAC assignments (Authz table)
	if s.authzRepo != nil {
		if err := s.authzRepo.DeleteByContextID(ctx, policy.ScopeTenant, tenantID); err != nil {
			return fmt.Errorf("failed to cascade rbac assignment deletion: %w", err)
		}
	}

	// 5. Delete tenant itself
	if err := s.repo.Delete(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeTenantDeleted,
		ActorID:    actorID,
		Resource:   audit.ResourceTenant,
		TargetName: tenantName,
		TargetID:   tenantID,
		Metadata: map[string]any{
			audit.AttrTenantID:   tenantID,
			audit.AttrTenantName: tenantName,
		},
	})
	return nil
}

// AssignRole assigns a role to a user in a tenant
func (s *Service) AssignRole(ctx context.Context, tenantID, userID, roleName string, grantedBy string) error {
	// 1. Persist in tenant_user_roles (Legacy/Primary)
	// Validate role
	if roleName != role.RoleTenantOwner && roleName != role.RoleTenantAdmin && roleName != role.RoleTenantMember {
		return fmt.Errorf("invalid role: %s", roleName)
	}

	if err := s.roleRepo.AssignRole(ctx, tenantID, userID, roleName, grantedBy); err != nil {
		return err
	}

	// 2. Ensure membership exists (Simple Link, no fingerprint)
	if s.membershipRepo != nil {
		// Just try to create, ignore if already exists (unique constraint handles it)
		_ = s.membershipRepo.AddMember(ctx, &Membership{
			ID:        id.NewUUIDv7(),
			TenantID:  tenantID,
			UserID:    userID,
			CreatedAt: time.Now(),
		})
	}

	// ALSO create an authz assignment for proper permission checking
	// Map tenant role name to the seeded authz role UUID from migration
	var authzRoleID string
	switch roleName {
	case role.RoleTenantOwner:
		authzRoleID = role.RoleIDTenantOwner
	case role.RoleTenantAdmin:
		authzRoleID = role.RoleIDTenantAdmin
	case role.RoleTenantMember:
		authzRoleID = role.RoleIDMember
	default:
		authzRoleID = roleName // Fallback to name, but this shouldn't happen
	}

	if s.authzRepo != nil && authzRoleID != "" {
		authzAssignment := &policy.Assignment{
			ID:             id.NewUUIDv7(),
			UserID:         userID,
			RoleID:         authzRoleID,
			Scope:          policy.ScopeTenant,
			ScopeContextID: &tenantID,
			GrantedAt:      time.Now(),
			GrantedBy:      grantedBy,
		}
		if err := s.authzRepo.Grant(ctx, authzAssignment); err != nil {
			return fmt.Errorf("failed to grant authz role: %w", err)
		}
	}

	// Audit role assignment
	// Try to get user email/name for TargetName
	targetName := userID
	if u, err := s.identityService.GetUser(ctx, userID); err == nil {
		if u.EmailPlain != nil {
			targetName = *u.EmailPlain
		}
		if u.Profile.Nickname != "" {
			targetName = fmt.Sprintf("%s (%s)", u.Profile.Nickname, targetName)
		}
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeRoleAssigned,
		TenantID:   tenantID,
		ActorID:    grantedBy,
		Resource:   roleName,
		TargetName: targetName,
		TargetID:   userID,
		Metadata:   map[string]any{audit.AttrActorID: userID},
	})

	return nil
}

// RevokeRole revokes a role from a user in a tenant
func (s *Service) RevokeRole(ctx context.Context, tenantID, userID, roleName string, actorID string) error {
	// 1. Security Check: Prevent self-revocation of tenant_owner role to avoid accidental lockouts.
	if userID == actorID && roleName == role.RoleTenantOwner {
		return fmt.Errorf("security violation: tenant owners cannot revoke their own owner role")
	}

	if err := s.roleRepo.RevokeRole(ctx, tenantID, userID, roleName); err != nil {
		return err
	}

	// Audit role revocation
	targetName := userID
	if u, err := s.identityService.GetUser(ctx, userID); err == nil {
		if u.EmailPlain != nil {
			targetName = *u.EmailPlain
		}
		if u.Profile.Nickname != "" {
			targetName = fmt.Sprintf("%s (%s)", u.Profile.Nickname, targetName)
		}
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeRoleRevoked,
		TenantID:   tenantID,
		ActorID:    actorID,
		Resource:   roleName,
		TargetName: targetName,
		TargetID:   userID,
		Metadata:   map[string]any{audit.AttrActorID: userID},
	})

	return nil
}

// GetUserRoles retrieves all roles a user has in a tenant
func (s *Service) GetUserRoles(ctx context.Context, tenantID, userID string) ([]*TenantUserRole, error) {
	return s.roleRepo.GetUserRoles(ctx, tenantID, userID)
}

// GetTenantUsers retrieves all users with roles in a tenant
func (s *Service) GetTenantUsers(ctx context.Context, tenantID string) ([]*TenantUserRole, error) {
	return s.roleRepo.GetTenantUsers(ctx, tenantID)
}

// UpdateUser updates a user's profile information
func (s *Service) UpdateUser(ctx context.Context, tenantID, userID string, profile user.Profile, actorID string) error {
	// 2. Update profile in identity service
	if err := s.identityService.UpdateProfile(ctx, userID, profile); err != nil {
		return err
	}

	// 3. Audit Log
	targetName := userID
	if u, err := s.identityService.GetUser(ctx, userID); err == nil {
		if u.EmailPlain != nil {
			targetName = *u.EmailPlain
		}
		if u.Profile.Nickname != "" {
			targetName = fmt.Sprintf("%s (%s)", u.Profile.Nickname, targetName)
		}
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeUserUpdated,
		TenantID:   tenantID,
		ActorID:    actorID,
		Resource:   audit.ResourceUser,
		TargetName: targetName,
		TargetID:   userID,
		Metadata: map[string]any{
			"user_id":  userID,
			"nickname": profile.Nickname,
		},
	})

	return nil
}
