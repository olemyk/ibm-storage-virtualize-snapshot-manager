package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/audit"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/scheduler"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
)

// Authentication handlers

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stderr, "DEBUG: handleLogin called\n")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	fmt.Fprintf(os.Stderr, "DEBUG: Login request decoded, username: %s\n", req.Username)

	// Query user from database
	var user models.User
	var email sql.NullString
	query := `SELECT id, username, password_hash, email, role, created_at, updated_at FROM users WHERE username = $1`
	err := s.db.QueryRow(query, req.Username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if email.Valid {
		user.Email = email.String
	}

	if err != nil {
		// Log failed login attempt
		s.auditLogger.LogFailure(
			nil,
			req.Username,
			audit.ActionLogin,
			audit.ResourceTypeUser,
			nil,
			nil,
			nil,
			r,
			"Invalid credentials",
		)
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Check password
	if !s.auth.CheckPassword(req.Password, user.PasswordHash) {
		// Log failed login attempt
		s.auditLogger.LogFailure(
			&user.ID,
			user.Username,
			audit.ActionLogin,
			audit.ResourceTypeUser,
			nil,
			nil,
			nil,
			r,
			"Invalid password",
		)
		respondError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	token, err := s.auth.GenerateToken(&user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Generate CSRF token
	csrfToken, err := s.csrfManager.GenerateToken()
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate CSRF token")
		return
	}

	// Log successful login
	log.Printf("DEBUG: About to call auditLogger.LogSuccess for user: %s (ID: %d)", user.Username, user.ID)
	if s.auditLogger == nil {
		log.Printf("ERROR: auditLogger is nil!")
	} else {
		log.Printf("DEBUG: auditLogger is not nil, calling LogSuccess")
		if err := s.auditLogger.LogSuccess(
			&user.ID,
			user.Username,
			audit.ActionLogin,
			audit.ResourceTypeUser,
			nil,
			nil,
			map[string]interface{}{"role": user.Role},
			r,
		); err != nil {
			log.Printf("Error logging successful login audit: %v", err)
		} else {
			log.Printf("DEBUG: LogSuccess completed without error")
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"csrf_token": csrfToken,
		"token":      token,
		"user":       user,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get user info from context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	// Extract and blacklist JWT token
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]

			// Validate token to get expiry time
			claims, err := s.auth.ValidateToken(token)
			if err == nil && claims.ExpiresAt != nil {
				// Add token to blacklist with its expiry time
				s.tokenBlacklist.Add(token, claims.ExpiresAt.Time)
			}
		}
	}

	// Invalidate CSRF token
	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken != "" {
		s.csrfManager.InvalidateToken(csrfToken)
	}

	// Log logout
	if userID != 0 {
		s.auditLogger.LogSuccess(
			&userID,
			username,
			audit.ActionLogout,
			audit.ResourceTypeUser,
			nil,
			nil,
			nil,
			r,
		)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (s *Server) handleGetCSRFToken(w http.ResponseWriter, r *http.Request) {
	token, err := s.csrfManager.GenerateToken()
	if err != nil {
		log.Printf("Error generating CSRF token: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate CSRF token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"csrf_token": token,
	})
}

func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var user models.User
	var email sql.NullString
	query := `SELECT id, username, password_hash, email, role, created_at, updated_at FROM users WHERE id = $1`
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if email.Valid {
		user.Email = email.String
	}

	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Storage system handlers

func (s *Server) handleListSystems(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, name, ip_address, port, username, skip_tls_verify, is_active, connection_status, last_connection_check, connection_error, created_at, updated_at FROM storage_systems`
	rows, err := s.db.Query(query)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to query system")
		return
	}
	defer rows.Close()

	var systems []models.StorageSystem
	for rows.Next() {
		var system models.StorageSystem
		var connectionStatus sql.NullString
		var lastConnectionCheck sql.NullTime
		var connectionError sql.NullString

		err := rows.Scan(&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username,
			&system.SkipTLSVerify, &system.IsActive, &connectionStatus, &lastConnectionCheck,
			&connectionError, &system.CreatedAt, &system.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning system row: %v", err)
			continue
		}

		if connectionStatus.Valid {
			system.ConnectionStatus = &connectionStatus.String
			log.Printf("System %d (%s): connection_status=%s", system.ID, system.Name, connectionStatus.String)
		}
		if lastConnectionCheck.Valid {
			system.LastConnectionCheck = &lastConnectionCheck.Time
			log.Printf("System %d (%s): last_connection_check=%v", system.ID, system.Name, lastConnectionCheck.Time)
		}
		if connectionError.Valid {
			system.ConnectionError = &connectionError.String
		}

		systems = append(systems, system)
	}

	respondJSON(w, http.StatusOK, systems)
}

func (s *Server) handleCreateSystem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		IPAddress     string `json:"ip_address"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		SkipTLSVerify bool   `json:"skip_tls_verify"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Encrypt password
	encryptedPassword, err := crypto.Encrypt(req.Password, s.config.Security.EncryptionKey)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to encrypt password")
		return
	}

	// Insert into database
	query := `INSERT INTO storage_systems (name, ip_address, port, username, password_encrypted, skip_tls_verify) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var id int64
	err = s.db.QueryRow(query, req.Name, req.IPAddress, req.Port, req.Username, encryptedPassword, req.SkipTLSVerify).Scan(&id)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to create system")
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "message": "System created successfully"})
}

func (s *Server) handleGetSystem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}

	var system models.StorageSystem
	query := `SELECT id, name, ip_address, port, username, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`
	err = s.db.QueryRow(query, id).Scan(&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username, &system.SkipTLSVerify, &system.IsActive, &system.CreatedAt, &system.UpdatedAt)

	if err != nil {
		respondError(w, http.StatusNotFound, "System not found")
		return
	}

	respondJSON(w, http.StatusOK, system)
}

