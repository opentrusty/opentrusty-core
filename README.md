# OpenTrusty Core

OpenTrusty Core is the **Domain Kernel** and **Security Kernel** of the OpenTrusty Identity Platform.

It contains the foundational domain models, security primitives, and repository interfaces that govern the entire system.

## Role & Responsibility

- **Authority**: Defines the canonical models for Tenants, Users, Roles, Clients, and Sessions.
- **Security**: Houses core cryptographic primitives and password hashing (Argon2id).
- **Persistence**: Contains the PostgreSQL implementation of repository interfaces (to be shared across planes).
- **No Side Effects**: Core is a pure library. It has NO HTTP listeners, NO CLI entrypoints, and NO UI code.

## Package Structure

- `user/`: Global identity models and service.
- `tenant/`: Multi-tenancy and membership models.
- `role/`: RBAC model and assignment interfaces.
- `policy/`: Authorization policy definitions.
- `client/`: OAuth2/OIDC client metadata.
- `session/`: Persistent session models.
- `store/`: Concrete persistence implementations (Postgres).
- `crypto/`: Cryptographic primitives for token signing and encryption.

## Getting Started

This repository is intended to be used as a Go module dependency for other OpenTrusty repositories.

```bash
go get github.com/opentrusty/opentrusty-core
```

## License

MIT
