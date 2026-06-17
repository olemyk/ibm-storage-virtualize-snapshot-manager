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

	// Check if columns already exist
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM pragma_table_info('volume_groups') 
		WHERE name IN ('partition_id', 'partition_name')
	`).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check existing columns: %v", err)
	}

	if count > 0 {
		fmt.Println("Partition columns already exist, skipping migration")
		return
	}

	// Add partition_id and partition_name columns
	migrations := []string{
		`ALTER TABLE volume_groups ADD COLUMN partition_id VARCHAR(50)`,
		`ALTER TABLE volume_groups ADD COLUMN partition_name VARCHAR(255)`,
		`CREATE INDEX IF NOT EXISTS idx_volume_groups_partition ON volume_groups(partition_id, partition_name)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			log.Fatalf("Failed to execute migration: %v\nSQL: %s", err, migration)
		}
	}

	fmt.Println("Successfully added partition_id and partition_name columns to volume_groups table")
}

//
