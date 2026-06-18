# Quick Start - Production Deployment

## One-Liner Startup Script

This guide provides a simple way to deploy the IBM Storage Virtualize Snapshot Manager using pre-built images from GitHub Container Registry.

### Prerequisites

- `podman` and `podman-compose` installed
- `.env` file configured (see below)

### Quick Start Command

**Option 1: Standalone Script (Recommended)**

Download and run the script - it will automatically create an installation directory, fetch all required files, and guide you through setup:

```bash
# Download the script
curl -fsSL https://raw.githubusercontent.com/olemyk/ibm-storage-virtualize-snapshot-manager/main/start-prod.sh -o start-prod.sh
chmod +x start-prod.sh

# Run it (interactive setup)
./start-prod.sh
```

**What happens:**
- Creates `ibm-virtualize-snapshot-manager-dir/` directory
- Downloads all required files into this directory
- All configuration and data stored in this directory
- Easy to backup, move, or delete entire installation

**Interactive Setup Process:**

The script will guide you through:

1. **Download Dependencies** (automatic)
   - Downloads `podman-compose.prod.yml` if missing
   - Downloads `backend/scripts/postgres-init.sql` if missing
   - Generates self-signed SSL certificates if missing

2. **Environment Configuration** (interactive)
   - Prompts to generate `.env` file with secure random keys
   - Auto-generates: `DB_PASSWORD`, `JWT_SECRET`, `ENCRYPTION_KEY`
   - Asks for admin username (default: admin)
   - Asks for admin password (default: admin123)
   - Detects server IP and configures CORS automatically

3. **Start Services** (automatic)
   - Pulls latest container images from GHCR
   - Starts PostgreSQL, Backend, and Frontend
   - Creates admin user with your credentials
   - Displays access URLs and credentials

**Directory Structure:**

All files are organized in a dedicated directory:

```
./
├── start-prod.sh                                    # Startup script (you download this)
└── ibm-virtualize-snapshot-manager-dir/             # Installation directory (auto-created)
    ├── podman-compose.prod.yml                      # Downloaded automatically
    ├── .env                                         # Generated or downloaded
    ├── backend/
    │   └── scripts/
    │       └── postgres-init.sql                    # Downloaded automatically
    └── ssl/
        ├── nginx-selfsigned.crt                     # Generated automatically
        └── nginx-selfsigned.key                     # Generated automatically
```

**Benefits:**
- ✅ All files in one directory - easy to backup
- ✅ Easy to move installation to another location
- ✅ Clean uninstall - just delete the directory
- ✅ Multiple installations possible (different directories)
- ✅ Persistent across restarts - all configuration preserved

**Example Interactive Session:**

```
ℹ Checking prerequisites...
ℹ Creating installation directory: ibm-virtualize-snapshot-manager-dir
ℹ Working in directory: /home/user/ibm-virtualize-snapshot-manager-dir
✓ Prerequisites check passed
⚠ .env file not found.

? Generate .env file with secure random keys? (Y/n): Y
ℹ Generating .env file with secure defaults...

ℹ Set admin user credentials:
? Admin username (default: admin): admin
? Admin password (default: admin123): MySecurePassword123

ℹ CORS Configuration:
ℹ Detected server IP: 10.33.3.104
? Use this IP for CORS? (Y/n): Y

✓ .env file created successfully
ℹ Configuration summary:
  - Database password: [HIDDEN]
  - Admin username: admin
  - Admin password: [HIDDEN]
  - Server IP: 10.33.3.104

ℹ Pulling latest images from GitHub Container Registry...
✓ Images pulled successfully
ℹ Starting IBM Storage Virtualize Snapshot Manager...
✓ Admin user created successfully
✓ Stack started successfully

ℹ Services:
  - Frontend (HTTPS): https://10.33.3.104
  - Admin username: admin
  - Admin password: [as configured in .env]
```

**What the script does automatically:**
- ✅ Downloads all required files
- ✅ Generates secure random keys (32-byte)
- ✅ Configures CORS with your server IP
- ✅ Creates admin user in database
- ✅ Pulls latest container images from GHCR
- ✅ Starts the entire stack
- ✅ Verifies services are healthy

**Option 2: Full Repository Clone**

```bash
# 1. Clone the repository
git clone https://github.com/olemyk/ibm-storage-virtualize-snapshot-manager.git
cd ibm-storage-virtualize-snapshot-manager

# 2. Create .env file (or run setup script)
cp .env.example .env
# Edit .env and set DB_PASSWORD, JWT_SECRET, ENCRYPTION_KEY

# 3. Run the startup script
./start-prod.sh
```

### Manual Deployment Steps

If you prefer to run commands manually:

```bash
# 1. Pull latest images
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/backend:latest
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/frontend:latest
podman pull docker.io/library/postgres:16-alpine

# 2. Start the stack
podman-compose -f podman-compose.prod.yml up -d

# 3. Check status
podman-compose -f podman-compose.prod.yml ps

# 4. View logs
podman-compose -f podman-compose.prod.yml logs -f
```

### Environment Configuration

Create a `.env` file with the following required variables:

```bash
# Database
DB_PASSWORD=your_secure_password_here

# Security (generate with: openssl rand -base64 32)
JWT_SECRET=your_jwt_secret_here
ENCRYPTION_KEY=your_encryption_key_here

# Optional: Custom ports
# FRONTEND_HTTP_PORT=80
# FRONTEND_HTTPS_PORT=443
```

### Quick Setup Script

Use the provided setup script to auto-generate secure keys:

```bash
./deploy/setup.sh
```

