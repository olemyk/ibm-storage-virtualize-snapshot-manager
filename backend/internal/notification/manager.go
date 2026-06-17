package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/db"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
)

// Manager manages notification channels and sending notifications
type Manager struct {
	db            *db.DB
	encryptionKey string
	channels      map[int]Channel
	mu            sync.RWMutex
}

// NewManager creates a new notification manager
func NewManager(database *db.DB, encryptionKey string) *Manager {
	return &Manager{
		db:            database,
		encryptionKey: encryptionKey,
		channels:      make(map[int]Channel),
	}
}

// LoadChannels loads all active channels from the database
func (m *Manager) LoadChannels() error {
	query := `SELECT id, name, type, is_active, config, created_at, updated_at 
	          FROM notification_channels WHERE is_active = TRUE`

	rows, err := m.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.channels = make(map[int]Channel)

	for rows.Next() {
		var nc NotificationChannel
		if err := rows.Scan(&nc.ID, &nc.Name, &nc.Type, &nc.IsActive, &nc.Config, &nc.CreatedAt, &nc.UpdatedAt); err != nil {
			return fmt.Errorf("failed to scan channel: %w", err)
		}

		channel, err := m.createChannelWithFactory(&nc)
		if err != nil {
			return fmt.Errorf("failed to create channel %d: %w", nc.ID, err)
		}

		m.channels[nc.ID] = channel
	}

	return nil
}

// GetChannel returns a channel by ID
func (m *Manager) GetChannel(id int) (Channel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channel, exists := m.channels[id]
	if !exists {
		return nil, fmt.Errorf("channel %d not found", id)
	}
	return channel, nil
}

// GetChannelByDBID loads a channel from database by ID
func (m *Manager) GetChannelByDBID(id int) (*NotificationChannel, error) {
	query := `SELECT id, name, type, is_active, config, created_at, updated_at 
	          FROM notification_channels WHERE id = $1`

	var nc NotificationChannel
	err := m.db.QueryRow(query, id).Scan(
		&nc.ID, &nc.Name, &nc.Type, &nc.IsActive, &nc.Config, &nc.CreatedAt, &nc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return &nc, nil
}

// createChannel creates a channel instance based on type
// This is a placeholder - actual channel creation will be done via dependency injection
func (m *Manager) createChannel(nc *NotificationChannel) (Channel, error) {
	// TODO: Use factory pattern or dependency injection to create channels
	// For now, return an error to avoid import cycle
	return nil, fmt.Errorf("channel creation not yet implemented - use RegisterChannelFactory")
}

// ChannelFactory is a function that creates a channel from a NotificationChannel
type ChannelFactory func(*NotificationChannel, string) (Channel, error)

var channelFactory ChannelFactory

// RegisterChannelFactory registers the channel factory function
func RegisterChannelFactory(factory ChannelFactory) {
	channelFactory = factory
}

// createChannelWithFactory creates a channel using the registered factory
func (m *Manager) createChannelWithFactory(nc *NotificationChannel) (Channel, error) {
	if channelFactory == nil {
		return nil, fmt.Errorf("channel factory not registered")
	}
	return channelFactory(nc, m.encryptionKey)
}

// CreateChannel creates a new notification channel in the database
func (m *Manager) CreateChannel(name string, channelType ChannelType, config interface{}) (*NotificationChannel, error) {
	// Encrypt sensitive fields based on channel type
	configJSON, err := m.encryptAndMarshalConfig(channelType, config)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO notification_channels (name, type, config) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err = m.db.QueryRow(query, name, channelType, configJSON).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Reload channels
	if err = m.LoadChannels(); err != nil {
		return nil, fmt.Errorf("failed to reload channels: %w", err)
	}

	return m.GetChannelByDBID(int(id))
}

// UpdateChannel updates an existing notification channel
func (m *Manager) UpdateChannel(id int, name string, channelType ChannelType, config interface{}, isActive bool) error {
	// Encrypt sensitive fields based on channel type
	configJSON, err := m.encryptAndMarshalConfig(channelType, config)
	if err != nil {
		return err
	}

	query := `UPDATE notification_channels SET name = $1, type = $2, config = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	_, err = m.db.Exec(query, name, channelType, configJSON, isActive, id)
	if err != nil {
		return fmt.Errorf("failed to update channel: %w", err)
	}

	// Reload channels
	return m.LoadChannels()
}

// DeleteChannel deletes a notification channel
func (m *Manager) DeleteChannel(id int) error {
	query := `DELETE FROM notification_channels WHERE id = $1`
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	// Remove from memory
	m.mu.Lock()
	delete(m.channels, id)
	m.mu.Unlock()

	return nil
}

// ListChannels returns all notification channels
func (m *Manager) ListChannels() ([]*NotificationChannel, error) {
	query := `SELECT id, name, type, is_active, config, created_at, updated_at 
	          FROM notification_channels ORDER BY created_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}
	defer rows.Close()

	var channels []*NotificationChannel
	for rows.Next() {
		var nc NotificationChannel
		if err := rows.Scan(&nc.ID, &nc.Name, &nc.Type, &nc.IsActive, &nc.Config, &nc.CreatedAt, &nc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan channel: %w", err)
		}
		channels = append(channels, &nc)
	}

	return channels, nil
}

