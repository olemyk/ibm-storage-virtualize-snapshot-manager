# Release Preparation Guide - v1.0.0

This document outlines all the steps needed to prepare the repository for the v1.0.0 release.

## ✅ Completed Steps

1. ✅ Created [`RELEASE_NOTES.md`](RELEASE_NOTES.md) - Comprehensive release notes
2. ✅ Created [`CHANGELOG.md`](CHANGELOG.md) - Detailed changelog following Keep a Changelog format

## 📋 Remaining Steps

### Step 1: Update .gitignore

Add the following entries to `.gitignore`:

```gitignore
# Process ID files
*.pid

# Server log files
*-server.log
backend/server*.log
frontend/frontend.log

# Coverage reports
backend/coverage.*

# Test outputs
backend/scripts/*.txt
backend/scripts/*.log

# Compiled binaries (additional)
backend/snapshot-manager

# Development artifacts
cfrest.schema.yaml
```

### Step 2: Create temp Folder Structure

Create the following directory structure:

```
temp/
├── logs/
├── test-outputs/
├── planning-docs/
└── development-artifacts/
```

### Step 3: Move Files to temp Folder

**Root Directory → temp/logs/:**
- `backend-server.log`
- `frontend-server.log`
- `frontend.log`

**Root Directory → temp/development-artifacts/:**
- `frontend.pid`
- `server.pid`
- `.env` (if exists - contains secrets)
- `cfrest.schema.yaml`

**Root Directory → temp/planning-docs/:**
- `CONTAINERIZATION_PLAN.md`
- `INTEGRATION_TEST_PLAN.md`
- `POSTGRESQL_MIGRATION_COMPLETE.md`
- `POSTGRESQL_MIGRATION_FIXES.md`
- `SECURITY_FIXES_TESTING.md`

**Backend Directory → temp/logs/:**
- `backend/.env` (if exists - contains secrets)
- `backend/.env.keys` (if exists - contains secrets)
- `backend/server_fresh.log`
- `backend/server_test.log`
- `backend/server.log`
- `backend/server.pid`

**Backend Directory → temp/development-artifacts/:**
- `backend/coverage.html`
- `backend/coverage.out`
- `backend/snapshot-manager` (compiled binary)

**Backend Scripts → temp/test-outputs/:**
- `backend/scripts/integration_test_complete.txt`
- `backend/scripts/integration_test_full.txt`
- `backend/scripts/integration_test_manual_output.txt`
- `backend/scripts/integration_test_manual.log`
- `backend/scripts/integration_test_output.txt`
- `backend/scripts/integration_test.log`

**Frontend Directory → temp/development-artifacts/:**
- `frontend/frontend.pid`

### Step 4: Update frontend/package.json

Change version from `0.0.0` to `1.0.0`:

```json
{
  "name": "frontend",
  "private": true,
  "version": "1.0.0",
  ...
}
```

### Step 5: Verify Documentation

Ensure these files exist and are up to date:
- ✅ README.md
- ✅ LICENSE
- ✅ QUICKSTART.md
- ✅ DEVELOPMENT.md
- ✅ DEPLOYMENT.md
- ✅ AGENTS.md
- ✅ RELEASE_NOTES.md (newly created)
- ✅ CHANGELOG.md (newly created)
- ✅ docs/API_NOTIFICATIONS.md
- ✅ docs/NOTIFICATIONS_USER_GUIDE.md
- ✅ docs/SMTP_SERVICES_GUIDE.md
- ✅ docs/INTEGRATION_TESTING.md
- ✅ docs/NOTIFICATION_TESTING_GUIDE.md

### Step 6: Git Preparation

**Commands to run:**

