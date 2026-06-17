-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Storage systems table
CREATE TABLE IF NOT EXISTS storage_systems (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    ip_address VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 7443,
    username VARCHAR(255) NOT NULL,
    password_encrypted TEXT NOT NULL,
    auth_token TEXT,
    token_expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_address, port)
);

-- Volume groups table
CREATE TABLE IF NOT EXISTS volume_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_system_id INTEGER NOT NULL,
    vg_id VARCHAR(255) NOT NULL,
    vg_name VARCHAR(255) NOT NULL,
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (storage_system_id) REFERENCES storage_systems(id) ON DELETE CASCADE,
    UNIQUE(storage_system_id, vg_id)
);

-- Snapshot schedules table
CREATE TABLE IF NOT EXISTS snapshot_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    volume_group_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    cron_expression VARCHAR(255) NOT NULL,
    retention_days INTEGER NOT NULL,
    retention_minutes INTEGER,
    safeguarded BOOLEAN DEFAULT FALSE,
    pool_name VARCHAR(255),
    snapshot_name_pattern VARCHAR(255) DEFAULT '{schedule_name}_{timestamp}',
    is_active BOOLEAN DEFAULT TRUE,
    last_executed_at TIMESTAMP,
    next_execution_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (volume_group_id) REFERENCES volume_groups(id) ON DELETE CASCADE
);

-- Snapshot executions table
CREATE TABLE IF NOT EXISTS snapshot_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL,
    volume_group_id INTEGER NOT NULL,
    snapshot_name VARCHAR(255),
    execution_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    snapshot_id VARCHAR(255),
    retention_days INTEGER,
    retention_minutes INTEGER,
    FOREIGN KEY (schedule_id) REFERENCES snapshot_schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (volume_group_id) REFERENCES volume_groups(id) ON DELETE CASCADE
);

-- Settings table
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- NTP servers table
CREATE TABLE IF NOT EXISTS ntp_servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_address VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    last_sync_at TIMESTAMP,
    sync_status VARCHAR(50),
    time_offset_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_volume_groups_system_id ON volume_groups(storage_system_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_schedules_vg_id ON snapshot_schedules(volume_group_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_schedules_active ON snapshot_schedules(is_active);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_schedule_id ON snapshot_executions(schedule_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_vg_id ON snapshot_executions(volume_group_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_status ON snapshot_executions(status);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_time ON snapshot_executions(execution_time);

-- 
