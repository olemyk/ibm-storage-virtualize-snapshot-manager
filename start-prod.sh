#!/bin/bash
#
# IBM Storage Virtualize Snapshot Manager - Production Startup Script
# 
# This script pulls the latest pre-built images from GitHub Container Registry
# and starts the entire stack using podman-compose.
#
# Prerequisites:
#   - podman and podman-compose installed
#   - .env file configured with DB_PASSWORD, JWT_SECRET, ENCRYPTION_KEY
#
# Usage:
#   ./start-prod.sh              # Start with latest images
#   ./start-prod.sh --rebuild    # Force rebuild and restart
#   ./start-prod.sh --stop       # Stop all services
#   ./start-prod.sh --logs       # View logs
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="ibm-virtualize-snapshot-manager-dir"
COMPOSE_FILE="podman-compose.prod.yml"
BACKEND_IMAGE="ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/backend:latest"
FRONTEND_IMAGE="ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/frontend:latest"
POSTGRES_IMAGE="docker.io/library/postgres:16-alpine"

# Detect script location and set absolute paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Determine installation directory (absolute path)
# If compose file exists in script directory, use it (development scenario)
# Otherwise, use installation subdirectory (production scenario)
if [ -f "$SCRIPT_DIR/$COMPOSE_FILE" ]; then
    INSTALL_DIR_ABS="$SCRIPT_DIR"
else
    INSTALL_DIR_ABS="$SCRIPT_DIR/$INSTALL_DIR"
fi

# Helper functions
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

