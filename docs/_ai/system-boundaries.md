# System Boundaries

This document defines the strict physical and logical boundaries of the OpenTrusty ecosystem.

## Repository Topology

OpenTrusty is split into five repositories to ensure maximum decoupling and security isolation.

1. **`opentrusty-core` (The Pure Core)**
   - **Registry**: `github.com/opentrusty/opentrusty-core`
   - **Role**: Domain authority, Cryptographic primitives, Repository interfaces.
   - **Constraint**: **ZERO** HTTP, **ZERO** CLI.
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
   - **Role**: Migrations, Semantic Bootstrap, Admin scripting.
5. **`opentrusty-control-panel` (The UI)**
   - **Registry**: `github.com/opentrusty/opentrusty-control-panel`
   - **Binary**: N/A (SPA assets)
   - **Role**: Administrative Frontend.

## What This Repository Owns

### Authentication Plane (`auth.*`)
- OIDC/OAuth2 protocol endpoints
- Server-rendered login, consent, and error pages
- Session cookie issuance and validation
- Token generation and signing
- **Constraint**: MUST NOT expose Management APIs (404 enforced)

### Management API Plane (`api.*`)
- Tenant lifecycle (create, read, update, delete)
- User provisioning and management
- OAuth client registration
- RBAC role and assignment management
- Audit log access
- **Constraint**: MUST NOT expose Login/OIDC endpoints (404 enforced)

### Shared Domain Core
- Identity service (user management)
- Session service (state management)
- Authorization service (RBAC enforcement)
- Tenant service (isolation logic)
- Database repositories

## What This Repository Does NOT Own

| Component | Owner | Interaction |
|-----------|-------|-------------|
| Control Panel UI | `opentrusty-control-panel` | Consumes Management API |
| Static SPA assets | `opentrusty-control-panel` | None |
| Frontend routing | `opentrusty-control-panel` | None |
| React/Vue/Tailwind | `opentrusty-control-panel` | None |

## Dependencies

### This Repo Depends On
- PostgreSQL (persistence)
- OpenTelemetry collector (observability, optional)

### Other Repos Depend On This Repo
- `opentrusty-control-panel` depends on Management API (`api.*`)

## Forbidden Cross-Overs

| Action | Status |
|--------|--------|
| Embedding SPA code in binary | ❌ FORBIDDEN |
| Serving static UI assets | ❌ FORBIDDEN |
| Importing frontend frameworks | ❌ FORBIDDEN |
| Storing frontend secrets | ❌ FORBIDDEN |
| Implementing UI routing | ❌ FORBIDDEN |
