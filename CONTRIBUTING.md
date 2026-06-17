# Contributing to IBM Storage Virtualize Snapshot Manager

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [CI/CD Pipeline](#cicd-pipeline)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

This project follows a standard code of conduct. Please be respectful and professional in all interactions.

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Node.js 20 or higher
- Git
- Docker or Podman (for container testing)
- IBM Storage Virtualize system (for integration testing)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/ibm-storage-virtualize-snapshot-manager.git
   cd ibm-storage-virtualize-snapshot-manager
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/ORIGINAL-OWNER/ibm-storage-virtualize-snapshot-manager.git
   ```

## Development Setup

### Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Install development tools
make install-tools

# Create environment file
cp .env.example .env
# Edit .env with your configuration

# Initialize database
go run scripts/init_db.go

# Run development server
make dev
```

### Frontend Setup

```bash
cd frontend

# Install dependencies
npm ci

# Run development server
npm run dev
```

### Full Stack Development

Use the provided script for easy development:

```bash
./dev-start.sh
```

This starts both backend and frontend in development mode.

## Making Changes

### Branch Naming

Create a descriptive branch name:

- `feature/add-snapshot-filtering` - New features
- `fix/token-expiry-bug` - Bug fixes
- `docs/update-api-guide` - Documentation
- `refactor/simplify-auth` - Code refactoring
- `test/add-scheduler-tests` - Test additions

```bash
git checkout -b feature/your-feature-name
```

### Commit Messages

Follow conventional commit format:

```
type(scope): subject

body (optional)

footer (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(scheduler): add support for custom snapshot naming

fix(auth): resolve token expiry calculation bug

docs(api): update authentication endpoint documentation

test(svc): add unit tests for IBM SVC client
```

### Keep Changes Focused

- One feature/fix per pull request
- Keep PRs small and reviewable
- Split large changes into multiple PRs

## Testing

### Running Tests Locally

**Before committing, always run:**

```bash
# Backend
cd backend
make ci-full  # Runs lint, test, and coverage

# Frontend
cd frontend
npm run ci  # Runs lint and build
```

### Writing Tests

#### Backend Tests (Go)

Place tests in `*_test.go` files next to the code:

```go
// internal/api/helpers_test.go
package api

import "testing"

func TestRespondJSON(t *testing.T) {
    tests := []struct {
        name string
        // test cases
    }{
        // ...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

**Run specific tests:**
```bash
go test -v ./internal/api/...
```

#### Frontend Tests (TypeScript)

Tests will be added in future. For now, ensure:
- Code builds without errors: `npm run build`
- No linting errors: `npm run lint`
- TypeScript compiles: `npx tsc --noEmit`

### Test Coverage

Aim for:
- **Backend:** > 70% coverage
- **Frontend:** > 60% coverage (when tests are implemented)

Check coverage:
```bash
cd backend
make test-coverage
open coverage.html
```

## Submitting Changes

### Before Submitting

1. **Update from upstream:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests:**
   ```bash
   cd backend && make ci-full
   cd ../frontend && npm run ci
   ```

3. **Check for security issues:**
   ```bash
   cd backend && make deps-audit
   cd ../frontend && npm audit
   ```

4. **Update documentation** if needed

5. **Add/update tests** for your changes

### Creating a Pull Request

1. Push your branch:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Go to GitHub and create a Pull Request

3. Fill out the PR template:
   - **Title:** Clear, descriptive title
   - **Description:** What changes were made and why
   - **Testing:** How you tested the changes
   - **Screenshots:** If UI changes
   - **Related Issues:** Link to related issues

4. Ensure CI checks pass:
   - ✅ Backend tests
   - ✅ Frontend build
   - ✅ Security scans
   - ✅ Linting

### PR Review Process

1. Maintainers will review your PR
2. Address feedback by pushing new commits
3. Once approved, maintainers will merge

**Tips:**
- Respond to feedback promptly
- Be open to suggestions
- Ask questions if unclear
- Keep discussions professional

## CI/CD Pipeline

### Automated Checks

Every PR triggers:

1. **Backend Tests**
   - Go fmt check
   - Go vet static analysis
   - Unit tests with coverage
   - Security vulnerability scan

2. **Frontend Build**
   - ESLint checks
   - TypeScript compilation
   - Production build
   - npm audit

3. **Container Build** (on main branch)
   - Multi-architecture builds
   - Push to GitHub Container Registry

### Manual Workflow Triggers

Some workflows can be triggered manually:

1. Go to **Actions** tab
2. Select workflow
3. Click **Run workflow**

### CI Failures

If CI fails:

1. **Check the logs** in GitHub Actions
2. **Reproduce locally:**
   ```bash
   make ci-full  # Backend
   npm run ci    # Frontend
   ```
3. **Fix the issue** and push
4. **CI will re-run** automatically

## Code Style

### Go Code Style

Follow standard Go conventions:

```bash
# Format code
make fmt

# Run linter
make vet

# Both
make lint
```

**Guidelines:**
- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions small and focused

### TypeScript/React Code Style

Follow the project's ESLint configuration:

```bash
# Check for issues
npm run lint

# Auto-fix issues
npm run lint -- --fix
```

**Guidelines:**
- Use TypeScript for type safety
- Follow React best practices
- Use functional components with hooks
- Keep components small and reusable
- Add JSDoc comments for complex functions

### SQL Code Style

- Use uppercase for SQL keywords
- Indent nested queries
- Add comments for complex queries
- Use meaningful table/column names

## Documentation

### When to Update Documentation

Update documentation when:
- Adding new features
- Changing APIs
- Modifying configuration
- Fixing bugs that affect usage
- Improving deployment process

### Documentation Files

- `README.md` - Project overview and quick start
- `DEPLOYMENT.md` - Deployment instructions
- `docs/CI_CD.md` - CI/CD pipeline documentation
- `docs/CONTAINER_USAGE.md` - Container usage guide
- `backend/TESTING.md` - Testing guide
- API documentation - In code comments

### Writing Good Documentation

- **Clear and concise** - Get to the point
- **Examples** - Show, don't just tell
- **Up-to-date** - Keep in sync with code
- **Tested** - Verify commands work
- **Organized** - Use headings and lists

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0) - Breaking changes
- **MINOR** (0.1.0) - New features, backward compatible
- **PATCH** (0.0.1) - Bug fixes, backward compatible

### Creating a Release

1. **Update version** in relevant files
2. **Update CHANGELOG.md** with changes
3. **Create and push tag:**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
4. **GitHub Actions** will automatically:
   - Build binaries
   - Build containers
   - Create GitHub release

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated
- [ ] Security scan clean
- [ ] Tested in staging environment

## Dependency Management

### Updating Dependencies

**Backend:**
```bash
cd backend
make deps-check    # Check for updates
make deps-update   # Update dependencies
make test          # Verify updates work
```

**Frontend:**
```bash
cd frontend
npm run deps:check           # Check for updates
npm run deps:update          # Update dependencies
npm run deps:update-interactive  # Interactive updates
npm test                     # Verify updates work
```

### Security Updates

**High priority** - Fix immediately:
```bash
# Backend
make deps-audit

# Frontend
npm audit fix
```

### Weekly Dependency Checks

The project has automated weekly dependency checks. Review and address:
- Outdated dependencies
- Security vulnerabilities
- Breaking changes

## Getting Help

### Resources

- [Project Documentation](docs/)
- [IBM Storage Virtualize REST API Docs](https://www.ibm.com/docs/en/flashsystem)
- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)

### Communication

- **GitHub Issues** - Bug reports, feature requests
- **GitHub Discussions** - Questions, ideas
- **Pull Requests** - Code review discussions

### Reporting Issues

When reporting bugs, include:

1. **Description** - What happened vs. what you expected
2. **Steps to reproduce** - Detailed steps
3. **Environment:**
   - OS and version
   - Go/Node.js version
   - IBM SVC version
4. **Logs** - Relevant error messages
5. **Screenshots** - If applicable

## Recognition

Contributors will be:
- Listed in release notes
- Credited in CHANGELOG.md
- Recognized in the project

Thank you for contributing! 🎉

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

*Last updated: 2026-06-17*