### Access the Application

- **HTTPS**: https://localhost (port 443)
- **Credentials**: As configured during setup (default: admin / admin123)

### Managing the Stack

```bash
# Stop services
./start-prod.sh --stop

# View logs
./start-prod.sh --logs

# Check status
./start-prod.sh --status

# Update to latest version
./start-prod.sh --rebuild

# Complete cleanup (removes containers, volumes, and images)
./start-prod.sh --clean

# Enable auto-start on boot (systemd user service)
./start-prod.sh --autostart

# Disable auto-start
./start-prod.sh --remove-autostart
```

### Auto-Start on Boot (Non-Root Users)

The application can be configured to start automatically on boot using systemd user services. This works for non-root users and doesn't require sudo privileges.

**Setup Auto-Start:**
```bash
./start-prod.sh --autostart
```

This will:
- Create a systemd user service at `~/.config/systemd/user/snapshot-manager.service`
- Enable user linger (allows service to run when user is not logged in)
- Enable the service to start on boot

**Manual Service Control:**
```bash
# Start service
systemctl --user start snapshot-manager

# Stop service
systemctl --user stop snapshot-manager

# Check status
systemctl --user status snapshot-manager

# View logs
journalctl --user -u snapshot-manager -f

# Disable auto-start
systemctl --user disable snapshot-manager
# Or use: ./start-prod.sh --remove-autostart
```

**Requirements:**
- systemd-based Linux distribution
- User linger enabled (script does this automatically)
- Podman and podman-compose installed

**Note:** The service runs in the user's context, so it will start when the system boots, even if the user is not logged in (thanks to linger).

### Manual Commands

```bash
# Stop services
podman-compose -f podman-compose.prod.yml down

# Update to latest version
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/backend:latest
podman pull ghcr.io/olemyk/ibm-storage-virtualize-snapshot-manager/frontend:latest
podman-compose -f podman-compose.prod.yml up -d --force-recreate
```

### Troubleshooting

**Login Issues (CORS or Password Problems):**

If you can't login, run the automated fix script from the installation directory:

```bash
# Navigate to installation directory
cd ibm-virtualize-snapshot-manager-dir

# Auto-detect server IP and fix both CORS and password
../fix-login-issues.sh

# Or specify server IP manually
../fix-login-issues.sh 10.33.3.104

# Or specify custom password
../fix-login-issues.sh 10.33.3.104 MyNewPassword123
```

**Note:** The fix scripts need to be in the same directory as start-prod.sh (parent directory of installation folder).

This script will:
- ✅ Add your server IP to ALLOWED_ORIGINS (fixes CORS errors)
- ✅ Restart backend container with new configuration
- ✅ Reset admin password to "admin123" (or custom password)
- ✅ Verify the configuration

**Manual CORS Fix:**

If you're accessing from a different IP address (e.g., 10.33.3.104):

```bash
# Stop services
./start-prod.sh --stop

# Add your server IP to .env file
echo "ALLOWED_ORIGINS=http://localhost,https://localhost,http://10.33.3.104,https://10.33.3.104,http://127.0.0.1,https://127.0.0.1" >> .env

# Restart
./start-prod.sh
```

**Password Reset Only:**

```bash
# Reset to default password (admin123)
./reset-admin-password.sh

# Or set custom password
./reset-admin-password.sh MyNewPassword123
```

**Check container status:**
```bash
podman ps -a
```

**View logs:**
```bash
podman logs snapshot-manager-backend
podman logs snapshot-manager-frontend
podman logs snapshot-manager-db
```

**Check CORS configuration:**
```bash
podman exec snapshot-manager-backend env | grep ALLOWED_ORIGINS
```

**Reset everything:**
```bash
podman-compose -f podman-compose.prod.yml down -v
podman-compose -f podman-compose.prod.yml up -d
```

### Backup and Restore

Since all files are in one directory, backup and restore is simple:

**Backup:**
```bash
# Stop services first
./start-prod.sh --stop

# Backup entire installation directory
tar -czf snapshot-manager-backup-$(date +%Y%m%d).tar.gz ibm-virtualize-snapshot-manager-dir/

# Restart services
./start-prod.sh
```

**Restore:**
```bash
# Extract backup
tar -xzf snapshot-manager-backup-20260618.tar.gz

# Start services
./start-prod.sh
```

**What's included in backup:**
- ✅ All configuration (.env file)
- ✅ Database data (PostgreSQL volume)
- ✅ SSL certificates
- ✅ All downloaded files
- ✅ Complete working installation

**Move to another server:**
```bash
# On old server - create backup
./start-prod.sh --stop
tar -czf snapshot-manager.tar.gz ibm-virtualize-snapshot-manager-dir/

# Transfer to new server
scp snapshot-manager.tar.gz user@newserver:~

# On new server - restore
tar -xzf snapshot-manager.tar.gz
cd ibm-virtualize-snapshot-manager-dir
../start-prod.sh
```

### Architecture

- **Frontend**: Nginx + React (HTTPS port 443 only)
- **Backend**: Go API (internal port 8080)
- **Database**: PostgreSQL 16 (internal port 5432)
- **Network**: Bridge network (snapshot-manager-net)
- **Installation**: All files in `ibm-virtualize-snapshot-manager-dir/`

### Security Notes

1. Change default admin password after first login
2. Use strong, unique values for JWT_SECRET and ENCRYPTION_KEY
3. Replace self-signed SSL certificates with proper certs for production
4. Keep database password secure (minimum 16 characters recommended)
5. Never commit `.env` file to version control
