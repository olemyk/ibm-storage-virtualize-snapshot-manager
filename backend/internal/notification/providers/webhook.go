package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
)

// WebhookChannel implements the Channel interface for generic webhook notifications
type WebhookChannel struct {
	id     int
	name   string
	config *notification.WebhookConfig
}

// NewWebhookChannel creates a new webhook channel
func NewWebhookChannel(id int, name string, config *notification.WebhookConfig) *WebhookChannel {
	return &WebhookChannel{
		id:     id,
		name:   name,
		config: config,
	}
}

// Type returns the channel type
func (w *WebhookChannel) Type() notification.ChannelType {
	return notification.ChannelTypeWebhook
}

// Send sends a webhook notification
func (w *WebhookChannel) Send(ctx context.Context, notif *notification.Notification) error {
	payload := w.buildPayload(notif)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	return w.sendWithRetry(ctx, jsonPayload)
}

// Test tests the webhook channel configuration
func (w *WebhookChannel) Test(ctx context.Context) error {
	payload := map[string]interface{}{
		"test":      true,
		"message":   "Test notification from IBM Storage Virtualize Snapshot Manager",
		"channel":   w.name,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal test payload: %w", err)
	}

	return w.sendRequest(ctx, jsonPayload)
}

// buildPayload builds the webhook payload from a notification
func (w *WebhookChannel) buildPayload(notif *notification.Notification) map[string]interface{} {
	event := notif.Event

	payload := map[string]interface{}{
		"event_type": event.Type,
		"severity":   event.Severity,
		"timestamp":  event.Timestamp.Format(time.RFC3339),
		"message":    event.Message,
	}

	if event.SystemID != nil {
		payload["system_id"] = *event.SystemID
	}
	if event.SystemName != nil {
		payload["system_name"] = *event.SystemName
	}
	if event.VolumeGroupID != nil {
		payload["volume_group_id"] = *event.VolumeGroupID
	}
	if event.VolumeGroupName != nil {
		payload["volume_group_name"] = *event.VolumeGroupName
	}
	if event.ScheduleID != nil {
		payload["schedule_id"] = *event.ScheduleID
	}
	if event.ScheduleName != nil {
		payload["schedule_name"] = *event.ScheduleName
	}
	if event.ExecutionID != nil {
		payload["execution_id"] = *event.ExecutionID
	}

	if len(event.Details) > 0 {
		payload["details"] = event.Details
	}

	// Add metadata
	payload["notification_channel"] = w.name
	payload["notification_channel_id"] = w.id

	return payload
}

// sendWithRetry sends the request with retry logic
func (w *WebhookChannel) sendWithRetry(ctx context.Context, payload []byte) error {
	var lastErr error

	retryCount := w.config.RetryCount
	if retryCount <= 0 {
		retryCount = 1 // At least one attempt
	}

	retryDelay := time.Duration(w.config.RetryDelay) * time.Second
	if retryDelay <= 0 {
		retryDelay = 5 * time.Second
	}

	for attempt := 0; attempt < retryCount; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		err := w.sendRequest(ctx, payload)
		if err == nil {
			return nil
		}

		lastErr = err

		// Exponential backoff for subsequent retries
		retryDelay *= 2
	}

	return fmt.Errorf("failed after %d attempts: %w", retryCount, lastErr)
}

// sendRequest sends a single HTTP request
func (w *WebhookChannel) sendRequest(ctx context.Context, payload []byte) error {
	method := w.config.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, w.config.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set default content type
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}

	timeout := time.Duration(w.config.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

//
