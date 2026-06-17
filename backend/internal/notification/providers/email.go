package providers

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
)

// EmailChannel implements the Channel interface for email notifications
type EmailChannel struct {
	id     int
	name   string
	config *notification.EmailConfig
}

// NewEmailChannel creates a new email channel
func NewEmailChannel(id int, name string, config *notification.EmailConfig) *EmailChannel {
	return &EmailChannel{
		id:     id,
		name:   name,
		config: config,
	}
}

// Type returns the channel type
func (e *EmailChannel) Type() notification.ChannelType {
	return notification.ChannelTypeEmail
}

// Send sends an email notification
func (e *EmailChannel) Send(ctx context.Context, notif *notification.Notification) error {
	// Generate email subject and body
	subject, body, err := e.generateEmail(notif)
	if err != nil {
		return fmt.Errorf("failed to generate email: %w", err)
	}

	// Build email message
	msg := e.buildMessage(subject, body)

	// Send email
	if err := e.sendEmail(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Test tests the email channel configuration
func (e *EmailChannel) Test(ctx context.Context) error {
	subject := "Test Notification from IBM Storage Virtualize Snapshot Manager"
	body := `<html>
<body>
<h2>Test Notification</h2>
<p>This is a test email from the IBM Storage Virtualize Snapshot Manager notification system.</p>
<p>If you received this email, your email notification channel is configured correctly.</p>
<p><strong>Channel:</strong> ` + e.name + `</p>
<p><strong>Time:</strong> ` + time.Now().Format(time.RFC1123) + `</p>
</body>
</html>`

	msg := e.buildMessage(subject, body)
	return e.sendEmail(msg)
}

// generateEmail generates email subject and body from notification
func (e *EmailChannel) generateEmail(notif *notification.Notification) (string, string, error) {
	event := notif.Event

	// Generate subject
	subject := fmt.Sprintf("[%s] %s - %s",
		strings.ToUpper(string(event.Severity)),
		event.Type,
		e.getSystemInfo(event))

	// Generate HTML body
	bodyTemplate := `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .alert { padding: 20px; border-radius: 5px; margin: 20px 0; }
        .info { background-color: #d1ecf1; border: 1px solid #bee5eb; }
        .warning { background-color: #fff3cd; border: 1px solid #ffc107; }
        .error { background-color: #f8d7da; border: 1px solid #f5c6cb; }
        .critical { background-color: #f8d7da; border: 2px solid #dc3545; }
        .details { background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin-top: 15px; }
        .label { font-weight: bold; color: #555; }
        h2 { margin-top: 0; color: #333; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <div class="alert {{.Severity}}">
        <h2>{{.EventTypeDisplay}}</h2>
        
        <div class="details">
            {{if .SystemName}}<p><span class="label">System:</span> {{.SystemName}}</p>{{end}}
            {{if .VolumeGroupName}}<p><span class="label">Volume Group:</span> {{.VolumeGroupName}}</p>{{end}}
            {{if .ScheduleName}}<p><span class="label">Schedule:</span> {{.ScheduleName}}</p>{{end}}
            <p><span class="label">Time:</span> {{.Timestamp}}</p>
            <p><span class="label">Severity:</span> {{.SeverityDisplay}}</p>
        </div>
        
        <p><span class="label">Message:</span> {{.Message}}</p>
        
        {{if .ErrorMessage}}
        <div class="details">
            <p><span class="label">Error Details:</span></p>
            <pre>{{.ErrorMessage}}</pre>
        </div>
        {{end}}
        
        {{if .Details}}
        <div class="details">
            <p><span class="label">Additional Details:</span></p>
            {{range $key, $value := .Details}}
            <p><span class="label">{{$key}}:</span> {{$value}}</p>
            {{end}}
        </div>
        {{end}}
    </div>
    
    <div class="footer">
        <p>This is an automated notification from IBM Storage Virtualize Snapshot Manager.</p>
        <p>Notification Channel: {{.ChannelName}}</p>
    </div>
</body>
</html>`

	tmpl, err := template.New("email").Parse(bodyTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := map[string]interface{}{
		"Severity":         string(event.Severity),
		"SeverityDisplay":  strings.ToUpper(string(event.Severity)),
		"EventType":        event.Type,
		"EventTypeDisplay": e.formatEventType(event.Type),
		"SystemName":       e.stringOrEmpty(event.SystemName),
		"VolumeGroupName":  e.stringOrEmpty(event.VolumeGroupName),
		"ScheduleName":     e.stringOrEmpty(event.ScheduleName),
		"Timestamp":        event.Timestamp.Format(time.RFC1123),
		"Message":          event.Message,
		"ErrorMessage":     "",
		"Details":          event.Details,
		"ChannelName":      e.name,
	}

	// Add error message if present in details
	if errMsg, ok := event.Details["error"]; ok {
		data["ErrorMessage"] = errMsg
	}

	var bodyBuf bytes.Buffer
	if err := tmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	return subject, bodyBuf.String(), nil
}

// buildMessage builds the email message with headers
func (e *EmailChannel) buildMessage(subject, body string) []byte {
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", e.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.config.To, ", ")))
	if len(e.config.CC) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(e.config.CC, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")

	// Body
	msg.WriteString(body)

	return msg.Bytes()
}

// sendEmail sends the email via SMTP
func (e *EmailChannel) sendEmail(msg []byte) error {
	addr := fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort)

	// Prepare authentication
	var auth smtp.Auth
	if e.config.Username != "" && e.config.Password != "" {
		auth = smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)
	}

	// Combine To and CC recipients
	recipients := append([]string{}, e.config.To...)
	recipients = append(recipients, e.config.CC...)

	// Send with or without TLS
	if e.config.UseTLS {
		return e.sendWithTLS(addr, auth, recipients, msg)
	}

	return smtp.SendMail(addr, auth, e.config.From, recipients, msg)
}

// sendWithTLS sends email with TLS connection
func (e *EmailChannel) sendWithTLS(addr string, auth smtp.Auth, recipients []string, msg []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName:         e.config.SMTPHost,
		InsecureSkipVerify: e.config.SkipVerify,
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(e.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

// Helper functions

func (e *EmailChannel) getSystemInfo(event *notification.Event) string {
	if event.SystemName != nil {
		return *event.SystemName
	}
	if event.VolumeGroupName != nil {
		return *event.VolumeGroupName
	}
	return "System"
}

func (e *EmailChannel) formatEventType(eventType notification.EventType) string {
	s := string(eventType)
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Split(s, " ")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func (e *EmailChannel) stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

//
