# Deployment Guide

## Quick Start

### Docker (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd weight-tracker

# Build and run with Docker Compose
docker-compose up -d

# Access the application
open http://localhost:8080
```

### Local Development

```bash
# Install Go 1.22+
# Clone the repository
git clone <repository-url>
cd weight-tracker

# Install dependencies
go mod download

# Run the application
make start

# Or build and run manually
go build -o bin/weight-tracker ./cmd/server
./bin/weight-tracker
```

## OpenMediaVault Deployment

### Prerequisites

- OpenMediaVault 6.x or later
- Docker plugin installed and enabled
- SSH access to your OMV system

### Installation Steps

1. **SSH into OMV**
```bash
ssh root@<your-omv-ip>
```

2. **Clone and navigate to project**
```bash
cd /opt
git clone <repository-url> weight-tracker
cd weight-tracker
```

3. **Set up data directory**
```bash
# Create persistent data directory
mkdir -p /srv/dev-disk-by-uuid-<disk-uuid>/docker-data/weight-tracker

# Set permissions
chmod 755 /srv/dev-disk-by-uuid-<disk-uuid>/docker-data/weight-tracker
```

4. **Create production docker-compose.yml**
```yaml
version: '3.8'

services:
  weight-tracker:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - /srv/dev-disk-by-uuid-<disk-uuid>/docker-data/weight-tracker:/home/appuser/data
    environment:
      - DB_PATH=/home/appuser/data/weights.db
      - PORT=8080
      - ENV=production
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - weight-tracker-network

  # Optional: Reverse proxy
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - /srv/dev-disk-by-uuid-<disk-uuid>/docker-data/weight-tracker/nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - weight-tracker
    restart: unless-stopped
    networks:
      - weight-tracker-network

networks:
  weight-tracker-network:
    driver: bridge
```

5. **Deploy**
```bash
# Build and start
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f weight-tracker
```

6. **Access**
- Local: `http://<your-omv-ip>:8080`
- With Nginx: `http://<your-domain>`

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database file path (default: ./data/weights.db)
- `ENV`: Environment mode (development/production)

### Database Backups

#### Manual Backup
```bash
# Backup SQLite database
cp /path/to/data/weights.db /path/to/backups/weights_backup_$(date +%Y%m%d_%H%M%S).db
```

#### Automated Backup (Cron)
```bash
# Add to crontab
crontab -e

# Daily backup at 2 AM
0 2 * * * cp /path/to/data/weights.db /path/to/backups/weights_backup_$(date +\%Y\%m\%d_\%H\%M\%S).db
```

## SSL/HTTPS Setup

### With Let's Encrypt

1. **Install certbot on OMV**
```bash
apt update
apt install certbot
```

2. **Generate certificates**
```bash
certbot certonly --standalone -d your-domain.com
```

3. **Configure Nginx**
```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    location / {
        proxy_pass http://weight-tracker:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Monitoring

### Health Check
```bash
# Check application health
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "database": "healthy",
  "version": "1.0.0"
}
```

### Docker Monitoring
```bash
# Check container status
docker ps

# View resource usage
docker stats

# View logs
docker logs weight-tracker
```

## Troubleshooting

### Common Issues

1. **Database Permission Error**
```bash
# Fix ownership
chown -R appuser:appuser /path/to/data
```

2. **Port Already in Use**
```bash
# Check what's using the port
netstat -tulpn | grep :8080
# Change port in docker-compose.yml
```

3. **Migration Issues**
```bash
# Remove database to recreate
rm /path/to/data/weights.db
# Restart container
docker-compose restart weight-tracker
```

### Logs

```bash
# Application logs
docker-compose logs weight-tracker

# Nginx logs (if using)
docker-compose logs nginx

# System logs
journalctl -u docker.service
```

## Performance Tuning

### Database Optimization
- Ensure proper indexes exist (included in schema)
- Regular maintenance: `VACUUM` and `ANALYZE`
- Monitor database size growth

### Resource Limits
```yaml
# Add to docker-compose.yml
services:
  weight-tracker:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 256M
        reservations:
          cpus: '0.25'
          memory: 128M
```

## Security

- Run as non-root user (configured in Dockerfile)
- Use HTTPS in production
- Regular updates of base images
- Monitor for security advisories
- Backup data regularly

## Support

For issues and questions:
1. Check the logs for error messages
2. Verify configuration
3. Test with a fresh database
4. Check system resources (memory, disk space)