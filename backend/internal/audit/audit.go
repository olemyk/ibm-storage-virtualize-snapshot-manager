package audit

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
)

// Logger handles audit logging
type Logger struct {
	db *sql.DB
}

// NewLogger creates a new audit logger
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// LogAction logs an action to the audit log
func (l *Logger) LogAction(
	userID *int,
	username string,
	action string,
	resourceType string,
	resourceID *string,
	resourceName *string,
	details interface{},
	r *http.Request,
	status string,
	errorMessage *string,
) error {
	var detailsJSON *string
	if details != nil {
		jsonBytes, err := json.Marshal(details)
		if err == nil {
			jsonStr := string(jsonBytes)
			detailsJSON = &jsonStr
		}
	}

	ipAddress := getIPAddress(r)
	userAgent := r.UserAgent()

	query := `
		INSERT INTO audit_logs (
			user_id, username, action, resource_type, resource_id, resource_name,
			details, ip_address, user_agent, status, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	log.Printf("Audit log: Attempting to log action=%s, user=%s, status=%s", action, username, status)
	_, err := l.db.Exec(
		query,
		userID,
		username,
		action,
		resourceType,
		resourceID,
		resourceName,
		detailsJSON,
		ipAddress,
		userAgent,
		status,
		errorMessage,
		time.Now(),
	)

	if err != nil {
		log.Printf("Audit log ERROR: Failed to insert audit log: %v", err)
	} else {
		log.Printf("Audit log: Successfully logged action=%s for user=%s", action, username)
	}

	return err
}

// LogSuccess logs a successful action
func (l *Logger) LogSuccess(
	userID *int,
	username string,
	action string,
	resourceType string,
	resourceID *string,
	resourceName *string,
	details interface{},
	r *http.Request,
) error {
	return l.LogAction(userID, username, action, resourceType, resourceID, resourceName, details, r, "success", nil)
}

// LogFailure logs a failed action
func (l *Logger) LogFailure(
	userID *int,
	username string,
	action string,
	resourceType string,
	resourceID *string,
	resourceName *string,
	details interface{},
	r *http.Request,
	errorMsg string,
) error {
	return l.LogAction(userID, username, action, resourceType, resourceID, resourceName, details, r, "failed", &errorMsg)
}

// ListAuditLogs retrieves audit logs with optional filters
func (l *Logger) ListAuditLogs(
	userID *int,
	action *string,
	resourceType *string,
	status *string,
	startDate *time.Time,
	endDate *time.Time,
	limit int,
	offset int,
) ([]models.AuditLog, error) {
	query := `
		SELECT 
			id, user_id, username, action, resource_type, resource_id, resource_name,
			details, ip_address, user_agent, status, error_message, created_at
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	paramCount := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", paramCount)
		args = append(args, *userID)
		paramCount++
	}

	if action != nil {
		query += fmt.Sprintf(" AND action = $%d", paramCount)
		args = append(args, *action)
		paramCount++
	}

	if resourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", paramCount)
		args = append(args, *resourceType)
		paramCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", paramCount)
		args = append(args, *status)
		paramCount++
	}

	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", paramCount)
		args = append(args, *startDate)
		paramCount++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", paramCount)
		args = append(args, *endDate)
		paramCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	args = append(args, limit, offset)

	rows, err := l.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Username,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&log.ResourceName,
			&log.Details,
			&log.IPAddress,
			&log.UserAgent,
			&log.Status,
			&log.ErrorMessage,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// CountAuditLogs counts audit logs matching filters
func (l *Logger) CountAuditLogs(
	userID *int,
	action *string,
	resourceType *string,
	status *string,
	startDate *time.Time,
	endDate *time.Time,
) (int, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	paramCount := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", paramCount)
		args = append(args, *userID)
		paramCount++
	}

	if action != nil {
		query += fmt.Sprintf(" AND action = $%d", paramCount)
		args = append(args, *action)
		paramCount++
	}

	if resourceType != nil {
		query += fmt.Sprintf(" AND resource_type = $%d", paramCount)
		args = append(args, *resourceType)
		paramCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", paramCount)
		args = append(args, *status)
		paramCount++
	}

	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", paramCount)
		args = append(args, *startDate)
		paramCount++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", paramCount)
		args = append(args, *endDate)
		paramCount++
	}

	var count int
	err := l.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// getIPAddress extracts the real IP address from the request
