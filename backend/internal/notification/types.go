package notification

import (
	"context"
	"time"
)

// Channel represents a notification channel interface
type Channel interface {
	Send(ctx context.Context, notification *Notification) error
	Test(ctx context.Context) error
	Type() ChannelType
}

// ChannelType represents the type of notification channel
type ChannelType string

const (
	ChannelTypeEmail   ChannelType = "email"
	ChannelTypeSNMP    ChannelType = "snmp"
	ChannelTypeSlack   ChannelType = "slack"
	ChannelTypeWebhook ChannelType = "webhook"
)

// EventType represents the type of event that triggers a notification
type EventType string

const (
	EventSnapshotSuccess      EventType = "snapshot_success"
	EventSnapshotFailure      EventType = "snapshot_failure"
	EventSnapshotWarning      EventType = "snapshot_warning"
	EventSystemConnectionLost EventType = "system_connection_lost"
	EventSchedulerError       EventType = "scheduler_error"
	EventConsecutiveFailures  EventType = "consecutive_failures"
)

// Severity represents the severity level of an event
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusSent      NotificationStatus = "sent"
	StatusFailed    NotificationStatus = "failed"
	StatusThrottled NotificationStatus = "throttled"
	StatusPending   NotificationStatus = "pending"
)

// NotificationChannel represents a notification channel configuration
type NotificationChannel struct {
	ID        int         `json:"id" db:"id"`
	Name      string      `json:"name" db:"name"`
	Type      ChannelType `json:"type" db:"type"`
	IsActive  bool        `json:"is_active" db:"is_active"`
	Config    string      `json:"config" db:"config"` // JSON string (encrypted for sensitive data)
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}

// AlertRule represents an alert rule configuration
type AlertRule struct {
	ID                     int        `json:"id" db:"id"`
	Name                   string     `json:"name" db:"name"`
	Description            *string    `json:"description,omitempty" db:"description"`
	IsActive               bool       `json:"is_active" db:"is_active"`
	EventType              EventType  `json:"event_type" db:"event_type"`
	Conditions             *string    `json:"conditions,omitempty" db:"conditions"` // JSON string
	Severity               Severity   `json:"severity" db:"severity"`
	NotificationChannelIDs string     `json:"notification_channel_ids" db:"notification_channel_ids"` // JSON array
	ThrottleMinutes        int        `json:"throttle_minutes" db:"throttle_minutes"`
	LastTriggeredAt        *time.Time `json:"last_triggered_at,omitempty" db:"last_triggered_at"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationHistory represents a notification history record
type NotificationHistory struct {
	ID                    int                `json:"id" db:"id"`
	AlertRuleID           *int               `json:"alert_rule_id,omitempty" db:"alert_rule_id"`
	NotificationChannelID int                `json:"notification_channel_id" db:"notification_channel_id"`
	EventType             EventType          `json:"event_type" db:"event_type"`
	EventData             *string            `json:"event_data,omitempty" db:"event_data"` // JSON string
	Status                NotificationStatus `json:"status" db:"status"`
	ErrorMessage          *string            `json:"error_message,omitempty" db:"error_message"`
	SentAt                time.Time          `json:"sent_at" db:"sent_at"`
}

// Event represents an event that can trigger notifications
type Event struct {
	Type            EventType              `json:"type"`
	Timestamp       time.Time              `json:"timestamp"`
	Severity        Severity               `json:"severity"`
	SystemID        *int                   `json:"system_id,omitempty"`
	SystemName      *string                `json:"system_name,omitempty"`
	VolumeGroupID   *int                   `json:"volume_group_id,omitempty"`
	VolumeGroupName *string                `json:"volume_group_name,omitempty"`
	ScheduleID      *int                   `json:"schedule_id,omitempty"`
	ScheduleName    *string                `json:"schedule_name,omitempty"`
	ExecutionID     *int                   `json:"execution_id,omitempty"`
	Message         string                 `json:"message"`
	Details         map[string]interface{} `json:"details,omitempty"`
}

// Notification represents a notification to be sent
type Notification struct {
	Event        *Event
	Rule         *AlertRule
	Channel      *NotificationChannel
	TemplateData map[string]interface{}
}

// EmailConfig represents email channel configuration
type EmailConfig struct {
	SMTPHost    string   `json:"smtp_host"`
	SMTPPort    int      `json:"smtp_port"`
	Username    string   `json:"username"`
	Password    string   `json:"password_encrypted"` // Encrypted
	From        string   `json:"from"`
	To          []string `json:"to"`
	CC          []string `json:"cc,omitempty"`
	UseTLS      bool     `json:"use_tls"`
	UseSTARTTLS bool     `json:"use_starttls"`
	SkipVerify  bool     `json:"skip_verify"`
}

// SNMPConfig represents SNMP channel configuration
type SNMPConfig struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Version       string `json:"version"` // v2c or v3
	Community     string `json:"community"`
	TrapOID       string `json:"trap_oid"`
	EnterpriseOID string `json:"enterprise_oid"`
	Timeout       int    `json:"timeout"`
}

// SlackConfig represents Slack channel configuration
type SlackConfig struct {
	WebhookURL     string   `json:"webhook_url_encrypted"` // Encrypted
	Channel        string   `json:"channel"`
	Username       string   `json:"username"`
	IconEmoji      string   `json:"icon_emoji,omitempty"`
	MentionUsers   []string `json:"mention_users,omitempty"`
	MentionChannel bool     `json:"mention_channel"`
}

// WebhookConfig represents generic webhook channel configuration
type WebhookConfig struct {
	URL        string            `json:"url"`
	Method     string            `json:"method"` // POST, PUT, PATCH
	Headers    map[string]string `json:"headers,omitempty"`
	Timeout    int               `json:"timeout"`
	RetryCount int               `json:"retry_count"`
	RetryDelay int               `json:"retry_delay"`
}

//
