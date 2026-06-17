# IBM Storage Virtualize Snapshot Manager

A web-based application for managing multiple snapshot schedules per volume group on IBM Storage Virtualize systems (FlashSystem).

> **⚠️ DISCLAIMER**: This is an unofficial, community-driven project and is NOT affiliated with, endorsed by, or supported by IBM Corporation. This software is provided "AS IS" without warranty of any kind. Use at your own risk.

## Overview

IBM Storage Virtualize allows only ONE snapshot policy per volume group. This application works around that limitation by managing multiple schedules and triggering snapshots directly via the REST API.

## Features

- **Multi-System Management**: Connect and manage multiple IBM Storage Virtualize systems
- **Multiple Schedules per Volume Group**: Create unlimited snapshot schedules for each volume group
- **Cron-based Scheduling**: Flexible scheduling using cron expressions
- **Retention Management**: Configure retention days and safeguarded snapshots
- **Execution History**: Track all snapshot operations with detailed logs
- **Web Interface**: Modern, responsive UI for easy management

## Architecture

- **Backend**: Go (Golang) with REST API
- **Frontend**: React/Vue.js (to be implemented)
- **Database**: SQLite (default) or PostgreSQL
- **Scheduler**: Cron-based snapshot execution
- **API Integration**: IBM Storage Virtualize REST API

## Prerequisites

- Go 1.21 or higher
- IBM Storage Virtualize system (FlashSystem) with REST API access
- SQLite (included) or PostgreSQL (optional)

## Installation

Choose one of the following installation methods:

### Option 1: Container Deployment (Recommended)

**Quick Start with Podman:**

```bash
# 1. Clone the repository
git clone <repository-url>
cd ibm-storage-virtualize-snapshot-manager

# 2. Run setup (generates keys and certificates)
./deploy/setup.sh

# 3. Review and update .env file
nano .env

# 4. Start the application
./deploy/start.sh

# 5. Access at https://localhost
```

**Features:**
- ✅ Production-ready with PostgreSQL database
- ✅ HTTPS with self-signed certificates (easily replaceable)
- ✅ Nginx reverse proxy with security headers
- ✅ Automatic health checks and restart policies
- ✅ Easy backup and restore scripts

**See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed instructions.**

### Option 2: Local Development Setup

### 1. Clone the repository

```bash
git clone <repository-url>
cd ibm-storage-virtualize-snapshot-manager
```

### 2. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Create data and logs directories
mkdir -p data logs

# Copy environment configuration
cp .env.example .env

# Edit .env and set required values:
# - JWT_SECRET: Generate a secure random string
# - ENCRYPTION_KEY: Generate a 32-byte key (see below)
```

### 3. Generate Encryption Key

```bash
# Generate a secure 32-byte encryption key
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { key := make([]byte, 32); rand.Read(key); fmt.Println(base64.StdEncoding.EncodeToString(key)) }'
```

Or use this helper:

```go
package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	key := make([]byte, 32)
	rand.Read(key)
	fmt.Println("ENCRYPTION_KEY=" + base64.StdEncoding.EncodeToString(key))
	
	secret := make([]byte, 32)
	rand.Read(secret)
	fmt.Println("JWT_SECRET=" + base64.StdEncoding.EncodeToString(secret))
}
```

### 4. Create Initial User

```bash
# After starting the server, create an initial user directly in the database
sqlite3 data/snapshots.db

# Generate password hash (use bcrypt)
# Then insert user:
INSERT INTO users (username, password_hash, email) 
VALUES ('admin', '<bcrypt_hash>', 'admin@example.com');
```

## Running the Application

### Development Mode

```bash
cd backend
go run cmd/server/main.go
```

### Production Build

```bash
cd backend
go build -o snapshot-manager cmd/server/main.go
./snapshot-manager
```

The server will start on `http://localhost:8080` by default.

## API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `GET /api/auth/me` - Get current user

### Storage Systems
- `GET /api/systems` - List all systems
- `POST /api/systems` - Add new system
- `GET /api/systems/:id` - Get system details
- `PUT /api/systems/:id` - Update system
- `DELETE /api/systems/:id` - Delete system
- `POST /api/systems/:id/test` - Test connection
- `GET /api/systems/:id/volumegroups` - List volume groups
- `POST /api/systems/:id/volumegroups/sync` - Sync volume groups

### Snapshot Schedules
- `GET /api/volumegroups/:id/schedules` - List schedules
- `POST /api/volumegroups/:id/schedules` - Create schedule
- `GET /api/schedules/:id` - Get schedule details
- `PUT /api/schedules/:id` - Update schedule
- `DELETE /api/schedules/:id` - Delete schedule
- `POST /api/schedules/:id/execute` - Manually trigger

