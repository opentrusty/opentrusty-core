// Copyright 2026 The OpenTrusty Authors
// SPDX-License-Identifier: MIT

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/opentrusty/opentrusty-core/role"
)

// SetupTestDB creates a connection to the test database and runs migrations.
func SetupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()

	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5434" // Default port in docker-compose.test.yml
	}

	cfg := Config{
		Host:         host,
		Port:         port,
		User:         "opentrusty",
		Password:     "opentrusty_test_password",
		Database:     "opentrusty_test",
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 10,
	}

	ctx := context.Background()
	db, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Clean up before starting (in case previous run failed badly)
	tables := []string{
		"audit_events",
		"audit_logs",
		"sessions",
		"rbac_assignments",
		"rbac_role_permissions",
		"rbac_roles",
		"rbac_permissions",
		"oauth2_clients",
		"clients",             // if exists
		"tokens",              // if exists
		"authorization_codes", // if exists
		"tenant_members",
		"memberships", // if exists
		"projects",
		"credentials",
		"users",
		"tenants",
	}
	for _, table := range tables {
		// Use IF EXISTS to avoid errors if schema is not yet created
		_, _ = db.pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}

	// Run initial schema
	if err := db.Migrate(ctx, InitialSchema); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Seed RBAC (Permissions & Roles)
	if err := seedRBAC(ctx, db); err != nil {
		db.Close()
		t.Fatalf("failed to seed RBAC: %v", err)
	}

	cleanup := func() {
		// Clean up tables
		tables := []string{
			"audit_logs",
			"sessions",
			"role_assignments",
			"clients",
			"tokens",
			"authorization_codes",
			"memberships",
			"users",
			"tenants",
			"roles",
		}
		for _, table := range tables {
			_, _ = db.pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		}
		db.Close()
	}

	return db, cleanup
}

func seedRBAC(ctx context.Context, db *DB) error {
	perms := []struct {
		ID   string
		Name string
	}{
		{"00000000-0000-0000-0000-000000000001", "platform:manage_tenants"},
		{"00000000-0000-0000-0000-000000000002", "tenant:manage_users"},
		{"00000000-0000-0000-0000-000000000003", "user:read_profile"},
	}

	for _, p := range perms {
		_, err := db.pool.Exec(ctx, `
			INSERT INTO rbac_permissions (id, name, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (name) DO NOTHING
		`, p.ID, p.Name)
		if err != nil {
			return err
		}
	}

	roles := []struct {
		ID    string
		Name  string
		Scope role.Scope
	}{
		{role.RoleIDPlatformAdmin, role.RolePlatformAdmin, role.ScopePlatform},
		{role.RoleIDTenantOwner, role.RoleTenantOwner, role.ScopeTenant},
		{role.RoleIDTenantAdmin, role.RoleTenantAdmin, role.ScopeTenant},
		{role.RoleIDMember, role.RoleTenantMember, role.ScopeTenant},
	}

	for _, r := range roles {
		_, err := db.pool.Exec(ctx, `
			INSERT INTO rbac_roles (id, name, scope, created_at, updated_at)
			VALUES ($1, $2, $3, NOW(), NOW())
			ON CONFLICT (scope, name) DO NOTHING
		`, r.ID, r.Name, string(r.Scope))
		if err != nil {
			return err
		}
	}

	// Link Permissions
	// 1. Platform Admin -> *
	// Ensure * permission exists
	_, err := db.pool.Exec(ctx, `INSERT INTO rbac_permissions (id, name, created_at) VALUES ($1, '*', NOW()) ON CONFLICT (name) DO NOTHING`, "00000000-0000-0000-0000-000000000099")
	if err != nil {
		return err
	}
	// Retrieve ID of * (in case it existed)
	var wildcardPermID string
	err = db.pool.QueryRow(ctx, "SELECT id FROM rbac_permissions WHERE name = '*'").Scan(&wildcardPermID)
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx, `
		INSERT INTO rbac_role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, role.RoleIDPlatformAdmin, wildcardPermID)
	if err != nil {
		return err
	}

	return nil
}
