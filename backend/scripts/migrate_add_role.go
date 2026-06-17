package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./snapshot_manager.db"
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Check if role column exists
	var columnExists bool
	row := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='role'")
	err = row.Scan(&columnExists)
	if err != nil {
		log.Fatalf("Failed to check for role column: %v", err)
	}

	if columnExists {
		fmt.Println("✓ Role column already exists")
		return
	}

	fmt.Println("Adding role column to users table...")

	// Add role column with default value
	_, err = db.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'viewer'")
	if err != nil {
		log.Fatalf("Failed to add role column: %v", err)
	}

	fmt.Println("✓ Role column added successfully")

	// Update existing users to have admin role (assuming first user should be admin)
	result, err := db.Exec("UPDATE users SET role = 'admin' WHERE id = (SELECT MIN(id) FROM users)")
	if err != nil {
		log.Printf("Warning: Failed to set first user as admin: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Println("✓ Set first user as admin")
		}
	}

	// Show current users
	rows, err := db.Query("SELECT id, username, role FROM users")
	if err != nil {
		log.Printf("Warning: Failed to query users: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\nCurrent users:")
	fmt.Println("ID | Username | Role")
	fmt.Println("---|----------|------")
	for rows.Next() {
		var id int
		var username, role string
		if err := rows.Scan(&id, &username, &role); err != nil {
			continue
		}
		fmt.Printf("%d  | %s | %s\n", id, username, role)
	}

	fmt.Println("\n✓ Migration completed successfully!")
}

//
