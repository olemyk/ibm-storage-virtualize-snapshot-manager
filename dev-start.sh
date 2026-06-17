#!/bin/bash

# IBM Storage Virtualize Snapshot Manager - Development Startup Script
# This script starts both backend and frontend services for development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}IBM Storage Virtualize Snapshot Manager${NC}"
echo -e "${BLUE}Development Environment Startup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if backend directory exists
if [ ! -d "backend" ]; then
    echo -e "${RED}Error: backend directory not found${NC}"
    exit 1
fi

# Check if frontend directory exists
if [ ! -d "frontend" ]; then
    echo -e "${RED}Error: frontend directory not found${NC}"
    exit 1
fi

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Shutting down services...${NC}"
    pkill -f "snapshot-manager" 2>/dev/null || true
    pkill -f "vite.*frontend" 2>/dev/null || true
    echo -e "${GREEN}Services stopped${NC}"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Check if database exists and is initialized
if [ ! -f "backend/data/snapshots.db" ] || [ ! -s "backend/data/snapshots.db" ]; then
    echo -e "${YELLOW}Database not found or empty. Initializing...${NC}"
    
    # Create database from schema
    mkdir -p backend/data
    cd backend
    sqlite3 data/snapshots.db < internal/db/schema.sql
    
    # Create default admin user
    echo -e "${BLUE}Creating default admin user...${NC}"
    HASH=$(go run scripts/genhash.go admin123 2>/dev/null)
    sqlite3 data/snapshots.db "INSERT INTO users (username, password_hash, email, role) VALUES ('admin', '$HASH', 'admin@example.com', 'admin');"
    
    cd ..
    echo -e "${GREEN}✓ Database initialized${NC}"
    echo -e "${GREEN}  Username: admin${NC}"
    echo -e "${GREEN}  Password: admin123${NC}"
    echo ""
else
    # Database exists, run migrations to ensure it's up to date
    echo -e "${BLUE}Checking database migrations...${NC}"
    cd backend
    
    # Run all migrations (they check if already applied)
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_skip_tls_verify.go 2>/dev/null || true
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_connection_status.go 2>/dev/null || true
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_partition_fields.go 2>/dev/null || true
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_audit_logs.go 2>/dev/null || true
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_settings.go 2>/dev/null || true
    DB_PATH=./data/snapshots.db go run scripts/migrate_add_notifications.go 2>/dev/null || true
    
    cd ..
    echo -e "${GREEN}✓ Database migrations complete${NC}"
fi

# Check backend environment
if [ ! -f "backend/.env" ]; then
    echo -e "${RED}Error: backend/.env not found${NC}"
    echo -e "${YELLOW}Please copy backend/.env.example to backend/.env and configure it${NC}"
    exit 1
fi

# Check frontend environment
if [ ! -f "frontend/.env" ]; then
    echo -e "${YELLOW}Warning: frontend/.env not found, creating from example...${NC}"
    cp frontend/.env.example frontend/.env
fi

# Verify frontend is pointing to correct backend port
BACKEND_PORT=$(grep "^PORT=" backend/.env | cut -d'=' -f2 | tr -d ' ')
BACKEND_PORT=${BACKEND_PORT:-8080}
FRONTEND_API_URL=$(grep "^VITE_API_URL=" frontend/.env | cut -d'=' -f2)

if [[ ! "$FRONTEND_API_URL" =~ ":$BACKEND_PORT" ]]; then
    echo -e "${YELLOW}Updating frontend API URL to match backend port $BACKEND_PORT...${NC}"
    sed -i.bak "s|VITE_API_URL=.*|VITE_API_URL=http://localhost:$BACKEND_PORT/api|" frontend/.env
    rm -f frontend/.env.bak
fi

# Stop any existing services
echo -e "${YELLOW}Stopping any existing services...${NC}"
pkill -f "snapshot-manager" 2>/dev/null || true
pkill -f "vite.*frontend" 2>/dev/null || true
sleep 1

# Start backend
echo -e "${BLUE}Starting backend server...${NC}"
cd backend
bash start-server.sh &
BACKEND_PID=$!
cd ..

# Wait for backend to be ready
echo -e "${YELLOW}Waiting for backend to start...${NC}"
for i in {1..30}; do
    if curl -s http://localhost:$BACKEND_PORT/api/systems >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Backend started on http://localhost:$BACKEND_PORT${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}Error: Backend failed to start${NC}"
        cleanup
        exit 1
    fi
    sleep 1
done

# Start frontend
echo -e "${BLUE}Starting frontend server...${NC}"
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..

# Wait for frontend to be ready
echo -e "${YELLOW}Waiting for frontend to start...${NC}"
sleep 3

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ Development environment ready!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Services:${NC}"
echo -e "  Backend:  ${GREEN}http://localhost:$BACKEND_PORT${NC}"
echo -e "  Frontend: ${GREEN}http://localhost:5173${NC}"
echo ""
echo -e "${BLUE}Login Credentials:${NC}"
echo -e "  Username: ${GREEN}admin${NC}"
echo -e "  Password: ${GREEN}admin123${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop all services${NC}"
echo ""

# Keep script running
wait

# 