check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check for podman
    if ! command -v podman &> /dev/null; then
        print_error "podman is not installed. Please install podman first."
        exit 1
    fi
    
    # Check for podman-compose
    if ! command -v podman-compose &> /dev/null; then
        print_error "podman-compose is not installed. Please install podman-compose first."
        exit 1
    fi
    
    # Create installation directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR_ABS" ]; then
        print_info "Creating installation directory: $INSTALL_DIR_ABS"
        mkdir -p "$INSTALL_DIR_ABS"
    fi
    
    # Change to installation directory
    cd "$INSTALL_DIR_ABS" || {
        print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
        exit 1
    }
    print_info "Working in directory: $(pwd)"
    
    # Check for compose file
    if [ ! -f "$COMPOSE_FILE" ]; then
        print_warning "Compose file not found: $COMPOSE_FILE"
        print_info "Attempting to download from GitHub..."
        
        if command -v curl &> /dev/null; then
            curl -fsSL "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/$COMPOSE_FILE" -o "$COMPOSE_FILE" || {
                print_error "Failed to download compose file"
                print_info "Please download manually from:"
                print_info "https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager/blob/main/$COMPOSE_FILE"
                exit 1
            }
            print_success "Compose file downloaded successfully"
        elif command -v wget &> /dev/null; then
            wget -q "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/$COMPOSE_FILE" -O "$COMPOSE_FILE" || {
                print_error "Failed to download compose file"
                print_info "Please download manually from:"
                print_info "https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager/blob/main/$COMPOSE_FILE"
                exit 1
            }
            print_success "Compose file downloaded successfully"
        else
            print_error "Neither curl nor wget found. Cannot download compose file."
            print_info "Please download manually from:"
            print_info "https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager/blob/main/$COMPOSE_FILE"
            exit 1
        fi
    fi
    
    # Check for backend/scripts/postgres-init.sql (required by compose file)
    if [ ! -f "backend/scripts/postgres-init.sql" ]; then
        print_warning "PostgreSQL init script not found"
        print_info "Creating directory structure..."
        mkdir -p backend/scripts
        
        print_info "Downloading postgres-init.sql..."
        if command -v curl &> /dev/null; then
            curl -fsSL "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/backend/scripts/postgres-init.sql" -o "backend/scripts/postgres-init.sql" || {
                print_error "Failed to download postgres-init.sql"
                print_info "Please download manually from:"
                print_info "https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager/tree/main/backend/scripts"
                exit 1
            }
        elif command -v wget &> /dev/null; then
            wget -q "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/backend/scripts/postgres-init.sql" -O "backend/scripts/postgres-init.sql" || {
                print_error "Failed to download postgres-init.sql"
                print_info "Please download manually from:"
                print_info "https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager/tree/main/backend/scripts"
                exit 1
            }
        else
            print_error "Neither curl nor wget found. Cannot download required files."
            exit 1
        fi
        print_success "PostgreSQL init script downloaded"
    fi
    

    
    # Check for SSL certificates directory
    if [ ! -d "ssl" ]; then
        print_warning "SSL directory not found. Creating with self-signed certificates..."
        mkdir -p ssl
        
        # Generate self-signed certificate
        if command -v openssl &> /dev/null; then
            openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
                -keyout ssl/nginx-selfsigned.key \
                -out ssl/nginx-selfsigned.crt \
                -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" &> /dev/null
            print_success "Self-signed SSL certificates generated"
        else
            print_warning "openssl not found. SSL certificates not generated."
            print_info "HTTPS may not work without certificates in ./ssl/"
        fi
    fi
    
    # Check for .env file and generate if needed
    if [ ! -f .env ]; then
        print_warning ".env file not found."
        echo ""
        
        # Ask user if they want to generate .env automatically
        read -p "$(echo -e ${BLUE}?${NC}) Generate .env file with secure random keys? (Y/n): " -n 1 -r
        echo ""
        
        if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
            print_info "Generating .env file with secure defaults..."
            
            # Generate secure random keys
            if command -v openssl &> /dev/null; then
                DB_PASSWORD=$(openssl rand -base64 32 | tr -d "=+/" | cut -c1-32)
                JWT_SECRET=$(openssl rand -base64 32)
                ENCRYPTION_KEY=$(openssl rand -base64 32)
            else
                print_error "openssl not found. Cannot generate secure keys."
                print_info "Please install openssl or create .env manually"
                exit 1
            fi
            
            # Ask for admin credentials
            echo ""
            print_info "Set admin user credentials:"
            read -p "$(echo -e ${BLUE}?${NC}) Admin username (default: admin): " ADMIN_USERNAME
            ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
            
            read -sp "$(echo -e ${BLUE}?${NC}) Admin password (default: admin123): " ADMIN_PASSWORD
            echo ""
            ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin123}
            
            # Ask for server IP for CORS
            echo ""
            print_info "CORS Configuration:"
            DETECTED_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
            if [ -n "$DETECTED_IP" ]; then
                print_info "Detected server IP: $DETECTED_IP"
                read -p "$(echo -e ${BLUE}?${NC}) Use this IP for CORS? (Y/n): " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
                    SERVER_IP=$DETECTED_IP
                else
                    read -p "$(echo -e ${BLUE}?${NC}) Enter server IP address: " SERVER_IP
                fi
            else
                read -p "$(echo -e ${BLUE}?${NC}) Enter server IP address (or press Enter to skip): " SERVER_IP
            fi
            
            # Build ALLOWED_ORIGINS
            ALLOWED_ORIGINS="http://localhost,https://localhost,http://127.0.0.1,https://127.0.0.1"
            if [ -n "$SERVER_IP" ]; then
                ALLOWED_ORIGINS="${ALLOWED_ORIGINS},http://${SERVER_IP},https://${SERVER_IP}"
            fi
            
            # Create .env file
            cat > .env << EOF
# IBM Storage Virtualize Snapshot Manager - Environment Configuration
# Generated on $(date)

# ============================================================================
# DATABASE CONFIGURATION
# ============================================================================
DB_NAME=snapshots
DB_USER=snapshots
DB_PASSWORD=${DB_PASSWORD}

# ============================================================================
# SECURITY CONFIGURATION (REQUIRED)
# ============================================================================
JWT_SECRET=${JWT_SECRET}
ENCRYPTION_KEY=${ENCRYPTION_KEY}

# ============================================================================
# ADMIN USER CREDENTIALS
# ============================================================================
ADMIN_USERNAME=${ADMIN_USERNAME}
ADMIN_PASSWORD=${ADMIN_PASSWORD}

# ============================================================================
# SERVER CONFIGURATION
# ============================================================================
LOG_LEVEL=info

