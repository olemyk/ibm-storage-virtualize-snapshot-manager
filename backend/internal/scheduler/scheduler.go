package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/ibm-storage-virtualize-snapshot-manager/internal/config"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/db"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/models"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/notification"
	"github.com/ibm-storage-virtualize-snapshot-manager/internal/svc"
	"github.com/ibm-storage-virtualize-snapshot-manager/pkg/crypto"
	"github.com/robfig/cron/v3"
)

// Scheduler manages snapshot schedules
type Scheduler struct {
	cron          *cron.Cron
	db            *db.DB
	svcClient     *svc.Client
	encryptionKey string
	jobs          map[int]cron.EntryID
	mu            sync.RWMutex
	notifier      *notification.Notifier
	executing     map[int]bool // Track schedules currently executing
	executingMu   sync.Mutex   // Mutex for executing map
}

// New creates a new scheduler
func New(database *db.DB, svcClient *svc.Client, encryptionKey string) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		db:            database,
		svcClient:     svcClient,
		encryptionKey: encryptionKey,
		jobs:          make(map[int]cron.EntryID),
		executing:     make(map[int]bool),
		notifier:      nil, // Will be set via SetNotifier
	}
}

// SetNotifier sets the notifier for the scheduler
func (s *Scheduler) SetNotifier(notifier *notification.Notifier) {
	s.notifier = notifier
}

