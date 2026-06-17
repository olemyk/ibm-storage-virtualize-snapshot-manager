package providers

import (
	"encoding/json"
	"fmt"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
)

// RegisterFactory registers the CreateChannel factory function with the notification package
// This must be called during initialization to set up the factory
func RegisterFactory() {
	notification.RegisterChannelFactory(CreateChannel)
}

// CreateChannel creates a channel instance based on the notification channel configuration
func CreateChannel(nc *notification.NotificationChannel, encryptionKey string) (notification.Channel, error) {
	switch nc.Type {
	case notification.ChannelTypeEmail:
		var config notification.EmailConfig
		if err := json.Unmarshal([]byte(nc.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse email config: %w", err)
		}

		// Decrypt password
		if config.Password != "" {
			decrypted, err := crypto.Decrypt(config.Password, encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt email password: %w", err)
			}
			config.Password = decrypted
		}

		return NewEmailChannel(nc.ID, nc.Name, &config), nil

	case notification.ChannelTypeSNMP:
		var config notification.SNMPConfig
		if err := json.Unmarshal([]byte(nc.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse SNMP config: %w", err)
		}
		return NewSNMPChannel(nc.ID, nc.Name, &config), nil

	case notification.ChannelTypeSlack:
		var config notification.SlackConfig
		if err := json.Unmarshal([]byte(nc.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse Slack config: %w", err)
		}

		// Decrypt webhook URL
		if config.WebhookURL != "" {
			decrypted, err := crypto.Decrypt(config.WebhookURL, encryptionKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt Slack webhook URL: %w", err)
			}
			config.WebhookURL = decrypted
		}

		return NewSlackChannel(nc.ID, nc.Name, &config), nil

	case notification.ChannelTypeWebhook:
		var config notification.WebhookConfig
		if err := json.Unmarshal([]byte(nc.Config), &config); err != nil {
			return nil, fmt.Errorf("failed to parse webhook config: %w", err)
		}
		return NewWebhookChannel(nc.ID, nc.Name, &config), nil

	default:
		return nil, fmt.Errorf("unsupported channel type: %s", nc.Type)
	}
}

//
