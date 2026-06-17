package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
)

// SlackChannel implements the Channel interface for Slack webhook notifications
type SlackChannel struct {
	id     int
	name   string
	config *notification.SlackConfig
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel(id int, name string, config *notification.SlackConfig) *SlackChannel {
	return &SlackChannel{
		id:     id,
		name:   name,
		config: config,
	}
}

// Type returns the channel type
func (s *SlackChannel) Type() notification.ChannelType {
	return notification.ChannelTypeSlack
}

// Send sends a Slack notification
func (s *SlackChannel) Send(ctx context.Context, notif *notification.Notification) error {
	message := s.buildSlackMessage(notif)

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	return nil
}

// Test tests the Slack channel configuration
func (s *SlackChannel) Test(ctx context.Context) error {
	message := map[string]interface{}{
		"text": "Test notification from IBM Storage Virtualize Snapshot Manager",
		"blocks": []map[string]interface{}{
			{
				"type": "header",
				"text": map[string]string{
					"type": "plain_text",
					"text": "✅ Test Notification",
				},
			},
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("This is a test message from the notification channel: *%s*\n\nIf you received this message, your Slack integration is working correctly.", s.name),
				},
			},
			{
				"type": "context",
				"elements": []map[string]string{
					{
						"type": "mrkdwn",
						"text": fmt.Sprintf("Sent at: %s", time.Now().Format(time.RFC1123)),
					},
				},
			},
		},
	}

	if s.config.Username != "" {
		message["username"] = s.config.Username
	}
	if s.config.IconEmoji != "" {
		message["icon_emoji"] = s.config.IconEmoji
	}
	if s.config.Channel != "" {
		message["channel"] = s.config.Channel
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal test message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send test message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned status %d", resp.StatusCode)
	}

	return nil
}

// buildSlackMessage builds a Slack message from a notification
func (s *SlackChannel) buildSlackMessage(notif *notification.Notification) map[string]interface{} {
	event := notif.Event

	// Determine emoji based on severity
	emoji := s.getSeverityEmoji(event.Severity)

	// Build text for fallback
	text := fmt.Sprintf("%s %s - %s", emoji, event.Type, event.Message)

	// Build blocks for rich formatting
	blocks := []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": fmt.Sprintf("%s %s", emoji, s.formatEventType(event.Type)),
			},
		},
	}

	// Add main details section
	fields := []map[string]string{}
	if event.SystemName != nil {
		fields = append(fields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*System:*\n%s", *event.SystemName),
		})
	}
	if event.VolumeGroupName != nil {
		fields = append(fields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Volume Group:*\n%s", *event.VolumeGroupName),
		})
	}
	if event.ScheduleName != nil {
		fields = append(fields, map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Schedule:*\n%s", *event.ScheduleName),
		})
	}
	fields = append(fields, map[string]string{
		"type": "mrkdwn",
		"text": fmt.Sprintf("*Time:*\n%s", event.Timestamp.Format(time.RFC1123)),
	})
	fields = append(fields, map[string]string{
		"type": "mrkdwn",
		"text": fmt.Sprintf("*Severity:*\n%s", strings.ToUpper(string(event.Severity))),
	})

	if len(fields) > 0 {
		blocks = append(blocks, map[string]interface{}{
			"type":   "section",
			"fields": fields,
		})
	}

	// Add message section
	blocks = append(blocks, map[string]interface{}{
		"type": "section",
		"text": map[string]string{
			"type": "mrkdwn",
			"text": fmt.Sprintf("*Message:*\n%s", event.Message),
		},
	})

	// Add details if present
	if len(event.Details) > 0 {
		detailsText := "*Additional Details:*\n"
		for key, value := range event.Details {
			detailsText += fmt.Sprintf("• *%s:* %v\n", key, value)
		}
		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": detailsText,
			},
		})
	}

	// Add context
	blocks = append(blocks, map[string]interface{}{
		"type": "context",
		"elements": []map[string]string{
			{
				"type": "mrkdwn",
				"text": fmt.Sprintf("Notification from IBM Storage Virtualize Snapshot Manager • Channel: %s", s.name),
			},
		},
	})

	message := map[string]interface{}{
		"text":   text,
		"blocks": blocks,
	}

	// Add optional configuration
	if s.config.Username != "" {
		message["username"] = s.config.Username
	}
	if s.config.IconEmoji != "" {
		message["icon_emoji"] = s.config.IconEmoji
	}
	if s.config.Channel != "" {
		message["channel"] = s.config.Channel
	}

	// Add mentions if configured
	if len(s.config.MentionUsers) > 0 || s.config.MentionChannel {
		mentionText := ""
		if s.config.MentionChannel {
			mentionText = "<!channel> "
		}
		for _, user := range s.config.MentionUsers {
			mentionText += fmt.Sprintf("<@%s> ", user)
		}
		if mentionText != "" {
			blocks = append([]map[string]interface{}{
				{
					"type": "section",
					"text": map[string]string{
						"type": "mrkdwn",
						"text": mentionText,
					},
				},
			}, blocks...)
			message["blocks"] = blocks
		}
	}

	return message
}

// getSeverityEmoji returns an emoji for the severity level
func (s *SlackChannel) getSeverityEmoji(severity notification.Severity) string {
	switch severity {
	case notification.SeverityInfo:
		return "ℹ️"
	case notification.SeverityWarning:
		return "⚠️"
	case notification.SeverityError:
		return "❌"
	case notification.SeverityCritical:
		return "🚨"
	default:
		return "📢"
	}
}

// formatEventType formats the event type for display
func (s *SlackChannel) formatEventType(eventType notification.EventType) string {
	str := string(eventType)
	str = strings.ReplaceAll(str, "_", " ")
	words := strings.Split(str, " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

//
