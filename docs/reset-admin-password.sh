#!/bin/bash
#
# IBM Storage Virtualize Snapshot Manager - Admin Password Reset Script
#
# This script resets the admin user password in the database.
# Default password will be set to: admin123
#
# Usage:
#   ./reset-admin-password.sh [new_password]
#
# If no password is provided, defaults to "admin123"
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

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Get password from argument or use default
NEW_PASSWORD="${1:-admin123}"

print_info "Resetting admin password..."

# Check if database container is running
if ! podman ps | grep -q snapshot-manager-db; then
    print_error "Database container is not running"
    print_info "Start the stack first: ./start-prod.sh"
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

print_success "Password hash generated"

# Update password in database
print_info "Updating password in database..."

podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
    "UPDATE users SET password_hash='${HASH}' WHERE username='admin';" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    print_success "Admin password reset successfully"
    print_info ""
    print_info "New credentials:"
    print_info "  Username: admin"
    print_info "  Password: ${NEW_PASSWORD}"
    print_info ""
    print_info "You can now login at: https://localhost"
else
    print_error "Failed to update password"
    print_info "Checking if admin user exists..."
    
    USER_EXISTS=$(podman exec snapshot-manager-db psql -U snapshots -d snapshots -t -c \
        "SELECT COUNT(*) FROM users WHERE username='admin';" 2>/dev/null | tr -d ' ')
    
    if [ "$USER_EXISTS" = "0" ]; then
        print_info "Admin user does not exist. Creating..."
        
        podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
            "INSERT INTO users (username, password_hash, email, role, created_at, updated_at) 
             VALUES ('admin', '${HASH}', 'admin@localhost', 'admin', NOW(), NOW());" > /dev/null 2>&1
        
        if [ $? -eq 0 ]; then
            print_success "Admin user created successfully"
            print_info ""
            print_info "New credentials:"
            print_info "  Username: admin"
            print_info "  Password: ${NEW_PASSWORD}"
        else
            print_error "Failed to create admin user"
            exit 1
        fi
    else
        print_error "Unknown error occurred"
        exit 1
    fi
fi
