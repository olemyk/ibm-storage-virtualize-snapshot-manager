# Container Usage Guide

## Overview

The IBM Storage Virtualize Snapshot Manager is available as pre-built container images for easy deployment. This guide covers pulling, running, and managing the containerized application.

## Container Images

### Available Images

**Backend:**
```
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:1.2.3
```

**Frontend:**
```
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest
ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:1.2.3
```

### Supported Architectures

- `linux/amd64` - Intel/AMD 64-bit (x86_64)
- `linux/arm64` - ARM 64-bit (Apple Silicon, ARM servers)

Docker/Podman automatically pulls the correct architecture for your system.

## Quick Start

### Using Podman Compose (Recommended)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/[org]/ibm-storage-virtualize-snapshot-manager.git
   cd ibm-storage-virtualize-snapshot-manager
   ```

2. **Run setup script:**
   ```bash
   ./deploy/setup.sh
   ```
   This generates encryption keys and SSL certificates.

3. **Edit `.env` file:**
   ```bash
   nano .env
   ```
   Set required variables (see Configuration section).

4. **Start the application:**
   ```bash
   ./deploy/start.sh
   ```

5. **Access the application:**
   - HTTPS: https://localhost
   - HTTP: http://localhost

### Using Docker Compose

Replace `podman-compose` with `docker-compose` in the commands above:

```bash
docker-compose -f podman-compose.yml up -d
```

## Pulling Images

### Latest Version

```bash
# Backend
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest

# Frontend
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest
```

### Specific Version

```bash
# Backend
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:1.2.3

# Frontend
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:1.2.3
```

### Specific Architecture

```bash
# ARM64 (Apple Silicon)
docker pull --platform linux/arm64 ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest

# AMD64 (Intel/AMD)
docker pull --platform linux/amd64 ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

## Running Containers

### Backend Container

**Minimal run:**
```bash
docker run -d \
  --name snapshot-manager-backend \
  -p 8080:8080 \
  -e DB_TYPE=sqlite \
  -e JWT_SECRET=your-secret-here \
  -e ENCRYPTION_KEY=your-key-here \
  -v $(pwd)/data:/app/data \
  ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

**With PostgreSQL:**
```bash
docker run -d \
  --name snapshot-manager-backend \
  -p 8080:8080 \
  -e DB_TYPE=postgres \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_NAME=snapshots \
  -e DB_USER=snapshots \
  -e DB_PASSWORD=secure-password \
  -e JWT_SECRET=your-secret-here \
  -e ENCRYPTION_KEY=your-key-here \
  --network snapshot-manager-net \
  ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

### Frontend Container

```bash
docker run -d \
  --name snapshot-manager-frontend \
  -p 80:80 \
  -p 443:443 \
  -v $(pwd)/ssl:/etc/nginx/ssl:ro \
  ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest
```

## Configuration

### Environment Variables

#### Backend Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `JWT_SECRET` | Secret for JWT token signing | `your-random-secret-here` |
| `ENCRYPTION_KEY` | 32-byte key for encrypting passwords | `base64-encoded-key` |

#### Backend Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `DB_TYPE` | `sqlite` | Database type (`sqlite` or `postgres`) |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_NAME` | `snapshots` | Database name |
| `DB_USER` | `snapshots` | Database user |
| `DB_PASSWORD` | - | Database password |
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `LOG_FILE` | `/app/logs/app.log` | Log file path |
| `ALLOWED_ORIGINS` | `*` | CORS allowed origins |

#### Frontend Variables

The frontend container uses Nginx and doesn't require environment variables. Configuration is done through:
- SSL certificates (mounted at `/etc/nginx/ssl`)
- Nginx configuration (mounted at `/etc/nginx/nginx.conf`)

### Generating Secrets

**JWT Secret:**
```bash
openssl rand -base64 32
```

**Encryption Key:**
```bash
openssl rand -base64 32
```

Or use the provided script:
```bash
cd backend
go run scripts/generate_keys.go
```

## Volume Mounts

### Backend Volumes

| Container Path | Purpose | Recommended Host Path |
|----------------|---------|----------------------|
| `/app/data` | SQLite database (if used) | `./data` |
| `/app/logs` | Application logs | `./logs` |

### Frontend Volumes

| Container Path | Purpose | Recommended Host Path |
|----------------|---------|----------------------|
| `/etc/nginx/ssl` | SSL certificates | `./ssl` |
| `/etc/nginx/nginx.conf` | Nginx config (optional) | `./nginx.conf` |

## Networking

### Creating a Network

```bash
docker network create snapshot-manager-net
```

### Connecting Containers

All containers should be on the same network:

```bash
docker run --network snapshot-manager-net ...
```

## Health Checks

### Backend Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

### Frontend Health Check

```bash
curl http://localhost:80/health
```

Expected response: HTTP 200

### Container Health Status

```bash
docker ps --format "table {{.Names}}\t{{.Status}}"
```

## Upgrading

### Using Compose

```bash
# Pull latest images
docker-compose -f podman-compose.yml pull

# Restart services
docker-compose -f podman-compose.yml up -d
```

### Manual Upgrade

