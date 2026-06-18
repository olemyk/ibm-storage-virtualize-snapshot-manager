#!/bin/bash
#
# IBM Storage Virtualize Snapshot Manager - Login Issues Fix Script
#
# This script fixes common login issues:
# 1. CORS configuration (adds server IP to allowed origins)
# 2. Admin password reset
#
# Usage:
#   ./fix-login-issues.sh [server_ip] [new_password]
#
# Examples:
#   ./fix-login-issues.sh 10.33.3.104
#   ./fix-login-issues.sh 10.33.3.104 MyNewPassword123
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Get server IP from argument or detect
SERVER_IP="${1}"
NEW_PASSWORD="${2:-admin123}"

if [ -z "$SERVER_IP" ]; then
    # Try to detect server IP
    print_info "Detecting server IP address..."
    SERVER_IP=$(hostname -I | awk '{print $1}')
    
    if [ -z "$SERVER_IP" ]; then
        print_error "Could not detect server IP"
        print_info "Usage: $0 <server_ip> [password]"
        print_info "Example: $0 10.33.3.104"
        exit 1
    fi
    
    print_info "Detected IP: $SERVER_IP"
fi

print_info "=== Fixing Login Issues ==="
print_info "Server IP: $SERVER_IP"
print_info "New Password: $NEW_PASSWORD"
echo ""

# Step 1: Check if .env file exists
if [ ! -f .env ]; then
    print_error ".env file not found"
    print_info "Please run this script from the deployment directory"
    exit 1
fi

# Step 2: Update ALLOWED_ORIGINS in .env
print_info "Step 1: Updating CORS configuration..."

# Check if ALLOWED_ORIGINS exists in .env
if grep -q "^ALLOWED_ORIGINS=" .env; then
    # Update existing line
    CURRENT_ORIGINS=$(grep "^ALLOWED_ORIGINS=" .env | cut -d'=' -f2-)
    
    # Check if IP is already in the list
    if echo "$CURRENT_ORIGINS" | grep -q "$SERVER_IP"; then
        print_success "Server IP already in ALLOWED_ORIGINS"
    else
        # Add server IP to existing origins
        NEW_ORIGINS="${CURRENT_ORIGINS},http://${SERVER_IP},https://${SERVER_IP}"
        sed -i.bak "s|^ALLOWED_ORIGINS=.*|ALLOWED_ORIGINS=${NEW_ORIGINS}|" .env
        print_success "Added server IP to ALLOWED_ORIGINS"
    fi
else
    # Add new ALLOWED_ORIGINS line
    echo "ALLOWED_ORIGINS=http://localhost,https://localhost,http://${SERVER_IP},https://${SERVER_IP},http://127.0.0.1,https://127.0.0.1" >> .env
    print_success "Created ALLOWED_ORIGINS configuration"
fi

# Step 3: Restart backend container to apply CORS changes
print_info "Step 2: Restarting backend container..."

if podman ps | grep -q snapshot-manager-backend; then
    podman restart snapshot-manager-backend > /dev/null 2>&1
    print_success "Backend container restarted"
    
    # Wait for backend to be healthy
    print_info "Waiting for backend to be ready..."
    sleep 5
    
    for i in {1..30}; do
        if podman exec snapshot-manager-backend wget -q --spider http://localhost:8080/health 2>/dev/null; then
            print_success "Backend is healthy"
            break
        fi
        
        if [ $i -eq 30 ]; then
            print_error "Backend did not become healthy in time"
            print_info "Check logs: podman logs snapshot-manager-backend"
            exit 1
        fi
        
        sleep 2
    done
else
    print_warning "Backend container not running"
    print_info "Starting the stack..."
    ./start-prod.sh
fi

# Step 4: Reset admin password
print_info "Step 3: Resetting admin password..."

# Check if database container is running
if ! podman ps | grep -q snapshot-manager-db; then
    print_error "Database container is not running"
    exit 1
fi

# Generate bcrypt hash using backend container (has Go bcrypt built-in)
print_info "Generating password hash using backend container..."

HASH=$(podman exec snapshot-manager-backend /bin/sh -c "cat > /tmp/genhash.go << 'EOFGO'
package main
import (
    \"fmt\"
    \"os\"
    \"golang.org/x/crypto/bcrypt\"
)
func main() {
    if len(os.Args) < 2 {
        os.Exit(1)
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    if err != nil {
        os.Exit(1)
    }
    fmt.Println(string(hash))
}
EOFGO
cd /tmp && go run genhash.go '${NEW_PASSWORD}' 2>/dev/null && rm -f genhash.go" 2>/dev/null)

if [ -z "$HASH" ]; then
    print_error "Failed to generate password hash using backend container"
    print_info "Please use the manual method:"
    print_info "1. Generate bcrypt hash online (cost 10): https://bcrypt-generator.com/"
    print_info "2. Run: podman exec -it snapshot-manager-db psql -U snapshots -d snapshots"
    print_info "3. Execute: UPDATE users SET password_hash='<hash>' WHERE username='admin';"
    exit 1
fi
    
    # Update or create admin user
    USER_EXISTS=$(podman exec snapshot-manager-db psql -U snapshots -d snapshots -t -c \
        "SELECT COUNT(*) FROM users WHERE username='admin';" 2>/dev/null | tr -d ' ')
    
    if [ "$USER_EXISTS" = "0" ]; then
        print_info "Creating admin user..."
        podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
            "INSERT INTO users (username, password_hash, email, role, created_at, updated_at) 
             VALUES ('admin', '${HASH}', 'admin@localhost', 'admin', NOW(), NOW());" > /dev/null 2>&1
        print_success "Admin user created"
    else
        print_info "Updating admin password..."
        podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
            "UPDATE users SET password_hash='${HASH}' WHERE username='admin';" > /dev/null 2>&1
        print_success "Admin password updated"
    fi
else
    print_error "Python3 not found. Cannot generate password hash."
    print_info "Please install Python3 or use online bcrypt generator (cost 10)"
    exit 1
fi

# Step 5: Verify configuration
print_info "Step 4: Verifying configuration..."

# Check ALLOWED_ORIGINS in backend
BACKEND_ORIGINS=$(podman exec snapshot-manager-backend env | grep ALLOWED_ORIGINS | cut -d'=' -f2-)
if echo "$BACKEND_ORIGINS" | grep -q "$SERVER_IP"; then
    print_success "CORS configuration verified"
else
    print_warning "Server IP not found in backend ALLOWED_ORIGINS"
    print_info "Current origins: $BACKEND_ORIGINS"
fi

# Final summary
echo ""
print_success "=== Login Issues Fixed ==="
echo ""
print_info "Access the application at:"
print_info "  - https://${SERVER_IP}"
print_info "  - http://${SERVER_IP}"
echo ""
print_info "Login credentials:"
print_info "  Username: admin"
print_info "  Password: ${NEW_PASSWORD}"
echo ""
print_info "If you still have issues:"
print_info "  1. Check backend logs: podman logs snapshot-manager-backend"
print_info "  2. Check frontend logs: podman logs snapshot-manager-frontend"
print_info "  3. Verify .env file: cat .env | grep ALLOWED_ORIGINS"
echo ""
