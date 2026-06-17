# Ask Mode Rules

## Project Structure

**Backend (Go):**
- `cmd/` - Application entry points
- `internal/` - Private application code
  - `api/` - REST API handlers
  - `auth/` - Authentication logic
  - `scheduler/` - Snapshot scheduler
  - `svc/` - IBM Storage Virtualize client
  - `db/` - Database models and queries
- `pkg/` - Public libraries (if any)
- `cfrest.schema.yaml` - IBM SVC REST API specification (29K lines)

**Frontend (React/Vue):**
- `src/components/` - UI components
- `src/pages/` - Page components
- `src/api/` - API client
- `src/store/` - State management

## Key Concepts

**Volume Groups:**
- Container for volumes/LUNs in IBM Storage Virtualize
- Can only have ONE snapshot policy (IBM limitation)
- This app bypasses limitation by managing schedules externally

**Snapshot Schedules:**
- Multiple schedules per volume group (core feature)
- Each schedule has: cron expression, retention days, safeguarded flag
- Scheduler triggers `/addsnapshot` REST API calls

**IBM Storage Virtualize API:**
- All endpoints use POST method (even for reads)
- Authentication returns JWT token with embedded expiry
- Strict rate limits: 3 auth/sec, 10 commands/sec, max 4 tokens

## Documentation References

- [`PROJECT_PLAN.md`](PROJECT_PLAN.md:1) - Complete project architecture and implementation plan
- [`cfrest.schema.yaml`](cfrest.schema.yaml:1) - IBM SVC REST API specification
- [`AGENTS.md`](AGENTS.md:1) - General project guidance