# Quick Start Guide

Get the IBM Storage Virtualize Snapshot Manager up and running in 5 minutes.

## Prerequisites

- Go 1.21 or higher installed
- IBM Storage Virtualize system (FlashSystem) with REST API access

## Step 1: Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd ibm-storage-virtualize-snapshot-manager

# Navigate to backend
cd backend

# Install dependencies
go mod download
```

## Step 2: Configure Environment

```bash
# Copy environment template
cp .env.example .env

# Generate secure keys
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { key := make([]byte, 32); rand.Read(key); fmt.Println("ENCRYPTION_KEY=" + base64.StdEncoding.EncodeToString(key)); secret := make([]byte, 32); rand.Read(secret); fmt.Println("JWT_SECRET=" + base64.StdEncoding.EncodeToString(secret)) }'

# Edit .env and add the generated keys
nano .env
```

Or use the setup script:

```bash
make setup
```

## Step 3: Create Directories

```bash
mkdir -p data logs
```

## Step 4: Build and Run

```bash
# Build the application
make build

# Or run directly
make run
```

The server will start on `http://localhost:8080`

## Step 5: Create Initial User

```bash
# In a new terminal
make create-user
```

Follow the prompts to create your admin user.

## Step 6: Test the API

```bash
# Health check
curl http://localhost:8080/health

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'

# Save the token from the response
export TOKEN="your-jwt-token-here"

# Test authenticated endpoint
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## Step 7: Add Your First Storage System

```bash
curl -X POST http://localhost:8080/api/systems \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "FlashSystem-01",
    "ip_address": "192.168.1.100",
    "port": 7443,
    "username": "superuser",
    "password": "your-svc-password"
  }'
```

## Step 8: Test Connection

```bash
# Replace {id} with the system ID from previous response
curl -X POST http://localhost:8080/api/systems/{id}/test \
  -H "Authorization: Bearer $TOKEN"
```

## Step 9: Sync Volume Groups

```bash
curl -X POST http://localhost:8080/api/systems/{id}/volumegroups/sync \
  -H "Authorization: Bearer $TOKEN"
```

## Step 10: Create a Snapshot Schedule

```bash
curl -X POST http://localhost:8080/api/volumegroups/{vg_id}/schedules \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Daily Backup",
    "cron_expression": "0 2 * * *",
    "retention_days": 7,
    "safeguarded": false
  }'
```

## Common Cron Expressions

```
0 2 * * *       # Daily at 2 AM
0 */6 * * *     # Every 6 hours
*/15 * * * *    # Every 15 minutes
0 0 * * 0       # Weekly on Sunday at midnight
0 3 1 * *       # Monthly on the 1st at 3 AM
```

## Troubleshooting

### Database Issues

```bash
# Check if database exists
ls -la data/

# Reset database (WARNING: Deletes all data)
make clean
make run  # Will recreate database
```

### Connection Issues

1. Verify IBM SVC is reachable: `ping <ip-address>`
2. Check port 7443 is open: `telnet <ip-address> 7443`
3. Verify credentials are correct
4. Check logs: `tail -f logs/app.log`

### Authentication Issues

```bash
# Recreate user
make create-user

# Check users in database
sqlite3 data/snapshots.db "SELECT * FROM users;"
```

## Next Steps

1. **Add more systems**: Repeat steps 7-9 for additional IBM SVC systems
2. **Create schedules**: Set up snapshot schedules for your volume groups
3. **Monitor executions**: Check `/api/executions` for snapshot history
4. **Configure retention**: Adjust retention days based on your backup policy

## Useful Commands

```bash
# View all systems
curl http://localhost:8080/api/systems \
  -H "Authorization: Bearer $TOKEN"

# View all schedules for a volume group
curl http://localhost:8080/api/volumegroups/{vg_id}/schedules \
  -H "Authorization: Bearer $TOKEN"

# View execution history
curl http://localhost:8080/api/executions \
  -H "Authorization: Bearer $TOKEN"

# Dashboard statistics
curl http://localhost:8080/api/dashboard/stats \
  -H "Authorization: Bearer $TOKEN"
```

## Production Deployment

For production use:

1. Use PostgreSQL instead of SQLite
2. Set up HTTPS with proper certificates
3. Configure firewall rules
4. Set up log rotation
5. Monitor the application with systemd or supervisor
6. Back up the database regularly
7. Use strong, unique keys for JWT_SECRET and ENCRYPTION_KEY

## Getting Help

- Check the [README.md](README.md) for detailed documentation
- Review [PROJECT_PLAN.md](PROJECT_PLAN.md) for architecture details
- Check logs in `logs/app.log`
- Review IBM Storage Virtualize REST API documentation

## What's Next?

The backend is now running! Next steps:

1. **Frontend Development**: Build the web UI (React/Vue.js)
2. **Enhanced Features**: Add email notifications, cleanup automation
3. **Monitoring**: Add Prometheus metrics
4. **Deployment**: Containerize with Docker

Happy snapshot scheduling! 🎉