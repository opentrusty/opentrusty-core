# Data Minimization Policy

> **Status**: Phase 6.2 planning document (not yet implemented)

## Core Principles

OpenTrusty stores the **minimum identity attributes** required for protocol correctness and administrative usability.

## Email Identity Model

### Mandatory: Tenant-Scoped HMAC

```
email_hmac = HMAC-SHA256(tenant_secret, lowercase(email))
```

- Used for: uniqueness, login lookup, identity binding
- **NEVER** exposed via API, logs, or UI
- Non-linkable across tenants (tenant-specific key)

### Optional: Encrypted Email

- Stored only if tenant policy enables
- Encrypted at rest (AES-GCM)
- **Only** for: login UX convenience, OIDC `email` claim release
- **NOT** required for protocol correctness

## Key Invariant

> **"Email collection is optional; email_hmac is mandatory."**

## Human Identification

- UI uses `display_name`, not email
- Email is **never** treated as a display identifier
- Users without stored email remain fully functional

## Scope vs. Storage

| Scope | Controls |
|-------|----------|
| `email` | Claim **release** in id_token |
| Tenant policy | Email **storage** |

These are independent. A client may request `email` scope, but if tenant policy disables email storage, no email claim is released.

## Legal Alignment

This design implements GDPR Art. 5(1)(c) data minimization.