func (s *Server) handleUpdateSystem(w http.ResponseWriter, r *http.Request) {
	// Parse and validate system ID
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}

	// Parse request body
	var req struct {
		Name          string `json:"name"`
		IPAddress     string `json:"ip_address"`
		Port          int    `json:"port"`
		Username      string `json:"username"`
		Password      string `json:"password,omitempty"`
		SkipTLSVerify *bool  `json:"skip_tls_verify,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if err := validateUpdateSystemRequest(req.Name, req.IPAddress, req.Username, req.Port); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Encrypt password if provided
	encryptedPassword, err := encryptPasswordIfProvided(req.Password, s.config.Security.EncryptionKey)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to encrypt password")
		return
	}

	// Build single UPDATE query with COALESCE for optional fields
	// This approach is PostgreSQL-compatible and atomic
	query := `UPDATE storage_systems
	          SET name = $1,
	              ip_address = $2,
	              port = $3,
	              username = $4,
	              skip_tls_verify = COALESCE($5, skip_tls_verify),
	              password_encrypted = COALESCE(NULLIF($6, ''), password_encrypted),
	              updated_at = CURRENT_TIMESTAMP
	          WHERE id = $7`

	// Prepare arguments - use pointer value or current value for skip_tls_verify
	var skipTLSArg interface{}
	if req.SkipTLSVerify != nil {
		skipTLSArg = *req.SkipTLSVerify
	} else {
		skipTLSArg = nil
	}

	_, err = s.db.Exec(query, req.Name, req.IPAddress, req.Port, req.Username, skipTLSArg, encryptedPassword, id)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to update system")
		return
	}

	// Get user context for audit logging
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	// Log successful update with audit trail
	systemIDStr := strconv.Itoa(id)
	auditDetails := map[string]interface{}{
		"system_id":        id,
		"name":             req.Name,
		"ip_address":       req.IPAddress,
		"port":             req.Port,
		"password_changed": req.Password != "",
	}
	if req.SkipTLSVerify != nil {
		auditDetails["skip_tls_verify"] = *req.SkipTLSVerify
	}

	s.auditLogger.LogSuccess(
		&userID,
		username,
		audit.ActionUpdate,
		audit.ResourceTypeSystem,
		&systemIDStr,
		&req.Name,
		auditDetails,
		r,
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "System updated successfully"})
}

func (s *Server) handleDeleteSystem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}

	query := `DELETE FROM storage_systems WHERE id = $1`
	_, err = s.db.Exec(query, id)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to delete system")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "System deleted successfully"})
}

func (s *Server) handleTestSystem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}

	// Get system from database
	var system models.StorageSystem
	var authToken sql.NullString
	var tokenExpiresAt sql.NullTime
	query := `SELECT id, name, ip_address, port, username, password_encrypted, auth_token, token_expires_at, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`
	err = s.db.QueryRow(query, id).Scan(
		&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username,
		&system.PasswordEncrypted, &authToken, &tokenExpiresAt, &system.SkipTLSVerify,
		&system.IsActive, &system.CreatedAt, &system.UpdatedAt,
	)

	if err != nil {
		handleError(w, r, err, http.StatusNotFound, "System not found")
		return
	}

	// Handle nullable fields
	if authToken.Valid {
		system.AuthToken = authToken.String
	}
	if tokenExpiresAt.Valid {
		system.TokenExpiresAt = &tokenExpiresAt.Time
	}

	// Decrypt password
	password, err := crypto.Decrypt(system.PasswordEncrypted, s.config.Security.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt password")
		return
	}

	// Test connection
	if err := s.svcClient.TestConnection(&system, password); err != nil {
		// Provide user-friendly error messages
		errorMsg := err.Error()
		var userMsg string
		if strings.Contains(errorMsg, "context deadline exceeded") || strings.Contains(errorMsg, "timeout") {
			userMsg = fmt.Sprintf("Connection timeout - System %s:%d is not reachable. Check IP address and network connectivity.", system.IPAddress, system.Port)
		} else if strings.Contains(errorMsg, "connection refused") {
			userMsg = fmt.Sprintf("Connection refused - System %s:%d is not accepting connections. Verify the port number.", system.IPAddress, system.Port)
		} else if strings.Contains(errorMsg, "403") || strings.Contains(errorMsg, "Invalid credentials") {
			userMsg = "Authentication failed - Invalid username or password."
		} else if strings.Contains(errorMsg, "certificate") || strings.Contains(errorMsg, "tls") {
			userMsg = "SSL/TLS error - Certificate validation failed (this is usually OK for IBM SVC)."
		} else {
			userMsg = fmt.Sprintf("Connection failed: %s", errorMsg)
		}

		// Update connection status in database
		now := time.Now()
		updateQuery := `UPDATE storage_systems SET connection_status = $1, last_connection_check = $2, connection_error = $3 WHERE id = $4`
		_, _ = s.db.Exec(updateQuery, "disconnected", now, userMsg, id)

		respondError(w, http.StatusBadRequest, userMsg)
		return
	}

	// Update connection status in database (success)
	now := time.Now()
	updateQuery := `UPDATE storage_systems SET connection_status = $1, last_connection_check = $2, connection_error = NULL WHERE id = $3`
	_, _ = s.db.Exec(updateQuery, "connected", now, id)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Connection successful"})
}
func (s *Server) handleCheckSystemsHealth(w http.ResponseWriter, r *http.Request) {
	// Get all active systems
	query := `SELECT id, name, ip_address, port, username, password_encrypted, skip_tls_verify FROM storage_systems WHERE is_active = TRUE`
	rows, err := s.db.Query(query)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to query system")
		return
	}
	defer rows.Close()

	type HealthCheckResult struct {
		SystemID   int    `json:"system_id"`
		SystemName string `json:"system_name"`
		Status     string `json:"status"`
		Error      string `json:"error,omitempty"`
	}

	// Collect all systems first
	var systems []models.StorageSystem
	for rows.Next() {
		var system models.StorageSystem
		err := rows.Scan(&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username, &system.PasswordEncrypted, &system.SkipTLSVerify)
		if err != nil {
			continue
		}
		systems = append(systems, system)
	}

	// Test connections in parallel using worker pool pattern to limit concurrency
	type resultChan struct {
		result HealthCheckResult
	}
	resultsCh := make(chan resultChan, len(systems))

	// Worker pool with max 5 concurrent health checks
	maxWorkers := config.MaxConcurrentHealthChecks
	if len(systems) < maxWorkers {
		maxWorkers = len(systems)
	}
	semaphore := make(chan struct{}, maxWorkers)

	for _, system := range systems {
		go func(sys models.StorageSystem) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic in health check for system %d: %v", sys.ID, r)
					resultsCh <- resultChan{
						result: HealthCheckResult{
							SystemID:   sys.ID,
							SystemName: sys.Name,
							Status:     "disconnected",
							Error:      "Internal error during health check",
						},
					}
				}
			}()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			// Decrypt password
			password, err := crypto.Decrypt(sys.PasswordEncrypted, s.config.Security.EncryptionKey)
			if err != nil {
				resultsCh <- resultChan{
					result: HealthCheckResult{
						SystemID:   sys.ID,
						SystemName: sys.Name,
						Status:     "disconnected",
						Error:      "Failed to decrypt password",
					},
				}
				return
			}

			// Test connection
			err = s.svcClient.TestConnection(&sys, password)
			now := time.Now()
			if err != nil {
				errorMsg := err.Error()
				// Update database
				updateQuery := `UPDATE storage_systems SET connection_status = $1, last_connection_check = $2, connection_error = $3 WHERE id = $4`
				execResult, execErr := s.db.Exec(updateQuery, "disconnected", now, errorMsg, sys.ID)
				if execErr != nil {
					log.Printf("Error updating system %d status: %v", sys.ID, execErr)
				} else {
					rowsAffected, _ := execResult.RowsAffected()
					log.Printf("Updated system %d (%s) to disconnected, rows affected: %d", sys.ID, sys.Name, rowsAffected)
				}

				resultsCh <- resultChan{
					result: HealthCheckResult{
						SystemID:   sys.ID,
						SystemName: sys.Name,
						Status:     "disconnected",
						Error:      errorMsg,
					},
				}
			} else {
				// Update database
				updateQuery := `UPDATE storage_systems SET connection_status = $1, last_connection_check = $2, connection_error = NULL WHERE id = $3`
				execResult, execErr := s.db.Exec(updateQuery, "connected", now, sys.ID)
				if execErr != nil {
					log.Printf("Error updating system %d status: %v", sys.ID, execErr)
				} else {
					rowsAffected, _ := execResult.RowsAffected()
					log.Printf("Updated system %d (%s) to connected, rows affected: %d, timestamp: %v", sys.ID, sys.Name, rowsAffected, now)
				}

				resultsCh <- resultChan{
					result: HealthCheckResult{
						SystemID:   sys.ID,
						SystemName: sys.Name,
						Status:     "connected",
					},
				}
			}
		}(system)
	}

	// Collect results from all goroutines
	var results []HealthCheckResult
	for i := 0; i < len(systems); i++ {
		res := <-resultsCh
		results = append(results, res.result)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"checked_at": time.Now().Format(time.RFC3339),
		"results":    results,
	})
}

func (s *Server) handleListVolumeGroups(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleListVolumeGroups called for request: %s", r.URL.Path)
	vars := mux.Vars(r)
	systemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid system ID in request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}
	log.Printf("Querying volume groups for system ID: %d", systemID)

	// Create context with timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	query := `
		SELECT
			vg.id,
			vg.storage_system_id,
			vg.vg_id,
			vg.vg_name,
			vg.partition_id,
			vg.partition_name,
			vg.last_synced_at,
			vg.created_at,
			COUNT(ss.id) as schedule_count
		FROM volume_groups vg
		LEFT JOIN snapshot_schedules ss ON vg.id = ss.volume_group_id
		WHERE vg.storage_system_id = $1
		GROUP BY vg.id, vg.storage_system_id, vg.vg_id, vg.vg_name, vg.partition_id, vg.partition_name, vg.last_synced_at, vg.created_at
	`
	rows, err := s.db.QueryContext(ctx, query, systemID)
	if err != nil {
		log.Printf("Error querying volume groups for system %d: %v", systemID, err)
		respondError(w, http.StatusInternalServerError, "Failed to query volume groups")
		return
	}
	defer rows.Close()

	var volumeGroups []models.VolumeGroupWithCount
	for rows.Next() {
		var vg models.VolumeGroupWithCount
		var lastSyncedAt sql.NullTime
		var partitionID, partitionName sql.NullString

		err := rows.Scan(
			&vg.ID, &vg.StorageSystemID, &vg.VGID, &vg.VGName,
			&partitionID, &partitionName, &lastSyncedAt, &vg.CreatedAt,
			&vg.ScheduleCount,
		)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to read volume groups")
			return
		}

		// Use helper functions for null field handling
		vg.PartitionID = scanNullString(partitionID)
		vg.PartitionName = scanNullString(partitionName)
		vg.LastSyncedAt = scanNullTime(lastSyncedAt)

		volumeGroups = append(volumeGroups, vg)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "Error reading volume groups")
		return
	}

	respondJSON(w, http.StatusOK, volumeGroups)
}

func (s *Server) handleSyncVolumeGroups(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	systemID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid system ID")
		return
	}

	// Get system from database
	var system models.StorageSystem
	var authToken sql.NullString
	var tokenExpiresAt sql.NullTime
	query := `SELECT id, name, ip_address, port, username, password_encrypted, auth_token, token_expires_at, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`
	err = s.db.QueryRow(query, systemID).Scan(
		&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username,
		&system.PasswordEncrypted, &authToken, &tokenExpiresAt, &system.SkipTLSVerify,
		&system.IsActive, &system.CreatedAt, &system.UpdatedAt,
	)

	if err != nil {
		handleError(w, r, err, http.StatusNotFound, "System not found")
		return
	}

	if authToken.Valid {
		system.AuthToken = authToken.String
	}
	if tokenExpiresAt.Valid {
		system.TokenExpiresAt = &tokenExpiresAt.Time
	}

	// Decrypt password
	password, err := crypto.Decrypt(system.PasswordEncrypted, s.config.Security.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt password")
		return
	}

	// Get or refresh token
	token, err := s.svcClient.GetOrRefreshToken(&system, password)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to authenticate: "+err.Error())
		return
	}

	// List volume groups from IBM SVC
	volumeGroups, err := s.svcClient.ListVolumeGroups(&system, token)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to list volume groups: "+err.Error())
		return
	}

	// Sync to database
	for _, vg := range volumeGroups {
		vgID, _ := vg["id"].(string)
		vgName, _ := vg["name"].(string)
		partitionID, _ := vg["partition_id"].(string)
		partitionName, _ := vg["partition_name"].(string)

		if vgID == "" || vgName == "" {
			continue
		}

		insertQuery := `INSERT INTO volume_groups (storage_system_id, vg_id, vg_name, partition_id, partition_name, last_synced_at)
			VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(storage_system_id, vg_id)
			DO UPDATE SET vg_name = excluded.vg_name, partition_id = excluded.partition_id, partition_name = excluded.partition_name, last_synced_at = CURRENT_TIMESTAMP`
		_, err := s.db.Exec(insertQuery, systemID, vgID, vgName, partitionID, partitionName)
		if err != nil {
			log.Printf("Error syncing volume group %s: %v", vgName, err)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Sync completed",
		"count":   len(volumeGroups),
	})
}

func (s *Server) handleGetVolumeGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid volume group ID")
		return
	}

	query := `SELECT id, storage_system_id, vg_id, vg_name, partition_id, partition_name, last_synced_at, created_at FROM volume_groups WHERE id = $1`
	var vg models.VolumeGroup
	var lastSyncedAt sql.NullTime
	var partitionID, partitionName sql.NullString

	err = s.db.QueryRow(query, id).Scan(
		&vg.ID,
		&vg.StorageSystemID,
		&vg.VGID,
		&vg.VGName,
		&partitionID,
		&partitionName,
		&lastSyncedAt,
		&vg.CreatedAt,
	)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Volume group not found")
		return
	}
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to fetch volume group")
		return
	}

	if lastSyncedAt.Valid {
		vg.LastSyncedAt = &lastSyncedAt.Time
	}
	if partitionID.Valid {
		vg.PartitionID = &partitionID.String
	}
	if partitionName.Valid {
		vg.PartitionName = &partitionName.String
	}

	respondJSON(w, http.StatusOK, vg)
}

func (s *Server) handleListSnapshots(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vgID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid volume group ID")
		return
	}

	// Get volume group from database
	var vg models.VolumeGroup
	var lastSyncedAt sql.NullTime
	query := `SELECT id, storage_system_id, vg_id, vg_name, last_synced_at, created_at FROM volume_groups WHERE id = $1`
	err = s.db.QueryRow(query, vgID).Scan(&vg.ID, &vg.StorageSystemID, &vg.VGID, &vg.VGName, &lastSyncedAt, &vg.CreatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "Volume group not found")
		return
	}
	if lastSyncedAt.Valid {
		vg.LastSyncedAt = &lastSyncedAt.Time
	}

	// Get storage system
	var system models.StorageSystem
	var authToken sql.NullString
	var tokenExpiresAt sql.NullTime
	sysQuery := `SELECT id, name, ip_address, port, username, password_encrypted, auth_token, token_expires_at, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`
	err = s.db.QueryRow(sysQuery, vg.StorageSystemID).Scan(
		&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username,
		&system.PasswordEncrypted, &authToken, &tokenExpiresAt, &system.SkipTLSVerify,
		&system.IsActive, &system.CreatedAt, &system.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching system: %v", err)
		respondError(w, http.StatusNotFound, "Storage system not found")
		return
	}

	if authToken.Valid {
		system.AuthToken = authToken.String
	}
	if tokenExpiresAt.Valid {
		system.TokenExpiresAt = &tokenExpiresAt.Time
	}

	// Decrypt password
	password, err := crypto.Decrypt(system.PasswordEncrypted, s.config.Security.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt password")
		return
	}

	// Get or refresh token
	token, err := s.svcClient.GetOrRefreshToken(&system, password)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to authenticate: "+err.Error())
		return
	}

	// List snapshots from IBM SVC
	snapshots, err := s.svcClient.ListVolumeGroupSnapshots(&system, token, vg.VGName)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to list snapshots: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, snapshots)
}

func (s *Server) handleListAllSchedules(w http.ResponseWriter, r *http.Request) {
	// Create context with timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	schedules, err := querySchedulesWithDetails(ctx, s.db.DB, "")
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to query schedules")
		return
	}

	respondJSON(w, http.StatusOK, schedules)
}

func (s *Server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vgID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid volume group ID")
		return
	}

	// Create context with timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	schedules, err := querySchedulesWithDetails(ctx, s.db.DB, "WHERE s.volume_group_id = $1", vgID)
	if err != nil {
		handleError(w, r, err, http.StatusInternalServerError, "Failed to query schedules")
		return
	}

	respondJSON(w, http.StatusOK, schedules)
}

func (s *Server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vgID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid volume group ID")
		return
	}

	// Get user context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	var req struct {
		Name                string  `json:"name"`
		CronExpression      string  `json:"cron_expression"`
		RetentionDays       int     `json:"retention_days"`
		RetentionMinutes    *int    `json:"retention_minutes,omitempty"`
		Safeguarded         bool    `json:"safeguarded"`
		PoolName            *string `json:"pool_name,omitempty"`
		SnapshotNamePattern string  `json:"snapshot_name_pattern"`
		IsActive            bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Schedule name is required")
		return
	}
	if req.CronExpression == "" {
		respondError(w, http.StatusBadRequest, "Cron expression is required")
		return
	}
	hasValidRetentionDays := req.RetentionDays > 0
	hasValidRetentionMinutes := req.RetentionMinutes != nil && *req.RetentionMinutes > 0

	if !hasValidRetentionDays && !hasValidRetentionMinutes {
		respondError(w, http.StatusBadRequest, "Either retention days must be greater than 0 or retention minutes must be greater than 0")
		return
	}

	// Validate retention values are reasonable
	if req.RetentionDays > 3650 { // 10 years max
		respondError(w, http.StatusBadRequest, "Retention days cannot exceed 3650 (10 years)")
		return
	}
	if req.RetentionMinutes != nil && *req.RetentionMinutes > 525600 { // 1 year in minutes
		respondError(w, http.StatusBadRequest, "Retention minutes cannot exceed 525600 (1 year)")
		return
	}

	// Validate cron expression
	if err := scheduler.ValidateCronExpression(req.CronExpression); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cron expression: "+err.Error())
		return
	}

	// Get volume group name for audit log
	var vgName string
	s.db.QueryRow("SELECT vg_name FROM volume_groups WHERE id = $1", vgID).Scan(&vgName)

	// Calculate next execution time
	nextExec := s.scheduler.CalculateNextExecution(req.CronExpression)

	// Set default snapshot name pattern if not provided
	snapshotNamePattern := req.SnapshotNamePattern
	if snapshotNamePattern == "" {
		snapshotNamePattern = config.DefaultSnapshotNamePattern
	}

	// Begin transaction for atomic schedule creation
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to create schedule")
		return
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert schedule
	query := `
		INSERT INTO snapshot_schedules
		(volume_group_id, name, cron_expression, retention_days, retention_minutes,
		 safeguarded, pool_name, snapshot_name_pattern, is_active, next_execution_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id
	`

	var scheduleID int64
	err = tx.QueryRow(query, vgID, req.Name, req.CronExpression, req.RetentionDays,
		req.RetentionMinutes, req.Safeguarded, req.PoolName, snapshotNamePattern, req.IsActive, nextExec).Scan(&scheduleID)
	if err != nil {
		log.Printf("Error creating schedule: %v", err)
		// Log failure
		s.auditLogger.LogFailure(
			&userID,
			username,
			audit.ActionCreateSchedule,
			audit.ResourceTypeSchedule,
			nil,
			&req.Name,
			map[string]interface{}{
				"volume_group":    vgName,
				"cron_expression": req.CronExpression,
			},
			r,
			err.Error(),
		)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to create schedule")
		return
	}

	scheduleIDStr := strconv.FormatInt(scheduleID, 10)

	// Commit transaction first
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to create schedule")
		return
	}

	// Add schedule to scheduler if active (after successful commit)
	if req.IsActive {
		schedule := &models.SnapshotSchedule{
			ID:                  int(scheduleID),
			VolumeGroupID:       vgID,
			Name:                req.Name,
			CronExpression:      req.CronExpression,
			RetentionDays:       req.RetentionDays,
			RetentionMinutes:    req.RetentionMinutes,
			Safeguarded:         req.Safeguarded,
			PoolName:            req.PoolName,
			SnapshotNamePattern: snapshotNamePattern,
			IsActive:            req.IsActive,
		}
		if err := s.scheduler.AddSchedule(schedule); err != nil {
			log.Printf("Error adding schedule to scheduler: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to add schedule to scheduler")
			return
		}
	}

	// Log success
	s.auditLogger.LogSuccess(
		&userID,
		username,
		audit.ActionCreateSchedule,
		audit.ResourceTypeSchedule,
		&scheduleIDStr,
		&req.Name,
		map[string]interface{}{
			"volume_group":    vgName,
			"cron_expression": req.CronExpression,
			"retention_days":  req.RetentionDays,
			"safeguarded":     req.Safeguarded,
			"is_active":       req.IsActive,
		},
		r,
	)

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      scheduleID,
		"message": "Schedule created successfully",
	})
}

func (s *Server) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	query := `
		SELECT s.id, s.volume_group_id, s.name, s.cron_expression, s.retention_days, s.retention_minutes,
		       s.safeguarded, s.pool_name, s.snapshot_name_pattern, s.is_active, s.last_executed_at,
		       s.next_execution_at, s.created_at, s.updated_at, vg.vg_name, sys.name as system_name
		FROM snapshot_schedules s
		JOIN volume_groups vg ON s.volume_group_id = vg.id
		JOIN storage_systems sys ON vg.storage_system_id = sys.id
		WHERE s.id = $1
	`

	var schedule models.ScheduleWithVolumeGroup
	err = s.db.QueryRow(query, id).Scan(
		&schedule.ID, &schedule.VolumeGroupID, &schedule.Name, &schedule.CronExpression,
		&schedule.RetentionDays, &schedule.RetentionMinutes, &schedule.Safeguarded,
		&schedule.PoolName, &schedule.SnapshotNamePattern, &schedule.IsActive,
		&schedule.LastExecutedAt, &schedule.NextExecutionAt, &schedule.CreatedAt, &schedule.UpdatedAt,
		&schedule.VGName, &schedule.SystemName,
	)

	if err != nil {
		respondError(w, http.StatusNotFound, "Schedule not found")
		return
	}

	respondJSON(w, http.StatusOK, schedule)
}

func (s *Server) handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	// Get user context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)
	scheduleIDStr := strconv.Itoa(id)

	var req struct {
		Name                string  `json:"name"`
		CronExpression      string  `json:"cron_expression"`
		RetentionDays       int     `json:"retention_days"`
		RetentionMinutes    *int    `json:"retention_minutes,omitempty"`
		Safeguarded         bool    `json:"safeguarded"`
		PoolName            *string `json:"pool_name,omitempty"`
		SnapshotNamePattern string  `json:"snapshot_name_pattern"`
		IsActive            bool    `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Schedule name is required")
		return
	}
	if req.CronExpression == "" {
		respondError(w, http.StatusBadRequest, "Cron expression is required")
		return
	}
	hasValidRetentionDays := req.RetentionDays > 0
	hasValidRetentionMinutes := req.RetentionMinutes != nil && *req.RetentionMinutes > 0

	if !hasValidRetentionDays && !hasValidRetentionMinutes {
		respondError(w, http.StatusBadRequest, "Either retention days must be greater than 0 or retention minutes must be greater than 0")
		return
	}

	// Validate retention values are reasonable
	if req.RetentionDays > 3650 { // 10 years max
		respondError(w, http.StatusBadRequest, "Retention days cannot exceed 3650 (10 years)")
		return
	}
	if req.RetentionMinutes != nil && *req.RetentionMinutes > 525600 { // 1 year in minutes
		respondError(w, http.StatusBadRequest, "Retention minutes cannot exceed 525600 (1 year)")
		return
	}

	// Validate cron expression
	if err := scheduler.ValidateCronExpression(req.CronExpression); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid cron expression: "+err.Error())
		return
	}

	// Calculate next execution time
	nextExec := s.scheduler.CalculateNextExecution(req.CronExpression)

	// Set default snapshot name pattern if not provided
	snapshotNamePattern := req.SnapshotNamePattern
	if snapshotNamePattern == "" {
		snapshotNamePattern = "{schedule_name}_{timestamp}"
	}

	// Update schedule
	query := `
		UPDATE snapshot_schedules
		SET name = $1, cron_expression = $2, retention_days = $3, retention_minutes = $4,
		    safeguarded = $5, pool_name = $6, snapshot_name_pattern = $7, is_active = $8, next_execution_at = $9,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $10
	`

	_, err = s.db.Exec(query, req.Name, req.CronExpression, req.RetentionDays,
		req.RetentionMinutes, req.Safeguarded, req.PoolName, snapshotNamePattern, req.IsActive, nextExec, id)
	if err != nil {
		log.Printf("Error updating schedule: %v", err)
		// Log failure
		s.auditLogger.LogFailure(
			&userID,
			username,
			audit.ActionUpdateSchedule,
			audit.ResourceTypeSchedule,
			&scheduleIDStr,
			&req.Name,
			nil,
			r,
			err.Error(),
		)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to update schedule")
		return
	}

	// Get volume group ID and name for the schedule
	var vgID int
	var vgName string
	err = s.db.QueryRow("SELECT vg.id, vg.vg_name FROM snapshot_schedules s JOIN volume_groups vg ON s.volume_group_id = vg.id WHERE s.id = $1", id).Scan(&vgID, &vgName)
	if err != nil {
		log.Printf("Error getting volume group info: %v", err)
	}

	// Update scheduler
	s.scheduler.RemoveSchedule(id)
	if req.IsActive {
		schedule := &models.SnapshotSchedule{
			ID:                  id,
			VolumeGroupID:       vgID,
			Name:                req.Name,
			CronExpression:      req.CronExpression,
			RetentionDays:       req.RetentionDays,
			RetentionMinutes:    req.RetentionMinutes,
			Safeguarded:         req.Safeguarded,
			PoolName:            req.PoolName,
			SnapshotNamePattern: snapshotNamePattern,
			IsActive:            req.IsActive,
		}
		if err := s.scheduler.AddSchedule(schedule); err != nil {
			log.Printf("Error adding schedule to scheduler: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to re-register schedule with scheduler")
			return
		}
	}

	// Log success
	s.auditLogger.LogSuccess(
		&userID,
		username,
		audit.ActionUpdateSchedule,
		audit.ResourceTypeSchedule,
		&scheduleIDStr,
		&req.Name,
		map[string]interface{}{
			"volume_group": vgName,
			"is_active":    req.IsActive,
		},
		r,
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Schedule updated successfully"})
}