```bash
# Create temp directory structure
mkdir -p temp/{logs,test-outputs,planning-docs,development-artifacts}

# Move files (use git mv for tracked files, mv for untracked)
# Root logs
mv backend-server.log temp/logs/ 2>/dev/null || true
mv frontend-server.log temp/logs/ 2>/dev/null || true
mv frontend.log temp/logs/ 2>/dev/null || true

# Root development artifacts
mv frontend.pid temp/development-artifacts/ 2>/dev/null || true
mv server.pid temp/development-artifacts/ 2>/dev/null || true
mv .env temp/development-artifacts/ 2>/dev/null || true
mv cfrest.schema.yaml temp/development-artifacts/ 2>/dev/null || true

# Planning docs
git mv CONTAINERIZATION_PLAN.md temp/planning-docs/ 2>/dev/null || true
git mv INTEGRATION_TEST_PLAN.md temp/planning-docs/ 2>/dev/null || true
git mv POSTGRESQL_MIGRATION_COMPLETE.md temp/planning-docs/ 2>/dev/null || true
git mv POSTGRESQL_MIGRATION_FIXES.md temp/planning-docs/ 2>/dev/null || true
git mv SECURITY_FIXES_TESTING.md temp/planning-docs/ 2>/dev/null || true

# Backend logs
mv backend/.env temp/development-artifacts/ 2>/dev/null || true
mv backend/.env.keys temp/development-artifacts/ 2>/dev/null || true
mv backend/server_fresh.log temp/logs/ 2>/dev/null || true
mv backend/server_test.log temp/logs/ 2>/dev/null || true
mv backend/server.log temp/logs/ 2>/dev/null || true
mv backend/server.pid temp/development-artifacts/ 2>/dev/null || true

# Backend development artifacts
mv backend/coverage.html temp/development-artifacts/ 2>/dev/null || true
mv backend/coverage.out temp/development-artifacts/ 2>/dev/null || true
mv backend/snapshot-manager temp/development-artifacts/ 2>/dev/null || true

# Backend test outputs
mv backend/scripts/integration_test_complete.txt temp/test-outputs/ 2>/dev/null || true
mv backend/scripts/integration_test_full.txt temp/test-outputs/ 2>/dev/null || true
mv backend/scripts/integration_test_manual_output.txt temp/test-outputs/ 2>/dev/null || true
mv backend/scripts/integration_test_manual.log temp/test-outputs/ 2>/dev/null || true
mv backend/scripts/integration_test_output.txt temp/test-outputs/ 2>/dev/null || true
mv backend/scripts/integration_test.log temp/test-outputs/ 2>/dev/null || true

# Frontend development artifacts
mv frontend/frontend.pid temp/development-artifacts/ 2>/dev/null || true

# Check git status
git status
```

### Step 7: Final Git Commit

**Commit Message:**

```
Release v1.0.0 - IBM Storage Virtualize Snapshot Manager

First official release of the IBM Storage Virtualize Snapshot Manager.

Major Features:
- Multi-system management with encrypted credentials
- Multiple snapshot schedules per volume group
- Cron-based flexible scheduling
- Comprehensive notification system (Email, Slack, Webhook, SNMP)
- Audit logging and execution history
- Modern React web interface
- Container deployment with Podman/Docker
- PostgreSQL and SQLite database support

Documentation:
- Added RELEASE_NOTES.md with comprehensive release information
- Added CHANGELOG.md following Keep a Changelog format
- Complete API documentation
- Deployment and development guides

Cleanup:
- Moved development artifacts to temp/ folder
- Updated .gitignore for better exclusions
- Organized planning documents

See RELEASE_NOTES.md for complete details.
```

### Step 8: Create Git Tag

```bash
# Create annotated tag
git tag -a v1.0.0 -m "Release v1.0.0 - First Official Release"

# Push tag to remote
git push origin v1.0.0
```

### Step 9: GitHub Release

Create a GitHub release with:
- **Tag:** v1.0.0
- **Title:** IBM Storage Virtualize Snapshot Manager v1.0.0
- **Description:** Copy content from RELEASE_NOTES.md
- **Assets:** None needed (users clone from git)

## 📊 Pre-Release Checklist

Before pushing to git:

- [ ] All tests pass (`cd backend && go test ./...`)
- [ ] Frontend builds successfully (`cd frontend && npm run build`)
- [ ] No sensitive data in repository (check .env files are excluded)
- [ ] All documentation is accurate and up-to-date
- [ ] Version numbers are updated (frontend/package.json)
- [ ] CHANGELOG.md is complete
- [ ] RELEASE_NOTES.md is comprehensive
- [ ] .gitignore excludes all development artifacts
- [ ] temp/ folder is in .gitignore
- [ ] No TODO or FIXME comments in production code
- [ ] All migration scripts are tested
- [ ] Container deployment works (`./deploy/start.sh`)
- [ ] Development setup works (`./dev-start.sh`)

## 🚀 Post-Release Tasks

After pushing v1.0.0:

1. **Announce Release**
   - Update project README with release badge
   - Post announcement in relevant communities
   - Update documentation links

2. **Monitor Issues**
   - Watch for bug reports
   - Respond to questions
   - Plan v1.0.1 patch if needed

3. **Plan Next Release**
   - Review roadmap in RELEASE_NOTES.md
   - Create milestone for v1.1
   - Prioritize features

## 📝 Notes

- The temp/ folder is already in .gitignore, so moved files won't be tracked
- Planning documents are moved to temp/ to keep the main directory clean
- All essential documentation remains in the repository
- Development artifacts are excluded from version control
- Secrets (.env files) are safely moved to temp/ folder

## 🔄 Switching to Code Mode

To execute these changes, switch to Code mode and run:

```bash
# The commands from Step 6 above
```

Or ask the AI to:
1. Update .gitignore with the new patterns
2. Create temp folder structure
3. Move all specified files
4. Update frontend/package.json version
5. Prepare the git commit

---

**Ready for Release:** Once all steps are completed, the repository will be clean and ready for the v1.0.0 release!