package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/audit"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
)

// Notification Channel Handlers

func (s *Server) handleListNotificationChannels(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	channels, err := s.notificationManager.ListChannels()
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, nil, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to list notification channels")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, nil, r)
	respondJSON(w, http.StatusOK, channels)
}

func (s *Server) handleGetNotificationChannel(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	channel, err := s.notificationManager.GetChannelByDBID(id)
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id}, r, err.Error())
		respondError(w, http.StatusNotFound, "Channel not found")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id}, r)
	respondJSON(w, http.StatusOK, channel)
}

func (s *Server) handleCreateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	var req struct {
		Name        string                 `json:"name"`
		Type        string                 `json:"type"`
		Config      map[string]interface{} `json:"config"`
		Description string                 `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Channel name is required")
		return
	}
	if req.Type == "" {
		respondError(w, http.StatusBadRequest, "Channel type is required")
		return
	}

	// Validate channel type
	validTypes := map[string]bool{
		"email":   true,
		"slack":   true,
		"webhook": true,
		"snmp":    true,
	}
	if !validTypes[req.Type] {
		respondError(w, http.StatusBadRequest, "Invalid channel type. Must be one of: email, slack, webhook, snmp")
		return
	}

	channelType := notification.ChannelType(req.Type)
	channel, err := s.notificationManager.CreateChannel(req.Name, channelType, req.Config)
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_name": req.Name}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to create notification channel: "+err.Error())
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": channel.ID, "channel_name": channel.Name}, r)
	respondJSON(w, http.StatusCreated, channel)
}

func (s *Server) handleUpdateNotificationChannel(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	// Get existing channel
	channel, err := s.notificationManager.GetChannelByDBID(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Channel not found")
		return
	}

	var req struct {
		Name        *string                 `json:"name"`
		Type        *string                 `json:"type"`
		Config      *map[string]interface{} `json:"config"`
		IsActive    *bool                   `json:"is_active"`
		Description *string                 `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields if provided
	name := channel.Name
	if req.Name != nil {
		name = *req.Name
	}

	channelType := channel.Type
	if req.Type != nil {
		// Validate channel type
		validTypes := map[string]bool{
			"email":   true,
			"slack":   true,
			"webhook": true,
			"snmp":    true,
		}
		if !validTypes[*req.Type] {
			respondError(w, http.StatusBadRequest, "Invalid channel type")
			return
		}
		channelType = notification.ChannelType(*req.Type)
	}

	config := make(map[string]interface{})
	if req.Config != nil {
		config = *req.Config
	} else {
		// Parse existing config
		if err := json.Unmarshal([]byte(channel.Config), &config); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to parse existing config")
			return
		}
	}

	isActive := channel.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := s.notificationManager.UpdateChannel(id, name, channelType, config, isActive); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionUpdate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to update notification channel")
		return
	}

	// Get updated channel
	updatedChannel, _ := s.notificationManager.GetChannelByDBID(id)

	s.auditLogger.LogSuccess(&userID, username, audit.ActionUpdate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id, "channel_name": name}, r)
	respondJSON(w, http.StatusOK, updatedChannel)
}

func (s *Server) handleDeleteNotificationChannel(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	if err := s.notificationManager.DeleteChannel(id); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionDelete, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to delete notification channel")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionDelete, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id}, r)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Channel deleted successfully"})
}

func (s *Server) handleTestNotificationChannel(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	if err := s.notificationManager.TestChannel(id); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id, "action": "test"}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to test notification channel: "+err.Error())
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": id, "action": "test"}, r)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Test notification sent successfully"})
}

// Alert Rule Handlers

func (s *Server) handleListAlertRules(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	rules, err := s.notificationManager.ListRules()
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"resource": "alert_rules"}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to list alert rules")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"resource": "alert_rules", "count": len(rules)}, r)
	respondJSON(w, http.StatusOK, rules)
}

func (s *Server) handleGetAlertRule(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	rule, err := s.notificationManager.GetRule(id)
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id}, r, err.Error())
		respondError(w, http.StatusNotFound, "Alert rule not found")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id}, r)
	respondJSON(w, http.StatusOK, rule)
}

func (s *Server) handleCreateAlertRule(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	var req notification.AlertRule

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Rule name is required")
		return
	}

	if err := s.notificationManager.CreateRule(&req); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_name": req.Name}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to create alert rule")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": req.ID, "rule_name": req.Name}, r)
	respondJSON(w, http.StatusCreated, req)
}

func (s *Server) handleUpdateAlertRule(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	// Get existing rule
	rule, err := s.notificationManager.GetRule(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Alert rule not found")
		return
	}

	// Decode update request into existing rule
	if err := json.NewDecoder(r.Body).Decode(rule); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID doesn't change
	rule.ID = id

	if err := s.notificationManager.UpdateRule(rule); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionUpdate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to update alert rule")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionUpdate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id, "rule_name": rule.Name}, r)
	respondJSON(w, http.StatusOK, rule)
}

func (s *Server) handleDeleteAlertRule(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid rule ID")
		return
	}

	if err := s.notificationManager.DeleteRule(id); err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionDelete, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to delete alert rule")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionDelete, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"rule_id": id}, r)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Alert rule deleted successfully"})
}

// Notification History Handlers

func (s *Server) handleListNotificationHistory(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	// Parse query parameters
	query := r.URL.Query()

	var channelID *int
	if channelIDStr := query.Get("channel_id"); channelIDStr != "" {
		if id, err := strconv.Atoi(channelIDStr); err == nil {
			channelID = &id
		}
	}

	var status *string
	if statusStr := query.Get("status"); statusStr != "" {
		status = &statusStr
	}

	var startTime, endTime *time.Time
	if startStr := query.Get("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = &t
		}
	}
	if endStr := query.Get("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = &t
		}
	}

	limit := 100
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	history, err := s.notificationManager.GetHistory(channelID, status, startTime, endTime, limit, offset)
	if err != nil {
		s.auditLogger.LogFailure(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"resource": "history"}, r, err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to retrieve notification history")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionRead, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"resource": "history", "count": len(history)}, r)
	respondJSON(w, http.StatusOK, history)
}

// Test Notification Handler

func (s *Server) handleSendTestNotification(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	var req struct {
		ChannelID int    `json:"channel_id"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ChannelID == 0 {
		respondError(w, http.StatusBadRequest, "Channel ID is required")
		return
	}

	message := req.Message
	if message == "" {
		message = "This is a test notification from IBM Storage Virtualize Snapshot Manager"
	}

	// Create a test event
	event := &notification.Event{
		Type:      "test",
		Timestamp: time.Now(),
		Severity:  notification.SeverityInfo,
		Message:   message,
		Details: map[string]interface{}{
			"test":         true,
			"triggered_by": username,
		},
	}

	// Send notification directly to the specified channel
	if s.notifier != nil {
		ctx := context.Background()
		if err := s.notifier.SendToChannel(ctx, req.ChannelID, event); err != nil {
			s.auditLogger.LogFailure(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": req.ChannelID, "action": "test"}, r, err.Error())
			respondError(w, http.StatusInternalServerError, "Failed to send test notification: "+err.Error())
			return
		}
	} else {
		respondError(w, http.StatusServiceUnavailable, "Notification service is not available")
		return
	}

	s.auditLogger.LogSuccess(&userID, username, audit.ActionCreate, audit.ResourceTypeNotification, nil, nil, map[string]interface{}{"channel_id": req.ChannelID, "action": "test"}, r)
	respondJSON(w, http.StatusOK, map[string]string{"message": "Test notification sent successfully"})
}

//