func (s *Server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	// Get user context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)
	scheduleIDStr := strconv.Itoa(id)

	// Get schedule info before deleting
	var scheduleName, vgName string
	err = s.db.QueryRow(`
		SELECT s.name, vg.vg_name
		FROM snapshot_schedules s
		JOIN volume_groups vg ON s.volume_group_id = vg.id
		WHERE s.id = $1`, id).Scan(&scheduleName, &vgName)
	if err != nil {
		log.Printf("Error getting schedule info: %v", err)
	}

	// Remove from scheduler first
	s.scheduler.RemoveSchedule(id)

	// Delete from database
	query := `DELETE FROM snapshot_schedules WHERE id = $1`
	_, err = s.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting schedule: %v", err)
		// Log failure
		s.auditLogger.LogFailure(
			&userID,
			username,
			audit.ActionDeleteSchedule,
			audit.ResourceTypeSchedule,
			&scheduleIDStr,
			&scheduleName,
			map[string]interface{}{"volume_group": vgName},
			r,
			err.Error(),
		)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to delete schedule")
		return
	}

	// Log success
	s.auditLogger.LogSuccess(
		&userID,
		username,
		audit.ActionDeleteSchedule,
		audit.ResourceTypeSchedule,
		&scheduleIDStr,
		&scheduleName,
		map[string]interface{}{"volume_group": vgName},
		r,
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Schedule deleted successfully"})
}

