-- Create ntp_servers table
CREATE TABLE IF NOT EXISTS ntp_servers (
    id SERIAL PRIMARY KEY,
    server_address VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    last_sync_at TIMESTAMP,
    sync_status VARCHAR(50),
    time_offset_ms INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_ntp_servers_active ON ntp_servers(is_active);
CREATE INDEX IF NOT EXISTS idx_ntp_servers_priority ON ntp_servers(priority);

-- Create trigger for updated_at
DROP TRIGGER IF EXISTS update_ntp_servers_updated_at ON ntp_servers;
CREATE TRIGGER update_ntp_servers_updated_at BEFORE UPDATE ON ntp_servers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
