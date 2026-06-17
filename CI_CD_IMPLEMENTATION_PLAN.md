# CI/CD Implementation Plan
## IBM Storage Virtualize Snapshot Manager

**Date:** 2026-06-17  
**Status:** Planning Phase  
**Goal:** Implement automated testing, dependency checking, and container builds

---

## Overview

This plan outlines the implementation of a complete CI/CD pipeline for the IBM Storage Virtualize Snapshot Manager project, including:

1. **Automated Testing** - Run tests on every push and pull request
2. **Dependency Checking** - Weekly automated checks for outdated packages
3. **Container Builds** - Multi-architecture container images published to GitHub Container Registry
4. **Documentation** - Comprehensive guides for contributors and users

---

## Architecture Decisions

### Container Registry
- **Primary Registry:** GitHub Container Registry (ghcr.io)
- **Image Naming:** `ghcr.io/[org]/snapshot-manager`
- **Rationale:** Native GitHub integration, free for public repos, supports multi-arch

### Testing Strategy
- **Trigger:** Every push to any branch + all pull requests
- **Backend Tests:** Go unit tests with coverage reporting
- **Frontend Tests:** npm test (when implemented)
- **Integration Tests:** Manual trigger only (requires IBM SVC system)

### Dependency Management
- **Schedule:** Weekly automated checks (Sundays at 00:00 UTC)
- **Tools:** 
  - Go: `go list -u -m all` and `govulncheck`
  - npm: `npm outdated` and `npm audit`
- **Notifications:** Create GitHub issues for outdated/vulnerable dependencies

### Container Build Strategy
- **Architectures:** linux/amd64, linux/arm64
- **Build Tool:** Docker Buildx with QEMU
- **Caching:** GitHub Actions cache for faster builds
- **Tagging Strategy:**
  - `latest` - Latest commit on main branch
  - `v1.2.3` - Semantic version tags
  - `sha-abc123` - Commit SHA for traceability
  - `pr-123` - Pull request builds (not pushed to registry)

---

## Implementation Tasks

### Phase 1: GitHub Actions Workflows

#### 1.1 Main CI Workflow (`.github/workflows/ci.yml`)

**Purpose:** Run tests and linting on every push and PR

**Jobs:**
- **Backend Tests**
  - Setup Go 1.25
  - Cache Go modules
  - Run `go mod download`
  - Run `make lint` (fmt + vet)
  - Run `make test-coverage`
  - Upload coverage to Codecov (optional)
  - Generate coverage badge

- **Frontend Tests**
  - Setup Node.js 20
  - Cache npm dependencies
  - Run `npm ci`
  - Run `npm run lint`
  - Run `npm run build` (validates TypeScript)
  - Run `npm test` (when tests exist)

- **Security Scanning**
  - Run `govulncheck` for Go vulnerabilities
  - Run `npm audit` for npm vulnerabilities
  - Fail on high/critical vulnerabilities

**Triggers:**
```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
```

#### 1.2 Container Build Workflow (`.github/workflows/container-build.yml`)

**Purpose:** Build and publish multi-arch container images

**Jobs:**
- **Build Backend Image**
  - Setup Docker Buildx
  - Login to ghcr.io
  - Build for linux/amd64 and linux/arm64
  - Tag with version/SHA/latest
  - Push to registry (only on main branch)
  - Generate SBOM (Software Bill of Materials)

- **Build Frontend Image**
  - Same as backend
  - Separate image for frontend

- **Build Complete Stack**
  - Test podman-compose.yml
  - Validate all services start correctly
  - Run smoke tests

**Triggers:**
```yaml
on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:
```

#### 1.3 Dependency Check Workflow (`.github/workflows/dependency-check.yml`)

**Purpose:** Weekly automated dependency audits

**Jobs:**
- **Check Go Dependencies**
  - Run `go list -u -m all`
  - Parse outdated modules
  - Run `govulncheck`
  - Create issue if updates available

- **Check npm Dependencies**
  - Run `npm outdated --json`
  - Run `npm audit --json`
  - Create issue if updates available

