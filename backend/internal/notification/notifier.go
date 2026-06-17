package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/db"
)

// Notifier is the main orchestrator for the notification system
type Notifier struct {
	db             *db.DB
	encryptionKey  string
	channelManager *Manager
	rules          []*AlertRule // Cache of active rules
	mu             sync.RWMutex
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// NewNotifier creates a new notification orchestrator
func NewNotifier(database *db.DB, encryptionKey string) (*Notifier, error) {
	channelManager := NewManager(database, encryptionKey)

	n := &Notifier{
		db:             database,
		encryptionKey:  encryptionKey,
		channelManager: channelManager,
		rules:          make([]*AlertRule, 0),
		stopCh:         make(chan struct{}),
	}

	// Load channels and rules
	if err := n.loadChannels(); err != nil {
		return nil, fmt.Errorf("failed to load channels: %w", err)
	}

	if err := n.loadRules(); err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	return n, nil
}

// Start begins the notification processing
func (n *Notifier) Start() {
	log.Println("Notification system started")
}

// Stop gracefully shuts down the notification system
func (n *Notifier) Stop() {
	close(n.stopCh)
	n.wg.Wait()
	log.Println("Notification system stopped")
}

// loadChannels loads all enabled notification channels using the manager
func (n *Notifier) loadChannels() error {
	return n.channelManager.LoadChannels()
}

// loadRules loads all enabled alert rules from the database
func (n *Notifier) loadRules() error {
	query := `
		SELECT id, name, description, is_active, event_type, conditions,
		       severity, notification_channel_ids, throttle_minutes,
		       last_triggered_at, created_at, updated_at
		FROM alert_rules
		WHERE is_active = TRUE
		ORDER BY name
	`

	rows, err := n.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query alert rules: %w", err)
	}
	defer rows.Close()

	n.mu.Lock()
	defer n.mu.Unlock()

	n.rules = make([]*AlertRule, 0)

	for rows.Next() {
		var rule AlertRule

		err := rows.Scan(
			&rule.ID, &rule.Name, &rule.Description, &rule.IsActive,
			&rule.EventType, &rule.Conditions, &rule.Severity,
			&rule.NotificationChannelIDs, &rule.ThrottleMinutes,
			&rule.LastTriggeredAt, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to scan alert rule: %w", err)
		}

		n.rules = append(n.rules, &rule)
		log.Printf("Loaded alert rule: %s (event: %s, severity: %s)", rule.Name, rule.EventType, rule.Severity)
	}

	return rows.Err()
}

// ReloadChannels reloads all notification channels from the database
func (n *Notifier) ReloadChannels() error {
	return n.loadChannels()
}

// ReloadRules reloads all alert rules from the database
func (n *Notifier) ReloadRules() error {
	return n.loadRules()
}

// SendEvent processes an event and sends notifications based on matching rules
func (n *Notifier) SendEvent(ctx context.Context, event *Event) error {
	n.mu.RLock()
	matchingRules := n.findMatchingRules(event)
	n.mu.RUnlock()

	if len(matchingRules) == 0 {
		log.Printf("No matching rules for event: %s (severity: %s)", event.Type, event.Severity)
		return nil
	}

	log.Printf("Found %d matching rules for event: %s", len(matchingRules), event.Type)

	// Process each matching rule
	for _, rule := range matchingRules {
		if err := n.processRule(ctx, event, rule); err != nil {
			log.Printf("Failed to process rule %s: %v", rule.Name, err)
			// Continue processing other rules
		}
	}

	return nil
}

// findMatchingRules finds all rules that match the given event
func (n *Notifier) findMatchingRules(event *Event) []*AlertRule {
	matching := make([]*AlertRule, 0)

	for _, rule := range n.rules {
		if n.ruleMatches(rule, event) {
			matching = append(matching, rule)
		}
	}

	return matching
}

// ruleMatches checks if a rule matches the given event
func (n *Notifier) ruleMatches(rule *AlertRule, event *Event) bool {
	// Check if rule is active
	if !rule.IsActive {
		return false
	}

	// Check event type
	if rule.EventType != event.Type {
		return false
	}

	// Check severity level (rule severity is minimum level)
	if !n.severityMatches(rule.Severity, event.Severity) {
		return false
	}

	// Check conditions if specified
	if rule.Conditions != nil && *rule.Conditions != "" {
		if !n.conditionsMatch(rule.Conditions, event) {
			return false
		}
	}

	// Check throttling
	if rule.ThrottleMinutes > 0 {
		if n.isThrottled(rule) {
			log.Printf("Rule %s is throttled", rule.Name)
			return false
		}
	}

	return true
}

// severityMatches checks if event severity meets rule's minimum severity
func (n *Notifier) severityMatches(ruleSeverity, eventSeverity Severity) bool {
	severityLevels := map[Severity]int{
		SeverityInfo:     1,
		SeverityWarning:  2,
		SeverityError:    3,
		SeverityCritical: 4,
	}

	ruleLevel := severityLevels[ruleSeverity]
	eventLevel := severityLevels[eventSeverity]

	return eventLevel >= ruleLevel
}

// conditionsMatch checks if event matches rule conditions (JSON-based filtering)
func (n *Notifier) conditionsMatch(conditions *string, event *Event) bool {
	if conditions == nil || *conditions == "" {
		return true
	}

	// Parse conditions JSON
	var cond map[string]interface{}
	if err := json.Unmarshal([]byte(*conditions), &cond); err != nil {
		log.Printf("Failed to parse conditions: %v", err)
		return false
	}

	// Check system_id filter
	if systemID, ok := cond["system_id"].(float64); ok && event.SystemID != nil {
		if int(systemID) != *event.SystemID {
			return false
		}
	}

	// Check volume_group_id filter
	if vgID, ok := cond["volume_group_id"].(float64); ok && event.VolumeGroupID != nil {
		if int(vgID) != *event.VolumeGroupID {
			return false
		}
	}

	// Check schedule_id filter
	if schedID, ok := cond["schedule_id"].(float64); ok && event.ScheduleID != nil {
		if int(schedID) != *event.ScheduleID {
			return false
		}
	}

	return true
}

