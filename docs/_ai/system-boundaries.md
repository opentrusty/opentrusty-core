# System Boundaries

This document defines the strict physical and logical boundaries of the OpenTrusty ecosystem.

## Repository Topology

OpenTrusty is split into five core repositories (plus a demo app) to ensure maximum decoupling and security isolation.

1. **`opentrusty-core` (The Pure Core)**
   - **Registry**: `github.com/opentrusty/opentrusty-core`
   - **Role**: Domain authority, Cryptographic primitives, Repository interfaces, Store implementations.
   - **Constraint**: **ZERO** HTTP, **ZERO** CLI, **ZERO** deploy artifacts.
2. **`opentrusty-auth` (The Auth Plane)**
   - **Registry**: `github.com/opentrusty/opentrusty-auth`
   - **Binary**: `authd`
   - **Role**: OIDC/OAuth2 Gateway, login/consent UI.
3. **`opentrusty-admin` (The Admin Plane)**
   - **Registry**: `github.com/opentrusty/opentrusty-admin`
   - **Binary**: `admind`
   - **Role**: Management API, Audit logs, Bootstrap hooks.
4. **`opentrusty-cli` (The Operator Tooling)**
   - **Registry**: `github.com/opentrusty/opentrusty-cli`
   - **Binary**: `opentrusty`
   - **Role**: Migrations, Semantic Bootstrap, Admin scripting, Deployment orchestration.
5. **`opentrusty-control-panel` (The UI)**
   - **Registry**: `github.com/opentrusty/opentrusty-control-panel`
   - **Binary**: N/A (SPA assets)
   - **Role**: Administrative Frontend.
6. **`opentrusty-demo-app` (E2E Test Client)**
   - **Registry**: `github.com/opentrusty/opentrusty-demo-app`
   - **Binary**: `demo-app`
   - **Role**: Simulates a real third-party OIDC Relying Party for end-to-end testing.

## What This Repository (opentrusty-core) Owns

### Domain Models & Business Logic
- Identity service (user management, password hashing)
- Session service (state management)
- Authorization service (RBAC enforcement, permission checks)
- Tenant service (lifecycle, membership, isolation logic)
- Client models (OAuth2 client definitions)
- Audit models (event structure)
- Project models (resource boundary for authorization)
- Policy models (scope, role, assignment definitions)
- Cryptographic primitives (Argon2id, HMAC)

### Data Access Layer
- `store/postgres/` — PostgreSQL repository implementations for all domain entities

## What This Repository Does NOT Own

| Component | Owner | Interaction |
|-----------|-------|-------------|
| HTTP transport (auth endpoints) | `opentrusty-auth` | Imports core packages |
| HTTP transport (admin endpoints) | `opentrusty-admin` | Imports core packages |
| CLI commands, migrations, bootstrap | `opentrusty-cli` | Imports core packages |
| Control Panel UI | `opentrusty-control-panel` | Consumes Admin API via HTTP |
| Deploy scripts, systemd units, Caddyfile | `opentrusty-cli` | Operator tooling |
| E2E test client | `opentrusty-demo-app` | Uses OIDC protocol |

## Dependencies

### This Repo Depends On
- PostgreSQL (persistence via pgx)
- golang.org/x/crypto (Argon2id)
- google/uuid

### Other Repos Depend On This Repo
- `opentrusty-auth` → Go module import
- `opentrusty-admin` → Go module import
- `opentrusty-cli` → Go module import

## Forbidden Cross-Overs

| Action | Status |
|--------|--------|
| Embedding SPA code in binary | ❌ FORBIDDEN |
| Serving static UI assets | ❌ FORBIDDEN |
| Importing frontend frameworks | ❌ FORBIDDEN |
| Storing frontend secrets | ❌ FORBIDDEN |
| Implementing UI routing | ❌ FORBIDDEN |
| HTTP handlers or routers | ❌ FORBIDDEN |
| CLI parsing or commands | ❌ FORBIDDEN |
| Deploy scripts or systemd units | ❌ FORBIDDEN |
