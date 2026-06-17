package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection details from environment
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres"
	}

	var connStr string
	if dbType == "postgres" {
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "snapshots"
		}
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "snapshots"
		}
		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			log.Fatal("DB_PASSWORD environment variable is required")
		}

		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
	} else {
		log.Fatal("This migration is only for PostgreSQL databases")
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Check if ntp_servers table already exists
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'ntp_servers'
		)
	`).Scan(&exists)
	if err != nil {
		log.Fatalf("Failed to check if ntp_servers table exists: %v", err)
	}

	if exists {
		log.Println("ntp_servers table already exists, skipping migration")
		return
	}

	log.Println("Creating ntp_servers table...")

	// Create ntp_servers table
	_, err = db.Exec(`
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
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create ntp_servers table: %v", err)
	}

	log.Println("ntp_servers table created successfully")

	// Create indexes
	log.Println("Creating indexes...")

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ntp_servers_active ON ntp_servers(is_active)
	`)
	if err != nil {
		log.Fatalf("Failed to create idx_ntp_servers_active index: %v", err)
	}

	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_ntp_servers_priority ON ntp_servers(priority)
	`)
	if err != nil {
		log.Fatalf("Failed to create idx_ntp_servers_priority index: %v", err)
	}

	log.Println("Indexes created successfully")

	// Create trigger for updated_at
	log.Println("Creating trigger for updated_at...")

	_, err = db.Exec(`
		DROP TRIGGER IF EXISTS update_ntp_servers_updated_at ON ntp_servers
	`)
	if err != nil {
		log.Fatalf("Failed to drop existing trigger: %v", err)
	}

	_, err = db.Exec(`
		CREATE TRIGGER update_ntp_servers_updated_at BEFORE UPDATE ON ntp_servers
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column()
	`)
	if err != nil {
		log.Fatalf("Failed to create trigger: %v", err)
	}

	log.Println("Trigger created successfully")
	log.Println("Migration completed successfully!")
}