- **Create Summary Report**
  - Aggregate all findings
  - Post to GitHub issue
  - Label with `dependencies`, `security`

**Triggers:**
```yaml
on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sunday
  workflow_dispatch:
```

#### 1.4 Release Workflow (`.github/workflows/release.yml`)

**Purpose:** Automated releases with changelog generation

**Jobs:**
- **Create Release**
  - Validate version tag format
  - Generate changelog from commits
  - Build release artifacts
  - Create GitHub release
  - Publish containers with version tag

**Triggers:**
```yaml
on:
  push:
    tags: [ 'v*.*.*' ]
```

---

### Phase 2: Enhanced Build Tools

#### 2.1 Backend Makefile Enhancements

Add new targets to [`backend/Makefile`](backend/Makefile):

```makefile
# Dependency management
deps-check: ## Check for outdated dependencies
	@echo "Checking Go dependencies..."
	@go list -u -m all | grep '\['

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-audit: ## Run security audit
	@echo "Running security audit..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Container builds
docker-build: ## Build Docker image
	docker build -t snapshot-manager-backend:latest .

docker-build-multiarch: ## Build multi-architecture image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-t snapshot-manager-backend:latest .

# CI targets
ci: lint test ## Run CI checks locally
	@echo "CI checks complete"

ci-full: clean deps-audit ci test-coverage ## Full CI pipeline
	@echo "Full CI pipeline complete"
```

#### 2.2 Frontend package.json Enhancements

Add new scripts to [`frontend/package.json`](frontend/package.json):

```json
{
  "scripts": {
    "deps:check": "npm outdated",
    "deps:audit": "npm audit",
    "deps:update": "npm update",
    "deps:update-interactive": "npx npm-check-updates -i",
    "test": "echo \"Tests not yet implemented\" && exit 0",
    "test:watch": "echo \"Tests not yet implemented\"",
    "ci": "npm run lint && npm run build",
    "docker:build": "docker build -t snapshot-manager-frontend:latest ."
  }
}
```

---

### Phase 3: Container Registry Setup

#### 3.1 GitHub Container Registry Configuration

**Repository Settings:**
1. Enable GitHub Packages
2. Set package visibility to public
3. Configure package permissions

**Required Secrets:**
- `GITHUB_TOKEN` (automatically provided)
- No additional secrets needed for ghcr.io

#### 3.2 Image Tagging Strategy

**Tag Format:**
```
ghcr.io/[org]/snapshot-manager-backend:latest
ghcr.io/[org]/snapshot-manager-backend:v1.2.3
ghcr.io/[org]/snapshot-manager-backend:sha-abc123
ghcr.io/[org]/snapshot-manager-frontend:latest
ghcr.io/[org]/snapshot-manager-frontend:v1.2.3
```

**Tagging Rules:**
- `latest` - Always points to the most recent main branch build
- `v*.*.*` - Semantic version from Git tags
- `sha-*` - Commit SHA for exact version tracking
- `pr-*` - Pull request builds (not pushed to registry)

#### 3.3 Multi-Architecture Support

**Supported Platforms:**
- `linux/amd64` - Intel/AMD 64-bit (primary)
- `linux/arm64` - ARM 64-bit (Apple Silicon, ARM servers)

