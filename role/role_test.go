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

package role

import (
	"testing"
)

func TestRoleHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       Role
		permission string
		want       bool
	}{
		{
			name: "exact match",
			role: Role{
				Permissions: []string{"read:users", "write:users"},
			},
			permission: "read:users",
			want:       true,
		},
		{
			name: "wildcard match",
			role: Role{
				Permissions: []string{"*"},
			},
			permission: "any:permission",
			want:       true,
		},
		{
			name: "no match",
			role: Role{
				Permissions: []string{"read:users"},
			},
			permission: "write:users",
			want:       false,
		},
		{
			name: "empty permissions",
			role: Role{
				Permissions: []string{},
			},
			permission: "read:users",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.HasPermission(tt.permission); got != tt.want {
				t.Errorf("Role.HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultRoleMappings(t *testing.T) {
	// Verify that default roles have the expected "anchor" permissions
	platformAdmin := Role{Permissions: PlatformAdminPermissions}
	if !platformAdmin.HasPermission("random:perm") {
		t.Error("Platform admin should have all permissions via wildcard")
	}

	tenantOwner := Role{Permissions: TenantOwnerPermissions}
	if !tenantOwner.HasPermission("tenant:manage_users") {
		t.Error("Tenant owner should have tenant:manage_users permission")
	}

	tenantMember := Role{Permissions: TenantMemberPermissions}
	if tenantMember.HasPermission("tenant:manage_users") {
		t.Error("Tenant member should NOT have tenant:manage_users permission")
	}
}
