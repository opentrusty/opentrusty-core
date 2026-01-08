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

package policy

import (
	"context"
	"errors"
	"time"
)

// Domain errors
var (
	ErrProjectNotFound         = errors.New("project not found")
	ErrProjectAlreadyExists    = errors.New("project already exists")
	ErrAssignmentNotFound      = errors.New("assignment not found")
	ErrAssignmentAlreadyExists = errors.New("assignment already exists")
	ErrRoleNotFound            = errors.New("role not found")
	ErrRoleAlreadyExists       = errors.New("role already exists")
	ErrAccessDenied            = errors.New("access denied")
	ErrInvalidPermission       = errors.New("invalid permission")
	ErrInvalidScope            = errors.New("invalid scope")
)

// Scope defines the level at which a role is assigned
type Scope string

const (
	ScopePlatform Scope = "platform"
	ScopeTenant   Scope = "tenant"
	ScopeClient   Scope = "client"
)

// Project represents a project/resource that users can access
type Project struct {
	ID          string
	Name        string
	Description string
	OwnerID     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Role represents a scoped role with associated permission names
type Role struct {
	ID          string
	Name        string
	Scope       Scope
	Description string
	Permissions []string // Names of permissions
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

// Assignment represents a role granted to a user at a specific scope
type Assignment struct {
	ID             string
	UserID         string
	RoleID         string
	Scope          Scope
	ScopeContextID *string // NULL for platform, tenant_id for tenant, etc.
	GrantedAt      time.Time
	GrantedBy      string
}

// AssignmentRepository defines the interface for RBAC assignments
type AssignmentRepository interface {
	// Grant assigns a role to a user
	Grant(ctx context.Context, assignment *Assignment) error

	// Revoke removes a role assignment
	Revoke(ctx context.Context, userID, roleID string, scope Scope, scopeContextID *string) error

	// ListForUser retrieves all assignments for a user
	ListForUser(ctx context.Context, userID string) ([]*Assignment, error)

	// ListByRole retrieves all users assigned a specific role at a scope
	ListByRole(ctx context.Context, roleID string, scope Scope, scopeContextID *string) ([]string, error)

	// CheckExists checks if a specific assignment exists
	CheckExists(ctx context.Context, roleID string, scope Scope, scopeContextID *string) (bool, error)

	// DeleteByContextID removes all assignments for a specific scope and context
	DeleteByContextID(ctx context.Context, scope Scope, contextID string) error
}

// ProjectRepository defines the interface for project persistence
type ProjectRepository interface {
	// Create creates a new project
	Create(ctx context.Context, project *Project) error

	// GetByID retrieves a project by ID
	GetByID(ctx context.Context, id string) (*Project, error)

	// GetByName retrieves a project by name
	GetByName(ctx context.Context, name string) (*Project, error)

	// Update updates project information
	Update(ctx context.Context, project *Project) error

	// Delete soft-deletes a project
	Delete(ctx context.Context, id string) error

	// ListByOwner retrieves all projects owned by a user
	ListByOwner(ctx context.Context, ownerID string) ([]*Project, error)

	// ListByUser retrieves all projects a user has access to
	ListByUser(ctx context.Context, userID string) ([]*Project, error)
}

// RoleRepository defines the interface for role persistence
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id string) (*Role, error)

	// GetByName retrieves a role by name and scope
	GetByName(ctx context.Context, name string, scope Scope) (*Role, error)

	// Update updates role information
	Update(ctx context.Context, role *Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id string) error

	// List retrieves all roles, optionally filtered by scope
	List(ctx context.Context, scope *Scope) ([]*Role, error)
}
