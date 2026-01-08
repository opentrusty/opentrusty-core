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

package project

import (
	"context"
	"time"
)

// Project represents a project/resource that users can access.
//
// Purpose: Entity representing a resource boundary for authorization.
// Domain: Platform
// Invariants: ID must be unique. OwnerID must exist.
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	OwnerID     string     `json:"owner_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// ProjectRepository defines the interface for project persistence.
//
// Purpose: Abstraction for managing resource boundary storage.
// Domain: Platform
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
