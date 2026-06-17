package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("=================================")
	fmt.Println("Create Initial User")
	fmt.Println("=================================")
	fmt.Println()

	// Get database path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/snapshots.db"
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Get user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		os.Exit(1)
	}
	password := string(passwordBytes)
	fmt.Println()

	fmt.Print("Confirm Password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nError reading password: %v\n", err)
		os.Exit(1)
	}
	confirm := string(confirmBytes)
	fmt.Println()

	if password != confirm {
		fmt.Println("Error: Passwords do not match")
		os.Exit(1)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	// Insert user
	query := `INSERT INTO users (username, password_hash, email) VALUES (?, ?, ?)`
	_, err = db.Exec(query, username, string(hashedPassword), email)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✓ User created successfully!")
	fmt.Printf("  Username: %s\n", username)
	fmt.Printf("  Email: %s\n", email)
	fmt.Println()
}

//
