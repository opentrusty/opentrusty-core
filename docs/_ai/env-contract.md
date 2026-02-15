# OpenTrusty Environment Variable Contract

This document defines the mandatory environment variable conventions and requirements for all OpenTrusty repositories.

## 🏷️ Naming Convention

All runtime environment variables MUST use the prefix **OPENTRUSTY_**.

> [!IMPORTANT]
> Generic names like `PORT`, `DATABASE_URL`, or `SECRET` are forbidden to avoid namespace collisions in shared environments or containers.

---

## 🗄️ Database Configuration (Discrete Fields)

All database-consuming binaries MUST use discrete fields. Connection string format (`OPENTRUSTY_DATABASE_URL`) is **deprecated**.

| Variable | Description | Default | Required |
| :--- | :--- | :--- | :--- |
| `OPENTRUSTY_DB_HOST` | PostgreSQL host | — | ✅ |
| `OPENTRUSTY_DB_PORT` | PostgreSQL port | `5432` | — |
| `OPENTRUSTY_DB_USER` | PostgreSQL user | — | ✅ |
| `OPENTRUSTY_DB_PASSWORD` | PostgreSQL password | — | ✅ (prod) |
| `OPENTRUSTY_DB_NAME` | PostgreSQL database name | — | ✅ |
| `OPENTRUSTY_DB_SSLMODE` | SSL mode (`disable`, `require`, `verify-full`) | `disable` | — |

---

## 🔐 Security Variables

| Variable | Description | Consumption |
| :--- | :--- | :--- |
| `OPENTRUSTY_IDENTITY_SECRET` | Shared HMAC key for PII hashing (MANDATORY in prod) | All DB-consuming binaries |
| `OPENTRUSTY_SESSION_SECRET` | Secret for session management | Auth, Admin |

---

## 🛡️ Security Requirements

1. **No Hardcoding**: Secrets (keys, passwords) must never be hardcoded in source code.
2. **Per-Binary `.env`**: Each binary owns its own `.env.example` and deployed `.env` file. No `shared.env`.
3. **Environment Files**: For production (`systemd`), use dedicated environment files (`EnvironmentFile=`).
4. **Data Minimization**: Do not pass unused variables to a process.

---

## 🏗️ Per-Binary Configuration

### Auth Plane (`authd`)
- `OPENTRUSTY_ENV` (Default: `dev`)
- `OPENTRUSTY_AUTH_LISTEN_ADDR` (Default: `:8080`)
- `OPENTRUSTY_LOG_LEVEL` (Default: `info`)
- `OPENTRUSTY_BASE_URL`
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_SESSION_SECRET`
- `OPENTRUSTY_AUTH_SESSION_NAMESPACE` (Default: `auth`)
- `OPENTRUSTY_AUTH_CSRF_ENABLED`
- `OPENTRUSTY_COOKIE_SECURE`, `OPENTRUSTY_COOKIE_HTTPONLY`, `OPENTRUSTY_COOKIE_SAMESITE`, `OPENTRUSTY_COOKIE_DOMAIN`, `OPENTRUSTY_COOKIE_NAME`
- `OPENTRUSTY_DB_HOST`, `OPENTRUSTY_DB_PORT`, `OPENTRUSTY_DB_USER`, `OPENTRUSTY_DB_PASSWORD`, `OPENTRUSTY_DB_NAME`, `OPENTRUSTY_DB_SSLMODE`

### Admin Plane (`admind`)
- `OPENTRUSTY_ENV` (Default: `dev`)
- `OPENTRUSTY_ADMIN_LISTEN_ADDR` (Default: `:8081`)
- `OPENTRUSTY_LOG_LEVEL` (Default: `info`)
- `OPENTRUSTY_BASE_URL`
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_SESSION_SECRET`
- `OPENTRUSTY_ADMIN_SESSION_NAMESPACE` (Default: `admin`)
- `OPENTRUSTY_COOKIE_SECURE`, `OPENTRUSTY_COOKIE_HTTPONLY`, `OPENTRUSTY_COOKIE_SAMESITE`, `OPENTRUSTY_COOKIE_DOMAIN`, `OPENTRUSTY_COOKIE_NAME`
- `OPENTRUSTY_DB_HOST`, `OPENTRUSTY_DB_PORT`, `OPENTRUSTY_DB_USER`, `OPENTRUSTY_DB_PASSWORD`, `OPENTRUSTY_DB_NAME`, `OPENTRUSTY_DB_SSLMODE`

### CLI (`opentrusty`)
- `OPENTRUSTY_LOG_LEVEL` (Default: `info`)
- `OPENTRUSTY_IDENTITY_SECRET`
- `OPENTRUSTY_BOOTSTRAP_ADMIN_EMAIL`
- `OPENTRUSTY_BOOTSTRAP_ADMIN_PASSWORD`
- `OPENTRUSTY_DB_HOST`, `OPENTRUSTY_DB_PORT`, `OPENTRUSTY_DB_USER`, `OPENTRUSTY_DB_PASSWORD`, `OPENTRUSTY_DB_NAME`, `OPENTRUSTY_DB_SSLMODE`

---

## 👑 Platform Admin Semantics

1. **No `tenant_id`**: Platform administrators NEVER belong to a specific tenant.
2. **Global Authority**:
   - `scope = platform`
   - `scope_context_id = NULL`
3. **Implicit vs Explicit**: Authorization checks for platform features must explicitly check for `scope = platform`.
4. **No Platform ID**: The system does NOT use a `platform_id` or `system_id` UUID to represent the platform; authority is derived solely from the `scope = platform` assignment.
