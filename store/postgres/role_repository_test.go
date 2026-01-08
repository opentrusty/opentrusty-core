// Copyright 2026 The OpenTrusty Authors
// SPDX-License-Identifier: MIT

package postgres

import (
	"context"
	"testing"

	"github.com/opentrusty/opentrusty-core/role"
)

func TestRoleRepository(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewRoleRepository(db)

	r := &role.Role{
		ID:          "00000000-0000-0000-0000-000000000201",
		Name:        "Platform Editor",
		Scope:       role.ScopePlatform,
		Description: "Can edit platform settings",
		Permissions: []string{"platform:manage_tenants"},
	}

	t.Run("Create and Get", func(t *testing.T) {
		err := repo.Create(ctx, r)
		if err != nil {
			t.Fatalf("failed to create role: %v", err)
		}

		got, err := repo.GetByID(ctx, r.ID)
		if err != nil {
			t.Fatalf("failed to get role: %v", err)
		}
		if got.Name != r.Name {
			t.Errorf("expected name %s, got %s", r.Name, got.Name)
		}
		if len(got.Permissions) != 1 || got.Permissions[0] != "platform:manage_tenants" {
			t.Errorf("expected permission platform:manage_tenants, got %v", got.Permissions)
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		got, err := repo.GetByName(ctx, r.Name, r.Scope)
		if err != nil {
			t.Fatalf("failed to get role by name: %v", err)
		}
		if got.ID != r.ID {
			t.Errorf("expected ID %s, got %s", r.ID, got.ID)
		}
	})

	t.Run("List", func(t *testing.T) {
		roles, err := repo.List(ctx, nil)
		if err != nil {
			t.Fatalf("failed to list roles: %v", err)
		}
		if len(roles) == 0 {
			t.Errorf("expected at least one role")
		}
	})

	t.Run("Update", func(t *testing.T) {
		r.Description = "Updated description"
		err := repo.Update(ctx, r)
		if err != nil {
			t.Fatalf("failed to update role: %v", err)
		}

		got, err := repo.GetByID(ctx, r.ID)
		if err != nil {
			t.Fatalf("failed to get role: %v", err)
		}
		if got.Description != "Updated description" {
			t.Errorf("expected updated description, got %s", got.Description)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, r.ID)
		if err != nil {
			t.Fatalf("failed to delete role: %v", err)
		}

		_, err = repo.GetByID(ctx, r.ID)
		if err == nil {
			t.Errorf("expected error after delete, got nil")
		}
	})
}