**Build Configuration:**
```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    platforms: linux/amd64,linux/arm64
    push: ${{ github.event_name != 'pull_request' }}
    tags: |
      ghcr.io/${{ github.repository }}/snapshot-manager-backend:latest
      ghcr.io/${{ github.repository }}/snapshot-manager-backend:${{ github.sha }}
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

---

### Phase 4: Documentation

#### 4.1 CI/CD Documentation (`docs/CI_CD.md`)

**Contents:**
- Overview of CI/CD pipeline
- Workflow descriptions
- How to trigger manual builds
- Troubleshooting common issues
- Badge integration guide

#### 4.2 Container Usage Guide (`docs/CONTAINER_USAGE.md`)

**Contents:**
- Pulling images from ghcr.io
- Running containers locally
- Multi-architecture support
- Environment variables
- Volume mounts
- Networking configuration
- Upgrading containers

#### 4.3 Contributing Guide (`CONTRIBUTING.md`)

**Contents:**
- Development setup
- Running tests locally
- Code style guidelines
- PR requirements (tests must pass)
- Dependency update process
- Release process

#### 4.4 Update README.md

**Add Sections:**
- CI/CD status badges
- Container registry links
- Quick start with containers
- Link to container usage guide

---

### Phase 5: Status Badges

#### 5.1 Badges to Add

Add to [`README.md`](README.md):

```markdown
[![CI](https://github.com/[org]/ibm-storage-virtualize-snapshot-manager/workflows/CI/badge.svg)](https://github.com/[org]/ibm-storage-virtualize-snapshot-manager/actions/workflows/ci.yml)
[![Container Build](https://github.com/[org]/ibm-storage-virtualize-snapshot-manager/workflows/Container%20Build/badge.svg)](https://github.com/[org]/ibm-storage-virtualize-snapshot-manager/actions/workflows/container-build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/[org]/ibm-storage-virtualize-snapshot-manager)](https://goreportcard.com/report/github.com/[org]/ibm-storage-virtualize-snapshot-manager)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
```

---

## File Structure

```
.github/
├── workflows/
│   ├── ci.yml                      # Main CI workflow
│   ├── container-build.yml         # Container builds
│   ├── dependency-check.yml        # Weekly dependency checks
│   └── release.yml                 # Release automation
├── dependabot.yml                  # Dependabot configuration
└── ISSUE_TEMPLATE/
    └── dependency-update.md        # Template for dependency issues

docs/
├── CI_CD.md                        # CI/CD documentation
└── CONTAINER_USAGE.md              # Container usage guide

CONTRIBUTING.md                     # Contribution guidelines
```

---

## Implementation Order

### Week 1: Foundation
1. ✅ Create implementation plan (this document)
2. Create basic CI workflow (`.github/workflows/ci.yml`)
3. Enhance backend Makefile with new targets
4. Add frontend npm scripts for dependency management
5. Test CI workflow with sample PR

### Week 2: Container Builds
6. Create container build workflow (`.github/workflows/container-build.yml`)
7. Configure GitHub Container Registry
8. Test multi-arch builds
9. Implement tagging strategy
10. Create container usage documentation

### Week 3: Dependency Management
11. Create dependency check workflow (`.github/workflows/dependency-check.yml`)
12. Test weekly dependency checks
13. Configure Dependabot (optional)
14. Document dependency update process

### Week 4: Documentation & Polish
15. Create CI/CD documentation
16. Update README with badges and container info
17. Create CONTRIBUTING.md
18. Create release workflow
19. Test complete pipeline end-to-end

---

## Testing Strategy

### Local Testing

**Before pushing changes:**

```bash
# Backend
cd backend
make ci-full

# Frontend
cd frontend
npm run ci

# Container builds
docker buildx build --platform linux/amd64 -t test:latest ./backend
docker buildx build --platform linux/amd64 -t test:latest ./frontend
```

### CI Testing

**Automated on every push:**
- Go tests with coverage
- Frontend build validation
- Linting (Go fmt/vet, ESLint)
- Security scanning (govulncheck, npm audit)

**Manual triggers:**
- Container builds (via workflow_dispatch)
- Dependency checks (via workflow_dispatch)
- Release creation (via Git tags)

---

## Security Considerations

### Secrets Management
- Use GitHub Secrets for sensitive data
- Never commit credentials to repository
- Rotate secrets regularly

### Container Security
- Use minimal base images (Alpine)
- Run as non-root user
- Enable security scanning
- Generate SBOM for compliance

### Dependency Security
- Weekly vulnerability scans
- Automated security updates via Dependabot
- Fail builds on high/critical vulnerabilities

---

## Monitoring & Maintenance

### Weekly Tasks
- Review dependency check results
- Update outdated packages
- Review security advisories

### Monthly Tasks
- Review CI/CD performance
- Optimize build times
- Update documentation
- Clean up old container images

### Quarterly Tasks
- Review and update CI/CD strategy
- Evaluate new tools and practices
- Update dependencies to latest major versions

---

## Success Metrics

### CI/CD Performance
- **Build Time:** < 10 minutes for full pipeline
- **Test Coverage:** > 70% for backend
- **Container Build Time:** < 15 minutes for multi-arch

### Quality Metrics
- **Test Pass Rate:** > 95%
- **Security Vulnerabilities:** 0 high/critical
- **Outdated Dependencies:** < 10% of total

### User Experience
- **Container Pull Time:** < 2 minutes on 100Mbps
- **Documentation Clarity:** User feedback
- **Contributor Onboarding:** < 30 minutes to first PR

---

## Rollout Plan

### Phase 1: Soft Launch (Week 1-2)
- Implement basic CI workflow
- Test with development team
- Gather feedback
- Fix issues

### Phase 2: Container Builds (Week 3-4)
- Enable container builds
- Publish to ghcr.io
- Update documentation
- Announce to users

### Phase 3: Full Automation (Week 5-6)
- Enable all workflows
- Configure weekly checks
- Monitor performance
- Optimize as needed

### Phase 4: Maintenance Mode (Ongoing)
- Regular reviews
- Continuous improvements
- User support

---

## Risks & Mitigations

### Risk: Build Failures Block Development
**Mitigation:** 
- Allow manual override for urgent fixes
- Implement retry logic for flaky tests
- Maintain fast feedback loops

### Risk: Container Registry Costs
**Mitigation:**
- Use GitHub Container Registry (free for public repos)
- Implement image retention policies
- Clean up old images regularly

### Risk: Dependency Update Breaking Changes
**Mitigation:**
- Test updates in separate branch first
- Use semantic versioning
- Maintain changelog
- Rollback capability

### Risk: Multi-Arch Build Complexity
**Mitigation:**
- Start with amd64 only
- Add arm64 when stable
- Document platform-specific issues
- Provide fallback images

---

## Future Enhancements

### Short Term (3-6 months)
- Add integration tests with mock IBM SVC
- Implement code coverage tracking
- Add performance benchmarks
- Create demo environment

### Medium Term (6-12 months)
- Kubernetes deployment manifests
- Helm charts
- Prometheus metrics
- Grafana dashboards

### Long Term (12+ months)
- Multi-cloud container registry support
- Automated security patching
- AI-powered code review
- Advanced monitoring and alerting

---

## Resources

### Documentation
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Buildx Documentation](https://docs.docker.com/buildx/working-with-buildx/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

### Tools
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [npm-check-updates](https://www.npmjs.com/package/npm-check-updates)
- [Dependabot](https://docs.github.com/en/code-security/dependabot)

### Best Practices
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Container Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [CI/CD Best Practices](https://docs.github.com/en/actions/learn-github-actions/best-practices-for-github-actions)

---

## Approval & Sign-off

**Plan Created By:** AI Assistant  
**Date:** 2026-06-17  
**Status:** Awaiting Review

**Approval Required From:**
- [ ] Project Maintainer
- [ ] Development Team
- [ ] DevOps Team (if applicable)

**Next Steps:**
1. Review this plan
2. Provide feedback or approval
3. Begin implementation (switch to Code mode)

---

## Appendix A: Example Workflow Files

### A.1 Basic CI Workflow Structure

```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - name: Run tests
        run: |
          cd backend
          make test-coverage

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install and test
        run: |
          cd frontend
          npm ci
          npm run build
```

### A.2 Container Build Workflow Structure

```yaml
name: Container Build

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build-backend:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: ./backend
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ghcr.io/${{ github.repository }}/backend:latest
```

---

## Appendix B: Dependency Check Script

```bash
#!/bin/bash
# scripts/check-dependencies.sh

echo "=== Go Dependencies ==="
cd backend
go list -u -m all | grep '\['

echo ""
echo "=== npm Dependencies ==="
cd ../frontend
npm outdated

echo ""
echo "=== Security Audit ==="
cd ../backend
govulncheck ./...

cd ../frontend
npm audit
```

---

*End of Implementation Plan*