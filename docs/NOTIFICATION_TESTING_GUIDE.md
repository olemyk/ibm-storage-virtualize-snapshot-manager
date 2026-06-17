# Notification System - Testing Guide

## Overview

This guide provides comprehensive testing procedures for the notification system, including manual testing steps, test scenarios, and validation criteria.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Manual Testing](#manual-testing)
3. [Test Scenarios](#test-scenarios)
4. [Validation Checklist](#validation-checklist)
5. [Performance Testing](#performance-testing)
6. [Security Testing](#security-testing)
7. [Integration Testing](#integration-testing)
8. [Troubleshooting Tests](#troubleshooting-tests)

---

## Prerequisites

### Required Setup

1. **Backend Running:**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

2. **Frontend Running:**
   ```bash
   cd frontend
   npm run dev
   ```

3. **Test SMTP Account:**
   - Gmail with app password, OR
   - SendGrid API key, OR
   - Local SMTP server

4. **Test Slack Workspace:**
   - Incoming webhook URL (optional)

5. **Test Data:**
   - At least one storage system configured
   - At least one volume group
   - At least one snapshot schedule

---

## Manual Testing

### Test 1: Create Email Channel

**Objective:** Verify email channel creation and testing

**Steps:**

1. Navigate to: `http://localhost:5173/notifications/channels`
2. Click "Add Channel"
3. Fill in form:
   ```
   Name: Test Email Channel
   Type: Email
   SMTP Host: smtp.gmail.com
   SMTP Port: 587
   Username: your-email@gmail.com
   Password: your-app-password
   From: your-email@gmail.com
   To: recipient@example.com
   Use TLS: ✓
   ```
4. Click "Create"
5. Click "Test" button on the created channel
6. Check recipient inbox for test email

**Expected Results:**
- ✅ Channel created successfully
- ✅ Channel appears in list
- ✅ Test button shows success message
- ✅ Test email received in inbox

**Validation:**
- Channel ID assigned
- Config encrypted in database
- Test email contains "Test notification" message
- Email headers correct (From, To, Subject)

---

### Test 2: Create Slack Channel

**Objective:** Verify Slack channel creation and testing

**Steps:**

1. Navigate to: `http://localhost:5173/notifications/channels`
2. Click "Add Channel"
3. Fill in form:
   ```
   Name: Test Slack Channel
   Type: Slack
   Webhook URL: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
   Channel: #test-alerts
   Username: Snapshot Bot
   Icon: :warning:
   ```
4. Click "Create"
5. Click "Test" button
6. Check Slack channel for test message

**Expected Results:**
- ✅ Channel created successfully
- ✅ Test message appears in Slack
- ✅ Message formatted correctly
- ✅ Bot name and icon displayed

---

### Test 3: Create Webhook Channel

**Objective:** Verify webhook channel creation

**Steps:**

1. Set up test webhook endpoint (use webhook.site or similar)
2. Navigate to: `http://localhost:5173/notifications/channels`
3. Click "Add Channel"
4. Fill in form:
   ```
   Name: Test Webhook
   Type: Webhook
   URL: https://webhook.site/your-unique-url
   Method: POST
   Headers: {"Content-Type": "application/json"}
   ```
5. Click "Create"
6. Click "Test" button
7. Check webhook.site for received request

**Expected Results:**
- ✅ Channel created
- ✅ Webhook receives POST request
- ✅ Payload is valid JSON
- ✅ Headers included correctly

---

### Test 4: Create Alert Rule

**Objective:** Verify alert rule creation

**Steps:**

1. Ensure at least one channel exists
2. Navigate to: `http://localhost:5173/notifications/rules`
3. Click "Add Rule"
4. Fill in form:
   ```
   Name: Test Snapshot Failures
   Description: Alert on any snapshot failure
   Event Type: Snapshot Failure
   Severity: Error
   Channels: [Select your test channel]
   Throttle: 60 minutes
   Active: ✓
   ```
5. Click "Create"

**Expected Results:**
- ✅ Rule created successfully
- ✅ Rule appears in list
- ✅ Selected channel displayed
- ✅ Rule marked as active

---

### Test 5: Trigger Notification

**Objective:** Verify end-to-end notification flow

**Steps:**

1. Create email channel (Test 1)
2. Create alert rule for "Snapshot Failure" (Test 4)
3. Create a snapshot schedule that will fail:
   - Use invalid credentials, OR
   - Use non-existent volume group, OR
   - Disconnect storage system
4. Wait for schedule to execute (or trigger manually)
5. Check notification history
6. Check email inbox

**Expected Results:**
- ✅ Snapshot execution fails
- ✅ Event triggers alert rule
- ✅ Notification sent to channel
- ✅ History entry created with status "sent"
- ✅ Email received with failure details

**Validation:**
- History shows correct event type
- Email contains system name, volume group, error message
- Timestamp accurate
- Severity correct

---

### Test 6: Throttling

**Objective:** Verify throttling prevents spam

**Steps:**

1. Create alert rule with 60-minute throttle
2. Trigger multiple failures within 60 minutes
3. Check notification history
4. Check email inbox

**Expected Results:**
- ✅ First notification sent immediately
- ✅ Subsequent notifications throttled
- ✅ History shows "throttled" status for suppressed notifications
- ✅ Only one email received

---

### Test 7: Multiple Channels

**Objective:** Verify notifications sent to multiple channels

**Steps:**

1. Create email channel
2. Create Slack channel
3. Create alert rule selecting BOTH channels
4. Trigger snapshot failure
5. Check both email and Slack

**Expected Results:**
- ✅ Notification sent to email
- ✅ Notification sent to Slack
- ✅ Both entries in history
- ✅ Both marked as "sent"

---

### Test 8: Channel Update

**Objective:** Verify channel updates work correctly

**Steps:**

1. Create email channel
2. Click "Edit" on channel
3. Change SMTP host to different server
4. Click "Update"
5. Click "Test" button
6. Verify test uses new configuration

**Expected Results:**
- ✅ Channel updated successfully
- ✅ New config saved
- ✅ Test uses updated settings
- ✅ Existing rules still reference channel

---

### Test 9: Channel Deletion

**Objective:** Verify channel deletion and impact on rules

**Steps:**

1. Create channel
2. Create rule using that channel
3. Delete the channel
4. Check rule still exists
5. Trigger event
6. Check notification history

**Expected Results:**
- ✅ Channel deleted successfully
- ✅ Rule still exists but references deleted channel
- ✅ Notification fails with error
- ✅ History shows "failed" status with error message

---

### Test 10: Notification History Filtering

**Objective:** Verify history filtering works

**Steps:**

1. Create multiple notifications (mix of sent/failed)
2. Navigate to: `http://localhost:5173/notifications/history`
3. Test each filter:
   - Status: Failed
   - Event Type: Snapshot Failure
   - Severity: Error
   - Date Range: Last 24 hours
4. Clear filters
5. Test pagination

**Expected Results:**
- ✅ Filters reduce result set correctly
- ✅ Multiple filters work together (AND logic)
- ✅ Clear filters resets to all results
- ✅ Pagination works (prev/next buttons)
- ✅ Page size selector works

---

### Test 11: Export to CSV

**Objective:** Verify CSV export functionality

**Steps:**

1. Ensure notification history has data
2. Navigate to: `http://localhost:5173/notifications/history`
3. Click "Export to CSV"
4. Open downloaded file

**Expected Results:**
- ✅ CSV file downloads
- ✅ File contains all visible history entries
- ✅ Headers correct (ID, Timestamp, Event Type, etc.)
- ✅ Data properly formatted
- ✅ Special characters escaped

---

## Test Scenarios

### Scenario 1: Production Deployment

**Goal:** Validate system ready for production

**Steps:**

1. Create production email channel (company SMTP)
2. Create production Slack channel (ops team)
3. Create critical failure rule:
   - Event: Snapshot Failure
   - Severity: Error
   - Channels: Email + Slack
   - Throttle: 15 minutes
4. Create success summary rule:
   - Event: Snapshot Success
   - Severity: Info
   - Channels: Email only
   - Throttle: 1440 minutes (daily)
5. Test both rules
6. Monitor for 24 hours

**Success Criteria:**
- All notifications delivered
- No false positives
- Throttling works as expected
- Performance acceptable

---

### Scenario 2: High Volume Testing

**Goal:** Verify system handles high notification volume

**Steps:**

1. Create 10 snapshot schedules
2. Create alert rule for all events
3. Set throttle to 0 (no throttling)
4. Trigger all schedules simultaneously
5. Monitor system performance
6. Check all notifications sent

**Success Criteria:**
- All notifications processed
- No notifications lost
- System remains responsive
- Database not overwhelmed

---

### Scenario 3: Failure Recovery

**Goal:** Verify system recovers from failures

**Steps:**

1. Create email channel
2. Create alert rule
3. Stop SMTP server (simulate failure)
4. Trigger snapshot event
5. Check notification history (should show "failed")
6. Restart SMTP server
7. Trigger another event
8. Verify notification sent successfully

**Success Criteria:**
- Failed notification logged with error
- System continues operating
- Subsequent notifications work
- No data corruption

---

### Scenario 4: Security Validation

**Goal:** Verify security measures work

**Steps:**

1. Create channel with password
2. Check database - verify password encrypted
3. Try to access API without authentication
4. Try to access other user's channels (if multi-user)
5. Check audit logs for all operations

**Success Criteria:**
- Passwords encrypted in database
- API requires authentication
- Users can only access own channels
- All operations audited

---

## Validation Checklist

### Functional Testing

- [ ] Can create email channel
- [ ] Can create Slack channel
- [ ] Can create webhook channel
- [ ] Can create SNMP channel
- [ ] Can test all channel types
- [ ] Can update channels
- [ ] Can delete channels
- [ ] Can create alert rules
- [ ] Can update alert rules
- [ ] Can delete alert rules
- [ ] Can enable/disable rules
- [ ] Notifications sent on snapshot success
- [ ] Notifications sent on snapshot failure
- [ ] Throttling prevents spam
- [ ] Multiple channels receive notifications
- [ ] History records all notifications
- [ ] History filtering works
- [ ] History pagination works
- [ ] CSV export works

### UI Testing

- [ ] All pages load correctly
- [ ] Forms validate input
- [ ] Error messages displayed
- [ ] Success messages displayed
- [ ] Loading states shown
- [ ] Buttons disabled during operations
- [ ] Dropdown menu works
- [ ] Navigation works
- [ ] Responsive design (mobile/tablet)

### API Testing

- [ ] All endpoints require authentication
- [ ] Invalid requests return 400
- [ ] Not found returns 404
- [ ] Server errors return 500
- [ ] Response format consistent
- [ ] Error messages helpful
- [ ] Rate limiting works (if implemented)

### Security Testing

- [ ] Passwords encrypted in database
- [ ] API tokens validated
- [ ] SQL injection prevented
- [ ] XSS prevented
- [ ] CSRF protection (if applicable)
- [ ] Audit logging works
- [ ] Sensitive data not logged

### Performance Testing

- [ ] Page load time < 2 seconds
- [ ] API response time < 500ms
- [ ] Notification sent within 5 seconds
- [ ] History query with 1000 records < 1 second
- [ ] No memory leaks
- [ ] Database queries optimized

---

## Performance Testing

### Load Testing

**Test 1: Concurrent Notifications**

```bash
# Send 100 notifications simultaneously
for i in {1..100}; do
  curl -X POST http://localhost:8090/api/notifications/test \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"channel_id": 1, "message": "Load test '$i'"}' &
done
wait
```

**Expected:** All notifications processed within 30 seconds

---

**Test 2: History Query Performance**

```bash
# Query history with large dataset
time curl -X GET "http://localhost:8090/api/notifications/history?limit=1000" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected:** Response time < 1 second

---

### Stress Testing

**Test 3: Rapid Channel Creation**

```bash
# Create 50 channels rapidly
for i in {1..50}; do
  curl -X POST http://localhost:8090/api/notifications/channels \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Channel '$i'",
      "type": "email",
      "config": {"smtp_host": "smtp.test.com", "smtp_port": 587}
    }'
done
```

**Expected:** All channels created successfully

---

## Security Testing

### Test 1: Authentication

```bash
# Try to access without token
curl -X GET http://localhost:8090/api/notifications/channels

# Expected: 401 Unauthorized
```

### Test 2: SQL Injection

```bash
# Try SQL injection in channel name
curl -X POST http://localhost:8090/api/notifications/channels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test'; DROP TABLE notification_channels; --",
    "type": "email",
    "config": {}
  }'

# Expected: Channel created with escaped name, no SQL executed
```

### Test 3: Password Encryption

```sql
-- Check database directly
SELECT config FROM notification_channels WHERE id = 1;

-- Expected: Encrypted string, not plaintext password
```

---

## Integration Testing

### Test 1: End-to-End Flow

**Scenario:** Complete notification flow from snapshot to delivery

1. Create storage system
2. Create volume group
3. Create snapshot schedule
4. Create notification channel
5. Create alert rule
6. Wait for schedule execution
7. Verify notification received
8. Check history entry

**Validation:**
- All components work together
- Data flows correctly
- No errors in logs

---

### Test 2: Multi-System Notifications

**Scenario:** Notifications from multiple storage systems

1. Create 3 storage systems
2. Create schedules on each
3. Create single alert rule for all
4. Trigger snapshots on all systems
5. Verify all notifications sent
6. Check history shows all systems

---

## Troubleshooting Tests

### Test 1: SMTP Connection Failure

**Simulate:** Use invalid SMTP host

**Expected Behavior:**
- Test fails with connection error
- Error message helpful
- System continues operating
- History shows failure

---

### Test 2: Invalid Webhook URL

**Simulate:** Use non-existent webhook URL

**Expected Behavior:**
- Test fails with timeout/404
- Error logged
- Retry attempted (if configured)
- History shows failure

---

### Test 3: Database Connection Loss

**Simulate:** Stop database temporarily

**Expected Behavior:**
- API returns 500 error
- Error logged
- System recovers when database returns
- No data corruption

---

## Automated Testing Script

```bash
#!/bin/bash
# notification-test.sh

TOKEN="your-jwt-token"
BASE_URL="http://localhost:8090/api"

echo "=== Notification System Test Suite ==="

# Test 1: Create Channel
echo "Test 1: Create Email Channel"
CHANNEL_ID=$(curl -s -X POST $BASE_URL/notifications/channels \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Channel",
    "type": "email",
    "config": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "test@example.com",
      "password": "password",
      "from_address": "test@example.com",
      "to_addresses": "admin@example.com",
      "use_tls": true
    }
  }' | jq -r '.id')

if [ -n "$CHANNEL_ID" ]; then
  echo "✓ Channel created: ID=$CHANNEL_ID"
else
  echo "✗ Failed to create channel"
  exit 1
fi

# Test 2: List Channels
echo "Test 2: List Channels"
CHANNEL_COUNT=$(curl -s -X GET $BASE_URL/notifications/channels \
  -H "Authorization: Bearer $TOKEN" | jq 'length')

if [ "$CHANNEL_COUNT" -gt 0 ]; then
  echo "✓ Found $CHANNEL_COUNT channels"
else
  echo "✗ No channels found"
  exit 1
fi

# Test 3: Create Rule
echo "Test 3: Create Alert Rule"
RULE_ID=$(curl -s -X POST $BASE_URL/notifications/rules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Rule",
    "is_active": true,
    "event_type": "snapshot_failure",
    "severity": "error",
    "notification_channel_ids": "['$CHANNEL_ID']",
    "throttle_minutes": 60
  }' | jq -r '.id')

if [ -n "$RULE_ID" ]; then
  echo "✓ Rule created: ID=$RULE_ID"
else
  echo "✗ Failed to create rule"
  exit 1
fi

# Test 4: Get History
echo "Test 4: Get Notification History"
HISTORY_COUNT=$(curl -s -X GET "$BASE_URL/notifications/history?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq 'length')

echo "✓ Found $HISTORY_COUNT history entries"

# Cleanup
echo "Cleanup: Deleting test data"
curl -s -X DELETE $BASE_URL/notifications/rules/$RULE_ID \
  -H "Authorization: Bearer $TOKEN" > /dev/null
curl -s -X DELETE $BASE_URL/notifications/channels/$CHANNEL_ID \
  -H "Authorization: Bearer $TOKEN" > /dev/null

echo "=== All Tests Passed ==="
```

---

## Test Report Template

```markdown
# Notification System Test Report

**Date:** YYYY-MM-DD
**Tester:** Name
**Version:** 1.0

## Summary
- Total Tests: X
- Passed: Y
- Failed: Z
- Success Rate: Y/X %

## Test Results

### Functional Tests
| Test | Status | Notes |
|------|--------|-------|
| Create Email Channel | ✅ Pass | |
| Create Slack Channel | ✅ Pass | |
| Create Alert Rule | ✅ Pass | |
| Trigger Notification | ✅ Pass | |
| Throttling | ✅ Pass | |

### Performance Tests
| Test | Expected | Actual | Status |
|------|----------|--------|--------|
| API Response Time | < 500ms | 250ms | ✅ Pass |
| Notification Delivery | < 5s | 2s | ✅ Pass |

### Security Tests
| Test | Status | Notes |
|------|--------|-------|
| Authentication | ✅ Pass | |
| Encryption | ✅ Pass | |
| SQL Injection | ✅ Pass | |

## Issues Found
1. Issue description
2. Issue description

## Recommendations
1. Recommendation
2. Recommendation

## Conclusion
System ready for production: YES/NO
```

---

**Last Updated:** June 11, 2026  
**For:** IBM Storage Virtualize Snapshot Manager

---

* - Your AI Software Engineer*