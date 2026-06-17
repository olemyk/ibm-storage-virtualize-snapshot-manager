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

	// Create settings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating settings table: %v", err)
	}

	// Insert default audit log retention settings
	_, err = db.Exec(`
		INSERT OR IGNORE INTO settings (key, value) VALUES 
		('audit_log_max_entries', '1000'),
		('audit_log_retention_days', '365')
	`)
	if err != nil {
		log.Fatalf("Error inserting default settings: %v", err)
	}

	log.Println("Settings table created and default values inserted successfully")
}

//
