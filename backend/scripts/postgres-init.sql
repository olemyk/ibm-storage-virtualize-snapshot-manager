-- PostgreSQL schema for IBM Storage Virtualize Snapshot Manager
-- Converted from SQLite schema with PostgreSQL-specific types

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(50) DEFAULT 'viewer',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Storage systems table
CREATE TABLE IF NOT EXISTS storage_systems (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    ip_address VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 7443,
    username VARCHAR(255) NOT NULL,
    password_encrypted TEXT NOT NULL,
    auth_token TEXT,
    token_expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    skip_tls_verify BOOLEAN DEFAULT FALSE,
    connection_status VARCHAR(50) DEFAULT 'unknown',
    last_connection_check TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_address, port)
);

-- Volume groups table
CREATE TABLE IF NOT EXISTS volume_groups (
    id SERIAL PRIMARY KEY,
    storage_system_id INTEGER NOT NULL,
    vg_id VARCHAR(255) NOT NULL,
    vg_name VARCHAR(255) NOT NULL,
    partition_id VARCHAR(255),
    partition_name VARCHAR(255),
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (storage_system_id) REFERENCES storage_systems(id) ON DELETE CASCADE,
    UNIQUE(storage_system_id, vg_id)
);

-- Snapshot schedules table
CREATE TABLE IF NOT EXISTS snapshot_schedules (
    id SERIAL PRIMARY KEY,
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
    id SERIAL PRIMARY KEY,
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

-- Notification channels table
CREATE TABLE IF NOT EXISTS notification_channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alert rules table
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    event_type VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    conditions JSONB,
    notification_channel_ids JSONB,
    throttle_minutes INTEGER DEFAULT 0,
    last_triggered_at TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Alert rule channels (many-to-many relationship)
CREATE TABLE IF NOT EXISTS alert_rule_channels (
    alert_rule_id INTEGER NOT NULL,
    notification_channel_id INTEGER NOT NULL,
    PRIMARY KEY (alert_rule_id, notification_channel_id),
    FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
);

-- Notification history table
CREATE TABLE IF NOT EXISTS notification_history (
    id SERIAL PRIMARY KEY,
    notification_channel_id INTEGER NOT NULL,
    alert_rule_id INTEGER,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE,
    FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE SET NULL
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    username VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    resource_name VARCHAR(255),
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(50),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Settings table
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_volume_groups_system_id ON volume_groups(storage_system_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_schedules_vg_id ON snapshot_schedules(volume_group_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_schedules_active ON snapshot_schedules(is_active);
CREATE INDEX IF NOT EXISTS idx_snapshot_schedules_next_exec ON snapshot_schedules(next_execution_at) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_schedule_id ON snapshot_executions(schedule_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_vg_id ON snapshot_executions(volume_group_id);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_status ON snapshot_executions(status);
CREATE INDEX IF NOT EXISTS idx_snapshot_executions_time ON snapshot_executions(execution_time DESC);
CREATE INDEX IF NOT EXISTS idx_notification_history_channel_id ON notification_history(notification_channel_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_sent_at ON notification_history(sent_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_storage_systems_active ON storage_systems(is_active);

-- Insert default admin user (password: admin123)
-- Password hash generated with bcrypt cost 10
INSERT INTO users (username, password_hash, email, role)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin@example.com', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at columns
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_storage_systems_updated_at ON storage_systems;
CREATE TRIGGER update_storage_systems_updated_at BEFORE UPDATE ON storage_systems
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_snapshot_schedules_updated_at ON snapshot_schedules;
CREATE TRIGGER update_snapshot_schedules_updated_at BEFORE UPDATE ON snapshot_schedules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_notification_channels_updated_at ON notification_channels;
CREATE TRIGGER update_notification_channels_updated_at BEFORE UPDATE ON notification_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_alert_rules_updated_at ON alert_rules;
CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_settings_updated_at ON settings;
CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();