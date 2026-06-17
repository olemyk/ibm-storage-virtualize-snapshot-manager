# Advanced Mode Rules

## IBM Storage Virtualize REST API Integration

**CRITICAL - Token Management:**
- MUST cache tokens (max 4 per cluster limit)
- MUST decode JWT to extract expiry time before reuse
- NEVER make auth requests in loops (3 req/sec limit)
- ALWAYS check token expiry before API calls

**API Call Pattern:**
```go
// CORRECT: Reuse token with expiry check
func (c *SVCClient) GetOrRefreshToken(system StorageSystem) (string, error) {
    if c.tokenCache[system.ID] != nil && !c.isTokenExpired(c.tokenCache[system.ID]) {
        return c.tokenCache[system.ID].Token, nil
    }
    return c.authenticate(system)
}

// WRONG: Don't authenticate on every call
func (c *SVCClient) ListVolumeGroups(system StorageSystem) error {
    token, _ := c.authenticate(system) // BAD: Hits rate limit
    // ...
}
```

**Snapshot Execution:**
- MUST use `/addsnapshot` endpoint (not snapshot policies)
- ALWAYS include `retentiondays` parameter (required)
- Use `safeguarded: true` for immutable snapshots
- Generate unique snapshot names with timestamp

**Error Handling:**
- HTTP 429: Implement exponential backoff
- HTTP 401: Token expired, refresh and retry once
- HTTP 403: Invalid credentials, don't retry
- Network errors: Retry with backoff (max 3 attempts)

**Database Operations:**
- ALWAYS encrypt storage system passwords before INSERT/UPDATE
- Use transactions for multi-table operations (schedules + executions)
- Index on `storage_system_id`, `volume_group_id` for performance

**Scheduler Implementation:**
- Use `github.com/robfig/cron/v3` for cron scheduling
- Calculate next execution time on schedule creation/update
- Store next_execution_at in database for UI display
- Handle scheduler restart: reload all active schedules

**Cron Expression Validation:**
```go
// MUST validate cron expressions before saving
func ValidateCronExpression(expr string) error {
    _, err := cron.ParseStandard(expr)
    return err
}
```

## MCP & Browser Integration

**Testing with Browser:**
- Use browser tools to test web UI after implementation
- Verify snapshot schedule creation workflow
- Test cron expression picker/helper UI
- Validate error messages display correctly