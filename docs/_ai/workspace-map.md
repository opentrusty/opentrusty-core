# Workspace Map

Overview of the OpenTrusty multi-repository workspace.

## Repository Topology

| # | Repository | Role | Plane | Primary Tech | Binary |
|---|-----------|------|-------|-------------|--------|
| 1 | `opentrusty-core` | Pure Domain Library | — | Go | — |
| 2 | `opentrusty-auth` | Authentication Data Plane | `auth.*` | Go | `authd` |
| 3 | `opentrusty-admin` | Management Control Plane API | `api.*` | Go | `admind` |
| 4 | `opentrusty-cli` | Operator / Bootstrap Tooling | — | Go | `opentrusty` |
| 5 | `opentrusty-control-panel` | Administrative Web UI | `console.*` | React + TS | SPA |
| 6 | `opentrusty-demo-app` | E2E Test Client (RP Simulation) | External | Go | `demo-app` |

## Dependency Graph

```
opentrusty-auth ──────┐
opentrusty-admin ─────┼──→ opentrusty-core
opentrusty-cli ───────┘
opentrusty-control-panel ──→ opentrusty-admin (HTTP API only)
opentrusty-demo-app ───────→ opentrusty-auth (OIDC protocol only)
```

## Deployment Ownership

| Artifact | Owner Repo |
|----------|-----------|
| Database migrations | `opentrusty-cli` |
| Platform bootstrap | `opentrusty-cli` |
| systemd service units | `opentrusty-cli` |
| Caddyfile template | `opentrusty-cli` |
| One-click installer (`bootstrap.sh`) | `opentrusty-cli` |
| `.env.example` per binary | Each respective binary repo |

## Known Gaps

See [known-gaps.md](./known-gaps.md) for tracked deficiencies.