# CORS Configuration - Add your server IP addresses here
ALLOWED_ORIGINS=${ALLOWED_ORIGINS}

# ============================================================================
# PORT CONFIGURATION (Optional)
# ============================================================================
FRONTEND_HTTPS_PORT=443

# ============================================================================
# SSL CERTIFICATE CONFIGURATION (Optional)
# ============================================================================
SSL_CERT_PATH=./ssl
EOF
            
            print_success ".env file created successfully"
            print_info ""
            print_info "Configuration summary:"
            print_info "  - Database password: [HIDDEN]"
            print_info "  - Admin username: ${ADMIN_USERNAME}"
            print_info "  - Admin password: [HIDDEN]"
            if [ -n "$SERVER_IP" ]; then
                print_info "  - Server IP: ${SERVER_IP}"
            fi
            print_info ""
            
        else
            print_info "Please create .env file manually or download .env.example"
            if [ -f .env.example ]; then
                cp .env.example .env
                print_info ".env.example copied to .env"
                print_warning "Please edit .env and set required values"
            else
                print_info "Downloading .env.example..."
                if command -v curl &> /dev/null; then
                    curl -fsSL "https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/.env.example" -o .env
                    print_success ".env.example downloaded"
                    print_warning "Please edit .env and set required values"
                fi
            fi
            exit 1
        fi
    fi
    
    # Check for required environment variables
    source .env
    if [ -z "$DB_PASSWORD" ] || [ -z "$JWT_SECRET" ] || [ -z "$ENCRYPTION_KEY" ]; then
        print_error "Missing required environment variables in .env file"
        print_info "Required: DB_PASSWORD, JWT_SECRET, ENCRYPTION_KEY"
        print_info "Delete .env and run this script again to regenerate"
        exit 1
    fi
    
    # Check if ALLOWED_ORIGINS is missing and add it
    if [ -z "$ALLOWED_ORIGINS" ]; then
        print_warning "ALLOWED_ORIGINS not found in .env file"
        
        # Detect server IP
        DETECTED_IP=$(hostname -I 2>/dev/null | awk '{print $1}')
        if [ -n "$DETECTED_IP" ]; then
            print_info "Detected server IP: $DETECTED_IP"
            read -p "$(echo -e ${BLUE}?${NC}) Add this IP to ALLOWED_ORIGINS? (Y/n): " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
                echo "" >> .env
                echo "# CORS Configuration - Add your server IP addresses here" >> .env
                echo "ALLOWED_ORIGINS=http://localhost,https://localhost,http://${DETECTED_IP},https://${DETECTED_IP},http://127.0.0.1,https://127.0.0.1" >> .env
                print_success "ALLOWED_ORIGINS added to .env"
            else
                echo "" >> .env
                echo "# CORS Configuration - Add your server IP addresses here" >> .env
                echo "ALLOWED_ORIGINS=http://localhost,https://localhost,http://127.0.0.1,https://127.0.0.1" >> .env
                print_success "ALLOWED_ORIGINS added to .env (localhost only)"
            fi
        else
            echo "" >> .env
            echo "# CORS Configuration - Add your server IP addresses here" >> .env
            echo "ALLOWED_ORIGINS=http://localhost,https://localhost,http://127.0.0.1,https://127.0.0.1" >> .env
            print_success "ALLOWED_ORIGINS added to .env (localhost only)"
        fi
    fi
    
    print_success "Prerequisites check passed"
}

pull_images() {
    print_info "Pulling latest images from GitHub Container Registry..."
    
    print_info "Pulling backend image..."
    podman pull "$BACKEND_IMAGE" || {
        print_warning "Failed to pull backend image. Will use local image if available."
    }
    
    print_info "Pulling frontend image..."
    podman pull "$FRONTEND_IMAGE" || {
        print_warning "Failed to pull frontend image. Will use local image if available."
    }
    
    print_info "Pulling PostgreSQL image..."
    podman pull "$POSTGRES_IMAGE" || {
        print_error "Failed to pull PostgreSQL image"
        exit 1
    }
    
    print_success "Images pulled successfully"
}

