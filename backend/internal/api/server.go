package api

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/audit"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/auth"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/db"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/scheduler"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/svc"
	"github.com/rs/cors"
)

// Server represents the API server
type Server struct {
	config              *config.Config
	db                  *db.DB
	auth                *auth.Service
	svcClient           *svc.Client
	scheduler           *scheduler.Scheduler
	auditLogger         *audit.Logger
	router              *mux.Router
	notificationManager *notification.Manager
	notifier            *notification.Notifier
	csrfManager         *CSRFTokenManager
	tokenBlacklist      *TokenBlacklist
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	db *db.DB,
	authService *auth.Service,
	svcClient *svc.Client,
	snapshotScheduler *scheduler.Scheduler,
	notificationMgr *notification.Manager,
	notif *notification.Notifier,
) *Server {
	s := &Server{
		config:              cfg,
		db:                  db,
		auth:                authService,
		svcClient:           svcClient,
		scheduler:           snapshotScheduler,
		auditLogger:         audit.NewLogger(db.DB),
		router:              mux.NewRouter(),
		notificationManager: notificationMgr,
		notifier:            notif,
		csrfManager:         NewCSRFTokenManager(),
		tokenBlacklist:      NewTokenBlacklist(),
	}

	// Load audit retention settings from database and start periodic cleanup
	maxEntries, retentionDays, err := s.loadAuditRetentionSettings()
	if err != nil {
		// Use defaults if settings not found
		maxEntries = 1000
		retentionDays = 365
	}
	s.auditLogger.StartPeriodicCleanup(maxEntries, retentionDays, 24)

	s.setupRoutes()
	return s
}

// loadAuditRetentionSettings loads audit retention settings from database
func (s *Server) loadAuditRetentionSettings() (int, int, error) {
	var maxEntries, retentionDays int

	err := s.db.DB.QueryRow(`
		SELECT value FROM settings WHERE key = 'audit_max_entries'
	`).Scan(&maxEntries)
	if err != nil {
		return 0, 0, err
	}

	err = s.db.DB.QueryRow(`
		SELECT value FROM settings WHERE key = 'audit_retention_days'
	`).Scan(&retentionDays)
	if err != nil {
		return 0, 0, err
	}

	return maxEntries, retentionDays, nil
}

// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Apply security headers to all routes
	s.router.Use(s.securityHeadersMiddleware)

	// Create rate limiter for login endpoint (5 attempts per minute per IP)
	loginRateLimiter := NewRateLimiter(5, time.Minute)

	// API v1 routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Authentication routes (public) with rate limiting
	loginRouter := api.PathPrefix("/auth").Subrouter()
	loginRouter.Use(s.rateLimitMiddleware(loginRateLimiter))

	// Wrap handleLogin with debug logging
	loginRouter.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("=== LOGIN REQUEST RECEIVED === Method: %s, Path: %s, RemoteAddr: %s", r.Method, r.URL.Path, r.RemoteAddr)
		s.handleLogin(w, r)
		log.Printf("=== LOGIN REQUEST COMPLETED ===")
	}).Methods("POST")

	api.HandleFunc("/auth/logout", s.handleLogout).Methods("POST")
	api.HandleFunc("/auth/csrf-token", s.handleGetCSRFToken).Methods("GET")

	// Protected routes with CSRF protection
	protected := api.PathPrefix("").Subrouter()
	protected.Use(s.authMiddleware)
	protected.Use(s.csrfMiddleware)

	// User routes
	protected.HandleFunc("/auth/me", s.handleGetCurrentUser).Methods("GET")

	adminOnly := protected.PathPrefix("").Subrouter()
	adminOnly.Use(s.requireRoles("admin"))
	adminOnly.HandleFunc("/users", s.handleListUsers).Methods("GET")
	adminOnly.HandleFunc("/users", s.handleCreateUser).Methods("POST")
	adminOnly.HandleFunc("/users/{id}", s.handleGetUser).Methods("GET")
	adminOnly.HandleFunc("/users/{id}", s.handleUpdateUser).Methods("PUT")
	adminOnly.HandleFunc("/users/{id}", s.handleDeleteUser).Methods("DELETE")

	operatorOrAdmin := protected.PathPrefix("").Subrouter()
	operatorOrAdmin.Use(s.requireRoles("operator", "admin"))

	// Storage system routes
	operatorOrAdmin.HandleFunc("/systems", s.handleListSystems).Methods("GET")
	operatorOrAdmin.HandleFunc("/systems", s.handleCreateSystem).Methods("POST")
	operatorOrAdmin.HandleFunc("/systems/{id}", s.handleGetSystem).Methods("GET")
	operatorOrAdmin.HandleFunc("/systems/{id}", s.handleUpdateSystem).Methods("PUT")
	operatorOrAdmin.HandleFunc("/systems/{id}", s.handleDeleteSystem).Methods("DELETE")
	operatorOrAdmin.HandleFunc("/systems/{id}/test", s.handleTestSystem).Methods("POST")
	operatorOrAdmin.HandleFunc("/systems/health-check", s.handleCheckSystemsHealth).Methods("POST")
	operatorOrAdmin.HandleFunc("/systems/{id}/volumegroups", s.handleListVolumeGroups).Methods("GET")
	operatorOrAdmin.HandleFunc("/systems/{id}/volumegroups/sync", s.handleSyncVolumeGroups).Methods("POST")

	// Volume group routes
	operatorOrAdmin.HandleFunc("/volumegroups/{id}", s.handleGetVolumeGroup).Methods("GET")
	operatorOrAdmin.HandleFunc("/volumegroups/{id}/snapshots", s.handleListSnapshots).Methods("GET")
	operatorOrAdmin.HandleFunc("/volumegroups/{id}/volumes", s.handleListVolumesInGroup).Methods("GET")
	operatorOrAdmin.HandleFunc("/volumegroups/{id}/schedules", s.handleListSchedules).Methods("GET")
	operatorOrAdmin.HandleFunc("/volumegroups/{id}/schedules", s.handleCreateSchedule).Methods("POST")

	// Schedule routes
	protected.HandleFunc("/schedules", s.handleListAllSchedules).Methods("GET")
	protected.HandleFunc("/schedules/{id}", s.handleGetSchedule).Methods("GET")
	protected.HandleFunc("/schedules/{id}", s.handleUpdateSchedule).Methods("PUT")
	protected.HandleFunc("/schedules/{id}", s.handleDeleteSchedule).Methods("DELETE")
	protected.HandleFunc("/schedules/{id}/execute", s.handleExecuteSchedule).Methods("POST")

	// Execution routes
	protected.HandleFunc("/executions", s.handleListExecutions).Methods("GET")
	protected.HandleFunc("/executions/{id}", s.handleGetExecution).Methods("GET")

	// Dashboard routes
	protected.HandleFunc("/dashboard/stats", s.handleGetDashboardStats).Methods("GET")

	// Audit log routes
	protected.HandleFunc("/audit-logs", s.handleListAuditLogs).Methods("GET")

	// Settings routes
	protected.HandleFunc("/settings/audit-retention", s.handleGetAuditRetentionSettings).Methods("GET")
	protected.HandleFunc("/settings/audit-retention", s.handleUpdateAuditRetentionSettings).Methods("PUT")

	// NTP Server routes
	protected.HandleFunc("/ntp/servers", s.handleListNTPServers).Methods("GET")
	protected.HandleFunc("/ntp/servers", s.handleCreateNTPServer).Methods("POST")
	protected.HandleFunc("/ntp/servers/{id}", s.handleUpdateNTPServer).Methods("PUT")
	protected.HandleFunc("/ntp/servers/{id}", s.handleDeleteNTPServer).Methods("DELETE")
	protected.HandleFunc("/ntp/servers/{id}/sync", s.handleSyncNTPServer).Methods("POST")
	protected.HandleFunc("/ntp/time", s.handleGetSystemTime).Methods("GET")
	protected.HandleFunc("/ntp/time", s.handleSetSystemTime).Methods("POST")
	protected.HandleFunc("/ntp/timezone", s.handleSetTimezone).Methods("PUT")

	// Notification routes (only if notification manager is available)
	if s.notificationManager != nil {
		notifications := protected.PathPrefix("/notifications").Subrouter()

		// Notification Channels
		notifications.HandleFunc("/channels", s.handleListNotificationChannels).Methods("GET")
		notifications.HandleFunc("/channels", s.handleCreateNotificationChannel).Methods("POST")
		notifications.HandleFunc("/channels/{id}", s.handleGetNotificationChannel).Methods("GET")
		notifications.HandleFunc("/channels/{id}", s.handleUpdateNotificationChannel).Methods("PUT")
		notifications.HandleFunc("/channels/{id}", s.handleDeleteNotificationChannel).Methods("DELETE")
		notifications.HandleFunc("/channels/{id}/test", s.handleTestNotificationChannel).Methods("POST")

		// Alert Rules
		notifications.HandleFunc("/rules", s.handleListAlertRules).Methods("GET")
		notifications.HandleFunc("/rules", s.handleCreateAlertRule).Methods("POST")
		notifications.HandleFunc("/rules/{id}", s.handleGetAlertRule).Methods("GET")
		notifications.HandleFunc("/rules/{id}", s.handleUpdateAlertRule).Methods("PUT")
		notifications.HandleFunc("/rules/{id}", s.handleDeleteAlertRule).Methods("DELETE")

		// Notification History
		notifications.HandleFunc("/history", s.handleListNotificationHistory).Methods("GET")

		// Test notification
		notifications.HandleFunc("/test", s.handleSendTestNotification).Methods("POST")
	}

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Setup CORS
	log.Printf("DEBUG: Configured AllowedOrigins: %v", s.config.Server.AllowedOrigins)
	c := cors.New(cors.Options{
		AllowedOrigins:   s.config.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
		Debug:            true,
	})

	handler := c.Handler(s.router)

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	return http.ListenAndServe(addr, handler)
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

//
