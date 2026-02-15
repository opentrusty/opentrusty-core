# Architecture Map

This document defines the current boundaries of the system.

## Domain Boundaries (Top-Level Packages)

The core library is structured by domain as **top-level public packages** (exported via Go module).
Cross-domain dependencies should be minimized and explicit.

| Package | Domain Responsibility | Dependencies (Allowed) |
| :--- | :--- | :--- |
| `audit/` | Audit logging (Who did what) | — |
| `authz/` | Authorization Enforcement (RBAC) | `policy`, `project`, `role` |
| `client/` | OAuth2 Client management | — |
| `crypto/` | Cryptographic primitives | — |
| `id/` | ID generation utilities | — |
| `password/` | Password hashing (Argon2id) | — |
| `policy/` | Policy models, Scope, Permissions | — |
| `project/` | Project/Resource boundary for authorization | — |
| `role/` | Role models and interfaces | — |
| `session/` | Session primitives and service | — |
| `tenant/` | Tenant lifecycle and membership | `user`, `client`, `role`, `audit` |
| `user/` | User management, credentials | `audit` |
| `store/postgres/` | PostgreSQL Data Access Layer | All domain packages |

## Layering

1.  **Transport Layers** (External Repos): Handles HTTP/GRPC. Includes `opentrusty-auth` and `opentrusty-admin`.
2.  **Domain Layer** (`opentrusty-core`): Pure business logic. **ZERO HTTP, ZERO CLI.**
3.  **Storage Layer** (`store/postgres/`): Database interactions.

## Protocol Logic

-   **OAuth2** logic resides strictly in `opentrusty-auth/internal/oauth2`.
-   **OIDC** logic resides strictly in `opentrusty-auth/internal/oidc`.
-   **Session** handling resides in `session/` (core primitives) and respective plane middleware.

## External Consumers

The core engine is consumed by the following physical planes:

| Plane | Repository | Relationship |
| :--- | :--- | :--- |
| **Auth Plane** | `github.com/opentrusty/opentrusty-auth` | OIDC/OAuth2 Protocol Gateway |
| **Admin Plane** | `github.com/opentrusty/opentrusty-admin` | Management API Gateway |
| **CLI Tools** | `github.com/opentrusty/opentrusty-cli` | Operator CLI (Migrate/Bootstrap) |
| **Control Panel** | `github.com/opentrusty/opentrusty-control-panel` | Administrative Frontend (Untrusted) |
| **Demo App** | `github.com/opentrusty/opentrusty-demo-app` | E2E Test Client (Third-Party Simulation) |

**Critical Rule**: The `opentrusty-core` repository contains **ZERO transport logic** (no HTTP handlers, no CLI parsing).
All interface logic resides in the respective external repositories.