start_stack() {
    print_info "Starting IBM Storage Virtualize Snapshot Manager..."
    
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" up -d
    
    print_success "Stack started successfully"
    
    # Wait for database to be ready
    print_info "Waiting for database to be ready..."
    sleep 10
    
    # Check if admin user credentials are set in .env
    source .env
    if [ -n "$ADMIN_USERNAME" ] && [ -n "$ADMIN_PASSWORD" ]; then
        print_info "Creating admin user with credentials from .env..."
        
        # Wait for backend to be healthy
        for i in {1..30}; do
            if podman exec snapshot-manager-backend wget -q --spider http://localhost:8080/health 2>/dev/null; then
                break
            fi
            if [ $i -eq 30 ]; then
                print_warning "Backend not ready, skipping admin user creation"
                print_info "You can create admin user manually later with: ./reset-admin-password.sh"
                break
            fi
            sleep 2
        done
        
        # Generate bcrypt hash - try multiple methods
        print_info "Generating password hash..."
        ADMIN_HASH=""
        
        # Method 1: Try using Go locally (if available and backend/scripts/genhash.go exists)
        if command -v go &> /dev/null && [ -f "backend/scripts/genhash.go" ]; then
            print_info "Using local Go to generate hash..."
            ADMIN_HASH=$(cd backend && go run scripts/genhash.go "${ADMIN_PASSWORD}" 2>/dev/null)
        fi
        
        # Method 2: Fall back to Python with bcrypt
        if [ -z "$ADMIN_HASH" ] && command -v python3 &> /dev/null; then
            print_info "Using Python to generate hash..."
            # Check if bcrypt module is available
            if python3 -c "import bcrypt" 2>/dev/null; then
                ADMIN_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw('${ADMIN_PASSWORD}'.encode('utf-8'), bcrypt.gensalt(rounds=10)).decode('utf-8'))" 2>/dev/null)
            else
                print_info "Installing Python bcrypt module..."
                pip3 install bcrypt --user --quiet 2>/dev/null || true
                if python3 -c "import bcrypt" 2>/dev/null; then
                    ADMIN_HASH=$(python3 -c "import bcrypt; print(bcrypt.hashpw('${ADMIN_PASSWORD}'.encode('utf-8'), bcrypt.gensalt(rounds=10)).decode('utf-8'))" 2>/dev/null)
                fi
            fi
        fi
        
        if [ -n "$ADMIN_HASH" ]; then
            # Check if admin user exists
            USER_EXISTS=$(podman exec snapshot-manager-db psql -U snapshots -d snapshots -t -c \
                "SELECT COUNT(*) FROM users WHERE username='${ADMIN_USERNAME}';" 2>/dev/null | tr -d ' ')
            
            if [ "$USER_EXISTS" = "0" ]; then
                # Create admin user
                podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
                    "INSERT INTO users (username, password_hash, email, role, created_at, updated_at) 
                     VALUES ('${ADMIN_USERNAME}', '${ADMIN_HASH}', '${ADMIN_USERNAME}@localhost', 'admin', NOW(), NOW());" > /dev/null 2>&1
                
                if [ $? -eq 0 ]; then
                    print_success "Admin user created successfully"
                else
                    print_warning "Failed to create admin user"
                    print_info "You can create it manually with: ./reset-admin-password.sh"
                fi
            else
                # Update existing admin user password
                podman exec snapshot-manager-db psql -U snapshots -d snapshots -c \
                    "UPDATE users SET password_hash='${ADMIN_HASH}' WHERE username='${ADMIN_USERNAME}';" > /dev/null 2>&1
                print_success "Admin user password updated"
            fi
        else
            print_warning "Failed to generate password hash"
            print_info "You can create admin user manually with: ./reset-admin-password.sh"
        fi
    fi
    
    print_info ""
    print_info "Services:"
    
    # Get server IP from .env if available
    source .env
    if [ -n "$ALLOWED_ORIGINS" ] && echo "$ALLOWED_ORIGINS" | grep -qE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+'; then
        SERVER_IP=$(echo "$ALLOWED_ORIGINS" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        print_info "  - Frontend (HTTPS): https://${SERVER_IP}"
    fi
    
    print_info "  - Frontend (HTTPS): https://localhost"
    
    if [ -n "$ADMIN_USERNAME" ]; then
        print_info "  - Admin username: ${ADMIN_USERNAME}"
        print_info "  - Admin password: [as configured in .env]"
    else
        print_info "  - Default credentials: admin / admin123"
    fi
    
    print_info ""
    print_info "To view logs: podman-compose -f $COMPOSE_FILE logs -f"
    print_info "To stop:      podman-compose -f $COMPOSE_FILE down"
}

