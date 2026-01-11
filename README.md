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

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/opentrusty/opentrusty-core/main/scripts/bootstrap.sh | sudo bash
```

See the [Deployment Guide](DEPLOYMENT.md) for detailed configuration and manual installation steps.

## License

Apache-2.0
