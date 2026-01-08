# OpenTrusty Master Test Plan

This README provides the high-level architecture of the test system.

## Test Layers

| Layer | Repo | Focus | Trigger |
| :--- | :--- | :--- | :--- |
| **UT** | Individual | Local Logic | `make test-unit` |
| **ST** | Individual | API + DB | `make test-service` |
| **E2E** | System | Browser Flows | `make test-e2e` |

## Repository Quick Links

- [opentrusty-core tests](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty-core/internal/test)
- [opentrusty-auth tests](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty-auth/internal/test)
- [opentrusty-admin tests](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty-admin/internal/test)
- [opentrusty-cli tests](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty-cli/internal/test)

## Global Constraints

- **Real HTTP Only**: No GIN/Fiber test-performers calling handlers in-memory.
- **Docker PG**: Use `docker compose up -d postgres` before running ST/E2E.
- **Playwright**: Installed via `npm install` in `opentrusty-control-panel`.
