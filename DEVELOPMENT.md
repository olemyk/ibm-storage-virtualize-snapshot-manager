# Development Guide

This guide explains how to set up and run the IBM Storage Virtualize Snapshot Manager for local development.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Manual Setup](#manual-setup)
- [Development Workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)

## Prerequisites

### Required Software

- **Go** 1.21 or higher
- **Node.js** 18.x or higher
- **npm** 9.x or higher
- **SQLite3** (usually pre-installed on macOS/Linux)
- **Git**

### Verify Installation

```bash
go version        # Should show 1.21+
node --version    # Should show v18+
npm --version     # Should show 9+
sqlite3 --version # Should show 3.x+
```

## Quick Start

The easiest way to start the development environment is using the provided script:

```bash
./dev-start.sh
```

This script will:
1. ✅ Check for required directories and files
2. ✅ Initialize the database if needed
3. ✅ Create a default admin user (username: `admin`, password: `admin123`)
4. ✅ Start the backend server on port 8080
5. ✅ Start the frontend dev server on port 5173
6. ✅ Verify both services are running

### Accessing the Application

Once started, open your browser to:
- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:8080

**Default Login:**
- Username: `admin`
- Password: `admin123`

### Stopping Services

Press `Ctrl+C` in the terminal where `dev-start.sh` is running. This will gracefully shut down both services.

## Manual Setup

If you prefer to set up and run services manually:

### 1. Backend Setup

```bash
cd backend

# Copy environment configuration
cp .env.example .env

# Edit .env and configure:
# - JWT_SECRET (generate with: openssl rand -base64 32)
# - ENCRYPTION_KEY (generate with: openssl rand -base64 32)
# - DB_PATH (default: ./data/snapshots.db)
# - PORT (default: 8080)

# Generate encryption keys (if not already done)
go run scripts/generate_keys.go

# Initialize database
mkdir -p data
sqlite3 data/snapshots.db < internal/db/schema.sql

# Create admin user
HASH=$(go run scripts/genhash.go admin123)
sqlite3 data/snapshots.db "INSERT INTO users (username, password_hash, email, role) VALUES ('admin', '$HASH', 'admin@example.com', 'admin');"

# Build and start backend
go build -o snapshot-manager cmd/server/main.go
./snapshot-manager
```

### 2. Frontend Setup

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Copy environment configuration
cp .env.example .env

# Edit .env and set:
# VITE_API_URL=http://localhost:8080/api

# Start development server
npm run dev
```

The frontend will be available at http://localhost:5173

## Development Workflow

### Backend Development

#### Running Tests

```bash
cd backend
go test ./...                    # Run all tests
go test ./internal/api/...       # Run specific package tests
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
```

#### Code Formatting

```bash
cd backend
go fmt ./...                     # Format all Go files
go vet ./...                     # Run Go vet
```

#### Building

```bash
cd backend
go build -o snapshot-manager cmd/server/main.go
```

#### Database Migrations

When schema changes are needed:

```bash
cd backend

# Create migration script in scripts/
# Example: scripts/migrate_add_new_field.go

# Run migration
DB_PATH=./data/snapshots.db go run scripts/migrate_add_new_field.go
```

**Available Migrations:**

The following migrations are available to update existing databases:

```bash
cd backend

# Add skip_tls_verify column to storage_systems
DB_PATH=./data/snapshots.db go run scripts/migrate_add_skip_tls_verify.go

# Add connection status fields to storage_systems
DB_PATH=./data/snapshots.db go run scripts/migrate_add_connection_status.go

# Add partition fields to volume_groups
DB_PATH=./data/snapshots.db go run scripts/migrate_add_partition_fields.go

# Add audit_logs table
DB_PATH=./data/snapshots.db go run scripts/migrate_add_audit_logs.go

# Add settings table
DB_PATH=./data/snapshots.db go run scripts/migrate_add_settings.go

# Add notification tables
DB_PATH=./data/snapshots.db go run scripts/migrate_add_notifications.go
```

**Note:** The `dev-start.sh` script automatically runs all migrations on startup, so manual migration is typically only needed when developing new migrations or troubleshooting.

### Frontend Development

#### Running Tests

```bash
cd frontend
npm test                         # Run tests
npm run test:watch              # Watch mode
```

#### Code Formatting

```bash
cd frontend
npm run lint                     # Run ESLint
npm run format                   # Format with Prettier (if configured)
```

#### Building for Production

```bash
cd frontend
npm run build                    # Creates dist/ directory
npm run preview                  # Preview production build
```

### Hot Reload

Both backend and frontend support hot reload during development:

- **Backend:** Restart the server after code changes
- **Frontend:** Vite automatically reloads on file changes

## Troubleshooting

### Port Already in Use

If you see "port already in use" errors:

```bash
# Find and kill processes using the ports
lsof -ti:8080 | xargs kill -9    # Backend port
lsof -ti:5173 | xargs kill -9    # Frontend port

# Or use the dev script which handles this automatically
./dev-start.sh
```

### Database Issues

#### Missing Columns or Tables (500 Errors)

If you see 500 errors like "no such column" or "no such table", run the migrations:

```bash
cd backend

# Run all migrations
DB_PATH=./data/snapshots.db go run scripts/migrate_add_skip_tls_verify.go
DB_PATH=./data/snapshots.db go run scripts/migrate_add_connection_status.go
DB_PATH=./data/snapshots.db go run scripts/migrate_add_partition_fields.go
DB_PATH=./data/snapshots.db go run scripts/migrate_add_audit_logs.go
DB_PATH=./data/snapshots.db go run scripts/migrate_add_settings.go
DB_PATH=./data/snapshots.db go run scripts/migrate_add_notifications.go

# Restart the server
pkill -f snapshot-manager
bash start-server.sh
```

Or simply restart using `dev-start.sh` which runs migrations automatically.

#### Reset Database

```bash
cd backend
rm -f data/snapshots.db*
sqlite3 data/snapshots.db < internal/db/schema.sql

# Recreate admin user
HASH=$(go run scripts/genhash.go admin123)
sqlite3 data/snapshots.db "INSERT INTO users (username, password_hash, email, role) VALUES ('admin', '$HASH', 'admin@example.com', 'admin');"
```

#### View Database Contents

```bash
cd backend
sqlite3 data/snapshots.db

# Inside sqlite3:
.tables                          # List all tables
.schema users                    # Show table schema
SELECT * FROM users;             # Query data
.quit                            # Exit
```

### CORS Errors

If you see CORS errors in the browser console:

1. Verify frontend `.env` has correct backend URL:
   ```
   VITE_API_URL=http://localhost:8080/api
   ```

2. Restart frontend dev server:
   ```bash
   cd frontend
   npm run dev
   ```

### Login Issues

If login fails:

1. Verify user exists:
   ```bash
   cd backend
   sqlite3 data/snapshots.db "SELECT username, email, role FROM users;"
   ```

2. Recreate admin user:
   ```bash
   cd backend
   sqlite3 data/snapshots.db "DELETE FROM users WHERE username='admin';"
   HASH=$(go run scripts/genhash.go admin123)
   sqlite3 data/snapshots.db "INSERT INTO users (username, password_hash, email, role) VALUES ('admin', '$HASH', 'admin@example.com', 'admin');"
   ```

3. Check backend logs for authentication errors

### Missing Dependencies

#### Backend

```bash
cd backend
go mod download                  # Download dependencies
go mod tidy                      # Clean up go.mod
```

#### Frontend

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install                      # Reinstall dependencies
```

## Project Structure

```
.
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go          # Application entry point
│   ├── internal/
│   │   ├── api/                 # HTTP handlers and routes
│   │   ├── auth/                # Authentication logic
│   │   ├── config/              # Configuration management
│   │   ├── db/                  # Database layer
│   │   ├── models/              # Data models
│   │   ├── scheduler/           # Snapshot scheduler
│   │   └── svc/                 # IBM SVC client
│   ├── scripts/                 # Utility scripts
│   ├── .env                     # Environment configuration
│   └── go.mod                   # Go dependencies
│
├── frontend/
│   ├── src/
│   │   ├── api/                 # API client
│   │   ├── components/          # React components
│   │   ├── pages/               # Page components
│   │   └── types/               # TypeScript types
│   ├── .env                     # Environment configuration
│   └── package.json             # npm dependencies
│
├── dev-start.sh                 # Development startup script
├── DEVELOPMENT.md               # This file
└── README.md                    # Project overview
```

## Environment Variables

### Backend (.env)

```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0

# Database
DB_PATH=./data/snapshots.db

# Security (generate with: openssl rand -base64 32)
JWT_SECRET=your-secret-key-here
ENCRYPTION_KEY=your-encryption-key-here

# Logging
LOG_LEVEL=info
```

### Frontend (.env)

```bash
# API Configuration
VITE_API_URL=http://localhost:8080/api
```

## Additional Resources

- [API Documentation](docs/API_NOTIFICATIONS.md)
- [Testing Guide](backend/TESTING.md)
- [Quick Start Guide](QUICKSTART.md)
- [Main README](README.md)

## Getting Help

If you encounter issues not covered in this guide:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review backend logs: `tail -f backend/logs/server.log`
3. Check browser console for frontend errors
4. Verify all prerequisites are installed correctly

## Contributing

When contributing code:

1. Create a feature branch from `main`
2. Make your changes
3. Run tests: `go test ./...` and `npm test`
4. Format code: `go fmt ./...` and `npm run lint`
5. Commit with clear messages
6. Create a pull request

---

**Note:** This is an unofficial, community-driven project. Not affiliated with, endorsed by, or supported by IBM Corporation.