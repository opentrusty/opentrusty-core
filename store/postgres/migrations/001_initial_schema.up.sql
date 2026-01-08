-- 001_initial_schema.up.sql
-- Optimized Version following OpenTrusty Identity & Authorization Model principles.

-- 1. Scoped RBAC Tables
CREATE TABLE IF NOT EXISTS rbac_permissions (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rbac_roles (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, scope)
);

CREATE TABLE IF NOT EXISTS rbac_role_permissions (
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES rbac_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 2. Core Identity Tables
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email_hash CHAR(64) NOT NULL UNIQUE,
    email_plain VARCHAR(255),
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    given_name VARCHAR(255),
    family_name VARCHAR(255),
    full_name VARCHAR(255),
    nickname VARCHAR(255),
    picture TEXT,
    locale VARCHAR(10),
    timezone VARCHAR(50),
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tenant_members (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);

CREATE TABLE IF NOT EXISTS credentials (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    namespace VARCHAR(50) NOT NULL DEFAULT ''
);

-- 3. Scoped RBAC Assignments
CREATE TABLE IF NOT EXISTS rbac_assignments (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES rbac_roles(id) ON DELETE CASCADE,
    scope VARCHAR(50) NOT NULL CHECK (scope IN ('platform', 'tenant', 'client')),
    scope_context_id UUID,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    granted_by UUID REFERENCES users(id),
    UNIQUE(user_id, role_id, scope, scope_context_id),
    CHECK ((scope = 'platform' AND scope_context_id IS NULL) OR (scope != 'platform' AND scope_context_id IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 4. OAuth2 & OIDC Tables (Omitted for brevity in this step, but should be fully migrated)
CREATE TABLE IF NOT EXISTS oauth2_clients (
    id UUID PRIMARY KEY,
    client_id UUID UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_secret_hash TEXT NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    client_uri VARCHAR(255),
    logo_uri VARCHAR(255),
    redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
    allowed_scopes JSONB NOT NULL DEFAULT '["openid"]'::jsonb,
    grant_types JSONB NOT NULL DEFAULT '["authorization_code"]'::jsonb,
    response_types JSONB NOT NULL DEFAULT '["code"]'::jsonb,
    token_endpoint_auth_method VARCHAR(50) NOT NULL DEFAULT 'client_secret_basic',
    access_token_lifetime INT NOT NULL DEFAULT 3600,
    refresh_token_lifetime INT NOT NULL DEFAULT 2592000,
    id_token_lifetime INT NOT NULL DEFAULT 3600,
    owner_id UUID REFERENCES users(id) ON DELETE SET NULL,
    is_trusted BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS authorization_codes (
    id UUID PRIMARY KEY,
    code TEXT UNIQUE NOT NULL,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    scope TEXT,
    state TEXT,
    nonce TEXT,
    code_challenge TEXT,
    code_challenge_method VARCHAR(50),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_tokens (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT,
    token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    access_token_id UUID,
    client_id UUID NOT NULL REFERENCES oauth2_clients(client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope TEXT,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Audit Events
CREATE TABLE IF NOT EXISTS audit_events (
    id UUID PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255),
    actor_id VARCHAR(255),
    resource VARCHAR(255) NOT NULL,
    target_name VARCHAR(255),
    target_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. Seed Initial RBAC Data
INSERT INTO rbac_permissions (id, name, created_at) VALUES 
('00000000-0000-0000-0000-000000000001', 'platform:manage_tenants', NOW()),
('00000000-0000-0000-0000-000000000002', 'tenant:manage_users', NOW()),
('00000000-0000-0000-0000-000000000003', 'user:read_profile', NOW()),
('00000000-0000-0000-0000-000000000004', 'control_plane:login', NOW()),
('00000000-0000-0000-0000-000000000005', 'tenant:view', NOW()),
('00000000-0000-0000-0000-000000000006', 'tenant:view_audit', NOW()),
('00000000-0000-0000-0000-000000000007', 'tenant:manage_clients', NOW()),
('00000000-0000-0000-0000-000000000099', '*', NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO rbac_roles (id, name, scope, created_at, updated_at) VALUES 
('00000000-0000-0000-0000-000000000001', 'platform_admin', 'platform', NOW(), NOW()),
('00000000-0000-0000-0000-000000000002', 'tenant_owner', 'tenant', NOW(), NOW()),
('00000000-0000-0000-0000-000000000003', 'tenant_admin', 'tenant', NOW(), NOW()),
('00000000-0000-0000-0000-000000000004', 'tenant_member', 'tenant', NOW(), NOW())
ON CONFLICT (scope, name) DO NOTHING;

-- Link Permissions
-- Platform Admin -> *
INSERT INTO rbac_role_permissions (role_id, permission_id) VALUES 
('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000099'), -- Platform Admin -> *
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000004'), -- Tenant Owner -> control_plane:login
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002'), -- Tenant Owner -> tenant:manage_users
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000003'), -- Tenant Owner -> user:read_profile
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000005'), -- Tenant Owner -> tenant:view
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000006'), -- Tenant Owner -> tenant:view_audit
('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000007'), -- Tenant Owner -> tenant:manage_clients
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000004'), -- Tenant Admin -> control_plane:login
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000002'), -- Tenant Admin -> tenant:manage_users
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000003'), -- Tenant Admin -> user:read_profile
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000005'), -- Tenant Admin -> tenant:view
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000006'), -- Tenant Admin -> tenant:view_audit
('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000007'), -- Tenant Admin -> tenant:manage_clients
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000005'), -- Tenant Member -> tenant:view
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000004'), -- Tenant Member -> control_plane:login (Assuming members can login)
('00000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000003')  -- Tenant Member -> user:read_profile
ON CONFLICT DO NOTHING;
