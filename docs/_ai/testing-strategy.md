# OpenTrusty Testing Strategy

This document defines the three-layer testing system for the OpenTrusty multi-repo architecture.

## Core Philosophy

Test hardening is built on three pillars: **Correctness, Isolation, and Verifiability**. 

1. **Internal Correctness**: Repositories must be correct in isolation.
2. **Plane Boundary Integrity**: Interfaces between planes (Real HTTP) must be validated.
3. **End-to-End User Journeys**: Full system verification through the browser.

---

## 🚦 Global Rules

- **NO in-memory cross-repo testing**: Testing `auth` by importing `core` packages is allowed (dependency), but testing `auth` by importing `admin` internal logic is forbidden.
- **NO shared mocks**: Each repo defines its own mocks for its dependencies.
- **Real HTTP for ST/E2E**: Service and E2E tests must bind to real network ports (localhost).
- **Control Panel is External**: The UI is treated as a black-box browser consumer.
- **Isolated Sessions**: Auth and Admin planes use different cookies and session namespaces.

---

## 🧱 Layer 1: Unit Tests (UT)

**Objective**: Validate pure logic, domain invariants, and mathematical correctness.

- **Mocking**: Use interfaces for repositories and external services.
- **No Side Effects**: No database access, no network calls, no disk I/O in UT.
- **Execution**: `make test-unit`

### Coverage Targets
- **Core**: RBAC (HasPermission), Policy evaluation, Email hashing.
- **Auth**: PKCE, OIDC Request validation, Token issuance rules.
- **Admin**: Tenant lifecycle validation, permission checks.

---

## 🔌 Layer 2: Service Tests (ST)

**Objective**: Validate a single plane/repository over a real HTTP boundary with a real database.

- **Infrastructure**: Real PostgreSQL (Dockerized), real HTTP server.
- **Isolation**: Each test suite resets the database state.
- **Execution**: `make test-service`

### Scenarios
- **Auth**: Full OIDC flow from `/authorize` to `/token`.
- **Admin**: Create tenant -> Verify audit log -> Update resource.

---

## 🌍 Layer 3: End-to-End Tests (E2E)

**Objective**: Prove the real user journey works through a real browser.

- **Tooling**: Playwright (Chromium).
- **Environment**: All services (`authd`, `admind`, `console`) must be running.
- **Execution**: `make test-e2e`

### Mandatory Journeys
1. **Bootstrap**: Fresh DB -> CLI Bootstrap -> Login to Admin UI.
2. **Provisioning**: Platform Admin creates Tenant -> Tenant Owner logs in.
3. **OIDC Integration**: Demo App -> Redirect to Auth -> Login -> Exchange Tokens.

---

## 📦 Directory Structure

Each repository follows this testing layout:

```text
internal/
  └── test/
      ├── unit/      # Unit tests (logic only)
      ├── service/   # Service tests (HTTP + DB)
      └── e2e/       # Local E2E tests (if applicable)
```

Global orchestration and cross-repo coordination is handled via `scripts/` or the root `Makefile` in the respective component repos.
