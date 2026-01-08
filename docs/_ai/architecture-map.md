# Architecture Map

This document defines the current boundaries of the system.

## Domain Boundaries (`internal/`)

The application is structured by domain. Cross-domain dependencies should be minimized and explicit.

| Directory | Domain Responsibility | Dependencies (Allowed) |
| :--- | :--- | :--- |
| `internal/audit` | Audit logging (Who did what) | `store`, `observability` |
| `internal/authz` | Authorization Enforcement (RBAC) | `store`, `identity` |
| `internal/config` | Configuration primitives | *None* |
| `internal/identity` | User management, Credentials | `store`, `tenant` |
| `internal/oauth2` | OAuth2 Domain logic | `store`, `identity`, `tenant` |
| `internal/observability` | Tracing, Metrics, Logging | *None* |
| `internal/oidc` | OIDC Domain logic | `store`, `oauth2` |
| `internal/session` | Session Primitives | `store`, `identity` |
| `internal/store` | Data Access Layer (PostgreSQL) | *None* (Leaf) |
| `internal/tenant` | Tenant Lifecycle | `store` |

## Layering

1.  **Transport Layers** (External Repos): Handles HTTP/GRPC. Includes `opentrusty-auth` and `opentrusty-admin`.
2.  **Domain Layer** (`opentrusty-core`): Pure business logic. **ZERO HTTP, ZERO CLI.**
3.  **Storage Layer** (`internal/store`): Database interactions.

## Protocol Logic

-   **OAuth2** logic resides strictly in `internal/oauth2`.
-   **OIDC** logic resides strictly in `internal/oidc`.
-   **Session** handling resides in `internal/session`.

## External Consumers

The core engine is consumed by the following physical planes:

| Plane | Repository | Relationship |
| :--- | :--- | :--- |
| **Auth Plane** | `github.com/opentrusty/opentrusty-auth` | OIDC/OAuth2 Protocol Gateway |
| **Admin Plane** | `github.com/opentrusty/opentrusty-admin` | Management API Gateway |
| **CLI Tools** | `github.com/opentrusty/opentrusty-cli` | Operator CLI (Migrate/Bootstrap) |
| **Control Panel** | `github.com/opentrusty/opentrusty-control-panel` | Administrative Frontend (Untrusted) |

**Critical Rule**: The `opentrusty-core` repository contains **ZERO transport logic** (no HTTP handlers, no CLI parsing).
All interface logic resides in the respective external repositories.

