# Testing Strategy

This document describes the testing architecture and strategy for OpenTrusty.

## Three-Layer Testing Model

### 1. Unit Tests
- **Location**: Co-located with source code using standard Go conventions (`*_test.go` alongside implementation files)
- **Pattern**: `package_name_test` for black-box testing of exported API
- **Scope**: Single function or method, no I/O, no database
- **Coverage Target**: ≥ 80% for security-critical packages (`password/`, `crypto/`, `session/`, `authz/`)

### 2. Service / Integration Tests
- **Location**: Co-located within each package (e.g., `store/postgres/user_repository_test.go`)
- **Scope**: Tests that interact with real PostgreSQL via test containers or dedicated test database
- **Guard**: Build tag `//go:build integration` to separate from unit tests

### 3. End-to-End (E2E) Tests
- **Location**: `opentrusty-demo-app` acts as the E2E test harness (Relying Party simulation)
- **Scope**: Full OIDC Authorization Code + PKCE flow across Auth Plane → Admin Plane → DB
- **Guard**: Requires a fully running stack

## Test Execution

```bash
# Unit tests only (fast, no dependencies)
go test ./... -short

# Integration tests (requires PostgreSQL)
go test ./... -tags integration

# E2E tests (requires running stack)
# See opentrusty-demo-app README
```

## Invariants

1. **No test should mutate shared state** — each test uses isolated contexts
2. **Security-critical code must never be tested with mocks** — real Argon2id, real HMAC
3. **Benchmark tests** required for `password/argon2.go` and `crypto/hashing.go`
4. **Race detection** must pass: `go test -race ./...`
