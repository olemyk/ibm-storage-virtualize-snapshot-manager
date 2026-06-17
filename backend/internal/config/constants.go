package config

import "time"

// Server defaults
const (
	DefaultServerPort = 8080
	DefaultServerHost = "0.0.0.0"
)

// Database defaults
const (
	DefaultPostgresPort = 5432
	DefaultSQLitePath   = "./data/snapshots.db"
)

// Query timeouts
const (
	DefaultQueryTimeoutSeconds = 5
)

// Audit log settings
const (
	MinAuditEntries  = 100
	MaxAuditEntries  = 100000
	MinRetentionDays = 1
	MaxRetentionDays = 3650 // 10 years
)

// User defaults
const (
	DefaultUserRole = "viewer"
)

// Execution defaults
const (
	DefaultExecutionLimit = 50
)

// Snapshot defaults
const (
	DefaultSnapshotNamePattern = "{schedule_name}_{timestamp}"
)

// Token settings
const (
	TokenRefreshBufferMinutes = 5
	JWTExpirationHours        = 24
)

// HTTP client settings
const (
	DefaultHTTPTimeoutSeconds = 30
	MinHTTPTimeoutSeconds     = 5
	MaxHTTPTimeoutSeconds     = 300
)

// Port validation
const (
	MinPortNumber = 1
	MaxPortNumber = 65535
)

// Health check settings
const (
	MaxConcurrentHealthChecks = 5
)

// IBM SVC API Rate Limits
const (
	IBMSVCAuthRateLimit       = 3  // Max 3 auth requests per second per cluster
	IBMSVCCmdRateLimit        = 10 // Max 10 command requests per second per cluster
	IBMSVCMaxTokensPerCluster = 4  // Max 4 different tokens per cluster
)

// Retry configuration
const (
	MaxRetries     = 3
	BaseRetryDelay = 1 // seconds
)

// Retention validation (MaxRetentionDays already defined above at line 27)
const (
	MaxRetentionMinutes = 525600 // 1 year in minutes
)

// Timeout durations
const (
	QueryTimeout = 5 * time.Second
)

//