### Monitoring
- `GET /api/executions` - List execution history
- `GET /api/dashboard/stats` - Dashboard statistics

## Configuration

### Environment Variables

See `.env.example` for all available configuration options.

### Cron Expression Examples

```
# Every day at 2 AM
0 2 * * *

# Every 6 hours
0 */6 * * *

# Every Monday at 3 AM
0 3 * * 1

# Every 15 minutes
*/15 * * * *
```

## IBM Storage Virtualize API Constraints

**CRITICAL - Rate Limits:**
- Auth endpoint: 3 requests/second maximum
- Command endpoints: 10 requests/second maximum
- Max 4 different tokens per cluster
- Returns HTTP 429 if limits exceeded

**API Quirks:**
- ALL endpoints use POST method (even for reading data)
- Auth uses `X-Auth-Username` and `X-Auth-Password` headers
- Subsequent requests use `X-Auth-Token` header
- HTTPS required on port 7443
- Token expiry is encoded in JWT

## Database Schema

The application uses the following tables:

- `users` - User accounts
- `storage_systems` - IBM SVC systems (with encrypted credentials)
- `volume_groups` - Cached volume group information
- `snapshot_schedules` - Snapshot schedules (multiple per VG)
- `snapshot_executions` - Execution history and audit log

## Security

- User passwords are hashed with bcrypt
- Storage system passwords are encrypted with AES-256-GCM
- JWT tokens for API authentication
- HTTPS recommended for production

## Troubleshooting

### Database Issues

```bash
# Check database
sqlite3 data/snapshots.db ".tables"

# Reset database (WARNING: Deletes all data)
rm data/snapshots.db
# Restart application to recreate
```

### Connection Issues

- Verify IBM SVC system is reachable
- Check firewall rules (port 7443)
- Verify credentials are correct
- Check rate limits (max 3 auth requests/second)

### Scheduler Issues

- Check logs for errors
- Verify cron expressions are valid
- Ensure schedules are marked as active
- Check system time and timezone

## Development

### Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/             # REST API handlers
│   ├── auth/            # Authentication service
│   ├── config/          # Configuration management
│   ├── db/              # Database connection and schema
│   ├── models/          # Data models
│   ├── scheduler/       # Snapshot scheduler
│   └── svc/             # IBM SVC REST API client
├── pkg/
│   └── crypto/          # Encryption utilities
└── go.mod
```

### Running Tests

```bash
cd backend
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

### Important Legal Notices

**NO WARRANTY**: This software is provided "AS IS", WITHOUT WARRANTY OF ANY KIND, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no event shall the authors or copyright holders be liable for any claim, damages or other liability, whether in an action of contract, tort or otherwise, arising from, out of or in connection with the software or the use or other dealings in the software.

**USE AT YOUR OWN RISK**: This application interacts with IBM Storage Virtualize systems and manages snapshot operations. Users are solely responsible for:
- Testing thoroughly in non-production environments before production use
- Ensuring proper backup and disaster recovery procedures are in place
- Verifying snapshot operations and retention policies
- Understanding the impact of snapshot operations on their storage systems
- Compliance with their organization's policies and procedures

**NOT AN IBM PRODUCT**: This is an independent, open-source project. IBM, IBM Storage, IBM Storage Virtualize, and FlashSystem are trademarks of IBM Corporation. This project is not affiliated with, endorsed by, or supported by IBM Corporation.

**COMMUNITY SUPPORT ONLY**: Support is provided on a best-effort basis by the community. There are no service level agreements (SLAs) or guaranteed response times.

## Support

For issues and questions:
- Check the troubleshooting section
- Review IBM Storage Virtualize REST API documentation
- Open an issue on GitHub

## Deployment Options

### Container Deployment (Production)
- ✅ **Podman/Docker support** - Full containerization with podman-compose.yml
- ✅ **PostgreSQL database** - Production-ready database
- ✅ **Nginx with HTTPS** - Secure frontend with SSL/TLS
- ✅ **Automated deployment** - One-command setup and start scripts
- ✅ **Backup scripts** - Automated database backup and restore

See [DEPLOYMENT.md](DEPLOYMENT.md) for complete deployment guide.

### Local Development
- ✅ **SQLite database** - Lightweight development database
- ✅ **Direct Go execution** - Fast development iteration
- ✅ **Hot reload** - Frontend development server

## Roadmap

- [x] Container deployment with Podman/Docker
- [x] PostgreSQL database support
- [x] HTTPS with Nginx reverse proxy
- [x] Automated backup scripts
- [ ] Email notifications for failures
- [ ] Snapshot cleanup automation
- [ ] Multi-tenancy support
- [ ] Prometheus metrics
- [ ] Kubernetes support