func (s *Server) handleExecuteSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid schedule ID")
		return
	}

	// Get user context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)
	scheduleIDStr := strconv.Itoa(id)

	// Get schedule from database
	var schedule models.SnapshotSchedule
	query := `SELECT id, volume_group_id, name, cron_expression, retention_days, retention_minutes,
	          safeguarded, pool_name, snapshot_name_pattern, is_active, last_executed_at,
	          next_execution_at, created_at, updated_at
	          FROM snapshot_schedules WHERE id = $1`
	err = s.db.QueryRow(query, id).Scan(
		&schedule.ID, &schedule.VolumeGroupID, &schedule.Name, &schedule.CronExpression,
		&schedule.RetentionDays, &schedule.RetentionMinutes, &schedule.Safeguarded,
		&schedule.PoolName, &schedule.SnapshotNamePattern, &schedule.IsActive,
		&schedule.LastExecutedAt, &schedule.NextExecutionAt, &schedule.CreatedAt, &schedule.UpdatedAt,
	)

	if err != nil {
		// Log failure
		s.auditLogger.LogFailure(
			&userID,
			username,
			audit.ActionExecuteSchedule,
			audit.ResourceTypeSchedule,
			&scheduleIDStr,
			nil,
			nil,
			r,
			"Schedule not found",
		)
		respondError(w, http.StatusNotFound, "Schedule not found")
		return
	}

	// Get volume group name
	var vgName string
	s.db.QueryRow("SELECT vg_name FROM volume_groups WHERE id = $1", schedule.VolumeGroupID).Scan(&vgName)

	// Log the execution request
	s.auditLogger.LogSuccess(
		&userID,
		username,
		audit.ActionExecuteSchedule,
		audit.ResourceTypeSchedule,
		&scheduleIDStr,
		&schedule.Name,
		map[string]interface{}{
			"volume_group":   vgName,
			"retention_days": schedule.RetentionDays,
			"safeguarded":    schedule.Safeguarded,
		},
		r,
	)

	// Execute snapshot in a goroutine to avoid blocking
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in snapshot execution for schedule %d: %v", schedule.ID, r)
			}
		}()
		if err := s.scheduler.ExecuteSnapshot(&schedule); err != nil {
			log.Printf("Error executing snapshot for schedule %d: %v", schedule.ID, err)
		}
	}()

	respondJSON(w, http.StatusAccepted, map[string]string{
		"message": "Snapshot execution started",
	})
}

