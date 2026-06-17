package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int       `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Email        string    `json:"email" db:"email"`
	Role         string    `json:"role" db:"role"` // viewer, operator, admin
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// StorageSystem represents an IBM Storage Virtualize system
type StorageSystem struct {
	ID                  int        `json:"id" db:"id"`
	Name                string     `json:"name" db:"name"`
	IPAddress           string     `json:"ip_address" db:"ip_address"`
	Port                int        `json:"port" db:"port"`
	Username            string     `json:"username" db:"username"`
	PasswordEncrypted   string     `json:"-" db:"password_encrypted"`
	AuthToken           string     `json:"-" db:"auth_token"`
	TokenExpiresAt      *time.Time `json:"-" db:"token_expires_at"`
	SkipTLSVerify       bool       `json:"skip_tls_verify" db:"skip_tls_verify"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	ConnectionStatus    *string    `json:"connection_status,omitempty" db:"connection_status"`
	LastConnectionCheck *time.Time `json:"last_connection_check,omitempty" db:"last_connection_check"`
	ConnectionError     *string    `json:"connection_error,omitempty" db:"connection_error"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// VolumeGroup represents a volume group from IBM SVC
type VolumeGroup struct {
	ID              int        `json:"id" db:"id"`
	StorageSystemID int        `json:"storage_system_id" db:"storage_system_id"`
	VGID            string     `json:"vg_id" db:"vg_id"`
	VGName          string     `json:"vg_name" db:"vg_name"`
	PartitionID     *string    `json:"partition_id,omitempty" db:"partition_id"`
	PartitionName   *string    `json:"partition_name,omitempty" db:"partition_name"`
	LastSyncedAt    *time.Time `json:"last_synced_at" db:"last_synced_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
}

// SnapshotSchedule represents a snapshot schedule for a volume group
type SnapshotSchedule struct {
	ID                  int        `json:"id" db:"id"`
	VolumeGroupID       int        `json:"volume_group_id" db:"volume_group_id"`
	Name                string     `json:"name" db:"name"`
	CronExpression      string     `json:"cron_expression" db:"cron_expression"`
	RetentionDays       int        `json:"retention_days" db:"retention_days"`
	RetentionMinutes    *int       `json:"retention_minutes,omitempty" db:"retention_minutes"`
	Safeguarded         bool       `json:"safeguarded" db:"safeguarded"`
	PoolName            *string    `json:"pool_name,omitempty" db:"pool_name"`
	SnapshotNamePattern string     `json:"snapshot_name_pattern" db:"snapshot_name_pattern"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	LastExecutedAt      *time.Time `json:"last_executed_at" db:"last_executed_at"`
	NextExecutionAt     *time.Time `json:"next_execution_at" db:"next_execution_at"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// SnapshotExecution represents a snapshot execution record
type SnapshotExecution struct {
	ID               int       `json:"id" db:"id"`
	ScheduleID       int       `json:"schedule_id" db:"schedule_id"`
	VolumeGroupID    int       `json:"volume_group_id" db:"volume_group_id"`
	SnapshotName     *string   `json:"snapshot_name" db:"snapshot_name"`
	ExecutionTime    time.Time `json:"execution_time" db:"execution_time"`
	Status           string    `json:"status" db:"status"` // success, failed, pending
	ErrorMessage     *string   `json:"error_message,omitempty" db:"error_message"`
	SnapshotID       *string   `json:"snapshot_id,omitempty" db:"snapshot_id"`
	RetentionDays    int       `json:"retention_days" db:"retention_days"`
	RetentionMinutes *int      `json:"retention_minutes,omitempty" db:"retention_minutes"`
}

// VolumeGroupWithSystem extends VolumeGroup with storage system info
type VolumeGroupWithSystem struct {
	VolumeGroup
	SystemName    string `json:"system_name" db:"system_name"`
	SystemIP      string `json:"system_ip" db:"system_ip"`
	SnapshotCount int    `json:"snapshot_count" db:"snapshot_count"`
}

// VolumeGroupWithCount extends VolumeGroup with schedule count
type VolumeGroupWithCount struct {
	VolumeGroup
	ScheduleCount int `json:"schedule_count" db:"schedule_count"`
}

// ScheduleWithVolumeGroup extends SnapshotSchedule with volume group info
type ScheduleWithVolumeGroup struct {
	SnapshotSchedule
	VGName     string `json:"vg_name" db:"vg_name"`
	SystemName string `json:"system_name" db:"system_name"`
}

// ExecutionWithDetails extends SnapshotExecution with related info
type ExecutionWithDetails struct {
	SnapshotExecution
	ScheduleName string `json:"schedule_name" db:"schedule_name"`
	VGName       string `json:"vg_name" db:"vg_name"`
	SystemName   string `json:"system_name" db:"system_name"`
}

// NTPServer represents an NTP server configuration
type NTPServer struct {
	ID            int        `json:"id" db:"id"`
	ServerAddress string     `json:"server_address" db:"server_address"`
	IsActive      bool       `json:"is_active" db:"is_active"`
	Priority      int        `json:"priority" db:"priority"`
	LastSyncAt    *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	SyncStatus    *string    `json:"sync_status,omitempty" db:"sync_status"`
	TimeOffsetMs  *int       `json:"time_offset_ms,omitempty" db:"time_offset_ms"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// SystemTimeInfo represents system time information
type SystemTimeInfo struct {
	CurrentTime    string  `json:"current_time"`
	Timezone       string  `json:"timezone"`
	TimezoneOffset string  `json:"timezone_offset"`
	NTPSyncEnabled bool    `json:"ntp_sync_enabled"`
	NTPSyncStatus  string  `json:"ntp_sync_status"`
	LastNTPSync    *string `json:"last_ntp_sync,omitempty"`
	SystemUptime   string  `json:"system_uptime"`
	TimeDriftMs    *int    `json:"time_drift_ms,omitempty"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           int       `json:"id" db:"id"`
	UserID       *int      `json:"user_id,omitempty" db:"user_id"`
	Username     string    `json:"username" db:"username"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   *string   `json:"resource_id,omitempty" db:"resource_id"`
	ResourceName *string   `json:"resource_name,omitempty" db:"resource_name"`
	Details      *string   `json:"details,omitempty" db:"details"`
	IPAddress    *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string   `json:"user_agent,omitempty" db:"user_agent"`
	Status       string    `json:"status" db:"status"` // success, failed
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

//
