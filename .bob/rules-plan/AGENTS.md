# Plan Mode Rules

## Architecture Constraints

**IBM Storage Virtualize Limitation:**
- Volume groups can only have ONE snapshot policy attached
- This is a hard platform limitation, not a configuration issue
- Solution: External scheduler triggers `/addsnapshot` API directly
- NEVER suggest using multiple snapshot policies per volume group

**Token Management Architecture:**
- Max 4 tokens per cluster (hard limit)
- Token expiry encoded in JWT (must decode to check)
- Auth endpoint: 3 req/sec limit (very restrictive)
- Design must cache and reuse tokens aggressively

**Rate Limit Constraints:**
- Auth: 3 requests/second
- Commands: 10 requests/second
- HTTP 429 returned when exceeded
- Must implement exponential backoff and retry logic

## Design Patterns

**Scheduler Design:**
- Use cron library for scheduling (not custom time-based loops)
- Store next_execution_at in database for UI display
- Reload all schedules on application restart
- Each schedule executes independently (no dependencies)

**Database Design:**
- Encrypt storage system credentials (passwords)
- Use foreign keys with CASCADE delete
- Index on frequently queried fields (system_id, vg_id)
- Audit log all snapshot operations (snapshot_executions table)

**API Design:**
- Backend REST API separate from IBM SVC REST API
- Frontend calls backend, backend calls IBM SVC
- Backend manages token caching and rate limiting
- Use JWT for backend authentication

## Non-Standard Patterns

**All IBM SVC Endpoints Use POST:**
- Even read operations (lsvolumegroup, lsvolumegroupsnapshot)
- This is counterintuitive but required by IBM SVC API
- Don't suggest GET/PUT/DELETE for IBM SVC calls

**Authentication Headers:**
- Initial auth: `X-Auth-Username` and `X-Auth-Password` headers
- Subsequent calls: `X-Auth-Token` header
- NOT in request body (common mistake)

**Snapshot Creation:**
- Must use `/addsnapshot` endpoint directly
- Cannot rely on snapshot policies (only one per VG)
- Retention days is mandatory parameter
- Safeguarded flag makes snapshots immutable

## Technology Stack Decisions

**Backend: Go (Golang)**
- Rationale: User mentioned existing Go project for IBM SVC
- Strong concurrency for scheduler
- Single binary deployment
- Excellent HTTP client libraries

**Database: SQLite or PostgreSQL**
- SQLite: Simple deployment, good for initial release
- PostgreSQL: Better for production, multi-user scenarios
- Design must support both (use database/sql interface)

**Frontend: React or Vue.js**
- Modern component-based architecture
- Good ecosystem for form handling (cron expressions)
- Easy integration with Go backend

## Implementation Phases

1. **Core Infrastructure** - Database, IBM SVC client, auth
2. **Storage System Management** - Add/edit systems, test connections
3. **Snapshot Scheduling** - Cron scheduler, execute snapshots
4. **Frontend Development** - UI for all features
5. **Monitoring & Polish** - Execution history, dashboard, testing