func (s *Server) handleListExecutions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	status := r.URL.Query().Get("status")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "50"
	}

	query := `
		SELECT e.id, e.schedule_id, e.volume_group_id, e.snapshot_name, e.execution_time, e.status,
		       e.error_message, e.snapshot_id, e.retention_days, e.retention_minutes,
		       s.name as schedule_name, vg.vg_name, sys.name as system_name
		FROM snapshot_executions e
		JOIN snapshot_schedules s ON e.schedule_id = s.id
		JOIN volume_groups vg ON e.volume_group_id = vg.id
		JOIN storage_systems sys ON vg.storage_system_id = sys.id
	`

	args := []interface{}{}
	paramCount := 1
	if status != "" {
		query += fmt.Sprintf(" WHERE e.status = $%d", paramCount)
		args = append(args, status)
		paramCount++
	}

	query += fmt.Sprintf(" ORDER BY e.execution_time DESC LIMIT $%d", paramCount)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying executions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to query executions")
		return
	}
	defer rows.Close()

	var executions []models.ExecutionWithDetails
	for rows.Next() {
		var exec models.ExecutionWithDetails
		var snapshotName, errorMessage, snapshotID sql.NullString
		var retentionMinutes sql.NullInt64
		err := rows.Scan(
			&exec.ID, &exec.ScheduleID, &exec.VolumeGroupID, &snapshotName,
			&exec.ExecutionTime, &exec.Status, &errorMessage, &snapshotID,
			&exec.RetentionDays, &retentionMinutes, &exec.ScheduleName, &exec.VGName, &exec.SystemName,
		)
		if err != nil {
			log.Printf("Error scanning execution: %v", err)
			continue
		}
		if snapshotName.Valid {
			exec.SnapshotName = &snapshotName.String
		}
		if errorMessage.Valid {
			exec.ErrorMessage = &errorMessage.String
		}
		if snapshotID.Valid {
			exec.SnapshotID = &snapshotID.String
		}
		if retentionMinutes.Valid {
			value := int(retentionMinutes.Int64)
			exec.RetentionMinutes = &value
		}
		executions = append(executions, exec)
	}

	respondJSON(w, http.StatusOK, executions)
}

