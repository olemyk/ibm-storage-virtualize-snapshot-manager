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

	// Add connection status fields to storage_systems table
	migrations := []string{
		`ALTER TABLE storage_systems ADD COLUMN connection_status VARCHAR(50) DEFAULT 'unknown'`,
		`ALTER TABLE storage_systems ADD COLUMN last_connection_check TIMESTAMP`,
		`ALTER TABLE storage_systems ADD COLUMN connection_error TEXT`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			// Ignore "duplicate column" errors
			if err.Error() != "duplicate column name: connection_status" &&
				err.Error() != "duplicate column name: last_connection_check" &&
				err.Error() != "duplicate column name: connection_error" {
				log.Printf("Migration warning: %v", err)
			}
		}
	}

	log.Println("Migration completed: Added connection status fields to storage_systems table")
}

//
