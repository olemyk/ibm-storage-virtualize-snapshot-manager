# Release Notes - IBM Storage Virtualize Snapshot Manager v1.0.0

**Release Date:** June 17, 2026

**⚠️ IMPORTANT DISCLAIMER:** This is an unofficial, community-driven project and is NOT affiliated with, endorsed by, or supported by IBM Corporation. This software is provided "AS IS" without warranty of any kind. Use at your own risk.

---

## 🎉 First Official Release

We're excited to announce the first official release of IBM Storage Virtualize Snapshot Manager! This release provides a complete, production-ready solution for managing multiple snapshot schedules per volume group on IBM Storage Virtualize systems (FlashSystem).

---

## 🌟 Key Features

### 🚀 NEW: One-Liner Production Deployment with Interactive Setup
- **Standalone Script**: Download single script - no git clone required!
- **Interactive Configuration**: Guided setup prompts for all settings
- **Auto-Generate Credentials**: Secure random keys for DB_PASSWORD, JWT_SECRET, ENCRYPTION_KEY
- **Custom Admin User**: Set your own admin username and password during setup
- **Auto-Detect Server IP**: Automatically configures CORS for your server
- **Admin User Creation**: Creates admin user in database with your credentials
- **No Default Password Issues**: Your chosen password works immediately - no more "admin123" problems!
- **Auto-Download Dependencies**: Automatically fetches compose file, init scripts, and SSL certs
- **Simple Startup**: `./start-prod.sh` - interactive guided deployment
- **Pre-built Images**: Available on GitHub Container Registry (GHCR)
- **Multi-Architecture**: Supports amd64 and arm64 (Apple Silicon, ARM servers)
- **Automatic Updates**: Pull latest images with `./start-prod.sh --rebuild`
- **Status Management**: Built-in commands for logs, status, stop, and restart
- **Environment Validation**: Checks prerequisites and validates configuration
- **Color-Coded Output**: Clear, helpful status messages throughout deployment

### Multi-System Management
- **Connect Multiple Systems**: Manage multiple IBM Storage Virtualize (FlashSystem) systems from a single interface
- **Encrypted Credentials**: Storage system passwords encrypted with AES-256-GCM
- **Connection Testing**: Test connectivity before saving system configurations
- **TLS Verification Control**: Option to skip TLS verification for self-signed certificates
- **Connection Status Monitoring**: Real-time connection status tracking for all systems

### Snapshot Scheduling
- **Multiple Schedules per Volume Group**: Overcome IBM SVC's one-policy-per-VG limitation
- **Flexible Cron Expressions**: Use standard cron syntax for scheduling (e.g., `0 2 * * *` for daily at 2 AM)
- **Retention Management**: Configure retention days for automatic cleanup
- **Safeguarded Snapshots**: Create immutable snapshots for ransomware protection
- **Custom Snapshot Names**: Define naming patterns with variables (system, VG, timestamp)
- **Manual Execution**: Trigger snapshots on-demand outside of schedules
- **Active/Inactive Schedules**: Enable or disable schedules without deletion

### Volume Group Management
- **Automatic Synchronization**: Sync volume groups from IBM SVC systems
- **Partition Support**: Track volume group partitions and ownership
- **Capacity Tracking**: Monitor used and total capacity
- **Pool Information**: View associated storage pools
- **Snapshot Count**: Track number of existing snapshots per VG

### Execution History & Audit Logs
- **Complete Audit Trail**: Track all snapshot operations with detailed logs
- **Execution Status**: Monitor success, failure, and error details
- **Filtering & Search**: Filter by system, volume group, status, and date range
- **Error Tracking**: Detailed error messages for troubleshooting
- **Performance Metrics**: Track execution duration and timestamps

### Notification System
- **Multi-Channel Support**: Email, Slack, Webhook, and SNMP notifications
- **Alert Rules**: Create rules based on event types and severity levels
- **Event Types**: 
  - Snapshot success/failure/warning
  - System connection lost
  - Scheduler errors
  - Consecutive failures
- **Throttling**: Prevent notification spam with configurable throttle periods
- **Notification History**: Track all sent notifications with status and error details
- **Test Notifications**: Verify channel configuration before activation

### Settings & Configuration
- **Global Settings**: Configure application-wide parameters
- **Retention Policies**: Set default retention days
- **Scheduler Configuration**: Adjust scheduler behavior
- **Security Settings**: Configure authentication and session timeouts
- **Database Settings**: PostgreSQL or SQLite support

### User Management & Security
- **Role-Based Access Control**: Admin and user roles
- **JWT Authentication**: Secure API access with JSON Web Tokens
- **Password Hashing**: bcrypt password hashing
- **Session Management**: Configurable session timeouts
- **CSRF Protection**: Cross-Site Request Forgery protection
- **Security Headers**: HSTS, X-Frame-Options, CSP, and more

### Web Interface
- **Modern React UI**: Responsive, intuitive interface
- **Dashboard**: Overview of systems, schedules, and recent executions
- **Real-time Updates**: Live status updates for operations
- **Dark Mode Support**: (Coming soon)
- **Mobile Responsive**: Works on tablets and mobile devices

