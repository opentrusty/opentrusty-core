# Known Gaps & Deferred Cleanups

The following items are identified as missing or incomplete following the repository split.

## 1. Automated Testing
- [ ] **opentrusty-core**: Port domain and repository unit tests.
- [ ] **opentrusty-auth**: Implement OIDC protocol compliance tests.
- [ ] **opentrusty-admin**: Implement API integration tests.
- [ ] **opentrusty-cli**: Implement migration rollback tests.

## 2. Infrastructure & Tooling
- [ ] **Linter**: Add `.golangci.yml` to each repository for strict linting.
- [ ] **CI/CD**: Configure GitHub Actions for per-repo builds and tests.
- [ ] **Versioning**: Establish a semantic versioning (SemVer) strategy for the core module.

## 3. Documentation
- [ ] **OpenAPI**: Finalize and place `openapi.yaml` in `opentrusty-admin/docs/api/`.
- [ ] **Deployment**: Create standard Dockerfiles for `authd` and `admind`.
- [ ] **Examples**: Update the demo application to use the new multi-repo setup.
