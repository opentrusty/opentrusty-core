# AI_CONTRACT — opentrusty-core

## Scope of Responsibility
- Domain authority (User, Tenant, Session, Audit models).
- Cryptographic primitives and security logic (Argon2id, Session validation).
- Repository interfaces for persistence.
- Platform-wide architectural invariants.

## Explicit Non-Goals
- **NO HTTP**: This repository must not contain any web handlers or router logic.
- **NO CLI**: This repository must not contain CLI parsing or command definitions.
- **NO Database Migrations**: Schema ownership resides in the CLI repository.

## Allowed Dependencies
- Standard Library.
- Infrastructure drivers (e.g., pgx) via internal store implementations.

## Forbidden Dependencies
- **NO dependencies** on other opentrusty repositories (auth, admin, cli, ui). Core is the leaf.

## Change Discipline
- Any change to domain models or security primitives MUST update docs/_ai/invariants.md.
- Modification of repository interfaces requires verification in all consuming repositories.

## Invariants
- **Argon2id ONLY** for password hashing.
- **Server-Side Sessions ONLY** for user state.
- **Tenant Isolation** is first-class; no implicit global fallback.
