# Identity Privacy & Data Minimization

> **Status**: Implemented (Phase 6.2)
> **Last Updated**: 2026-01-02

## Overview

OpenTrusty employs a **Global Privacy-Preserving Identity Model**. Users are identified globally by a secure hash of their email address (`email_hash`). The actual email address (`email_plain`) is treated as sensitive PII metadata, not as a lookup key.

## The Global Identity Model

### 1. Global Users (`users` table)
- **Primary Identifier**: `id` (UUIDv7)
- **Global Identity Key**: `email_hash` (CHAR(64) UNIQUE NOT NULL).
  - Derived via `HMAC-SHA256(OPENTRUSTY_IDENTITY_HMAC_KEY, Normalized_Email)`.
  - Used for all authentication lookups.
- **PII Metadata**: `email_plain` (TEXT, NULLABLE, No Index).
  - Stores the actual email address if provided.
  - NEVER used for database lookups (unindexed).
  - ONLY used for communication (sending emails) or display.
- **Scope**: Platform-wide.

### 2. Tenant Membership (`tenant_members` table)
- **Primary Identifier**: `id` (UUIDv7)
- **Link**: `tenant_id` + `user_id`.
- **Purpose**: Defines authorization context.
- **No Private Data**: Does NOT store email, fingerprint, or any PII. Simply links a global user ID to a tenant ID.

## Privacy & Security Guarantees

### 1. Protection Against Enumeration
By indexing only `email_hash` (HMAC) and relying on a server-side secret key:
- Attackers cannot verify if an email exists in the database without the secret key (mitigates timing/enumeration attacks if the key remains secure).
- Database dumps do not reveal cleartext emails immediately (requires brute-forcing the HMAC if key is compromised, or checking against `email_plain` if populated). Note: `email_plain` is present but unindexed, discouraging its misuse in queries.

### 2. Global Login Consistency
- Login is **ALWAYS Global**.
- The system computes the hash of the provided email and looks up the user.
- Tenant context is **strictly authorization**, not identification. Authenticate first (Global), then Authorize (Tenant).

### 3. Data Minimization
- The system prefers usage of `user_id` or `email_hash` for internal operations.
- `email_plain` is accessed only when necessary for user-facing interactions.

## Authentication Flows

### Login (Universal)
1. User provides `email` and `password`.
2. Backend computes `hash = HMAC-SHA256(GlobalKey, email)`.
3. Backend performs lookup: `SELECT * FROM users WHERE email_hash = ?`.
4. If found, verifies password (Argon2id).
5. If successful, session creates.

### Tenant Authorization
- After login, the user's session is established.
- Access to specific tenant resources requires a check against `tenant_members` (or RBAC assignments) for that `user_id` and `tenant_id`.

## Audit & Compliance
- **Logs**: Prefer `user_id` or `email_hash` in machine-readable log fields.
- **PII**: `email_plain` should only be logged in access-controlled audit trails where human readability is required and legally compliant.