stop_stack() {
    print_info "Stopping IBM Storage Virtualize Snapshot Manager..."
    
    if [ ! -d "$INSTALL_DIR_ABS" ]; then
        print_error "Installation directory not found: $INSTALL_DIR_ABS"
        print_info "Please run './start-prod.sh' first to create the installation"
        exit 1
    fi
    
    cd "$INSTALL_DIR_ABS" || {
        print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
        exit 1
    }
    
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" down
    print_success "Stack stopped"
}

clean_stack() {
    print_warning "This will stop and remove all containers, volumes, and images"
    read -p "$(echo -e ${YELLOW}⚠${NC}) Are you sure? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [ ! -d "$INSTALL_DIR_ABS" ]; then
            print_error "Installation directory not found: $INSTALL_DIR_ABS"
            exit 1
        fi
        
        cd "$INSTALL_DIR_ABS" || {
            print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
            exit 1
        }
        
        print_info "Stopping and removing containers..."
        set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" down -v
        
        print_info "Removing images..."
        podman rmi "$BACKEND_IMAGE" 2>/dev/null || true
        podman rmi "$FRONTEND_IMAGE" 2>/dev/null || true
        podman rmi "$POSTGRES_IMAGE" 2>/dev/null || true
        
        print_success "Cleanup completed"
        print_info "All containers, volumes, and images have been removed"
    else
        print_info "Cleanup cancelled"
    fi
}

rebuild_stack() {
    print_info "Rebuilding and restarting stack..."
    
    if [ ! -d "$INSTALL_DIR_ABS" ]; then
        print_error "Installation directory not found: $INSTALL_DIR_ABS"
        exit 1
    fi
    
    cd "$INSTALL_DIR_ABS" || {
        print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
        exit 1
    }
    
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" down
    pull_images
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" up -d --force-recreate
    print_success "Stack rebuilt and restarted"
}

show_logs() {
    print_info "Showing logs (Ctrl+C to exit)..."
    
    if [ ! -d "$INSTALL_DIR_ABS" ]; then
        print_error "Installation directory not found: $INSTALL_DIR_ABS"
        print_info "Please run './start-prod.sh' first to create the installation"
        exit 1
    fi
    
    cd "$INSTALL_DIR_ABS" || {
        print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
        exit 1
    }
    
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" logs -f
}

show_status() {
    print_info "Container status:"
    
    if [ ! -d "$INSTALL_DIR_ABS" ]; then
        print_error "Installation directory not found: $INSTALL_DIR_ABS"
        print_info "Please run './start-prod.sh' first to create the installation"
        exit 1
    fi
    
    cd "$INSTALL_DIR_ABS" || {
        print_error "Failed to change to installation directory: $INSTALL_DIR_ABS"
        exit 1
    }
    
    # Export env vars from .env file before calling podman-compose
    set -a
    [ -f .env ] && source .env
    set +a
    
    set -a; [ -f .env ] && source .env; set +a; podman-compose -f "$COMPOSE_FILE" ps
}