// isThrottled checks if a rule has been triggered recently (within throttle period)
func (n *Notifier) isThrottled(rule *AlertRule) bool {
	if rule.LastTriggeredAt == nil {
		return false
	}

	throttleDuration := time.Duration(rule.ThrottleMinutes) * time.Minute
	return time.Since(*rule.LastTriggeredAt) < throttleDuration
}

// processRule processes a single rule for an event
func (n *Notifier) processRule(ctx context.Context, event *Event, rule *AlertRule) error {
	// Parse channel IDs from JSON
	var channelIDs []int
	if err := json.Unmarshal([]byte(rule.NotificationChannelIDs), &channelIDs); err != nil {
		return fmt.Errorf("failed to parse channel IDs: %w", err)
	}

	// Get channels for this rule
	channels := n.getChannelsForRule(channelIDs)
	if len(channels) == 0 {
		log.Printf("No active channels for rule: %s", rule.Name)
		return nil
	}

	// Create notification
	notification := &Notification{
		Event: event,
		Rule:  rule,
	}

	// Send to each channel
	for channelID, channel := range channels {
		n.wg.Add(1)
		go func(chID int, ch Channel, nc *NotificationChannel) {
			defer n.wg.Done()

			// Set channel info in notification
			notification.Channel = nc

			if err := n.sendNotification(ctx, ch, notification, rule.ID, chID); err != nil {
				log.Printf("Failed to send notification via channel %d: %v", chID, err)
			}
		}(channelID, channel, n.getChannelConfig(channelID))
	}

	// Update last triggered time
	if err := n.updateLastTriggered(rule.ID); err != nil {
		log.Printf("Failed to update last triggered time for rule %d: %v", rule.ID, err)
	}

	return nil
}

// getChannelsForRule returns the active channels for given channel IDs
func (n *Notifier) getChannelsForRule(channelIDs []int) map[int]Channel {
	channels := make(map[int]Channel)
	for _, channelID := range channelIDs {
		channel, err := n.channelManager.GetChannel(channelID)
		if err != nil {
			log.Printf("Channel %d not found: %v", channelID, err)
			continue
		}
		channels[channelID] = channel
	}

	return channels
}

// getChannelConfig returns the channel configuration for a given channel ID
func (n *Notifier) getChannelConfig(channelID int) *NotificationChannel {
	nc, err := n.channelManager.GetChannelByDBID(channelID)
	if err != nil {
		log.Printf("Failed to get channel config for %d: %v", channelID, err)
		return nil
	}
	return nc
}

// sendNotification sends a notification via a channel and logs the result
func (n *Notifier) sendNotification(ctx context.Context, channel Channel, notification *Notification, ruleID, channelID int) error {
	startTime := time.Now()

	// Send notification
	err := channel.Send(ctx, notification)

	duration := time.Since(startTime)

	// Log to notification history
	status := StatusSent
	var errorMsg *string
	if err != nil {
		status = StatusFailed
		errStr := err.Error()
		errorMsg = &errStr
	}

	if err := n.logNotification(ruleID, channelID, notification, status, errorMsg, duration); err != nil {
		log.Printf("Failed to log notification: %v", err)
	}

	return err
}

// logNotification logs a notification to the history table
func (n *Notifier) logNotification(ruleID, channelID int, notification *Notification, status NotificationStatus, errorMsg *string, duration time.Duration) error {
	// Serialize event data to JSON
	eventDataJSON, err := json.Marshal(notification.Event)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}
	eventDataStr := string(eventDataJSON)

	query := `
		INSERT INTO notification_history (
			alert_rule_id, notification_channel_id, event_type,
			event_data, status, error_message, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`

	_, err = n.db.Exec(
		query,
		ruleID,
		channelID,
		notification.Event.Type,
		eventDataStr,
		status,
		errorMsg,
	)

	return err
}

// updateLastTriggered updates the last_triggered_at timestamp for a rule
func (n *Notifier) updateLastTriggered(ruleID int) error {
	query := `UPDATE alert_rules SET last_triggered_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := n.db.Exec(query, ruleID)
	return err
}

// SendToChannel sends a notification directly to a specific channel (for testing)
func (n *Notifier) SendToChannel(ctx context.Context, channelID int, event *Event) error {
	// Get the channel
	channel, err := n.channelManager.GetChannel(channelID)
	if err != nil {
		return fmt.Errorf("channel %d not found: %w", channelID, err)
	}

	// Get channel config
	channelConfig := n.getChannelConfig(channelID)
	if channelConfig == nil {
		return fmt.Errorf("failed to get channel config for %d", channelID)
	}

	// Create notification
	notification := &Notification{
		Event:   event,
		Channel: channelConfig,
	}

	// Send notification
	err = channel.Send(ctx, notification)

	// Log to notification history
	status := StatusSent
	var errorMsg *string
	if err != nil {
		status = StatusFailed
		errStr := err.Error()
		errorMsg = &errStr
	}

	// Log without rule ID (direct send)
	if logErr := n.logNotification(0, channelID, notification, status, errorMsg, 0); logErr != nil {
		log.Printf("Failed to log notification: %v", logErr)
	}

	return err
}

//