---

## 🏗️ Architecture

### Backend
- **Language**: Go 1.25
- **Framework**: Gorilla Mux for routing
- **Database**: SQLite (development) or PostgreSQL (production)
- **Scheduler**: Cron-based with robfig/cron/v3
- **API**: RESTful API with JWT authentication

### Frontend
- **Framework**: React 19.2
- **Build Tool**: Vite 8.0
- **State Management**: TanStack Query (React Query)
- **HTTP Client**: Axios
- **Routing**: React Router v7

### Deployment
- **Containerization**: Podman/Docker support
- **Reverse Proxy**: Nginx with HTTPS
- **Database**: PostgreSQL 16 for production
- **SSL/TLS**: Self-signed or custom certificates

---

## 📦 Installation Methods

### Option 1: Production Deployment (Recommended)

**🚀 One-Liner Quick Start:**

```bash
# Pull latest pre-built images and start the entire stack
./start-prod.sh
```

**What it does:**
- ✅ Pulls latest images from GitHub Container Registry (GHCR)
- ✅ Validates environment configuration (.env file)
- ✅ Starts PostgreSQL, Backend, and Frontend containers
- ✅ Provides helpful status messages and next steps

**Pre-built Container Images Available:**

```bash
# Backend (Go API)
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/backend:latest

# Frontend (React + Nginx)
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/frontend:latest

# Database (PostgreSQL 16)
podman pull docker.io/library/postgres:16-alpine
```

**Supported Architectures:**
- `linux/amd64` - Intel/AMD 64-bit
- `linux/arm64` - ARM 64-bit (Apple Silicon, ARM servers)

**Step-by-Step Setup:**

```bash
# 1. Clone repository
git clone <repository-url>
cd ibm-storage-virtualize-snapshot-manager

# 2. Generate secure keys and certificates
./deploy/setup.sh

# 3. Review and update .env file (optional)
nano .env

# 4. Start the application (pulls images automatically)
./start-prod.sh

# 5. Access at https://localhost
# Default credentials: admin / admin123
```

**Additional Commands:**

```bash
# View logs (follow mode)
./start-prod.sh --logs

# Check container status
./start-prod.sh --status

# Stop all services
./start-prod.sh --stop

# Force rebuild and restart
./start-prod.sh --rebuild
```

**Features:**
- ✅ Production-ready PostgreSQL database
- ✅ HTTPS with Nginx reverse proxy
- ✅ Automated health checks and restart policies
- ✅ Backup and restore scripts included
- ✅ One-command deployment and updates
- ✅ Multi-architecture support (amd64, arm64)
- ✅ Pre-built images from GitHub Container Registry
- ✅ Automatic image pulling and validation
- ✅ Color-coded status messages

**See [QUICK_START_PRODUCTION.md](QUICK_START_PRODUCTION.md) for complete documentation.**

### Option 2: Local Development

```bash
# Clone repository
git clone <repository-url>
cd ibm-storage-virtualize-snapshot-manager

# Start development environment
./dev-start.sh

# Access at http://localhost:5173
```

**Features:**
- ✅ SQLite database for quick setup
- ✅ Hot reload for frontend
- ✅ Automatic migrations
- ✅ Default admin user (admin/admin123)

---

## 🔒 Security Features

### Authentication & Authorization
- JWT-based authentication with configurable expiry
- Role-based access control (Admin/User)
- Secure password hashing with bcrypt
- Token blacklisting for logout

### Data Protection
- AES-256-GCM encryption for storage system credentials
- Encrypted database fields for sensitive data
- Secure key generation utilities
- Environment variable protection

### Network Security
- HTTPS support with TLS 1.2+
- CORS configuration
- Rate limiting on authentication endpoints
- Security headers (HSTS, CSP, X-Frame-Options)
- CSRF protection

### Audit & Compliance
- Complete audit trail of all operations
- User action logging
- Execution history with timestamps
- Failed login attempt tracking

---

## 📊 IBM Storage Virtualize API Integration

### Rate Limit Handling
- **Auth Endpoint**: 3 requests/second (strictly enforced)
- **Command Endpoints**: 10 requests/second
- **Token Management**: Max 4 tokens per cluster
- **Automatic Retry**: Exponential backoff on 429 errors
- **Token Caching**: Reuse tokens until expiry

### API Quirks Handled
- All endpoints use POST method (even for reads)
- Authentication via custom headers (X-Auth-Username, X-Auth-Password)
- Token-based subsequent requests (X-Auth-Token)
- JWT token expiry decoding and management
- HTTPS required on port 7443

---

## 🚀 Performance & Scalability

### Scheduler Performance
- Concurrent snapshot execution
- Efficient cron expression parsing
- Next execution time calculation
- Automatic schedule reload on restart

### Database Optimization
- Indexed queries for fast lookups
- Foreign key constraints with CASCADE delete
- Connection pooling
- Prepared statements

