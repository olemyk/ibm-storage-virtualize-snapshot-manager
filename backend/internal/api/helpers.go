package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	userIDKey   contextKey = "userID"
	usernameKey contextKey = "username"
	userRoleKey contextKey = "userRole"
)

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// handleError logs an error and sends an error response
func handleError(w http.ResponseWriter, r *http.Request, err error, status int, message string) {
	// Log detailed error internally with request context
	if err != nil {
		log.Printf("[%s %s] %s: %v", r.Method, r.URL.Path, message, err)
	} else {
		log.Printf("[%s %s] %s", r.Method, r.URL.Path, message)
	}
	// Return sanitized error to client (no internal details)
	respondError(w, status, message)
}

// getUserIDFromContext gets the user ID from the request context
func getUserIDFromContext(r *http.Request) int {
	userID, ok := r.Context().Value(userIDKey).(int)
	if !ok {
		return 0
	}
	return userID
}

// getUsernameFromContext gets the username from the request context
func getUsernameFromContext(r *http.Request) string {
	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		return ""
	}
	return username
}

func getUserRoleFromContext(r *http.Request) string {
	role, ok := r.Context().Value(userRoleKey).(string)
	if !ok {
		return ""
	}
	return role
}

func (s *Server) requireRoles(allowedRoles ...string) mux.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := allowed[getUserRoleFromContext(r)]; !ok {
				respondError(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// authMiddleware validates JWT tokens
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check if token format is valid
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Check if token is blacklisted
		if s.tokenBlacklist.IsBlacklisted(token) {
			respondError(w, http.StatusUnauthorized, "Token has been revoked")
			return
		}

		// Get real client IP (prefer RemoteAddr over X-Forwarded-For to prevent spoofing)
		clientIP := r.RemoteAddr
		if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
			clientIP = clientIP[:idx]
		}

		// Log authentication attempt with real IP
		log.Printf("Auth attempt from IP: %s", clientIP)

		// Validate token
		claims, err := s.auth.ValidateToken(token)
		if err != nil {
			log.Printf("Invalid token from IP %s: %v", clientIP, err)
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		var role string
		err = s.db.QueryRow("SELECT role FROM users WHERE id = $1", claims.UserID).Scan(&role)
		if err != nil {
			log.Printf("Failed to load user role for user %d from IP %s: %v", claims.UserID, clientIP, err)
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user ID, username, and role to context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)
		ctx = context.WithValue(ctx, userRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateUpdateSystemRequest validates the update system request fields
func validateUpdateSystemRequest(name, ipAddress, username string, port int) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if ipAddress == "" {
		return fmt.Errorf("ip_address is required")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// encryptPasswordIfProvided encrypts a password if it's not empty
func encryptPasswordIfProvided(password, encryptionKey string) (string, error) {
	if password == "" {
		return "", nil
	}
	return crypto.Encrypt(password, encryptionKey)
}

// querySchedulesWithDetails queries snapshot schedules with volume group and system details.
// It handles the common query logic shared between handleListSchedules and handleListAllSchedules.
// The whereClause parameter should include the WHERE keyword if needed (e.g., "WHERE s.volume_group_id = ?").
func querySchedulesWithDetails(ctx context.Context, db *sql.DB, whereClause string, args ...interface{}) ([]models.ScheduleWithVolumeGroup, error) {
	baseQuery := `
		SELECT s.id, s.volume_group_id, s.name, s.cron_expression, s.retention_days, s.retention_minutes,
		       s.safeguarded, s.pool_name, s.snapshot_name_pattern, s.is_active, s.last_executed_at,
		       s.next_execution_at, s.created_at, s.updated_at, vg.vg_name, sys.name as system_name
		FROM snapshot_schedules s
		JOIN volume_groups vg ON s.volume_group_id = vg.id
		JOIN storage_systems sys ON vg.storage_system_id = sys.id`

	query := baseQuery
	if whereClause != "" {
		query += " " + whereClause
	}
	query += " ORDER BY s.created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []models.ScheduleWithVolumeGroup
	for rows.Next() {
		var schedule models.ScheduleWithVolumeGroup
		err := rows.Scan(
			&schedule.ID, &schedule.VolumeGroupID, &schedule.Name, &schedule.CronExpression,
			&schedule.RetentionDays, &schedule.RetentionMinutes, &schedule.Safeguarded,
			&schedule.PoolName, &schedule.SnapshotNamePattern, &schedule.IsActive,
			&schedule.LastExecutedAt, &schedule.NextExecutionAt, &schedule.CreatedAt, &schedule.UpdatedAt,
			&schedule.VGName, &schedule.SystemName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule row: %w", err)
		}
		schedules = append(schedules, schedule)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schedule rows: %w", err)
	}

	// Return empty slice instead of nil for consistent JSON responses
	if schedules == nil {
		schedules = []models.ScheduleWithVolumeGroup{}
	}

	return schedules, nil
}

// scanNullString converts sql.NullString to *string
func scanNullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

// scanNullTime converts sql.NullTime to *time.Time
func scanNullTime(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

// scanNullInt64 converts sql.NullInt64 to *int
func scanNullInt64(ni sql.NullInt64) *int {
	if ni.Valid {
		value := int(ni.Int64)
		return &value
	}
	return nil
}

//
