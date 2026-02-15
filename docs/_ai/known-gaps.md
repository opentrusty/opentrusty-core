# Known Gaps

Tracked deficiencies and deferred items for the OpenTrusty project.

## Resolved

- [x] Committed binaries in auth/cli repos
- [x] AI_CONTRACT defects (auth: missing deps/typo, control-panel: duplicate line)
- [x] Env contract contradiction (`shared.env` vs discrete DB fields)
- [x] Ghost `internal/` directories in core
- [x] Empty `platform/` directory in core
- [x] Dead handler/router code in auth
- [x] Deploy artifacts misplaced in core (moved to CLI)
- [x] Demo app PKCE using `math/rand` (migrated to `crypto/rand`)
- [x] Architecture docs referencing non-existent internal packages
- [x] System boundaries mixing core/auth/admin descriptions

## Open

### Critical
- [ ] Auth `config.go` does not call `os.Getenv` for `LogLevel`, `IdentitySecret`, `SessionSecret`, `BaseURL`, `CSRFEnabled` — reads struct fields but never populates them from environment
- [ ] `project/` package: `Project` struct defined in both `project/project.go` AND `policy/models.go` — needs deduplication
- [ ] Rate limiting not implemented in any plane middleware (mentioned in threat model)

### High
- [ ] No automated test suite (CI/CD with GitHub Actions)
- [ ] No linter configuration (`.golangci.yml`)
- [ ] OIDC discovery document format undocumented
- [ ] Test coverage minimal across all repos (especially `store/` layer)

### Medium
- [ ] No versioning strategy documented (semver tagging cadence)
- [ ] OpenAPI / Swagger specifications not generated
- [ ] Dependency version skew between admin/auth (pseudo-version) and cli (tagged)
- [ ] Empty docs subdirectories across admin, auth, cli repos

### Low / Deferred
- [ ] Docker deployment (systemd-only for now — by design decision)
- [ ] CSRF protection not verified in auth plane
