#!/bin/bash
set -e

BACKUP_DIR="./backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/snapshot-manager-backup-$TIMESTAMP.sql"

echo "=========================================="
echo "IBM Storage Virtualize Snapshot Manager"
echo "Database Backup"
echo "=========================================="
echo ""

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check if database container is running
if ! podman ps | grep -q snapshot-manager-db; then
    echo "❌ Error: Database container is not running"
    echo "   Start services with: ./deploy/start.sh"
    exit 1
fi

echo "Creating database backup..."
echo "Backup file: ${BACKUP_FILE}.gz"
echo ""

# Backup PostgreSQL database
podman exec snapshot-manager-db pg_dump -U ${DB_USER:-snapshots} ${DB_NAME:-snapshots} > "$BACKUP_FILE"

# Compress backup
gzip "$BACKUP_FILE"

# Get file size
BACKUP_SIZE=$(du -h "${BACKUP_FILE}.gz" | cut -f1)

echo ""
echo "✓ Backup created successfully!"
echo ""
echo "Backup details:"
echo "  File: ${BACKUP_FILE}.gz"
echo "  Size: $BACKUP_SIZE"
echo "  Time: $(date)"
echo ""
echo "To restore this backup:"
echo "  gunzip ${BACKUP_FILE}.gz"
echo "  podman exec -i snapshot-manager-db psql -U snapshots -d snapshots < $BACKUP_FILE"
echo ""

# Clean up old backups (keep last 7 days)
echo "Cleaning up old backups (keeping last 7 days)..."
find "$BACKUP_DIR" -name "snapshot-manager-backup-*.sql.gz" -mtime +7 -delete
echo "✓ Cleanup complete"
echo ""
