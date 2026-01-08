# Interface Documentation Standard

This document defines the standard for documenting public interfaces (structs, interfaces, functions, methods) in OpenTrusty Go repositories. This is a security-critical requirement to ensure system determinism, auditability, and correct protocol implementation.

## Context

OpenTrusty follows a strict **Identity-as-a-Kernel** and **Plane Separation** architecture. Documentation must clearly identify the domain and security implications of every public element.

## Annotation Format

All public interfaces MUST be preceded by a structured doc comment block.

### 1. API Handlers (OpenAPI/swag)

Functions serving as HTTP handlers MUST include `swag` annotations for OpenAPI specification generation.

```go
// Name [A concise single-line description]
// @Summary [Friendly name for the operation]
// @Description [Detailed behavior and constraints]
// @Tags [Component or domain classification]
// @Accept json
// @Produce json
// @Param name type format requirement "description"
// @Success code {type} description
// @Failure code {type} description
// @Router /path [method]
func (h *Handler) Name(w http.ResponseWriter, r *http.Request) { ... }
```

### 2. Services and Logic Functions

Functions and methods containing business logic MUST include domain and security metadata.

```go
// Name [A concise single-line description]
//
// Purpose: [Contextual explanation of why this exists]
// Domain: [Identity | Tenant | Authz | Session | Client | Audit | Platform]
// Security: [Security considerations, invariants enforced, or potential risks]
// Audited: [Yes/No - does this function emit an audit event?]
// Errors: [List of specific domain errors that may be returned]
func Name(args...) (returns...)
```

### 3. Structs and Interfaces

```go
// Name [Concise description of the entity or interface]
//
// Purpose: [Description of the role this plays in the system]
// Domain: [Identity | Tenant | Authz | Session | Client | Audit | Platform]
// Invariants: [Structural or business invariants this type MUST satisfy]
type Name struct { ... }
```

## Mandatory Fields

| Field | Requirement | Description |
| :--- | :--- | :--- |
| **Purpose** | Mandatory | Explains the "Why" and "Where" this fits in the architecture. |
| **Domain** | Mandatory | Maps the element to one of the canonical system domains. |
| **Security** | Mandatory | Explicitly documents the security boundary or invariant being enforced. |
| **Audited** | Functions Only | Clear indicator for compliance and audit trail verification. |
| **Errors** | Functions Only | Documents expected error states, aiding integrator error prevention. |
| **Invariants** | Types Only | Documents the constraints that must hold true for the object to be valid. |

## Examples

### Good API Handler Documentation

```go
// Login handles user login
// @Summary Login
// @Description Authenticate admin user and create a session (tenant derived from user record)
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string "non-admin user"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) { ... }
```

### Good Service Documentation

```go
// Authenticate verifies a user's credentials and returns the Identity.
//
// Purpose: Entry point for user authentication across all planes.
// Domain: Identity
// Security: Enforces Argon2id password verification and lockout policies. Redacts secrets from logs.
// Audited: Yes (audit.TypeLoginSuccess or audit.TypeLoginFailed)
// Errors: ErrInvalidCredentials, ErrAccountLocked, ErrUserNotFound
func (s *Service) Authenticate(ctx context.Context, email, password string) (*User, error) { ... }
```

### Good Struct Documentation

```go
// User represents a persistent digital representation of an actor.
//
// Purpose: The core identity entity within OpenTrusty.
// Domain: Identity
// Invariants: ID must be a UUIDv7. EmailHash must be a valid HMAC-SHA256 of the normalized email.
type User struct { ... }
```

## Enforcement

-   This standard is a **Beta Gate** requirement.
-   CI will eventually enforce mandatory fields via custom linting.
-   Maintainers will reject PRs with "incomplete or semantically inaccurate" documentation.