func getIPAddress(r *http.Request) *string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return &forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return &realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	return &ip
}

// CleanupOldLogs removes audit logs older than the retention period
// Keeps either the last maxEntries or entries within retentionDays, whichever is more restrictive
func (l *Logger) CleanupOldLogs(maxEntries int, retentionDays int) error {
	// First, delete entries older than retention period
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	deleteOldQuery := `DELETE FROM audit_logs WHERE created_at < $1`
	result, err := l.db.Exec(deleteOldQuery, cutoffDate)
	if err != nil {
		return err
	}

	deletedOld, _ := result.RowsAffected()
	if deletedOld > 0 {
		log.Printf("Audit log cleanup: Deleted %d entries older than %d days", deletedOld, retentionDays)
	}

	// Then, keep only the last maxEntries using a more efficient subquery
	deleteExcessQuery := `
		DELETE FROM audit_logs 
		WHERE id NOT IN (
			SELECT id FROM audit_logs 
			ORDER BY created_at DESC 
			LIMIT $1
		)
	`
	result, err = l.db.Exec(deleteExcessQuery, maxEntries)
	if err != nil {
		return err
	}

	deletedExcess, _ := result.RowsAffected()
	if deletedExcess > 0 {
		log.Printf("Audit log cleanup: Deleted %d entries to maintain max %d entries", deletedExcess, maxEntries)
	}

	return nil
}

// StartPeriodicCleanup starts a background goroutine that periodically cleans up old audit logs
// maxEntries: maximum number of audit log entries to keep (default: 1000)
// retentionDays: maximum age of audit logs in days (default: 365)
// cleanupInterval: how often to run cleanup in hours (default: 24)
func (l *Logger) StartPeriodicCleanup(maxEntries, retentionDays, cleanupIntervalHours int) {
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	if retentionDays <= 0 {
		retentionDays = 365
	}
	if cleanupIntervalHours <= 0 {
		cleanupIntervalHours = 24
	}

	log.Printf("Starting audit log cleanup: max %d entries, %d days retention, cleanup every %d hours",
		maxEntries, retentionDays, cleanupIntervalHours)

	// Run cleanup immediately on startup
	go func() {
		if err := l.CleanupOldLogs(maxEntries, retentionDays); err != nil {
			log.Printf("Error during initial audit log cleanup: %v", err)
		}
	}()

	// Then run periodically
	ticker := time.NewTicker(time.Duration(cleanupIntervalHours) * time.Hour)
	go func() {
		for range ticker.C {
			if err := l.CleanupOldLogs(maxEntries, retentionDays); err != nil {
				log.Printf("Error during periodic audit log cleanup: %v", err)
			}
		}
	}()
}

// Action constants
const (
	ActionLogin            = "login"
	ActionLogout           = "logout"
	ActionCreate           = "create"
	ActionRead             = "read"
	ActionUpdate           = "update"
	ActionDelete           = "delete"
	ActionCreateSystem     = "create_system"
	ActionUpdateSystem     = "update_system"
	ActionDeleteSystem     = "delete_system"
	ActionSyncVolumeGroups = "sync_volume_groups"
	ActionCreateSchedule   = "create_schedule"
	ActionUpdateSchedule   = "update_schedule"
	ActionDeleteSchedule   = "delete_schedule"
	ActionExecuteSchedule  = "execute_schedule"
	ActionCreateUser       = "create_user"
	ActionUpdateUser       = "update_user"
	ActionDeleteUser       = "delete_user"
	ActionUpdateSettings   = "update_settings"
	ActionCreateNTPServer  = "create_ntp_server"
	ActionUpdateNTPServer  = "update_ntp_server"
	ActionDeleteNTPServer  = "delete_ntp_server"
)

// ResourceType constants
const (
	ResourceTypeSystem       = "storage_system"
	ResourceTypeVolumeGroup  = "volume_group"
	ResourceTypeSchedule     = "snapshot_schedule"
	ResourceTypeExecution    = "snapshot_execution"
	ResourceTypeUser         = "user"
	ResourceTypeSettings     = "settings"
	ResourceTypeNTPServer    = "ntp_server"
	ResourceTypeNotification = "notification"
)

//
