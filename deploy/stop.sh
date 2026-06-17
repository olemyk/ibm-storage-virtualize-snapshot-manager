#!/bin/bash
set -e

echo "=========================================="
echo "IBM Storage Virtualize Snapshot Manager"
echo "Stopping Container Stack"
echo "=========================================="
echo ""

# Stop and remove containers
echo "Stopping services..."
podman-compose down

echo ""
echo "✓ All services stopped"
echo ""
echo "Note: Data volumes are preserved"
echo "To remove volumes as well, run: podman-compose down -v"
echo ""