### API Performance
- Efficient JSON serialization
- Minimal database queries
- Caching where appropriate
- Streaming responses for large datasets

---

## 📚 Documentation

### Included Documentation
- **README.md**: Project overview and quick start
- **QUICKSTART.md**: 5-minute setup guide
- **DEVELOPMENT.md**: Development environment setup
- **DEPLOYMENT.md**: Production deployment guide
- **docs/API_NOTIFICATIONS.md**: Notification API documentation
- **docs/NOTIFICATIONS_USER_GUIDE.md**: User guide for notifications
- **docs/SMTP_SERVICES_GUIDE.md**: SMTP configuration guide
- **AGENTS.md**: AI agent development guidelines

### API Documentation
- Complete REST API reference
- Authentication flow
- Request/response examples
- Error codes and handling
- Rate limiting details

---

## 🔧 Configuration

### Environment Variables

**Backend (.env):**
```bash
# Server
PORT=8080
HOST=0.0.0.0

# Database
DB_TYPE=postgres  # or sqlite
DB_HOST=postgres
DB_PORT=5432
DB_NAME=snapshots
DB_USER=snapshots
DB_PASSWORD=<secure-password>

# Security
JWT_SECRET=<base64-encoded-32-bytes>
ENCRYPTION_KEY=<base64-encoded-32-bytes>

# Logging
LOG_LEVEL=info
```

**Frontend (.env):**
```bash
VITE_API_URL=http://localhost:8080/api
```

---

## 🐛 Known Issues & Limitations

### IBM SVC Limitations
- Volume groups can only have ONE snapshot policy (this app works around this)
- Rate limits are strictly enforced (3 auth req/sec, 10 cmd req/sec)
- Max 4 different tokens per cluster
- All API calls must use POST method

### Current Limitations
- No multi-tenancy support (planned for v2.0)
- No Kubernetes deployment (planned for v2.0)
- No Prometheus metrics export (planned for v1.1)
- Email notifications require external SMTP server

---

## 🔄 Upgrade Path

This is the first release, so no upgrade path is needed. Future releases will include:
- Database migration scripts
- Backward compatibility notes
- Breaking changes documentation

---

## 🛠️ Troubleshooting

### Common Issues

**Database Connection Errors:**
- Verify PostgreSQL is running: `podman ps`
- Check credentials in `.env` file
- Ensure database is initialized: `./deploy/setup.sh`

**Authentication Failures:**
- Verify JWT_SECRET is set correctly
- Check token expiry (default 24 hours)
- Ensure user exists in database

**Snapshot Execution Failures:**
- Verify IBM SVC connectivity
- Check rate limits (max 3 auth/sec, 10 cmd/sec)
- Review execution logs in Audit Logs page
- Verify volume group still exists

**Scheduler Not Running:**
- Check backend logs: `podman logs snapshot-manager-backend`
- Verify cron expressions are valid
- Ensure schedules are marked as active

---

## 📈 Roadmap

### v1.1 (Planned)
- [ ] Prometheus metrics export
- [ ] Grafana dashboard templates
- [ ] Enhanced notification templates
- [ ] Snapshot cleanup automation
- [ ] Dark mode UI theme

### v2.0 (Future)
- [ ] Kubernetes deployment
- [ ] Advanced RBAC
- [ ] REST API v2 with GraphQL

---

## 🤝 Contributing

We welcome contributions! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

See [DEVELOPMENT.md](DEVELOPMENT.md) for development setup.

---

## 📄 License

Apache License 2.0 - See [LICENSE](LICENSE) file for details.

---

## ⚠️ Important Legal Notices

**NO WARRANTY**: This software is provided "AS IS", WITHOUT WARRANTY OF ANY KIND, express or implied, including but not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement.

**USE AT YOUR OWN RISK**: This application interacts with IBM Storage Virtualize systems and manages snapshot operations. Users are solely responsible for:
- Testing thoroughly in non-production environments
- Ensuring proper backup and disaster recovery procedures
- Verifying snapshot operations and retention policies
- Understanding the impact on their storage systems
- Compliance with organizational policies

**NOT AN IBM PRODUCT**: This is an independent, open-source project. IBM, IBM Storage, IBM Storage Virtualize, and FlashSystem are trademarks of IBM Corporation. This project is not affiliated with, endorsed by, or supported by IBM Corporation.

**COMMUNITY SUPPORT ONLY**: Support is provided on a best-effort basis by the community. No SLAs or guaranteed response times.

---

## 🙏 Acknowledgments

- IBM Storage Virtualize REST API documentation
- Go community for excellent libraries
- React and Vite teams for modern web tooling
- All contributors and testers

---

## 📞 Support

For issues and questions:
- **GitHub Issues**: [Create an issue](https://github.com/your-repo/issues)
- **Documentation**: Check README.md and docs/ folder
- **Community**: Join discussions on GitHub

---

**Made with ❤️ by the community**

**Version:** 1.0.0  
**Release Date:** June 17, 2026  
**Go Version:** 1.25  
**React Version:** 19.2