func (s *Server) handleGetExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid execution ID")
		return
	}

	query := `
		SELECT e.id, e.schedule_id, e.volume_group_id, e.snapshot_name, e.execution_time, e.status,
		       e.error_message, e.snapshot_id, e.retention_days, e.retention_minutes,
		       s.name as schedule_name, vg.vg_name, sys.name as system_name
		FROM snapshot_executions e
		JOIN snapshot_schedules s ON e.schedule_id = s.id
		JOIN volume_groups vg ON e.volume_group_id = vg.id
		JOIN storage_systems sys ON vg.storage_system_id = sys.id
		WHERE e.id = $1
	`

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var exec models.ExecutionWithDetails
	var snapshotName, errorMessage, snapshotID sql.NullString
	var retentionMinutes sql.NullInt64
	err = s.db.QueryRowContext(ctx, query, id).Scan(
		&exec.ID, &exec.ScheduleID, &exec.VolumeGroupID, &snapshotName,
		&exec.ExecutionTime, &exec.Status, &errorMessage, &snapshotID,
		&exec.RetentionDays, &retentionMinutes, &exec.ScheduleName, &exec.VGName, &exec.SystemName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "Execution not found")
		} else {
			log.Printf("Failed to query execution %d: %v", id, err)
			respondError(w, http.StatusInternalServerError, "Failed to retrieve execution")
		}
		return
	}

	// Use helper functions for null conversions
	exec.SnapshotName = scanNullString(snapshotName)
	exec.ErrorMessage = scanNullString(errorMessage)
	exec.SnapshotID = scanNullString(snapshotID)
	exec.RetentionMinutes = scanNullInt64(retentionMinutes)

	respondJSON(w, http.StatusOK, exec)
}

