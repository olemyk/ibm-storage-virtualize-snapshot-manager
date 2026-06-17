#!/bin/bash
set -e

echo "=========================================="
echo "IBM Storage Virtualize Snapshot Manager"
echo "Starting Container Stack"
echo "=========================================="
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "❌ Error: .env file not found"
    echo "   Run ./deploy/setup.sh first"
    exit 1
fi

# Load environment variables
echo "Loading environment configuration..."
export $(cat .env | grep -v '^#' | xargs)

# Validate required environment variables
if [ -z "$JWT_SECRET" ]; then
    echo "❌ Error: JWT_SECRET not set in .env file"
    exit 1
fi

if [ -z "$ENCRYPTION_KEY" ]; then
    echo "❌ Error: ENCRYPTION_KEY not set in .env file"
    exit 1
fi

if [ -z "$DB_PASSWORD" ]; then
    echo "❌ Error: DB_PASSWORD not set in .env file"
    exit 1
fi

echo "✓ Environment configuration loaded"
echo ""

# Build containers
echo "Building container images..."
echo "This may take a few minutes on first run..."
echo ""

podman-compose build --no-cache

echo ""
echo "✓ Container images built successfully"
echo ""

# Start services
echo "Starting services..."
podman-compose up -d

echo ""
echo "Waiting for services to be healthy..."
echo "This may take 30-60 seconds..."
echo ""

# Wait for services to be healthy
MAX_WAIT=120
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $MAX_WAIT ]; do
    # Check if all containers are running
    RUNNING=$(podman-compose ps --format json 2>/dev/null | grep -c '"State":"running"' || echo "0")
    
    if [ "$RUNNING" -eq "3" ]; then
        echo "✓ All services are running"
        break
    fi
    
    echo "  Waiting... ($ELAPSED/$MAX_WAIT seconds)"
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))
done

if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo "⚠️  Warning: Services took longer than expected to start"
    echo "   Check logs with: podman-compose logs"
fi

echo ""
echo "Checking service status..."
podman-compose ps

echo ""
echo "=========================================="
echo "Deployment Complete!"
echo "=========================================="
echo ""
echo "Access the application at:"
echo "  🔒 HTTPS: https://localhost"
echo "  🔓 HTTP:  http://localhost (redirects to HTTPS)"
echo ""
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: admin123"
echo ""
echo "⚠️  IMPORTANT: Change the default password after first login!"
echo ""
echo "Useful commands:"
echo "  View logs:        podman-compose logs -f"
echo "  View backend:     podman-compose logs -f backend"
echo "  View frontend:    podman-compose logs -f frontend"
echo "  View database:    podman-compose logs -f postgres"
echo "  Stop services:    ./deploy/stop.sh"
echo "  Restart:          podman-compose restart"
echo "  Shell (backend):  podman exec -it snapshot-manager-backend sh"
echo "  Shell (database): podman exec -it snapshot-manager-db psql -U snapshots -d snapshots"
echo ""