```bash
# Stop containers
docker stop snapshot-manager-backend snapshot-manager-frontend

# Remove old containers
docker rm snapshot-manager-backend snapshot-manager-frontend

# Pull new images
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
docker pull ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/frontend:latest

# Start new containers
docker run ... (use same commands as before)
```

### Backup Before Upgrade

```bash
# Backup database
./deploy/backup.sh

# Or manually
docker exec snapshot-manager-db pg_dump -U snapshots snapshots > backup.sql
```

## Troubleshooting

### Container Won't Start

**Check logs:**
```bash
docker logs snapshot-manager-backend
docker logs snapshot-manager-frontend
```

**Check environment variables:**
```bash
docker inspect snapshot-manager-backend | grep -A 20 Env
```

**Verify network:**
```bash
docker network inspect snapshot-manager-net
```

### Database Connection Issues

**Test PostgreSQL connection:**
```bash
docker exec snapshot-manager-backend nc -zv postgres 5432
```

**Check PostgreSQL logs:**
```bash
docker logs snapshot-manager-db
```

### Permission Issues

**Fix volume permissions:**
```bash
sudo chown -R 1000:1000 ./data ./logs
```

**Check container user:**
```bash
docker exec snapshot-manager-backend id
```

### SSL Certificate Issues

**Generate self-signed certificate:**
```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/nginx-selfsigned.key \
  -out ssl/nginx-selfsigned.crt \
  -subj "/CN=localhost"
```

**Verify certificate:**
```bash
openssl x509 -in ssl/nginx-selfsigned.crt -text -noout
```

## Performance Tuning

### Resource Limits

**Set CPU and memory limits:**
```bash
docker run -d \
  --cpus="2" \
  --memory="2g" \
  --name snapshot-manager-backend \
  ...
```

**In compose file:**
```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

### Database Performance

**PostgreSQL tuning:**
```yaml
postgres:
  environment:
    POSTGRES_SHARED_BUFFERS: 256MB
    POSTGRES_EFFECTIVE_CACHE_SIZE: 1GB
    POSTGRES_MAX_CONNECTIONS: 100
```

## Security

### Running as Non-Root

Both images run as non-root user (UID 1000) by default.

**Verify:**
```bash
docker exec snapshot-manager-backend whoami
# Output: appuser
```

### Read-Only Filesystem

**Enable read-only mode:**
```bash
docker run -d \
  --read-only \
  --tmpfs /tmp \
  --tmpfs /app/logs \
  ...
```

### Security Scanning

**Scan images for vulnerabilities:**
```bash
docker scan ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

## Monitoring

### Container Stats

```bash
docker stats snapshot-manager-backend snapshot-manager-frontend
```

### Logs

**Follow logs:**
```bash
docker logs -f snapshot-manager-backend
```

**Last 100 lines:**
```bash
docker logs --tail 100 snapshot-manager-backend
```

**Since timestamp:**
```bash
docker logs --since 2026-06-17T10:00:00 snapshot-manager-backend
```

## Backup and Restore

### Backup

**Using provided script:**
```bash
./deploy/backup.sh
```

**Manual PostgreSQL backup:**
```bash
docker exec snapshot-manager-db pg_dump -U snapshots snapshots > backup-$(date +%Y%m%d).sql
```

**Manual SQLite backup:**
```bash
docker cp snapshot-manager-backend:/app/data/snapshots.db ./backup-$(date +%Y%m%d).db
```

### Restore

**PostgreSQL:**
```bash
docker exec -i snapshot-manager-db psql -U snapshots snapshots < backup.sql
```

**SQLite:**
```bash
docker cp backup.db snapshot-manager-backend:/app/data/snapshots.db
docker restart snapshot-manager-backend
```

## Multi-Architecture Deployment

### Building for Specific Architecture

```bash
docker buildx build \
  --platform linux/arm64 \
  -t snapshot-manager-backend:arm64 \
  ./backend
```

### Running on ARM (Apple Silicon)

No special configuration needed. Docker automatically pulls ARM64 images:

```bash
docker run ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
# Automatically uses linux/arm64 variant
```

### Cross-Platform Testing

```bash
# Test AMD64 on ARM (using QEMU)
docker run --platform linux/amd64 \
  ghcr.io/[org]/ibm-storage-virtualize-snapshot-manager/backend:latest
```

## Production Deployment

### Recommended Setup

1. **Use PostgreSQL** (not SQLite) for production
2. **Enable HTTPS** with valid SSL certificates
3. **Set resource limits** for containers
4. **Enable health checks** in compose file
5. **Configure log rotation** for container logs
6. **Set up monitoring** (Prometheus, Grafana)
7. **Regular backups** (automated daily)
8. **Use secrets management** (Docker secrets, Vault)

### Example Production Compose

See [`podman-compose.yml`](../podman-compose.yml) for a production-ready configuration.

## Additional Resources

- [Deployment Guide](../DEPLOYMENT.md)
- [CI/CD Documentation](CI_CD.md)
- [Docker Documentation](https://docs.docker.com/)
- [Podman Documentation](https://docs.podman.io/)

## Support

For container-related issues:
1. Check container logs
2. Verify environment variables
3. Test network connectivity
4. Review this documentation
5. Create a GitHub issue with:
   - Container version/tag
   - Docker/Podman version
   - Error messages
   - Steps to reproduce

---

*Last updated: 2026-06-17*