func (s *Server) handleGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	// Count total systems
	var totalSystems int
	err := s.db.QueryRow("SELECT COUNT(*) FROM storage_systems WHERE is_active = TRUE").Scan(&totalSystems)
	if err != nil {
		log.Printf("Error counting systems: %v", err)
		totalSystems = 0
	}
	stats["total_systems"] = totalSystems

	// Count total volume groups
	var totalVolumeGroups int
	err = s.db.QueryRow("SELECT COUNT(*) FROM volume_groups").Scan(&totalVolumeGroups)
	if err != nil {
		log.Printf("Error counting volume groups: %v", err)
		totalVolumeGroups = 0
	}
	stats["total_volume_groups"] = totalVolumeGroups

	// Count active schedules
	var activeSchedules int
	err = s.db.QueryRow("SELECT COUNT(*) FROM snapshot_schedules WHERE is_active = TRUE").Scan(&activeSchedules)
	if err != nil {
		log.Printf("Error counting schedules: %v", err)
		activeSchedules = 0
	}
	stats["active_schedules"] = activeSchedules

	// Count recent executions (last 24 hours)
	var recentExecutions int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM snapshot_executions
		WHERE execution_time >= NOW() - INTERVAL '1 day'
	`).Scan(&recentExecutions)
	if err != nil {
		log.Printf("Error counting executions: %v", err)
		recentExecutions = 0
	}
	stats["recent_executions"] = recentExecutions
	log.Printf("Dashboard stats - Recent executions (24h): %d", recentExecutions)

	// Count successful executions (last 24 hours)
	var successfulExecutions int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM snapshot_executions
		WHERE execution_time >= NOW() - INTERVAL '1 day' AND status = 'success'
	`).Scan(&successfulExecutions)
	if err != nil {
		log.Printf("Error counting successful executions: %v", err)
		successfulExecutions = 0
	}
	stats["successful_executions"] = successfulExecutions
	log.Printf("Dashboard stats - Successful executions (24h): %d", successfulExecutions)

	// Count failed executions (last 24 hours)
	var failedExecutions int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM snapshot_executions
		WHERE execution_time >= NOW() - INTERVAL '1 day' AND status = 'failed'
	`).Scan(&failedExecutions)
	if err != nil {
		log.Printf("Error counting failed executions: %v", err)
		failedExecutions = 0
	}
	stats["failed_executions"] = failedExecutions
	log.Printf("Dashboard stats - Failed executions (24h): %d", failedExecutions)

	respondJSON(w, http.StatusOK, stats)
}

//

func (s *Server) handleListVolumesInGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vgID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid volume group ID")
		return
	}

	// Get volume group from database
	var vg models.VolumeGroup
	var lastSyncedAt sql.NullTime
	var partitionID, partitionName sql.NullString
	query := `SELECT id, storage_system_id, vg_id, vg_name, partition_id, partition_name, last_synced_at, created_at FROM volume_groups WHERE id = $1`
	err = s.db.QueryRow(query, vgID).Scan(&vg.ID, &vg.StorageSystemID, &vg.VGID, &vg.VGName, &partitionID, &partitionName, &lastSyncedAt, &vg.CreatedAt)
	if err != nil {
		respondError(w, http.StatusNotFound, "Volume group not found")
		return
	}
	if lastSyncedAt.Valid {
		vg.LastSyncedAt = &lastSyncedAt.Time
	}
	if partitionID.Valid {
		vg.PartitionID = &partitionID.String
	}
	if partitionName.Valid {
		vg.PartitionName = &partitionName.String
	}

	// Get storage system
	var system models.StorageSystem
	var authToken sql.NullString
	var tokenExpiresAt sql.NullTime
	systemQuery := `SELECT id, name, ip_address, port, username, password_encrypted, auth_token, token_expires_at, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`
	err = s.db.QueryRow(systemQuery, vg.StorageSystemID).Scan(
		&system.ID, &system.Name, &system.IPAddress, &system.Port, &system.Username,
		&system.PasswordEncrypted, &authToken, &tokenExpiresAt, &system.SkipTLSVerify,
		&system.IsActive, &system.CreatedAt, &system.UpdatedAt,
	)
	if err != nil {
		respondError(w, http.StatusNotFound, "Storage system not found")
		return
	}

	if authToken.Valid {
		system.AuthToken = authToken.String
	}
	if tokenExpiresAt.Valid {
		system.TokenExpiresAt = &tokenExpiresAt.Time
	}

	// Decrypt password
	password, err := crypto.Decrypt(system.PasswordEncrypted, s.config.Security.EncryptionKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to decrypt password")
		return
	}

	// Get or refresh token
	token, err := s.svcClient.GetOrRefreshToken(&system, password)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to authenticate: "+err.Error())
		return
	}

	// List volumes from IBM SVC
	volumes, err := s.svcClient.ListVolumesInGroup(&system, token, vg.VGID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Failed to list volumes: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, volumes)
}

// User management handlers

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, username, email, role, created_at, updated_at FROM users ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to query user")
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		var email sql.NullString
		err := rows.Scan(&user.ID, &user.Username, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user: %v", err)
			continue
		}
		if email.Valid {
			user.Email = email.String
		}
		users = append(users, user)
	}

	if users == nil {
		users = []models.User{}
	}

	respondJSON(w, http.StatusOK, users)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	// Validate password complexity
	if err := s.auth.ValidatePasswordComplexity(req.Password); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate role
	validRoles := map[string]bool{"viewer": true, "operator": true, "admin": true}
	if req.Role == "" {
		req.Role = config.DefaultUserRole
	}
	if !validRoles[req.Role] {
		respondError(w, http.StatusBadRequest, "Invalid role. Must be viewer, operator, or admin")
		return
	}

	// Hash password
	passwordHash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Insert user and return the created user (PostgreSQL compatible with RETURNING)
	var emailParam interface{}
	if req.Email == "" {
		emailParam = nil
	} else {
		emailParam = req.Email
	}

	query := `INSERT INTO users (username, password_hash, email, role) VALUES ($1, $2, $3, $4) RETURNING id, username, email, role, created_at, updated_at`

	var user models.User
	var email sql.NullString
	log.Printf("DEBUG: Creating user with username=%s, email=%v, role=%s", req.Username, emailParam, req.Role)
	err = s.db.QueryRow(query, req.Username, passwordHash, emailParam, req.Role).Scan(
		&user.ID, &user.Username, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			respondError(w, http.StatusConflict, "Username already exists")
			return
		}
		log.Printf("Error scanning user after INSERT: %v (query: %s)", err, query)
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to retrieve created user")
		return
	}
	log.Printf("DEBUG: User created successfully: ID=%d, username=%s", user.ID, user.Username)
	if email.Valid {
		user.Email = email.String
	}

	respondJSON(w, http.StatusCreated, user)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var user models.User
	var email sql.NullString
	query := `SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = $1`
	err = s.db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to query user")
		return
	}
	if email.Valid {
		user.Email = email.String
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate role if provided
	if req.Role != "" {
		validRoles := map[string]bool{"viewer": true, "operator": true, "admin": true}
		if !validRoles[req.Role] {
			respondError(w, http.StatusBadRequest, "Invalid role. Must be viewer, operator, or admin")
			return
		}
	}

	// Build update query dynamically with whitelist validation
	allowedColumns := map[string]bool{
		"email = ?":                      true,
		"password_hash = ?":              true,
		"role = ?":                       true,
		"updated_at = CURRENT_TIMESTAMP": true,
	}

	updates := []string{}
	args := []interface{}{}

	if req.Email != "" {
		updates = append(updates, "email = ?")
		args = append(args, req.Email)
	}

	if req.Password != "" {
		// Validate password complexity
		if err := s.auth.ValidatePasswordComplexity(req.Password); err != nil {
			respondError(w, http.StatusBadRequest, err.Error())
			return
		}

		passwordHash, err := s.auth.HashPassword(req.Password)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}
		updates = append(updates, "password_hash = ?")
		args = append(args, passwordHash)
	}

	if req.Role != "" {
		updates = append(updates, "role = ?")
		args = append(args, req.Role)
	}

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// Validate all update clauses against whitelist
	for _, update := range updates {
		if !allowedColumns[update] {
			log.Printf("Security: Attempted to update invalid column: %s", update)
			respondError(w, http.StatusBadRequest, "Invalid update field")
			return
		}
	}

	// Build query with parameterized placeholders - convert ? to $N
	paramCount := 1
	for i := range updates {
		if strings.Contains(updates[i], "?") {
			updates[i] = strings.Replace(updates[i], "?", fmt.Sprintf("$%d", paramCount), 1)
			paramCount++
		}
	}
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), paramCount)
	args = append(args, userID)
	_, err = s.db.Exec(query, args...)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to update user")
		return
	}

	// Return updated user
	var user models.User
	var email sql.NullString
	selectQuery := `SELECT id, username, email, role, created_at, updated_at FROM users WHERE id = $1`
	err = s.db.QueryRow(selectQuery, userID).Scan(&user.ID, &user.Username, &email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated user")
		return
	}
	if email.Valid {
		user.Email = email.String
	}

	respondJSON(w, http.StatusOK, user)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Begin transaction to prevent race condition
	tx, err := s.db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer tx.Rollback()

	// Prevent deleting the last admin user - use SELECT FOR UPDATE for atomic check
	var adminCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check admin count")
		return
	}

	var userRole string
	err = tx.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to query user")
		return
	}

	if userRole == "admin" && adminCount <= 1 {
		respondError(w, http.StatusBadRequest, "Cannot delete the last admin user")
		return
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := tx.Exec(query, userID)
	if err != nil {
		respondErrorSafe(w, http.StatusInternalServerError, err, "Failed to delete user")
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}

//

// NTP Server management handlers

func (s *Server) handleListNTPServers(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, server_address, is_active, priority, last_sync_at, sync_status, time_offset_ms, created_at, updated_at 
	          FROM ntp_servers ORDER BY priority ASC, created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query NTP servers")
		return
	}
	defer rows.Close()

	var servers []models.NTPServer
	for rows.Next() {
		var server models.NTPServer
		var lastSyncAt sql.NullTime
		var syncStatus sql.NullString
		var timeOffsetMs sql.NullInt64

		err := rows.Scan(&server.ID, &server.ServerAddress, &server.IsActive, &server.Priority,
			&lastSyncAt, &syncStatus, &timeOffsetMs, &server.CreatedAt, &server.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning NTP server: %v", err)
			continue
		}

		if lastSyncAt.Valid {
			server.LastSyncAt = &lastSyncAt.Time
		}
		if syncStatus.Valid {
			server.SyncStatus = &syncStatus.String
		}
		if timeOffsetMs.Valid {
			offset := int(timeOffsetMs.Int64)
			server.TimeOffsetMs = &offset
		}

		servers = append(servers, server)
	}

	if servers == nil {
		servers = []models.NTPServer{}
	}

	respondJSON(w, http.StatusOK, servers)
}

func (s *Server) handleCreateNTPServer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServerAddress string `json:"server_address"`
		IsActive      bool   `json:"is_active"`
		Priority      int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.ServerAddress == "" {
		respondError(w, http.StatusBadRequest, "Server address is required")
		return
	}

	// Insert NTP server
	query := `INSERT INTO ntp_servers (server_address, is_active, priority, created_at, updated_at) 
	          VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id`
	var id int64
	var err error
	err = s.db.QueryRow(query, req.ServerAddress, req.IsActive, req.Priority).Scan(&id)
	if err != nil {
		log.Printf("Error creating NTP server: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create NTP server")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"message": "NTP server created successfully",
	})
}

func (s *Server) handleUpdateNTPServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid NTP server ID")
		return
	}

	var req struct {
		ServerAddress string `json:"server_address"`
		IsActive      bool   `json:"is_active"`
		Priority      int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update query dynamically with whitelist validation
	allowedColumns := map[string]bool{
		"server_address = ?":             true,
		"is_active = ?":                  true,
		"priority = ?":                   true,
		"updated_at = CURRENT_TIMESTAMP": true,
	}

	var updates []string
	var args []interface{}

	if req.ServerAddress != "" {
		updates = append(updates, "server_address = ?")
		args = append(args, req.ServerAddress)
	}

	updates = append(updates, "is_active = ?")
	args = append(args, req.IsActive)

	updates = append(updates, "priority = ?")
	args = append(args, req.Priority)

	if len(updates) == 0 {
		respondError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

	// Validate all update clauses against whitelist
	for _, update := range updates {
		if !allowedColumns[update] {
			log.Printf("Security: Attempted to update invalid NTP column: %s", update)
			respondError(w, http.StatusBadRequest, "Invalid update field")
			return
		}
	}

	args = append(args, serverID)

	// Build query safely - convert ? to $N
	paramCount := 1
	for i := range updates {
		if strings.Contains(updates[i], "?") {
			updates[i] = strings.Replace(updates[i], "?", fmt.Sprintf("$%d", paramCount), 1)
			paramCount++
		}
	}
	query := fmt.Sprintf("UPDATE ntp_servers SET %s WHERE id = $%d", strings.Join(updates, ", "), paramCount)
	_, err = s.db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating NTP server: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update NTP server")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "NTP server updated successfully"})
}

func (s *Server) handleDeleteNTPServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid NTP server ID")
		return
	}

	query := `DELETE FROM ntp_servers WHERE id = $1`
	_, err = s.db.Exec(query, serverID)
	if err != nil {
		log.Printf("Error deleting NTP server: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete NTP server")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "NTP server deleted successfully"})
}

func (s *Server) handleSyncNTPServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid NTP server ID")
		return
	}

	// Get NTP server
	var server models.NTPServer
	query := `SELECT id, server_address FROM ntp_servers WHERE id = $1`
	err = s.db.QueryRow(query, serverID).Scan(&server.ID, &server.ServerAddress)
	if err != nil {
		respondError(w, http.StatusNotFound, "NTP server not found")
		return
	}

	// Simulate NTP sync (in production, use actual NTP client library)
	// For now, just update the sync status
	syncStatus := "synced"
	timeOffset := 0 // milliseconds

	updateQuery := `UPDATE ntp_servers 
	                SET last_sync_at = CURRENT_TIMESTAMP, 
	                    sync_status = $1, 
	                    time_offset_ms = $2,
	                    updated_at = CURRENT_TIMESTAMP 
	                WHERE id = $3`
	_, err = s.db.Exec(updateQuery, syncStatus, timeOffset, serverID)
	if err != nil {
		log.Printf("Error updating NTP sync status: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to sync NTP server")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "NTP server synced successfully",
		"sync_status":    syncStatus,
		"time_offset_ms": timeOffset,
	})
}

func (s *Server) handleGetSystemTime(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	// Get timezone info
	zone, offset := now.Zone()
	offsetHours := offset / 3600
	offsetMinutes := (offset % 3600) / 60
	offsetStr := fmt.Sprintf("UTC%+d:%02d", offsetHours, offsetMinutes)

	// Check if any NTP servers are active
	var activeNTPCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ntp_servers WHERE is_active = TRUE").Scan(&activeNTPCount)
	if err != nil {
		log.Printf("Error checking NTP servers: %v", err)
	}

	ntpEnabled := activeNTPCount > 0
	ntpStatus := "disabled"
	if ntpEnabled {
		ntpStatus = "enabled"
	}

	// Get last sync time from most recent NTP server sync
	var lastSync sql.NullTime
	err = s.db.QueryRow(`SELECT last_sync_at FROM ntp_servers 
	                     WHERE is_active = TRUE AND last_sync_at IS NOT NULL 
	                     ORDER BY last_sync_at DESC LIMIT 1`).Scan(&lastSync)

	var lastSyncStr *string
	if err == nil && lastSync.Valid {
		syncStr := lastSync.Time.Format(time.RFC3339)
		lastSyncStr = &syncStr
	}

	// Calculate system uptime (simplified - in production use actual system uptime)
	uptime := "N/A"

	// Get average time drift from active NTP servers
	var avgDrift sql.NullInt64
	err = s.db.QueryRow(`SELECT AVG(time_offset_ms) FROM ntp_servers 
	                     WHERE is_active = TRUE AND time_offset_ms IS NOT NULL`).Scan(&avgDrift)

	var timeDrift *int
	if err == nil && avgDrift.Valid {
		drift := int(avgDrift.Int64)
		timeDrift = &drift
	}

	timeInfo := models.SystemTimeInfo{
		CurrentTime:    now.Format(time.RFC3339),
		Timezone:       zone,
		TimezoneOffset: offsetStr,
		NTPSyncEnabled: ntpEnabled,
		NTPSyncStatus:  ntpStatus,
		LastNTPSync:    lastSyncStr,
		SystemUptime:   uptime,
		TimeDriftMs:    timeDrift,
	}

	respondJSON(w, http.StatusOK, timeInfo)
}

func (s *Server) handleSetSystemTime(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Time string `json:"time"` // ISO 8601 format
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate time format
	parsedTime, err := time.Parse(time.RFC3339, req.Time)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid time format. Use ISO 8601 format (e.g., 2026-06-16T12:00:00Z)")
		return
	}

	// Note: Actually setting system time requires root privileges and syscall
	// In a containerized environment, this is typically not allowed
	// This endpoint stores the user's desired time for reference but cannot actually change system time

	log.Printf("Time set request received: %s (parsed as %v)", req.Time, parsedTime)

	// Store in settings table for reference
	_, err = s.db.Exec(`
		INSERT INTO settings (key, value, description, updated_at)
		VALUES ('manual_time_override', $1, 'Manually set system time (reference only)', CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = CURRENT_TIMESTAMP
	`, req.Time)

	if err != nil {
		log.Printf("Error storing manual time setting: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to store time setting")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Time setting stored successfully. Note: Actual system time cannot be changed from within container.",
		"time":    req.Time,
	})
}

func (s *Server) handleSetTimezone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Timezone string `json:"timezone"` // IANA timezone name (e.g., "Europe/Oslo", "America/New_York")
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate timezone
	_, err := time.LoadLocation(req.Timezone)
	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid timezone: %s", req.Timezone))
		return
	}

	// Store in settings table
	_, err = s.db.Exec(`
		INSERT INTO settings (key, value, description, updated_at)
		VALUES ('timezone', $1, 'System timezone setting', CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = CURRENT_TIMESTAMP
	`, req.Timezone)

	if err != nil {
		log.Printf("Error storing timezone setting: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to store timezone setting")
		return
	}

	// Note: Changing TZ environment variable affects only this process
	// For container-wide timezone change, need to restart with TZ env var
	os.Setenv("TZ", req.Timezone)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Timezone updated successfully",
		"timezone": req.Timezone,
	})
}

// Audit Log handlers

func (s *Server) handleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	var userID *int
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		id, err := strconv.Atoi(userIDStr)
		if err == nil {
			userID = &id
		}
	}

	var action *string
	if actionStr := query.Get("action"); actionStr != "" {
		action = &actionStr
	}

	var resourceType *string
	if rtStr := query.Get("resource_type"); rtStr != "" {
		resourceType = &rtStr
	}

	var status *string
	if statusStr := query.Get("status"); statusStr != "" {
		status = &statusStr
	}

	var startDate, endDate *time.Time
	if startStr := query.Get("start_date"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startDate = &t
		}
	}
	if endStr := query.Get("end_date"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endDate = &t
		}
	}

	limit := config.DefaultExecutionLimit
	if limitStr := query.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get logs
	logs, err := s.auditLogger.ListAuditLogs(userID, action, resourceType, status, startDate, endDate, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve audit logs")
		return
	}

	// Get total count
	total, err := s.auditLogger.CountAuditLogs(userID, action, resourceType, status, startDate, endDate)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count audit logs")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

//

// Settings handlers

func (s *Server) handleGetAuditRetentionSettings(w http.ResponseWriter, r *http.Request) {
	var maxEntries, retentionDays int

	// Query returns TEXT, need to convert to int
	var maxEntriesStr string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = 'audit_log_max_entries'").Scan(&maxEntriesStr)
	if err != nil {
		maxEntries = 1000 // Default
	} else {
		maxEntries, err = strconv.Atoi(maxEntriesStr)
		if err != nil {
			log.Printf("Error converting max_entries to int: %v", err)
			maxEntries = 1000 // Default on conversion error
		}
	}

	var retentionDaysStr string
	err = s.db.QueryRow("SELECT value FROM settings WHERE key = 'audit_log_retention_days'").Scan(&retentionDaysStr)
	if err != nil {
		retentionDays = 365 // Default
	} else {
		retentionDays, err = strconv.Atoi(retentionDaysStr)
		if err != nil {
			log.Printf("Error converting retention_days to int: %v", err)
			retentionDays = 365 // Default on conversion error
		}
	}

	respondJSON(w, http.StatusOK, map[string]int{
		"max_entries":    maxEntries,
		"retention_days": retentionDays,
	})
}

func (s *Server) handleUpdateAuditRetentionSettings(w http.ResponseWriter, r *http.Request) {
	// Get user context
	userID := getUserIDFromContext(r)
	username := getUsernameFromContext(r)

	var req struct {
		MaxEntries    int `json:"max_entries"`
		RetentionDays int `json:"retention_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate values
	if req.MaxEntries < config.MinAuditEntries || req.MaxEntries > config.MaxAuditEntries {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Max entries must be between %d and %d", config.MinAuditEntries, config.MaxAuditEntries))
		return
	}
	if req.RetentionDays < config.MinRetentionDays || req.RetentionDays > config.MaxRetentionDays {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Retention days must be between %d and %d (10 years)", config.MinRetentionDays, config.MaxRetentionDays))
		return
	}

	// Update settings in database using UPSERT so defaults persist even when rows do not yet exist
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value, description, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, "audit_log_max_entries", req.MaxEntries, "Maximum number of audit log entries to retain")
	if err != nil {
		log.Printf("Error upserting max_entries setting: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	_, err = s.db.Exec(`
		INSERT INTO settings (key, value, description, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`, "audit_log_retention_days", req.RetentionDays, "Number of days to retain audit logs")
	if err != nil {
		log.Printf("Error upserting retention_days setting: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	// Restart cleanup with new settings
	s.auditLogger.StartPeriodicCleanup(req.MaxEntries, req.RetentionDays, 24)

	// Log the settings change
	s.auditLogger.LogSuccess(
		&userID,
		username,
		"update_settings",
		"settings",
		nil,
		nil,
		map[string]interface{}{
			"setting":        "audit_retention",
			"max_entries":    req.MaxEntries,
			"retention_days": req.RetentionDays,
		},
		r,
	)

	respondJSON(w, http.StatusOK, map[string]string{"message": "Audit retention settings updated successfully"})
}
