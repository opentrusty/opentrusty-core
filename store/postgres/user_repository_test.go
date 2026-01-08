// Copyright 2026 The OpenTrusty Authors
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"testing"

	"github.com/opentrusty/opentrusty-core/user"
)

func TestUserRepository(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewUserRepository(db)

	u := &user.User{
		ID:         "00000000-0000-0000-0000-000000000101",
		EmailHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // sha256 of empty string
		EmailPlain: stringPtr("user1@example.com"),
		Profile: user.Profile{
			FullName: "User One",
		},
	}

	t.Run("Create and Get", func(t *testing.T) {
		err := repo.Create(ctx, u)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		got, err := repo.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if got.EmailHash != u.EmailHash {
			t.Errorf("expected hash %s, got %s", u.EmailHash, got.EmailHash)
		}
	})

	t.Run("Update", func(t *testing.T) {
		u.Profile.FullName = "User One Updated"
		err := repo.Update(ctx, u)
		if err != nil {
			t.Fatalf("failed to update user: %v", err)
		}

		got, err := repo.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if got.Profile.FullName != "User One Updated" {
			t.Errorf("expected updated name, got %s", got.Profile.FullName)
		}
	})

	t.Run("Credentials", func(t *testing.T) {
		c := &user.Credentials{
			UserID:       u.ID,
			PasswordHash: "passhash",
		}
		err := repo.AddCredentials(ctx, c)
		if err != nil {
			t.Fatalf("failed to add credentials: %v", err)
		}

		got, err := repo.GetCredentials(ctx, u.ID)
		if err != nil {
			t.Fatalf("failed to get credentials: %v", err)
		}
		if got.PasswordHash != "passhash" {
			t.Errorf("expected passhash, got %s", got.PasswordHash)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, u.ID)
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}

		_, err = repo.GetByID(ctx, u.ID)
		if err != user.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
