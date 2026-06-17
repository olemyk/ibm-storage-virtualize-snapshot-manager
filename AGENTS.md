# AGENTS.md

This file provides guidance to agents when working with code in this repository.

## Project: IBM Storage Virtualize Snapshot Manager

Web application for managing multiple snapshot schedules per volume group on IBM Storage Virtualize systems (FlashSystem).

**IMPORTANT**: This is an unofficial, community-driven project. NOT affiliated with, endorsed by, or supported by IBM Corporation. Licensed under Apache 2.0. Use at your own risk.

### Commands

**Build & Run:**
```bash
# Backend (Go)
cd backend && go run main.go

# Frontend (React/Vue)
cd frontend && npm run dev

# Build for production
cd backend && go build -o snapshot-manager
cd frontend && npm run build
```

**Testing:**
```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test
```

### IBM Storage Virtualize REST API Constraints

**CRITICAL - Rate Limits:**
- Auth endpoint: 3 requests/second maximum
- Command endpoints: 10 requests/second maximum
- Max 4 different tokens per cluster
- Returns HTTP 429 if limits exceeded
- Token expiry is encoded in JWT - must be decoded and cached

**API Quirks:**
- ALL endpoints use POST method (even for listing/reading data)
- Auth uses `X-Auth-Username` and `X-Auth-Password` headers (not body)
- Subsequent requests use `X-Auth-Token` header
- HTTPS required on port 7443
- Volume groups can only have ONE snapshot policy (this app works around this limitation)

### Architecture Notes

**Snapshot Limitation Workaround:**
- IBM SVC allows only ONE snapshot policy per volume group
- This app manages multiple schedules by triggering `/addsnapshot` REST API calls directly
- Scheduler runs independently and executes snapshots based on cron expressions
- Each schedule stores: cron expression, retention days, safeguarded flag, pool name

**Token Management:**
- Must cache and reuse tokens (max 4 per cluster)
- Decode JWT to check expiry before reuse
- Implement token refresh logic to avoid auth rate limits

**Database Schema:**
- `storage_systems` - stores encrypted credentials for multiple IBM SVC systems
- `volume_groups` - cached volume group info from systems
- `snapshot_schedules` - multiple schedules per volume group (core feature)
- `snapshot_executions` - audit log of all snapshot operations

### Code Style & Conventions

**Go Backend:**
- Use standard Go project layout (`cmd/`, `internal/`, `pkg/`)
- Error handling: always check errors, wrap with context
- Use `context.Context` for cancellation in long-running operations
- Encrypt storage system passwords before storing in database

**Frontend:**
- Component-based architecture (React/Vue)
- API calls through centralized client
- Handle JWT token storage and refresh

### Testing

**Integration Tests:**
- Mock IBM SVC REST API responses (don't hit real systems in tests)
- Test rate limit handling (429 responses)
- Test token expiry and refresh logic

**Scheduler Tests:**
- Test cron expression parsing
- Test next execution time calculation
- Mock time for deterministic tests