// Start starts the scheduler and loads all active schedules
func (s *Scheduler) Start() error {
	// Load all active schedules from database
	schedules, err := s.loadActiveSchedules()
	if err != nil {
		return fmt.Errorf("failed to load schedules: %w", err)
	}

	// Add each schedule to cron
	for _, schedule := range schedules {
		if err := s.AddSchedule(&schedule); err != nil {
			log.Printf("Failed to add schedule %d: %v", schedule.ID, err)
		}
	}

	s.cron.Start()
	log.Printf("Scheduler started with %d active schedules", len(schedules))
	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

// AddSchedule adds a schedule to the cron scheduler
func (s *Scheduler) AddSchedule(schedule *models.SnapshotSchedule) error {
	if !schedule.IsActive {
		return nil
	}

	// Validate cron expression
	if _, err := cron.ParseStandard(schedule.CronExpression); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	// Remove existing job if any
	s.RemoveSchedule(schedule.ID)

	// Add new job
	entryID, err := s.cron.AddFunc(schedule.CronExpression, func() {
		if err := s.ExecuteSnapshot(schedule); err != nil {
			log.Printf("Failed to execute snapshot for schedule %d: %v", schedule.ID, err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.mu.Lock()
	s.jobs[schedule.ID] = entryID
	s.mu.Unlock()

	// Update next execution time
	nextExec := s.calculateNextExecution(schedule.CronExpression)
	if err := s.updateNextExecution(schedule.ID, nextExec); err != nil {
		log.Printf("Failed to update next execution time: %v", err)
	}

	log.Printf("Added schedule %d (%s) with cron expression: %s", schedule.ID, schedule.Name, schedule.CronExpression)
	return nil
}

// RemoveSchedule removes a schedule from the cron scheduler
func (s *Scheduler) RemoveSchedule(scheduleID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entryID, exists := s.jobs[scheduleID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, scheduleID)
		log.Printf("Removed schedule %d", scheduleID)
	}
}

// ExecuteSnapshot executes a snapshot for a schedule (exported for manual execution)
func (s *Scheduler) ExecuteSnapshot(schedule *models.SnapshotSchedule) error {
	// Check if already executing (idempotency)
	s.executingMu.Lock()
	if s.executing[schedule.ID] {
		log.Printf("Schedule %d (%s) is already executing, skipping", schedule.ID, schedule.Name)
		s.executingMu.Unlock()
		return nil
	}
	s.executing[schedule.ID] = true
	s.executingMu.Unlock()

	// Ensure we mark as not executing when done
	defer func() {
		s.executingMu.Lock()
		delete(s.executing, schedule.ID)
		s.executingMu.Unlock()
	}()

	log.Printf("Executing snapshot for schedule %d (%s)", schedule.ID, schedule.Name)

	// Get volume group
	vg, err := s.getVolumeGroup(schedule.VolumeGroupID)
	if err != nil {
		return s.logExecution(schedule, "", "failed", fmt.Sprintf("Failed to get volume group: %v", err))
	}

	// Get storage system
	system, err := s.getStorageSystem(vg.StorageSystemID)
	if err != nil {
		return s.logExecution(schedule, "", "failed", fmt.Sprintf("Failed to get storage system: %v", err))
	}

	// Decrypt password
	password, err := crypto.Decrypt(system.PasswordEncrypted, s.encryptionKey)
	if err != nil {
		return s.logExecution(schedule, "", "failed", fmt.Sprintf("Failed to decrypt password: %v", err))
	}

	// Get or refresh token
	token, err := s.svcClient.GetOrRefreshToken(system, password)
	if err != nil {
		return s.logExecution(schedule, "", "failed", fmt.Sprintf("Failed to authenticate: %v", err))
	}

	// Generate snapshot name using pattern
	snapshotName := s.generateSnapshotName(schedule, vg)

	// Validate retention parameters - at least one must be set
	retentionMinutesSet := schedule.RetentionMinutes != nil && *schedule.RetentionMinutes > 0
	if schedule.RetentionDays == 0 && !retentionMinutesSet {
		return s.logExecution(schedule, snapshotName, "failed", "Either RetentionDays or RetentionMinutes must be set")
	}

	// Execute snapshot
	var retentionDays *int
	if schedule.RetentionDays > 0 {
		retentionDays = &schedule.RetentionDays
	}

	req := svc.AddSnapshotRequest{
		VolumeGroup:      vg.VGName,
		RetentionDays:    retentionDays,
		RetentionMinutes: schedule.RetentionMinutes,
		Safeguarded:      schedule.Safeguarded,
		Name:             snapshotName,
		Pool:             schedule.PoolName,
	}

	result, err := s.svcClient.AddSnapshot(system, token, req)
	if err != nil {
		// Log failed execution
		s.logExecution(schedule, snapshotName, "failed", fmt.Sprintf("Failed to create snapshot: %v", err))

		// Send failure notification
		if s.notifier != nil {
			systemName := system.Name
			vgName := vg.VGName
			scheduleName := schedule.Name
			event := &notification.Event{
				Type:            notification.EventSnapshotFailure,
				Timestamp:       time.Now(),
				Severity:        notification.SeverityError,
				SystemID:        &system.ID,
				SystemName:      &systemName,
				VolumeGroupID:   &vg.ID,
				VolumeGroupName: &vgName,
				ScheduleID:      &schedule.ID,
				ScheduleName:    &scheduleName,
				Message:         fmt.Sprintf("Snapshot failed for volume group '%s' on system '%s'", vgName, systemName),
				Details: map[string]interface{}{
					"error":         err.Error(),
					"snapshot_name": snapshotName,
					"schedule_name": scheduleName,
					"volume_group":  vgName,
					"system":        systemName,
				},
			}
			go func() {
				if notifyErr := s.notifier.SendEvent(context.Background(), event); notifyErr != nil {
					log.Printf("Failed to send failure notification for schedule %d: %v", schedule.ID, notifyErr)
				}
			}()
		}

		return err
	}

	// Extract snapshot ID from result if available
	snapshotID := ""
	if id, ok := result["id"].(string); ok {
		snapshotID = id
	}

	// Log successful execution
	if err := s.logExecution(schedule, snapshotName, "success", ""); err != nil {
		log.Printf("Failed to log execution: %v", err)
	}

	// Update last executed time
	if err := s.updateLastExecution(schedule.ID, time.Now()); err != nil {
		log.Printf("Failed to update last execution time: %v", err)
	}

	// Update next execution time
	nextExec := s.calculateNextExecution(schedule.CronExpression)
	if err := s.updateNextExecution(schedule.ID, nextExec); err != nil {
		log.Printf("Failed to update next execution time: %v", err)
	}

	// Send success notification
	if s.notifier != nil {
		systemName := system.Name
		vgName := vg.VGName
		scheduleName := schedule.Name
		event := &notification.Event{
			Type:            notification.EventSnapshotSuccess,
			Timestamp:       time.Now(),
			Severity:        notification.SeverityInfo,
			SystemID:        &system.ID,
			SystemName:      &systemName,
			VolumeGroupID:   &vg.ID,
			VolumeGroupName: &vgName,
			ScheduleID:      &schedule.ID,
			ScheduleName:    &scheduleName,
			Message:         fmt.Sprintf("Snapshot created successfully for volume group '%s' on system '%s'", vgName, systemName),
			Details: map[string]interface{}{
				"snapshot_id":    snapshotID,
				"snapshot_name":  snapshotName,
				"schedule_name":  scheduleName,
				"volume_group":   vgName,
				"system":         systemName,
				"retention_days": schedule.RetentionDays,
				"safeguarded":    schedule.Safeguarded,
			},
		}
		go s.notifier.SendEvent(context.Background(), event)
	}

	log.Printf("Successfully created snapshot %s (ID: %s) for schedule %d", snapshotName, snapshotID, schedule.ID)
	return nil
}

// logExecution logs a snapshot execution to the database
func (s *Scheduler) logExecution(schedule *models.SnapshotSchedule, snapshotName, status, errorMsg string) error {
	var snapshotNamePtr *string
	if snapshotName != "" {
		snapshotNamePtr = &snapshotName
	}

	var errorMsgPtr *string
	if errorMsg != "" {
		errorMsgPtr = &errorMsg
	}

	query := `
		INSERT INTO snapshot_executions (schedule_id, volume_group_id, snapshot_name, status, error_message, retention_days, retention_minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(
		query,
		schedule.ID,
		schedule.VolumeGroupID,
		snapshotNamePtr,
		status,
		errorMsgPtr,
		schedule.RetentionDays,
		schedule.RetentionMinutes,
	)
	return err
}

// loadActiveSchedules loads all active schedules from the database
func (s *Scheduler) loadActiveSchedules() ([]models.SnapshotSchedule, error) {
	query := `SELECT id, volume_group_id, name, cron_expression, retention_days, retention_minutes,
	          safeguarded, pool_name, snapshot_name_pattern, is_active, last_executed_at,
	          next_execution_at, created_at, updated_at
	          FROM snapshot_schedules WHERE is_active = TRUE`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.SnapshotSchedule
	for rows.Next() {
		var schedule models.SnapshotSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.VolumeGroupID,
			&schedule.Name,
			&schedule.CronExpression,
			&schedule.RetentionDays,
			&schedule.RetentionMinutes,
			&schedule.Safeguarded,
			&schedule.PoolName,
			&schedule.SnapshotNamePattern,
			&schedule.IsActive,
			&schedule.LastExecutedAt,
			&schedule.NextExecutionAt,
			&schedule.CreatedAt,
			&schedule.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// getVolumeGroup gets a volume group by ID
func (s *Scheduler) getVolumeGroup(id int) (*models.VolumeGroup, error) {
	query := `SELECT id, storage_system_id, vg_id, vg_name, partition_id, partition_name, last_synced_at, created_at FROM volume_groups WHERE id = $1`

	var vg models.VolumeGroup
	err := s.db.QueryRow(query, id).Scan(
		&vg.ID,
		&vg.StorageSystemID,
		&vg.VGID,
		&vg.VGName,
		&vg.PartitionID,
		&vg.PartitionName,
		&vg.LastSyncedAt,
		&vg.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &vg, nil
}

// getStorageSystem gets a storage system by ID
func (s *Scheduler) getStorageSystem(id int) (*models.StorageSystem, error) {
	query := `SELECT id, name, ip_address, port, username, password_encrypted, auth_token, token_expires_at, skip_tls_verify, is_active, created_at, updated_at FROM storage_systems WHERE id = $1`

	var system models.StorageSystem
	var authToken sql.NullString
	var tokenExpiresAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&system.ID,
		&system.Name,
		&system.IPAddress,
		&system.Port,
		&system.Username,
		&system.PasswordEncrypted,
		&authToken,
		&tokenExpiresAt,
		&system.SkipTLSVerify,
		&system.IsActive,
		&system.CreatedAt,
		&system.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if authToken.Valid {
		system.AuthToken = authToken.String
	}
	if tokenExpiresAt.Valid {
		system.TokenExpiresAt = &tokenExpiresAt.Time
	}

	return &system, nil
}

// updateLastExecution updates the last execution time for a schedule
func (s *Scheduler) updateLastExecution(scheduleID int, execTime time.Time) error {
	query := `UPDATE snapshot_schedules SET last_executed_at = $1 WHERE id = $2`
	_, err := s.db.Exec(query, execTime, scheduleID)
	return err
}

// updateNextExecution updates the next execution time for a schedule
func (s *Scheduler) updateNextExecution(scheduleID int, nextExec time.Time) error {
	query := `UPDATE snapshot_schedules SET next_execution_at = $1 WHERE id = $2`
	_, err := s.db.Exec(query, nextExec, scheduleID)
	return err
}

// CalculateNextExecution calculates the next execution time for a cron expression
func (s *Scheduler) CalculateNextExecution(cronExpr string) time.Time {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Now()
	}
	return schedule.Next(time.Now())
}

// calculateNextExecution is an internal helper that calls the exported method
func (s *Scheduler) calculateNextExecution(cronExpr string) time.Time {
	return s.CalculateNextExecution(cronExpr)
}

// ValidateCronExpression validates a cron expression
func ValidateCronExpression(expr string) error {
	_, err := cron.ParseStandard(expr)
	return err
}

//

// generateSnapshotName generates a snapshot name based on the pattern
// Supported placeholders:
// {schedule_name} - Name of the schedule
// {vg_name} - Volume group name
// {timestamp} - Current timestamp in format YYYYMMDD_HHMMSS
// {date} - Current date in format YYYYMMDD
// {time} - Current time in format HHMMSS
// {year} - Current year (YYYY)
// {month} - Current month (MM)
// {day} - Current day (DD)
func (s *Scheduler) generateSnapshotName(schedule *models.SnapshotSchedule, vg *models.VolumeGroup) string {
	pattern := schedule.SnapshotNamePattern
	if pattern == "" {
		pattern = config.DefaultSnapshotNamePattern
	}

	now := time.Now()

	// Replace placeholders
	pattern = strings.ReplaceAll(pattern, "{schedule_name}", schedule.Name)
	pattern = strings.ReplaceAll(pattern, "{vg_name}", vg.VGName)
	pattern = strings.ReplaceAll(pattern, "{timestamp}", now.Format("20060102_150405"))
	pattern = strings.ReplaceAll(pattern, "{date}", now.Format("20060102"))
	pattern = strings.ReplaceAll(pattern, "{time}", now.Format("150405"))
	pattern = strings.ReplaceAll(pattern, "{year}", now.Format("2006"))
	pattern = strings.ReplaceAll(pattern, "{month}", now.Format("01"))
	pattern = strings.ReplaceAll(pattern, "{day}", now.Format("02"))

	return pattern
}
