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
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if column already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('snapshot_schedules') WHERE name='snapshot_name_pattern'").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check column existence: %v", err)
	}

	if count > 0 {
		log.Println("Column 'snapshot_name_pattern' already exists, skipping migration")
		return
	}

	// Add snapshot_name_pattern column
	_, err = db.Exec("ALTER TABLE snapshot_schedules ADD COLUMN snapshot_name_pattern VARCHAR(255) DEFAULT '{schedule_name}_{timestamp}'")
	if err != nil {
		log.Fatalf("Failed to add snapshot_name_pattern column: %v", err)
	}

	log.Println("Successfully added snapshot_name_pattern column to snapshot_schedules table")
}

//