// TestChannel tests a notification channel
func (m *Manager) TestChannel(id int) error {
	channel, err := m.GetChannel(id)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return channel.Test(ctx)
}

// SendNotification sends a notification through specified channels
func (m *Manager) SendNotification(ctx context.Context, event *Event, channelIDs []int) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastErr error
	successCount := 0

	for _, channelID := range channelIDs {
		channel, exists := m.channels[channelID]
		if !exists {
			lastErr = fmt.Errorf("channel %d not found", channelID)
			continue
		}

		notif := &Notification{
			Event: event,
		}

		if err := channel.Send(ctx, notif); err != nil {
			lastErr = err
			// Log error but continue with other channels
			m.logNotificationHistory(channelID, event, StatusFailed, err.Error())
			continue
		}

		successCount++
		m.logNotificationHistory(channelID, event, StatusSent, "")
	}

	if successCount == 0 && lastErr != nil {
		return fmt.Errorf("all notifications failed, last error: %w", lastErr)
	}

	return nil
}

// encryptAndMarshalConfig encrypts sensitive fields and marshals config to JSON
func (m *Manager) encryptAndMarshalConfig(channelType ChannelType, config interface{}) (string, error) {
	switch channelType {
	case ChannelTypeEmail:
		emailConfig := config.(*EmailConfig)
		if emailConfig.Password != "" {
			encrypted, err := crypto.Encrypt(emailConfig.Password, m.encryptionKey)
			if err != nil {
				return "", fmt.Errorf("failed to encrypt email password: %w", err)
			}
			emailConfig.Password = encrypted
		}
		configBytes, err := json.Marshal(emailConfig)
		if err != nil {
			return "", fmt.Errorf("failed to marshal email config: %w", err)
		}
		return string(configBytes), nil

	case ChannelTypeSlack:
		slackConfig := config.(*SlackConfig)
		if slackConfig.WebhookURL != "" {
			encrypted, err := crypto.Encrypt(slackConfig.WebhookURL, m.encryptionKey)
			if err != nil {
				return "", fmt.Errorf("failed to encrypt Slack webhook URL: %w", err)
			}
			slackConfig.WebhookURL = encrypted
		}
		configBytes, err := json.Marshal(slackConfig)
		if err != nil {
			return "", fmt.Errorf("failed to marshal Slack config: %w", err)
		}
		return string(configBytes), nil

	default:
		configBytes, err := json.Marshal(config)
		if err != nil {
			return "", fmt.Errorf("failed to marshal config: %w", err)
		}
		return string(configBytes), nil
	}
}

// logNotificationHistory logs a notification to the history table
func (m *Manager) logNotificationHistory(channelID int, event *Event, status NotificationStatus, errorMsg string) {
	eventDataBytes, _ := json.Marshal(event)
	eventData := string(eventDataBytes)

	var errorMsgPtr *string
	if errorMsg != "" {
		errorMsgPtr = &errorMsg
	}

	query := `INSERT INTO notification_history (notification_channel_id, event_type, event_data, status, error_message)
	          VALUES ($1, $2, $3, $4, $5)`

	_, err := m.db.Exec(query, channelID, event.Type, eventData, status, errorMsgPtr)
	if err != nil {
		// Log error but don't fail the notification
		fmt.Printf("Failed to log notification history: %v\n", err)
	}
}

// Alert Rule Methods

// ListRules returns all alert rules
func (m *Manager) ListRules() ([]*AlertRule, error) {
	query := `SELECT id, name, description, is_active, event_type, conditions,
	          severity, notification_channel_ids, throttle_minutes,
	          last_triggered_at, created_at, updated_at
	          FROM alert_rules ORDER BY created_at DESC`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	var rules []*AlertRule
	for rows.Next() {
		var rule AlertRule
		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.IsActive,
			&rule.EventType, &rule.Conditions, &rule.Severity,
			&rule.NotificationChannelIDs, &rule.ThrottleMinutes,
			&rule.LastTriggeredAt, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert rule: %w", err)
		}
		rules = append(rules, &rule)
	}

	return rules, nil
}