setup_autostart() {
    print_info "Setting up auto-start on boot (systemd user service)..."
    
    # Create systemd user directory if it doesn't exist
    SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
    mkdir -p "$SYSTEMD_USER_DIR"
    
    # Use absolute path to installation directory
    INSTALL_PATH="$INSTALL_DIR_ABS"
    COMPOSE_FILE_PATH="$INSTALL_PATH/$COMPOSE_FILE"
    
    # Detect actual path to podman-compose
    PODMAN_COMPOSE_PATH=$(command -v podman-compose)
    if [ -z "$PODMAN_COMPOSE_PATH" ]; then
        print_error "podman-compose not found in PATH"
        exit 1
    fi
    print_info "Using podman-compose at: $PODMAN_COMPOSE_PATH"
    
    # Create systemd service file
    SERVICE_FILE="$SYSTEMD_USER_DIR/snapshot-manager.service"
    
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=IBM Storage Virtualize Snapshot Manager
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$INSTALL_PATH
ExecStart=$PODMAN_COMPOSE_PATH -f $COMPOSE_FILE_PATH up -d
ExecStop=$PODMAN_COMPOSE_PATH -f $COMPOSE_FILE_PATH down
TimeoutStartSec=0

[Install]
WantedBy=default.target
EOF
    
    print_success "Systemd service file created: $SERVICE_FILE"
    
    # Enable linger for user (allows services to run when user is not logged in)
    if command -v loginctl &> /dev/null; then
        print_info "Enabling user linger (allows service to run without login)..."
        loginctl enable-linger "$USER" 2>/dev/null || {
            print_warning "Could not enable linger. Service may not start on boot."
            print_info "Run manually: loginctl enable-linger $USER"
        }
    fi
    
    # Reload systemd and enable service
    print_info "Enabling service..."
    systemctl --user daemon-reload
    systemctl --user enable snapshot-manager.service
    
    print_success "Auto-start configured successfully"
    print_info ""
    print_info "Service commands:"
    print_info "  Start:   systemctl --user start snapshot-manager"
    print_info "  Stop:    systemctl --user stop snapshot-manager"
    print_info "  Status:  systemctl --user status snapshot-manager"
    print_info "  Disable: systemctl --user disable snapshot-manager"
    print_info ""
    print_info "The service will now start automatically on boot"
}

remove_autostart() {
    print_info "Removing auto-start configuration..."
    
    SYSTEMD_USER_DIR="$HOME/.config/systemd/user"
    SERVICE_FILE="$SYSTEMD_USER_DIR/snapshot-manager.service"
    
    if [ -f "$SERVICE_FILE" ]; then
        systemctl --user stop snapshot-manager.service 2>/dev/null || true
        systemctl --user disable snapshot-manager.service 2>/dev/null || true
        rm -f "$SERVICE_FILE"
        systemctl --user daemon-reload
        print_success "Auto-start removed"
    else
        print_info "No auto-start configuration found"
    fi
}

# Main script logic
case "${1:-}" in
    --stop)
        stop_stack
        ;;
    --clean)
        clean_stack
        ;;
    --rebuild)
        check_prerequisites
        rebuild_stack
        ;;
    --logs)
        show_logs
        ;;
    --status)
        show_status
        ;;
    --autostart)
        setup_autostart
        ;;
    --remove-autostart)
        remove_autostart
        ;;
    --help|-h)
        echo "Usage: $0 [OPTION]"
        echo ""
        echo "This script can be run from two locations:"
        echo "  1. From parent directory (where start-prod.sh is located)"
        echo "  2. From inside ibm-virtualize-snapshot-manager-dir/"
        echo ""
        echo "Options:"
        echo "  (no option)         Pull latest images and start stack"
        echo "  --rebuild           Force rebuild and restart all services"
        echo "  --stop              Stop all services"
        echo "  --clean             Stop and remove all containers, volumes, and images"
        echo "  --logs              View logs (follow mode)"
        echo "  --status            Show container status"
        echo "  --autostart         Configure auto-start on boot (systemd user service)"
        echo "  --remove-autostart  Remove auto-start configuration"
        echo "  --help              Show this help message"
        echo ""
        echo "Examples:"
        echo "  $0                       # Start with latest images"
        echo "  $0 --rebuild             # Force rebuild"
        echo "  $0 --stop                # Stop services"
        echo "  $0 --clean               # Complete cleanup (removes everything)"
        echo "  $0 --logs                # View logs"
        echo "  $0 --autostart           # Enable auto-start on boot"
        echo "  $0 --remove-autostart    # Disable auto-start"
        echo ""
        echo "Note: Management commands (--stop, --logs, --status, --clean, --rebuild)"
        echo "      work from both the parent directory and installation directory"
        ;;
    "")
        check_prerequisites
        pull_images
        start_stack
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
