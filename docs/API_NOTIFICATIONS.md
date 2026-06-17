# Notification System - API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Notification Channels API](#notification-channels-api)
4. [Alert Rules API](#alert-rules-api)
5. [Notification History API](#notification-history-api)
6. [Test Notification API](#test-notification-api)
7. [Error Codes](#error-codes)
8. [Rate Limiting](#rate-limiting)
9. [Webhooks](#webhooks)

---

## Overview

The Notification API provides endpoints for managing notification channels, alert rules, and viewing notification history.

**Base URL:** `http://localhost:8090/api`

**Content Type:** `application/json`

**API Version:** 1.0

---

## Authentication

All notification endpoints require authentication via JWT token.

### Obtaining a Token

```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your_password"
}
```

**Response:**

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "admin",
    "email": "admin@company.com",
    "role": "admin"
  }
}
```

### Using the Token

Include the token in the `Authorization` header:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Notification Channels API

### List All Channels

Retrieve all notification channels.

```http
GET /api/notifications/channels
Authorization: Bearer {token}
```

**Response:** `200 OK`

```json
[
  {
    "id": 1,
    "name": "IT Team Email",
    "type": "email",
    "is_active": true,
    "config": "{\"smtp_host\":\"smtp.gmail.com\",\"smtp_port\":587,...}",
    "created_at": "2026-06-10T10:00:00Z",
    "updated_at": "2026-06-10T10:00:00Z"
  },
  {
    "id": 2,
    "name": "Slack Alerts",
    "type": "slack",
    "is_active": true,
    "config": "{\"webhook_url\":\"https://hooks.slack.com/...\"}",
    "created_at": "2026-06-10T11:00:00Z",
    "updated_at": "2026-06-10T11:00:00Z"
  }
]
```

---

### Get Channel by ID

Retrieve a specific notification channel.

```http
GET /api/notifications/channels/{id}
Authorization: Bearer {token}
```

**Path Parameters:**
- `id` (integer, required) - Channel ID

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "IT Team Email",
  "type": "email",
  "is_active": true,
  "config": "{\"smtp_host\":\"smtp.gmail.com\",\"smtp_port\":587,\"username\":\"alerts@company.com\",\"password\":\"encrypted_password\",\"from_address\":\"noreply@company.com\",\"to_addresses\":\"admin@company.com\",\"use_tls\":true}",
  "created_at": "2026-06-10T10:00:00Z",
  "updated_at": "2026-06-10T10:00:00Z"
}
```

**Error Response:** `404 Not Found`

```json
{
  "error": "Channel not found"
}
```

---

### Create Channel

Create a new notification channel.

```http
POST /api/notifications/channels
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "Production Alerts",
  "type": "email",
  "config": {
    "smtp_host": "smtp.office365.com",
    "smtp_port": 587,
    "username": "alerts@company.com",
    "password": "your_password",
    "from_address": "noreply@company.com",
    "to_addresses": "ops@company.com,admin@company.com",
    "use_tls": true
  },
  "description": "Email alerts for production environment"
}
```

**Channel Type Configurations:**

#### Email Configuration

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "username": "alerts@company.com",
  "password": "app_specific_password",
  "from_address": "noreply@company.com",
  "to_addresses": "admin@company.com,ops@company.com",
  "use_tls": true
}
```

#### Slack Configuration

```json
{
  "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX",
  "channel": "#storage-alerts",
  "username": "Snapshot Bot",
  "icon_emoji": ":warning:"
}
```

#### Webhook Configuration

```json
{
  "url": "https://api.company.com/webhooks/alerts",
  "method": "POST",
  "headers": {
    "Authorization": "Bearer your_token",
    "Content-Type": "application/json"
  },
  "auth_type": "bearer",
  "auth_token": "your_bearer_token"
}
```

#### SNMP Configuration

```json
{
  "host": "192.168.1.100",
  "port": 162,
  "community": "public",
  "version": "v2c",
  "trap_oid": "1.3.6.1.4.1.99999.1"
}
```

**Response:** `201 Created`

```json
{
  "id": 3,
  "name": "Production Alerts",
  "type": "email",
  "is_active": true,
  "config": "{\"smtp_host\":\"smtp.office365.com\",...}",
  "created_at": "2026-06-11T09:00:00Z",
  "updated_at": "2026-06-11T09:00:00Z"
}
```

**Error Response:** `400 Bad Request`

```json
{
  "error": "Invalid configuration: smtp_host is required"
}
```

---

### Update Channel

Update an existing notification channel.

```http
PUT /api/notifications/channels/{id}
Authorization: Bearer {token}
Content-Type: application/json
```

**Path Parameters:**
- `id` (integer, required) - Channel ID

**Request Body:**

```json
{
  "name": "Updated Channel Name",
  "is_active": false,
  "config": {
    "smtp_host": "smtp.newserver.com",
    "smtp_port": 587,
    "username": "newalerts@company.com",
    "password": "new_password",
    "from_address": "noreply@company.com",
    "to_addresses": "newadmin@company.com",
    "use_tls": true
  }
}
```

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Updated Channel Name",
  "type": "email",
  "is_active": false,
  "config": "{\"smtp_host\":\"smtp.newserver.com\",...}",
  "created_at": "2026-06-10T10:00:00Z",
  "updated_at": "2026-06-11T09:30:00Z"
}
```

---

### Delete Channel

Delete a notification channel.

```http
DELETE /api/notifications/channels/{id}
Authorization: Bearer {token}
```

**Path Parameters:**
- `id` (integer, required) - Channel ID

**Response:** `204 No Content`

**Error Response:** `404 Not Found`

```json
{
  "error": "Channel not found"
}
```

**Warning:** Deleting a channel will cause alert rules using this channel to fail.

---

### Test Channel

Send a test notification through a channel.

```http
POST /api/notifications/channels/{id}/test
Authorization: Bearer {token}
```

**Path Parameters:**
- `id` (integer, required) - Channel ID

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Test notification sent successfully"
}
```

**Error Response:** `400 Bad Request`

```json
{
  "success": false,
  "error": "Failed to send test notification: SMTP authentication failed"
}
```

---

## Alert Rules API

### List All Rules

Retrieve all alert rules.

```http
GET /api/notifications/rules
Authorization: Bearer {token}
```

**Response:** `200 OK`

```json
[
  {
    "id": 1,
    "name": "Critical Snapshot Failures",
    "description": "Alert on any snapshot failure",
    "is_active": true,
    "event_type": "snapshot_failure",
    "conditions": null,
    "severity": "error",
    "notification_channel_ids": "[1,2]",
    "throttle_minutes": 60,
    "last_triggered_at": "2026-06-11T08:30:00Z",
    "created_at": "2026-06-10T10:00:00Z",
    "updated_at": "2026-06-10T10:00:00Z"
  }
]
```

---

### Get Rule by ID

Retrieve a specific alert rule.

```http
GET /api/notifications/rules/{id}
Authorization: Bearer {token}
```

**Path Parameters:**
- `id` (integer, required) - Rule ID

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Critical Snapshot Failures",
  "description": "Alert on any snapshot failure",
  "is_active": true,
  "event_type": "snapshot_failure",
  "conditions": null,
  "severity": "error",
  "notification_channel_ids": "[1,2]",
  "throttle_minutes": 60,
  "last_triggered_at": "2026-06-11T08:30:00Z",
  "created_at": "2026-06-10T10:00:00Z",
  "updated_at": "2026-06-10T10:00:00Z"
}
```

---

### Create Rule

Create a new alert rule.

```http
POST /api/notifications/rules
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "Production Snapshot Failures",
  "description": "Alert when production snapshots fail",
  "is_active": true,
  "event_type": "snapshot_failure",
  "severity": "error",
  "notification_channel_ids": "[1,2,3]",
  "throttle_minutes": 30
}
```

**Field Descriptions:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | Yes | Rule name |
| description | string | No | Rule description |
| is_active | boolean | Yes | Enable/disable rule |
| event_type | string | Yes | Event type to match |
| conditions | object | No | Additional conditions (JSON) |
| severity | string | Yes | Minimum severity level |
| notification_channel_ids | string | Yes | JSON array of channel IDs |
| throttle_minutes | integer | Yes | Minimum time between notifications |

**Event Types:**
- `snapshot_success`
- `snapshot_failure`
- `snapshot_warning`
- `system_connection_lost`
- `scheduler_error`
- `consecutive_failures`

**Severity Levels:**
- `info` - All events
- `warning` - Warning, error, critical
- `error` - Error, critical
- `critical` - Only critical

**Response:** `201 Created`

```json
{
  "id": 4,
  "name": "Production Snapshot Failures",
  "description": "Alert when production snapshots fail",
  "is_active": true,
  "event_type": "snapshot_failure",
  "conditions": null,
  "severity": "error",
  "notification_channel_ids": "[1,2,3]",
  "throttle_minutes": 30,
  "last_triggered_at": null,
  "created_at": "2026-06-11T09:45:00Z",
  "updated_at": "2026-06-11T09:45:00Z"
}
```

---

### Update Rule

Update an existing alert rule.

```http
PUT /api/notifications/rules/{id}
Authorization: Bearer {token}
Content-Type: application/json
```

**Path Parameters:**
- `id` (integer, required) - Rule ID

**Request Body:**

```json
{
  "name": "Updated Rule Name",
  "is_active": false,
  "throttle_minutes": 120
}
```

**Response:** `200 OK`

```json
{
  "id": 1,
  "name": "Updated Rule Name",
  "description": "Alert on any snapshot failure",
  "is_active": false,
  "event_type": "snapshot_failure",
  "conditions": null,
  "severity": "error",
  "notification_channel_ids": "[1,2]",
  "throttle_minutes": 120,
  "last_triggered_at": "2026-06-11T08:30:00Z",
  "created_at": "2026-06-10T10:00:00Z",
  "updated_at": "2026-06-11T09:50:00Z"
}
```

---

### Delete Rule

Delete an alert rule.

```http
DELETE /api/notifications/rules/{id}
Authorization: Bearer {token}
```

**Path Parameters:**
- `id` (integer, required) - Rule ID

**Response:** `204 No Content`

---

## Notification History API

### List Notification History

Retrieve notification history with optional filters.

```http
GET /api/notifications/history?status=failed&limit=50&offset=0
Authorization: Bearer {token}
```

**Query Parameters:**

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| channel_id | integer | Filter by channel ID | 1 |
| status | string | Filter by status | sent, failed, pending |
| event_type | string | Filter by event type | snapshot_failure |
| severity | string | Filter by severity | error |
| from_date | string | Start date (ISO 8601) | 2026-06-01T00:00:00Z |
| to_date | string | End date (ISO 8601) | 2026-06-11T23:59:59Z |
| limit | integer | Results per page (default: 50) | 100 |
| offset | integer | Pagination offset (default: 0) | 50 |

**Response:** `200 OK`

```json
[
  {
    "id": 123,
    "alert_rule_id": 1,
    "rule_name": "Critical Snapshot Failures",
    "notification_channel_id": 2,
    "channel_name": "Slack Alerts",
    "event_type": "snapshot_failure",
    "severity": "error",
    "message": "Snapshot failed for VG_Production",
    "event_details": "{\"system\":\"FlashSystem-01\",\"volume_group\":\"VG_Production\",\"error\":\"Connection timeout\"}",
    "status": "sent",
    "error_message": null,
    "sent_at": "2026-06-11T08:30:00Z",
    "created_at": "2026-06-11T08:30:00Z"
  },
  {
    "id": 124,
    "alert_rule_id": 1,
    "rule_name": "Critical Snapshot Failures",
    "notification_channel_id": 1,
    "channel_name": "IT Team Email",
    "event_type": "snapshot_failure",
    "severity": "error",
    "message": "Snapshot failed for VG_Production",
    "event_details": "{\"system\":\"FlashSystem-01\",\"volume_group\":\"VG_Production\",\"error\":\"Connection timeout\"}",
    "status": "failed",
    "error_message": "SMTP authentication failed",
    "sent_at": null,
    "created_at": "2026-06-11T08:30:00Z"
  }
]
```

**Status Values:**
- `sent` - Successfully sent
- `failed` - Failed to send
- `throttled` - Suppressed due to throttling
- `pending` - Queued for sending

---

## Test Notification API

### Send Test Notification

Send a test notification without creating a rule.

```http
POST /api/notifications/test
Authorization: Bearer {token}
Content-Type: application/json
```

**Request Body:**

```json
{
  "channel_id": 1,
  "message": "This is a test notification",
  "severity": "info"
}
```

**Response:** `200 OK`

```json
{
  "success": true,
  "message": "Test notification sent successfully"
}
```

**Error Response:** `400 Bad Request`

```json
{
  "success": false,
  "error": "Channel not found"
}
```

---

## Error Codes

### HTTP Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created successfully |
| 204 | No Content | Resource deleted successfully |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 409 | Conflict | Resource already exists |
| 500 | Internal Server Error | Server error |

### Error Response Format

```json
{
  "error": "Detailed error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| INVALID_CONFIG | Channel configuration invalid | Check required fields |
| CHANNEL_NOT_FOUND | Channel ID doesn't exist | Verify channel ID |
| RULE_NOT_FOUND | Rule ID doesn't exist | Verify rule ID |
| AUTH_FAILED | Authentication failed | Check credentials |
| SMTP_ERROR | Email sending failed | Verify SMTP settings |
| WEBHOOK_ERROR | Webhook request failed | Check webhook URL |
| THROTTLED | Notification throttled | Wait for throttle period |

---

## Rate Limiting

### Limits

- **API Requests:** 100 requests per minute per user
- **Test Notifications:** 10 per minute per channel
- **Notification Sending:** No limit (controlled by throttling)

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1686480000
```

### Rate Limit Exceeded

**Response:** `429 Too Many Requests`

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

---

## Webhooks

### Outgoing Webhook Format

When using webhook channels, the following payload is sent:

```json
{
  "event_type": "snapshot_failure",
  "severity": "error",
  "timestamp": "2026-06-11T08:30:00Z",
  "message": "Snapshot failed for VG_Production",
  "details": {
    "system_id": 1,
    "system_name": "FlashSystem-01",
    "system_ip": "192.168.1.100",
    "volume_group_id": 5,
    "volume_group_name": "VG_Production",
    "schedule_id": 12,
    "schedule_name": "Hourly Backup",
    "snapshot_name": "snap_VG_Production_20260611_083000",
    "error": "Connection timeout",
    "retention_days": 7
  }
}
```

### Webhook Retry Logic

- **Initial Attempt:** Immediate
- **Retry 1:** After 5 seconds
- **Retry 2:** After 15 seconds
- **Retry 3:** After 45 seconds
- **Max Retries:** 3

### Webhook Timeout

- **Connection Timeout:** 10 seconds
- **Response Timeout:** 30 seconds

### Expected Response

Webhooks should return:
- **Status Code:** 200-299 (success)
- **Content:** Any (ignored)

---

## Examples

### Complete Workflow Example

#### 1. Create Email Channel

```bash
curl -X POST http://localhost:8090/api/notifications/channels \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Operations Team",
    "type": "email",
    "config": {
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "alerts@company.com",
      "password": "app_password",
      "from_address": "noreply@company.com",
      "to_addresses": "ops@company.com",
      "use_tls": true
    }
  }'
```

#### 2. Test Channel

```bash
curl -X POST http://localhost:8090/api/notifications/channels/1/test \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 3. Create Alert Rule

```bash
curl -X POST http://localhost:8090/api/notifications/rules \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Snapshot Failure Alerts",
    "is_active": true,
    "event_type": "snapshot_failure",
    "severity": "error",
    "notification_channel_ids": "[1]",
    "throttle_minutes": 60
  }'
```

#### 4. View History

```bash
curl -X GET "http://localhost:8090/api/notifications/history?limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## SDK Examples

### JavaScript/TypeScript

```typescript
import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8090/api',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

// Create channel
const channel = await api.post('/notifications/channels', {
  name: 'Slack Alerts',
  type: 'slack',
  config: {
    webhook_url: 'https://hooks.slack.com/services/...',
    channel: '#alerts'
  }
});

// Create rule
const rule = await api.post('/notifications/rules', {
  name: 'Critical Failures',
  is_active: true,
  event_type: 'snapshot_failure',
  severity: 'critical',
  notification_channel_ids: `[${channel.data.id}]`,
  throttle_minutes: 30
});

// Get history
const history = await api.get('/notifications/history', {
  params: { status: 'failed', limit: 50 }
});
```

### Python

```python
import requests

BASE_URL = 'http://localhost:8090/api'
headers = {
    'Authorization': f'Bearer {token}',
    'Content-Type': 'application/json'
}

# Create channel
response = requests.post(
    f'{BASE_URL}/notifications/channels',
    headers=headers,
    json={
        'name': 'Email Alerts',
        'type': 'email',
        'config': {
            'smtp_host': 'smtp.gmail.com',
            'smtp_port': 587,
            'username': 'alerts@company.com',
            'password': 'password',
            'from_address': 'noreply@company.com',
            'to_addresses': 'admin@company.com',
            'use_tls': True
        }
    }
)
channel = response.json()

# Test channel
test_response = requests.post(
    f'{BASE_URL}/notifications/channels/{channel["id"]}/test',
    headers=headers
)
print(test_response.json())
```

---

## Changelog

### Version 1.0 (2026-06-11)

- Initial release
- Notification channels API
- Alert rules API
- Notification history API
- Test notification API

---

**API Version:** 1.0  
**Last Updated:** June 11, 2026  
**For:** IBM Storage Virtualize Snapshot Manager

---

* - Your AI Software Engineer*