// GetRule returns an alert rule by ID
func (m *Manager) GetRule(id int) (*AlertRule, error) {
	query := `SELECT id, name, description, is_active, event_type, conditions,
	          severity, notification_channel_ids, throttle_minutes,
	          last_triggered_at, created_at, updated_at
	          FROM alert_rules WHERE id = $1`

	var rule AlertRule
	err := m.db.QueryRow(query, id).Scan(
		&rule.ID, &rule.Name, &rule.Description, &rule.IsActive,
		&rule.EventType, &rule.Conditions, &rule.Severity,
		&rule.NotificationChannelIDs, &rule.ThrottleMinutes,
		&rule.LastTriggeredAt, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert rule: %w", err)
	}

	return &rule, nil
}

// CreateRule creates a new alert rule
func (m *Manager) CreateRule(rule *AlertRule) error {
	// Convert channel IDs to JSON
	channelIDsJSON, err := json.Marshal(rule.NotificationChannelIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal channel IDs: %w", err)
	}

	query := `INSERT INTO alert_rules (name, description, is_active, event_type, conditions,
	          severity, notification_channel_ids, throttle_minutes)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err = m.db.QueryRow(query,
		rule.Name, rule.Description, rule.IsActive, rule.EventType,
		rule.Conditions, rule.Severity, string(channelIDsJSON), rule.ThrottleMinutes,
	).Scan(&rule.ID)
	if err != nil {
		return fmt.Errorf("failed to create alert rule: %w", err)
	}
	return nil
}

// UpdateRule updates an existing alert rule
func (m *Manager) UpdateRule(rule *AlertRule) error {
	// Convert channel IDs to JSON
	channelIDsJSON, err := json.Marshal(rule.NotificationChannelIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal channel IDs: %w", err)
	}

	query := `UPDATE alert_rules SET name = $1, description = $2, is_active = $3,
	          event_type = $4, conditions = $5, severity = $6,
	          notification_channel_ids = $7, throttle_minutes = $8,
	          updated_at = CURRENT_TIMESTAMP
	          WHERE id = $9`

	_, err = m.db.Exec(query,
		rule.Name, rule.Description, rule.IsActive, rule.EventType,
		rule.Conditions, rule.Severity, string(channelIDsJSON),
		rule.ThrottleMinutes, rule.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update alert rule: %w", err)
	}

	return nil
}

// DeleteRule deletes an alert rule
func (m *Manager) DeleteRule(id int) error {
	query := `DELETE FROM alert_rules WHERE id = $1`
	_, err := m.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert rule: %w", err)
	}

	return nil
}

// Notification History Methods

// GetHistory returns notification history with optional filters
func (m *Manager) GetHistory(channelID *int, status *string, startTime, endTime *time.Time, limit, offset int) ([]*NotificationHistory, error) {
	query := `SELECT id, alert_rule_id, notification_channel_id, event_type,
	          event_data, status, error_message, sent_at
	          FROM notification_history WHERE 1=1`

	args := []interface{}{}

	paramCount := 1
	if channelID != nil {
		query += fmt.Sprintf(` AND notification_channel_id = $%d`, paramCount)
		args = append(args, *channelID)
		paramCount++
	}

	if status != nil {
		query += fmt.Sprintf(` AND status = $%d`, paramCount)
		args = append(args, *status)
		paramCount++
	}

	if startTime != nil {
		query += fmt.Sprintf(` AND sent_at >= $%d`, paramCount)
		args = append(args, *startTime)
		paramCount++
	}

	if endTime != nil {
		query += fmt.Sprintf(` AND sent_at <= $%d`, paramCount)
		args = append(args, *endTime)
		paramCount++
	}

	query += fmt.Sprintf(` ORDER BY sent_at DESC LIMIT $%d OFFSET $%d`, paramCount, paramCount+1)
	args = append(args, limit, offset)

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notification history: %w", err)
	}
	defer rows.Close()

	var history []*NotificationHistory
	for rows.Next() {
		var h NotificationHistory
		err := rows.Scan(
			&h.ID, &h.AlertRuleID, &h.NotificationChannelID,
			&h.EventType, &h.EventData, &h.Status,
			&h.ErrorMessage, &h.SentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification history: %w", err)
		}
		history = append(history, &h)
	}

	return history, nil
}

//
