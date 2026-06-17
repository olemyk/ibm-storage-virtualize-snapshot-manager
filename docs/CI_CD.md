# CI/CD Pipeline Documentation

## Overview

The IBM Storage Virtualize Snapshot Manager uses GitHub Actions for continuous integration and deployment. This document describes the CI/CD pipeline, workflows, and how to use them.

## Workflows

### 1. CI Workflow (`.github/workflows/ci.yml`)

**Purpose:** Automated testing and quality checks on every push and pull request.

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**

#### Backend Tests
- Sets up Go 1.25
- Runs `go fmt` to check code formatting
- Runs `go vet` for static analysis
- Executes all tests with coverage reporting
- Uploads coverage to Codecov (optional)

#### Backend Security
- Installs and runs `govulncheck`
- Scans for known vulnerabilities in dependencies
- Fails build if vulnerabilities found

#### Frontend Build & Lint
- Sets up Node.js 20
- Installs dependencies with `npm ci`
- Runs ESLint for code quality
- Builds the frontend application
- Checks for TypeScript errors

#### Frontend Security
- Runs `npm audit` for security vulnerabilities
- Reports high and critical vulnerabilities

**Status:** ✅ Required for merging PRs

### 2. Container Build Workflow (`.github/workflows/container-build.yml`)

**Purpose:** Build and publish multi-architecture container images.

**Triggers:**
- Push to `main` branch
- Version tags (`v*.*.*`)
- Pull requests (build only, no push)
- Manual trigger via workflow_dispatch

**Jobs:**

#### Build Backend Container
- Builds for `linux/amd64` and `linux/arm64`
- Pushes to GitHub Container Registry (ghcr.io)
- Tags: `latest`, `v1.2.3`, `sha-abc123`

#### Build Frontend Container
- Builds for `linux/amd64` and `linux/arm64`
- Pushes to GitHub Container Registry (ghcr.io)
- Tags: `latest`, `v1.2.3`, `sha-abc123`

#### Test Compose
- Validates `podman-compose.yml`
- Ensures all services can start (PR only)

**Container Images:**
```
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest
```

### 3. Dependency Check Workflow (`.github/workflows/dependency-check.yml`)

**Purpose:** Weekly automated dependency audits.

**Triggers:**
- Schedule: Every Sunday at 00:00 UTC
- Manual trigger via workflow_dispatch

**Jobs:**

#### Check Go Dependencies
- Lists outdated Go modules
- Runs `govulncheck` for security vulnerabilities
- Generates dependency report

#### Check npm Dependencies
- Lists outdated npm packages
- Runs `npm audit` for security issues
- Generates dependency report

#### Create Issue
- Creates or updates a GitHub issue with findings
- Labels: `dependencies`, `automated`, `maintenance`
- Provides update instructions

**Example Issue:** [Weekly Dependency Check - 2026-06-17](#)

### 4. Release Workflow (`.github/workflows/release.yml`)

**Purpose:** Automated release creation and artifact publishing.

**Triggers:**
- Push of version tags (`v1.2.3`)

**Jobs:**

#### Validate Release
- Validates tag format (must be `v*.*.*`)
- Ensures tag matches semantic versioning

#### Build Artifacts
- Builds binaries for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64)
- Generates SHA256 checksums

#### Build Containers
- Builds and pushes versioned container images
- Tags with version number and `latest`

#### Create Release
- Generates changelog from git commits
- Creates GitHub release with notes
- Attaches binary artifacts
- Links to container images

**Release Assets:**
- `snapshot-manager-linux-amd64`
- `snapshot-manager-linux-arm64`
- `snapshot-manager-darwin-amd64`
- `snapshot-manager-darwin-arm64`
- `snapshot-manager-windows-amd64.exe`
- `checksums.txt`

## Using the CI/CD Pipeline

### Running Tests Locally

Before pushing code, run tests locally to catch issues early:

**Backend:**
```bash
cd backend
make ci-full  # Run full CI pipeline locally
```

**Frontend:**
```bash
cd frontend
npm run ci  # Run linting and build
```

### Triggering Manual Builds

Some workflows can be triggered manually:

1. Go to **Actions** tab in GitHub
2. Select the workflow (e.g., "Container Build")
3. Click **Run workflow**
4. Select branch and click **Run workflow**

### Creating a Release

To create a new release:

1. Update version in relevant files
2. Commit changes
3. Create and push a version tag:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
4. GitHub Actions will automatically:
   - Build binaries
   - Build containers
   - Create GitHub release
   - Publish artifacts

### Checking Workflow Status

**View all workflows:**
- Go to repository → **Actions** tab

**Check specific run:**
- Click on workflow name
- Click on specific run
- View job details and logs

**Status badges:**
- See README.md for current status

## Container Registry

### Pulling Images

**Latest version:**
```bash
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest
```

**Specific version:**
```bash
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:1.2.3
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:1.2.3
```

**Specific architecture:**
```bash
docker pull --platform linux/arm64 ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

### Authentication

For private repositories, authenticate with GitHub:

```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

## Dependency Management

### Checking for Updates

**Backend:**
```bash
cd backend
make deps-check  # List outdated dependencies
make deps-audit  # Check for vulnerabilities
```

**Frontend:**
```bash
cd frontend
npm run deps:check  # List outdated packages
npm run deps:audit  # Check for vulnerabilities
```

### Updating Dependencies

**Backend:**
```bash
cd backend
make deps-update  # Update all dependencies
make test        # Verify updates work
```

**Frontend:**
```bash
cd frontend
npm run deps:update              # Update within semver ranges
npm run deps:update-interactive  # Interactive major updates
npm test                         # Verify updates work
```

### Security Fixes

**Backend:**
```bash
cd backend
make deps-audit  # Identify vulnerabilities
# Manually update vulnerable packages
go get -u github.com/vulnerable/package@latest
go mod tidy
make test
```

**Frontend:**
```bash
cd frontend
npm audit fix           # Automatic fixes
npm audit fix --force   # Force major version updates
npm test
```

## Troubleshooting

### CI Failures

**Test failures:**
1. Check test logs in GitHub Actions
2. Run tests locally: `make test` or `npm test`
3. Fix failing tests
4. Push changes

**Linting failures:**
1. Run linter locally: `make lint` or `npm run lint`
2. Fix issues automatically: `make fmt` or `npm run lint --fix`
3. Push changes

**Build failures:**
1. Check build logs
2. Verify dependencies: `go mod download` or `npm ci`
3. Build locally: `make build` or `npm run build`
4. Fix issues and push

### Container Build Failures

**Multi-arch build issues:**
- Ensure QEMU is set up correctly
- Check platform-specific dependencies
- Test locally with Docker Buildx

**Registry authentication:**
- Verify `GITHUB_TOKEN` has package write permissions
- Check repository settings → Actions → General → Workflow permissions

**Image size issues:**
- Review Dockerfile for unnecessary files
- Use `.dockerignore` to exclude files
- Consider multi-stage builds

### Dependency Check Issues

**False positives:**
- Review vulnerability details
- Check if vulnerability applies to your usage
- Document exceptions if needed

**Update conflicts:**
- Check for breaking changes in changelogs
- Update code to match new API
- Run full test suite

## Best Practices

### Before Committing

1. Run tests locally: `make ci-full` or `npm run ci`
2. Check code formatting: `make fmt` or `npm run lint`
3. Review changes: `git diff`
4. Write clear commit messages

### Pull Requests

1. Ensure all CI checks pass
2. Request reviews from maintainers
3. Address review feedback
4. Keep PRs focused and small
5. Update documentation if needed

### Releases

1. Update CHANGELOG.md
2. Update version numbers
3. Test thoroughly before tagging
4. Use semantic versioning
5. Write clear release notes

### Security

1. Never commit secrets or credentials
2. Use GitHub Secrets for sensitive data
3. Keep dependencies up to date
4. Review security advisories regularly
5. Enable Dependabot alerts

## Monitoring

### Workflow Metrics

Monitor these metrics for CI/CD health:

- **Build time:** Should be < 10 minutes
- **Test pass rate:** Should be > 95%
- **Container build time:** Should be < 15 minutes
- **Dependency updates:** Review weekly

### Alerts

Set up notifications for:
- Failed CI runs
- Security vulnerabilities
- Dependency updates
- Release failures

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Buildx Documentation](https://docs.docker.com/buildx/)
- [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Semantic Versioning](https://semver.org/)

## Support

For CI/CD issues:
1. Check workflow logs in GitHub Actions
2. Review this documentation
3. Search existing GitHub issues
4. Create a new issue with:
   - Workflow name
   - Run ID
   - Error messages
   - Steps to reproduce

---

*Last updated: 2026-06-17*