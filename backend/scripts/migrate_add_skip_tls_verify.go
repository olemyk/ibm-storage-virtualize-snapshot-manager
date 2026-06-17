package main

import (
	"database/sql"
	"fmt"
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
	var columnExists bool
	row := db.QueryRow(`
		SELECT COUNT(*) 
		FROM pragma_table_info('storage_systems') 
		WHERE name='skip_tls_verify'
	`)
	if err := row.Scan(&columnExists); err != nil {
		log.Fatalf("Failed to check if column exists: %v", err)
	}

	if columnExists {
		fmt.Println("Column 'skip_tls_verify' already exists in storage_systems table")
		return
	}

	// Add skip_tls_verify column (default TRUE for backward compatibility)
	_, err = db.Exec(`
		ALTER TABLE storage_systems 
		ADD COLUMN skip_tls_verify BOOLEAN DEFAULT TRUE
	`)
	if err != nil {
		log.Fatalf("Failed to add skip_tls_verify column: %v", err)
	}

	fmt.Println("Successfully added 'skip_tls_verify' column to storage_systems table")
	fmt.Println("Default value is TRUE (skip verification) for backward compatibility")
}

//
