# Integration Testing Guide

This guide explains how to run integration tests for the IBM Storage Virtualize Snapshot Manager.

## Prerequisites

- Two IBM Storage Virtualize (FlashSystem) systems accessible on your network
- Volume groups created on each system:
  - System 1: `snapshotmanager_vg_svc1_01`
  - System 2: `snapshotmanager_vg_svc2_01`
- User account with snapshot permissions: `snapshotmanager` / `snapshotmanager`
- Application running in Podman containers (backend, frontend, database)

## Test Scripts

Two integration test scripts are available:

### 1. Automated Test with Cleanup (`integration_test_runner.go`)

**Purpose:** Full end-to-end test that automatically cleans up all created resources.

**What it tests:**
- User authentication and authorization
- Storage system management (add, test connection, delete)
- Volume group synchronization
- Snapshot schedule creation and execution
- User management (create users with different roles)
- Audit logging
- NTP server configuration

**Run the test:**
```bash
cd backend/scripts
go run integration_test_runner.go
```

**Output:**
- Console output with step-by-step progress
- Log file: `integration_test_complete.txt`
- Detailed log: `integration_test.log`

**Cleanup:** Automatically deletes all test resources at the end.

---

### 2. Manual Verification Test (`integration_test_manual.go`)

**Purpose:** Creates test resources that remain in the system for manual verification via the web UI.

**What it creates:**
- Two storage systems: `SVC01_Manual` and `SVC02_Manual`
- Two snapshot schedules: `Manual_Test_Daily_02_00` and `Manual_Test_Daily_03_00`
- Executes both schedules to create snapshots

**Run the test:**
```bash
cd backend/scripts
go run integration_test_manual.go
```

**Output:**
- Console output with resource IDs
- Log file: `integration_test_manual_output.txt`
- Detailed log: `integration_test_manual.log`

**Cleanup:** Resources are **NOT** deleted. You must manually delete them via the UI when done.

---

## Manual Verification Steps

After running `integration_test_manual.go`:

1. **Open the Web UI:**
   ```
   https://localhost
   ```

2. **Login:**
   - Username: `admin`
   - Password: `admin123`

3. **Verify Storage Systems:**
   - Navigate to **Systems** page
   - Confirm `SVC01_Manual` (10.33.7.80) and `SVC02_Manual` (10.33.7.81) are listed
   - Check connection status is "Connected"

4. **Verify Schedules:**
   - Navigate to **Schedules** page
   - Confirm `Manual_Test_Daily_02_00` and `Manual_Test_Daily_03_00` are listed
   - Check schedule details (cron expression, retention, etc.)

5. **Verify Snapshot Executions:**
   - Navigate to **Executions** page
   - Confirm 2 successful snapshot executions
   - Check snapshot names follow pattern: `manual_snap_YYYYMMDD_HHMMSS`

6. **Verify Audit Logs:**
   - Navigate to **Audit Logs** page
   - Confirm actions are logged (system creation, schedule creation, snapshot execution)

7. **Cleanup Test Resources:**
   - Delete the two schedules
   - Delete the two storage systems
   - Verify volume groups are also removed

---

## Configuration

### Storage System IPs

Edit the test scripts to match your environment:

**`integration_test_runner.go`:**
```go
const (
    SVC01IP   = "10.33.7.80"  // Change to your SVC01 IP
    SVC02IP   = "10.33.7.81"  // Change to your SVC02 IP
    SVC01Port = 7443
    SVC02Port = 7443
)
```

**`integration_test_manual.go`:**
```go
svc1Body := map[string]interface{}{
    "ip_address": "10.33.7.80",  // Change to your SVC01 IP
    // ...
}

svc2Body := map[string]interface{}{
    "ip_address": "10.33.7.81",  // Change to your SVC02 IP
    // ...
}
```

### Volume Group Names

If your volume groups have different names, update:

```go
// Look for these lines in Step4_SyncVolumeGroups
if vg["vg_name"].(string) == "snapshotmanager_vg_svc1_01" {
    // Change to your VG name
}
```

---

## Troubleshooting

### Test Fails at "Test system connections"

**Error:** `HTTP 400: cluster has reached maximum of 4 tokens`

**Solution:** IBM SVC has a hard limit of 4 active tokens per cluster. 

### Test Fails at "Sync volume groups"

**Error:** Volume groups not found

**Solution:**
1. Verify volume groups exist on the IBM SVC systems
2. Check volume group names match exactly (case-sensitive)
3. Ensure user has permissions to list volume groups

### Test Fails at "Execute schedules"

**Error:** Snapshot creation failed

**Solution:**
1. Check IBM SVC has available storage capacity
2. Verify snapshot naming pattern is valid (no spaces, starts with letter/underscore)
3. Check IBM SVC logs for detailed error messages

### Frontend/Backend Connection Issues

**Error:** `context deadline exceeded` or `502 Bad Gateway`

**Solution:**
1. Restart containers:
   ```bash
   podman restart snapshot-manager-frontend
   podman restart snapshot-manager-backend
   sleep 10
   ```

2. Check container status:
   ```bash
   podman ps
   podman logs snapshot-manager-backend
   podman logs snapshot-manager-frontend
   ```

---

## Test Coverage

### Automated Test (`integration_test_runner.go`)

- ✅ User authentication (login)
- ✅ Storage system CRUD operations
- ✅ Connection testing with retry logic
- ✅ Volume group synchronization
- ✅ Snapshot schedule creation
- ✅ Manual snapshot execution
- ✅ Execution status verification
- ✅ NTP server configuration
- ✅ Audit log verification
- ✅ User management (create, login with different roles)
- ✅ Schedule updates
- ✅ Resource deletion
- ✅ Automatic cleanup

### Manual Test (`integration_test_manual.go`)

- ✅ Storage system creation
- ✅ Connection testing
- ✅ Volume group synchronization
- ✅ Snapshot schedule creation
- ✅ Manual snapshot execution
- ✅ Resources left for UI verification

---

## IBM Storage Virtualize API Constraints

The tests are designed to work within IBM SVC API limitations:

- **Token Limit:** Max 4 tokens per cluster (enforced by retry logic)
- **Rate Limits:**
  - Auth endpoint: 3 requests/second
  - Command endpoints: 10 requests/second
- **Snapshot Naming:** Must start with letter/underscore, alphanumeric + underscores/dashes only
- **API Method:** All endpoints use POST (even for read operations)

---

## Continuous Integration

To run tests in CI/CD pipelines:

```bash
#!/bin/bash
set -e

# Start containers
podman-compose up -d
sleep 15

# Run automated test
cd backend/scripts
go run integration_test_runner.go

# Check exit code
if [ $? -eq 0 ]; then
    echo "✅ Integration tests passed"
    exit 0
else
    echo "❌ Integration tests failed"
    exit 1
fi
```

---

## Support

For issues or questions:
- Check logs: `backend/scripts/integration_test.log`
- Review IBM SVC REST API documentation
- Check container logs: `podman logs snapshot-manager-backend`
