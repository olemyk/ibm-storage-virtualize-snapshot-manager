# Notification System - User Guide

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Notification Channels](#notification-channels)
4. [Alert Rules](#alert-rules)
5. [Notification History](#notification-history)
6. [Event Types](#event-types)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Overview

The Notification System allows you to receive real-time alerts about snapshot operations and system events. You can configure multiple notification channels (Email, Slack, Webhook, SNMP) and create flexible alert rules to control when and how notifications are sent.

### Key Features

- **Multiple Channel Types:** Email, Slack, Webhook, SNMP
- **Flexible Alert Rules:** Configure which events trigger notifications
- **Severity Levels:** Filter by info, warning, error, or critical
- **Throttling:** Prevent notification spam
- **Comprehensive History:** Track all sent notifications
- **Test Functionality:** Verify channels before use

---

## Getting Started

### Quick Start (5 Minutes)

1. **Create a Notification Channel**
   - Navigate to: Notifications → Channels
   - Click "Add Channel"
   - Choose channel type and configure
   - Test the channel

2. **Create an Alert Rule**
   - Navigate to: Notifications → Alert Rules
   - Click "Add Rule"
   - Select event type and severity
   - Choose your channel
   - Save and activate

3. **Verify Setup**
   - Trigger a snapshot (or wait for scheduled execution)
   - Check Notifications → History for sent notifications

---

## Notification Channels

Channels define **where** notifications are sent. You can create multiple channels of different types.

### Channel Types

#### 1. Email (SMTP)

Send notifications via email using your SMTP server.

**Configuration:**

| Field | Description | Example |
|-------|-------------|---------|
| Name | Friendly name for the channel | "IT Team Email" |
| SMTP Host | Mail server hostname | smtp.gmail.com |
| SMTP Port | Mail server port | 587 (TLS) or 465 (SSL) |
| Username | SMTP authentication username | alerts@company.com |
| Password | SMTP authentication password | (encrypted) |
| From Address | Sender email address | noreply@company.com |
| To Addresses | Recipient emails (comma-separated) | admin@company.com, ops@company.com |
| Use TLS | Enable TLS encryption | ✓ Recommended |

**Example Setup:**

```
Name: Production Alerts
SMTP Host: smtp.office365.com
SMTP Port: 587
Username: alerts@company.com
Password: ••••••••
From: snapshot-manager@company.com
To: storage-team@company.com, ops@company.com
Use TLS: ✓
```

**Testing:**
Click "Test" button to send a test email. Check your inbox (and spam folder) for the test message.

---

#### 2. Slack

Send notifications to Slack channels using incoming webhooks.

**Configuration:**

| Field | Description | Example |
|-------|-------------|---------|
| Name | Friendly name | "Slack #storage-alerts" |
| Webhook URL | Slack incoming webhook URL | https://hooks.slack.com/services/... |
| Channel | Target channel (optional) | #storage-alerts |
| Username | Bot display name (optional) | Snapshot Manager |
| Icon Emoji | Bot icon (optional) | :floppy_disk: |

**Setup Steps:**

1. In Slack, go to: Apps → Incoming Webhooks
2. Click "Add to Slack"
3. Choose channel and click "Add Incoming WebHooks integration"
4. Copy the Webhook URL
5. Paste into Snapshot Manager

**Example Setup:**

```
Name: Slack Storage Team
Webhook URL: https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX
Channel: #storage-alerts
Username: Snapshot Bot
Icon: :warning:
```

**Testing:**
Click "Test" to send a test message to your Slack channel.

---

#### 3. Webhook

Send HTTP requests to custom endpoints for integration with other systems.

**Configuration:**

| Field | Description | Example |
|-------|-------------|---------|
| Name | Friendly name | "ServiceNow Integration" |
| URL | Target endpoint URL | https://api.company.com/alerts |
| Method | HTTP method | POST |
| Headers | Custom headers (JSON) | {"Authorization": "Bearer token"} |
| Authentication | Auth type | Bearer Token / Basic Auth / None |

**Example Setup:**

```
Name: PagerDuty Integration
URL: https://events.pagerduty.com/v2/enqueue
Method: POST
Headers: {
  "Content-Type": "application/json",
  "Authorization": "Token token=your_integration_key"
}
```

**Payload Format:**

The webhook sends a JSON payload:

```json
{
  "event_type": "snapshot_failure",
  "severity": "error",
  "message": "Snapshot failed for VG_Production",
  "timestamp": "2026-06-11T10:30:00Z",
  "details": {
    "system": "FlashSystem-01",
    "volume_group": "VG_Production",
    "error": "Connection timeout"
  }
}
```

---

#### 4. SNMP

Send SNMP traps to network monitoring systems.

**Configuration:**

| Field | Description | Example |
|-------|-------------|---------|
| Name | Friendly name | "Nagios SNMP" |
| Host | SNMP manager hostname/IP | 192.168.1.100 |
| Port | SNMP port | 162 (default) |
| Community | SNMP community string | public |
| Version | SNMP version | v2c |
| Trap OID | Custom trap OID | 1.3.6.1.4.1.99999.1 |

**Example Setup:**

```
Name: Network Monitoring
Host: monitoring.company.com
Port: 162
Community: private
Version: v2c
Trap OID: 1.3.6.1.4.1.12345.1.1
```

---

### Managing Channels

#### Create Channel

1. Click "Add Channel"
2. Fill in configuration
3. Click "Create"
4. Test the channel

#### Edit Channel

1. Click "Edit" on existing channel
2. Modify configuration
3. Click "Update"
4. Re-test if needed

#### Delete Channel

1. Click "Delete" on channel
2. Confirm deletion
3. **Note:** Rules using this channel will fail

#### Test Channel

1. Click "Test" button
2. Check destination for test message
3. Verify message received correctly

---

## Alert Rules

Alert Rules define **when** and **what** notifications are sent.

### Rule Components

1. **Event Type:** Which events trigger the rule
2. **Severity:** Minimum severity level
3. **Channels:** Where to send notifications
4. **Throttling:** How often to send notifications

### Creating a Rule

#### Step 1: Basic Information

```
Name: Critical Snapshot Failures
Description: Alert on any snapshot failure
```

#### Step 2: Event Configuration

**Event Type:** Choose from:
- Snapshot Success
- Snapshot Failure ← Most common
- Snapshot Warning
- System Connection Lost
- Scheduler Error
- Consecutive Failures

**Severity:** Choose minimum level:
- Info (all events)
- Warning (warning, error, critical)
- Error (error, critical)
- Critical (only critical)

#### Step 3: Select Channels

Check one or more channels:
- ☑ IT Team Email
- ☑ Slack #storage-alerts
- ☐ PagerDuty Webhook

**Note:** Select at least one channel.

#### Step 4: Throttling

Set minimum time between notifications:

```
Throttle: 60 minutes
```

This prevents spam if multiple failures occur rapidly.

**Recommended Values:**
- Info events: 30-60 minutes
- Warning events: 15-30 minutes
- Error events: 5-15 minutes
- Critical events: 0-5 minutes (immediate)

#### Step 5: Activate

Toggle "Active" to enable the rule.

---

### Example Rules

#### Rule 1: Critical Failures

```yaml
Name: Critical Snapshot Failures
Event Type: Snapshot Failure
Severity: Error
Channels: 
  - IT Team Email
  - Slack #storage-alerts
  - PagerDuty
Throttle: 5 minutes
Active: Yes
```

**Use Case:** Immediate notification of any snapshot failure.

---

#### Rule 2: Success Notifications

```yaml
Name: Daily Success Summary
Event Type: Snapshot Success
Severity: Info
Channels:
  - IT Team Email
Throttle: 1440 minutes (24 hours)
Active: Yes
```

**Use Case:** Daily confirmation that snapshots are working.

---

#### Rule 3: Connection Issues

```yaml
Name: Storage System Offline
Event Type: System Connection Lost
Severity: Critical
Channels:
  - Slack #storage-alerts
  - PagerDuty
Throttle: 15 minutes
Active: Yes
```

**Use Case:** Alert when storage system becomes unreachable.

---

### Managing Rules

#### Edit Rule

1. Click "Edit" on rule
2. Modify settings
3. Click "Update"

#### Delete Rule

1. Click "Delete"
2. Confirm deletion

#### Enable/Disable Rule

Toggle the "Active" checkbox in the rule form.

**Tip:** Disable rules temporarily during maintenance windows.

---

## Notification History

View all sent notifications with detailed information.

### Accessing History

Navigate to: **Notifications → History**

### History Information

Each entry shows:
- **Status:** Sent, Failed, Throttled, Pending
- **Severity:** Info, Warning, Error, Critical
- **Timestamp:** When notification was sent
- **Event Type:** What triggered it
- **Channel:** Where it was sent
- **Message:** Notification content
- **Error:** If failed, why it failed

### Filtering History

Use filters to find specific notifications:

| Filter | Options |
|--------|---------|
| Status | All, Sent, Failed, Pending |
| Event Type | All event types |
| Severity | All, Info, Warning, Error, Critical |
| Channel ID | Specific channel |
| Date Range | From/To dates |

**Example:** Find all failed notifications in the last 7 days:
```
Status: Failed
From Date: 2026-06-04
To Date: 2026-06-11
```

### Pagination

- **Per Page:** 25, 50, 100, or 200 entries
- **Navigation:** Previous/Next buttons
- **Current Range:** Shows "Showing 1-50"

### Viewing Details

Click "Show Details" to see full event data in JSON format.

### Exporting History

Click "Export to CSV" to download history as a spreadsheet.

**CSV Columns:**
- ID
- Timestamp
- Event Type
- Severity
- Channel
- Status
- Message
- Error

---

## Event Types

### Snapshot Success

**Triggered When:** Snapshot created successfully

**Severity:** Info

**Details Include:**
- System name
- Volume group name
- Snapshot name
- Retention days
- Safeguarded status

**Example Message:**
```
Snapshot created successfully for VG_Production on FlashSystem-01
```

---

### Snapshot Failure

**Triggered When:** Snapshot creation fails

**Severity:** Error or Critical

**Details Include:**
- System name
- Volume group name
- Error message
- Attempted snapshot name

**Example Message:**
```
Snapshot failed for VG_Production: Connection timeout to storage system
```

**Common Causes:**
- Storage system unreachable
- Insufficient space
- Invalid credentials
- Volume group not found

---

### Snapshot Warning

**Triggered When:** Snapshot succeeds with warnings

**Severity:** Warning

**Example Message:**
```
Snapshot created with warnings: Retention time adjusted due to policy
```

---

### System Connection Lost

**Triggered When:** Cannot connect to storage system

**Severity:** Critical

**Details Include:**
- System name
- Last successful connection
- Error details

**Example Message:**
```
Lost connection to FlashSystem-01 (192.168.1.100)
```

---

### Scheduler Error

**Triggered When:** Scheduler encounters an error

**Severity:** Error

**Example Message:**
```
Scheduler error: Failed to load schedules from database
```

---

### Consecutive Failures

**Triggered When:** Multiple snapshots fail in a row

**Severity:** Critical

**Details Include:**
- Number of consecutive failures
- Time range
- Affected schedules

**Example Message:**
```
3 consecutive snapshot failures detected for schedule 'Hourly Backup'
```

---

## Best Practices

### Channel Configuration

1. **Use Multiple Channels**
   - Email for detailed reports
   - Slack for quick alerts
   - PagerDuty for critical issues

2. **Test Regularly**
   - Test channels after creation
   - Re-test after configuration changes
   - Verify credentials haven't expired

3. **Secure Credentials**
   - Use dedicated service accounts
   - Rotate passwords regularly
   - Use app-specific passwords when available

### Alert Rule Design

1. **Start Simple**
   - Begin with one rule for failures
   - Add more rules as needed
   - Don't over-alert

2. **Use Appropriate Severity**
   - Info: Routine operations
   - Warning: Potential issues
   - Error: Failed operations
   - Critical: Urgent attention needed

3. **Set Reasonable Throttling**
   - Prevent alert fatigue
   - Balance responsiveness vs. spam
   - Adjust based on experience

4. **Name Rules Clearly**
   - Use descriptive names
   - Include severity in name
   - Mention target audience

### Monitoring

1. **Check History Regularly**
   - Review failed notifications
   - Identify patterns
   - Fix recurring issues

2. **Maintain Channels**
   - Remove unused channels
   - Update contact information
   - Verify channels still work

3. **Review Rules Periodically**
   - Disable unnecessary rules
   - Adjust throttling as needed
   - Update channel assignments

---

## Troubleshooting

### Notifications Not Received

**Problem:** Created rule but no notifications arrive.

**Solutions:**

1. **Check Rule is Active**
   - Navigate to Alert Rules
   - Verify "Active" status
   - Enable if disabled

2. **Verify Channel Works**
   - Go to Notification Channels
   - Click "Test" on the channel
   - Check if test notification arrives

3. **Check Event Occurred**
   - Go to Notification History
   - Look for matching events
   - Verify event type matches rule

4. **Review Throttling**
   - Check rule throttle setting
   - May be suppressed due to recent notification
   - Check "Last Triggered" timestamp

5. **Verify Severity Match**
   - Event severity must meet rule minimum
   - Example: "Warning" rule won't trigger for "Info" events

---

### Test Notification Fails

**Problem:** Test button shows error.

**Solutions:**

#### Email Issues

1. **Check SMTP Settings**
   - Verify host and port
   - Confirm username/password
   - Try different port (587 vs 465)

2. **Firewall/Network**
   - Ensure outbound SMTP allowed
   - Check if port is blocked
   - Verify DNS resolution

3. **Authentication**
   - Use app-specific password (Gmail)
   - Enable "Less secure apps" if needed
   - Check account not locked

#### Slack Issues

1. **Verify Webhook URL**
   - Copy entire URL
   - Check for extra spaces
   - Ensure webhook still active

2. **Check Permissions**
   - Webhook must have channel access
   - Verify app not removed from workspace

#### Webhook Issues

1. **Check URL**
   - Verify endpoint is reachable
   - Test with curl/Postman
   - Check for typos

2. **Authentication**
   - Verify token/credentials
   - Check header format
   - Ensure not expired

3. **SSL/TLS**
   - Verify certificate valid
   - Check if self-signed cert accepted

#### SNMP Issues

1. **Network Connectivity**
   - Ping SNMP manager
   - Verify port 162 open
   - Check firewall rules

2. **Community String**
   - Verify correct community
   - Check case sensitivity
   - Ensure manager accepts traps

---

### Notifications Sent But Not Received

**Problem:** History shows "Sent" but notification not received.

**Solutions:**

1. **Email**
   - Check spam/junk folder
   - Verify recipient address correct
   - Check email server logs

2. **Slack**
   - Verify channel exists
   - Check if channel archived
   - Ensure bot has permissions

3. **Webhook**
   - Check endpoint logs
   - Verify payload format accepted
   - Test endpoint independently

---

### Too Many Notifications

**Problem:** Receiving excessive notifications.

**Solutions:**

1. **Increase Throttling**
   - Edit rule
   - Set higher throttle value
   - Example: 60 → 240 minutes

2. **Adjust Severity**
   - Change from "Info" to "Warning"
   - Filter out routine events
   - Focus on errors only

3. **Disable Unnecessary Rules**
   - Review all active rules
   - Disable redundant rules
   - Keep only essential alerts

---

### Missing Notifications

**Problem:** Some notifications not in history.

**Solutions:**

1. **Check Filters**
   - Clear all filters
   - Expand date range
   - Check all statuses

2. **Database Issues**
   - Verify database accessible
   - Check disk space
   - Review application logs

3. **Throttling**
   - Notifications may be suppressed
   - Check rule throttle settings
   - Review "Last Triggered" times

---

## Support

### Getting Help

1. **Check Logs**
   - Backend: `backend/logs/server.log`
   - Look for notification errors
   - Check timestamps

2. **Review History**
   - Check Notification History
   - Look for error messages
   - Identify patterns

3. **Test Components**
   - Test channels individually
   - Verify rules are active
   - Check event generation

### Common Error Messages

| Error | Meaning | Solution |
|-------|---------|----------|
| "Connection refused" | Cannot reach destination | Check network/firewall |
| "Authentication failed" | Invalid credentials | Update username/password |
| "Timeout" | Destination not responding | Check host is online |
| "Invalid configuration" | Missing required field | Review channel config |
| "Channel not found" | Channel deleted | Update rule to use valid channel |

---

## Appendix

### Notification Payload Examples

#### Email

```
Subject: [ERROR] Snapshot Failed - VG_Production

Snapshot Manager Alert

Event: Snapshot Failure
Severity: Error
Time: 2026-06-11 10:30:00 UTC

System: FlashSystem-01
Volume Group: VG_Production
Error: Connection timeout

Details:
- Schedule: Hourly Backup
- Attempted Snapshot: snap_VG_Production_20260611_103000
- Retention: 7 days
```

#### Slack

```json
{
  "text": "🔴 Snapshot Failed",
  "attachments": [{
    "color": "danger",
    "fields": [
      {"title": "System", "value": "FlashSystem-01", "short": true},
      {"title": "Volume Group", "value": "VG_Production", "short": true},
      {"title": "Error", "value": "Connection timeout"}
    ]
  }]
}
```

#### Webhook

```json
{
  "event_type": "snapshot_failure",
  "severity": "error",
  "timestamp": "2026-06-11T10:30:00Z",
  "message": "Snapshot failed for VG_Production",
  "details": {
    "system_id": 1,
    "system_name": "FlashSystem-01",
    "volume_group_id": 5,
    "volume_group_name": "VG_Production",
    "schedule_id": 12,
    "schedule_name": "Hourly Backup",
    "error": "Connection timeout",
    "snapshot_name": "snap_VG_Production_20260611_103000"
  }
}
```

---

**Document Version:** 1.0  
**Last Updated:** June 11, 2026  
**For:** IBM Storage Virtualize Snapshot Manager

---

* - Your AI Software Engineer*