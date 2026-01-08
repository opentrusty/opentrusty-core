# Initialization Lifecycle

OpenTrusty enforces a strict separation between schema management, semantic initialization, and runtime operation.

## The Three Stages

### 1. Schema Migration (`migrate up`)
- **Action**: Applies SQL migrations to the database.
- **Rules**:
    - Focuses solely on table structures, indexes, and constraints.
    - MUST NOT create users, roles, or any business-level data.
    - Safe to run multiple times (idempotent at the schema level).

### 2. Semantic Initialization (`bootstrap`)
- **Action**: Initializes the system's first administrative identity and core RBAC data.
- **Rules**:
    - Creates the first `platform_admin`.
    - MUST BE idempotent.
    - Refuses to execute if a `platform_admin` already exists (`ErrAlreadyBootstrapped`).
    - MUST emit audit events for the bootstrap action.
    - Use explicit environment variables for credentials:
        - `OPENTRUSTY_BOOTSTRAP_ADMIN_EMAIL`
        - `OPENTRUSTY_BOOTSTRAP_ADMIN_PASSWORD`

### 3. Runtime Operation (`serve`)
- **Action**: Starts the HTTP server (auth, admin, or all).
- **Rules**:
    - **Fail-Fast**: MUST verify that the system is bootstrapped before starting. If not, the process exits with an error.
    - MUST NOT perform schema changes or user creation.
    - Assumes a ready and initialized environment.

## Design Rationale

- **Security**: Prevents accidental user creation or schema changes in production runtimes.
- **Predictability**: Deployment pipelines can clearly separate infra-level changes from application-level logic.
- **Observability**: Every state-changing initialization step is tracked and auditable.
