package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./snapshot_manager.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create notification_channels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			config TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating notification_channels table: %v", err)
	}

	// Create alert_rules table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS alert_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			event_type VARCHAR(50) NOT NULL,
			conditions TEXT,
			severity VARCHAR(20) DEFAULT 'info',
			notification_channel_ids TEXT NOT NULL,
			throttle_minutes INTEGER DEFAULT 0,
			last_triggered_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating alert_rules table: %v", err)
	}

	// Create notification_history table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notification_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alert_rule_id INTEGER,
			notification_channel_id INTEGER NOT NULL,
			event_type VARCHAR(50) NOT NULL,
			event_data TEXT,
			status VARCHAR(20) NOT NULL,
			error_message TEXT,
			sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (alert_rule_id) REFERENCES alert_rules(id) ON DELETE SET NULL,
			FOREIGN KEY (notification_channel_id) REFERENCES notification_channels(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		log.Fatalf("Error creating notification_history table: %v", err)
	}

	// Create indexes
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notification_history_sent_at ON notification_history(sent_at);
		CREATE INDEX IF NOT EXISTS idx_notification_history_status ON notification_history(status);
		CREATE INDEX IF NOT EXISTS idx_notification_history_event_type ON notification_history(event_type);
		CREATE INDEX IF NOT EXISTS idx_alert_rules_event_type ON alert_rules(event_type);
		CREATE INDEX IF NOT EXISTS idx_alert_rules_active ON alert_rules(is_active);
	`)
	if err != nil {
		log.Fatalf("Error creating indexes: %v", err)
	}

	log.Println("Notification tables created successfully")
}

//
