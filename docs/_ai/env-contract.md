# OpenTrusty Environment Variable Contract

This document defines the mandatory environment variable conventions and requirements for all OpenTrusty repositories.

## 🏷️ Naming Convention

All runtime environment variables MUST use the prefix **OPENTRUSTY_**.

> [!IMPORTANT]
> Generic names like `PORT`, `DATABASE_URL`, or `SECRET` are forbidden to avoid namespace collisions in shared environments or containers.

---

## 🛠️ Shared Variables

| Variable | Description | Default | Consumption |
| :--- | :--- | :--- | :--- |
| `OPENTRUSTY_DB_URL` | PostgreSQL connection string | `postgres://...` | Auth, Admin, CLI |
| `OPENTRUSTY_PORT` | HTTP listener port | See repo docs | Auth, Admin |
| `OPENTRUSTY_LOG_LEVEL` | Logging verbosity (debug, info, error) | `info` | All |
| `OPENTRUSTY_IDENTITY_SECRET` | Shared HMAC key for PII hashing (MANDATORY in prod) | - | All |

---

## 🛡️ Security Requirements

1. **No Hardcoding**: Secrets (keys, passwords) must never be hardcoded in `main.go` or templates.
2. **Environment Files**: For production (`systemd`, `docker`), use dedicated environment files (`EnvironmentFile=` or `--env-file`).
3. **No Sharing**: Each plane (Auth, Admin) must use its own distinct environment file to maintain plane isolation.
4. **Data Minimization**: Do not pass unused variables to a process.

---

## 🏗️ Repository consumption

### Auth Plane (`authd`)
- `OPENTRUSTY_DB_URL`
- `OPENTRUSTY_PORT` (Default: 8080)
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_AUTH_SIGNING_KEY`

### Admin Plane (`admind`)
- `OPENTRUSTY_DB_URL`
- `OPENTRUSTY_PORT` (Default: 8081)
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_ADMIN_SIGNING_KEY`

### CLI (`opentrusty`)
- `OPENTRUSTY_DB_URL`
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_BOOTSTRAP_ADMIN_EMAIL`
- `OPENTRUSTY_BOOTSTRAP_ADMIN_PASSWORD`

---

## 👑 Platform Admin Semantics

1. **No `tenant_id`**: Platform administrators NEVER belong to a specific tenant.
2. **Global Authority**:
   - `scope = platform`
   - `scope_context_id = NULL`
3. **Implicit vs Explicit**: Authorization checks for platform features must explicitly check for `scope = platform`.
4. **No Platform ID**: The system does NOT use a `platform_id` or `system_id` UUID to represent the platform; authority is derived solely from the `scope = platform` assignment.
