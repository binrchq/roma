# ROMA Deployment Guide

This document provides detailed instructions for deploying ROMA in various environments.

[中文版本](./DEPLOYMENT_CN.md)

---

## Deployment Options

- [Docker Compose (Recommended)](#docker-compose-deployment)
- [Kubernetes](#kubernetes-deployment)
- [Binary](#binary-deployment)
- [Production Configuration](#production-configuration)

---

## Docker Compose Deployment

### Quick Start (SQLite)

Simplest deployment, suitable for testing and small environments:

```bash
# 1. Download configuration
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. Start services
docker compose -f quickstart.yaml up -d

# 3. View logs
docker compose -f quickstart.yaml logs -f

# 4. Access services
# Web UI: http://localhost:7000
# API: http://localhost:6999
# SSH: localhost:2200
```

### MySQL Deployment

Suitable for medium-sized and production environments:

```bash
# 1. Clone repository
git clone https://github.com/binrchq/roma.git
cd roma/deployment

# 2. Start services
docker compose -f quickstart.mysql.yaml up -d

# 3. Check service status
docker compose -f quickstart.mysql.yaml ps
```

**Configuration file location:** `deployment/quickstart.mysql.yaml`

### PostgreSQL Deployment

```bash
docker compose -f quickstart.pgsql.yaml up -d
```

### Custom Configuration

Create `.env` file to override defaults:

```bash
# .env
TAG=latest
WEB_PORT=8080
ROMA_SSH_PORT=2200
ROMA_API_PORT=6999
ROMA_USER_1ST_USERNAME=admin
ROMA_USER_1ST_PASSWORD=YourStrongPassword123!
ROMA_USER_1ST_EMAIL=admin@example.com
```

Start with environment variables:

```bash
docker compose -f quickstart.yaml up -d
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (version ≥ 1.20)
- kubectl tool
- Helm 3 (optional)

### Using Helm (Recommended)

```bash
# 1. Add Helm repository
helm repo add roma https://charts.binrc.com
helm repo update

# 2. Install
helm install roma roma/roma \
  --namespace roma \
  --create-namespace \
  --set image.tag=latest \
  --set database.type=mysql

# 3. Check deployment
kubectl get pods -n roma
```

### Using YAML Manifests

```bash
# 1. Clone repository
git clone https://github.com/binrchq/roma.git
cd roma/deployment/k8s

# 2. Modify configuration
vim roma-configmap.yaml
vim roma-secret.yaml

# 3. Deploy
kubectl apply -f namespace.yaml
kubectl apply -f roma-configmap.yaml
kubectl apply -f roma-secret.yaml
kubectl apply -f roma-deployment.yaml
kubectl apply -f roma-service.yaml

# 4. Check status
kubectl get all -n roma
```

### Ingress Configuration

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: roma-ingress
  namespace: roma
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - roma.example.com
    secretName: roma-tls
  rules:
  - host: roma.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: roma-web
            port:
              number: 80
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: roma-api
            port:
              number: 6999
```

---

## Binary Deployment

### Build

```bash
# 1. Clone repository
git clone https://github.com/binrchq/roma.git
cd roma

# 2. Install dependencies
go mod download

# 3. Build
go build -o roma cmd/roma/main.go

# Or use Makefile
make build
```

### Configuration

Create configuration file `configs/config.toml`:

```toml
[api]
host = '0.0.0.0'
port = '6999'

[common]
port = '2200'
prompt = 'roma'

[database]
type = 'mysql'  # sqlite, mysql, postgresql
cdb_url = 'user:password@tcp(localhost:3306)/roma?charset=utf8mb4&parseTime=True&loc=Local'

[security]
jwt_secret = 'your-jwt-secret-change-me'
encryption_key = 'your-32-byte-encryption-key-here'

[apikey]
prefix = 'apikey.'
key = 'your-api-key-here'

[user_1st]
username = 'admin'
email = 'admin@example.com'
password = 'ChangeMe123!'
public_key = '''
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC...
'''
roles = "super,system,ops"

[rate_limit]
enabled = true
requests_per_second = 100
burst = 200

[ip_blacklist]
enabled = true
auto_ban_threshold = 5
```

### Start Service

```bash
# Run directly
./roma -c configs/config.toml

# Using systemd (recommended for production)
sudo cp roma /usr/local/bin/
sudo cp deployment/roma.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable roma
sudo systemctl start roma
sudo systemctl status roma
```

### Systemd Service File

Create `/etc/systemd/system/roma.service`:

```ini
[Unit]
Description=ROMA Jump Server
After=network.target

[Service]
Type=simple
User=roma
Group=roma
WorkingDirectory=/usr/local/roma
ExecStart=/usr/local/bin/roma -c /usr/local/roma/configs/config.toml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

---

## Production Configuration

### Database

#### MySQL Configuration

```toml
[database]
type = 'mysql'
cdb_url = 'roma:SecurePassword@tcp(mysql-host:3306)/roma?charset=utf8mb4&parseTime=True&loc=Local'
max_idle_conns = 10
max_open_conns = 100
conn_max_lifetime = 3600
```

#### PostgreSQL Configuration

```toml
[database]
type = 'postgresql'
cdb_url = 'host=postgres-host port=5432 user=roma password=SecurePassword dbname=roma sslmode=require'
```

### Security Configuration

```toml
[security]
# JWT secret (32+ bytes)
jwt_secret = 'change-this-to-a-secure-random-string-32-bytes-or-more'

# AES encryption key (32 bytes)
encryption_key = '12345678901234567890123456789012'

# SSH host key
ssh_host_key_path = '/usr/local/roma/keys/id_rsa'

[rate_limit]
enabled = true
requests_per_second = 100
burst = 200
per_ip_concurrent_limit = 10

[ip_blacklist]
enabled = true
auto_ban_threshold = 5
auto_ban_duration = 3600
```

### HTTPS/TLS Configuration

Using Nginx as reverse proxy:

```nginx
server {
    listen 443 ssl http2;
    server_name roma.example.com;

    ssl_certificate /etc/letsencrypt/live/roma.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/roma.example.com/privkey.pem;

    # Web UI
    location / {
        proxy_pass http://localhost:7000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # API
    location /api {
        proxy_pass http://localhost:6999;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

### Backup Strategy

#### Database Backup

```bash
#!/bin/bash
# backup-roma.sh

BACKUP_DIR="/backup/roma"
DATE=$(date +%Y%m%d_%H%M%S)

# MySQL backup
mysqldump -u roma -p roma > "$BACKUP_DIR/roma_$DATE.sql"

# Compress backup
gzip "$BACKUP_DIR/roma_$DATE.sql"

# Keep last 7 days backups
find "$BACKUP_DIR" -name "roma_*.sql.gz" -mtime +7 -delete
```

Add to crontab:

```bash
# Backup at 2 AM daily
0 2 * * * /usr/local/bin/backup-roma.sh
```

---

## Upgrade

### Docker Upgrade

```bash
# 1. Pull latest images
docker compose -f quickstart.yaml pull

# 2. Restart services
docker compose -f quickstart.yaml up -d

# 3. Clean old images
docker image prune -f
```

### Binary Upgrade

```bash
# 1. Backup current version
cp /usr/local/bin/roma /usr/local/bin/roma.backup

# 2. Download new version
wget https://github.com/binrchq/roma/releases/download/v1.x.x/roma-linux-amd64
chmod +x roma-linux-amd64

# 3. Replace binary
sudo mv roma-linux-amd64 /usr/local/bin/roma

# 4. Restart service
sudo systemctl restart roma

# 5. Check version
roma --version
```

---

## Troubleshooting

### Service Won't Start

```bash
# Check logs
docker compose -f quickstart.yaml logs roma

# Or systemd
sudo journalctl -u roma -f

# Common issues:
# 1. Port occupied
sudo netstat -tlnp | grep -E '6999|2200|7000'

# 2. Database connection failed
# Check database configuration

# 3. Permission issues
sudo chown -R roma:roma /usr/local/roma
```

### SSH Connection Failed

```bash
# Test SSH connection
ssh -vvv user@roma-server -p 2200

# Check SSH host key
ls -la /usr/local/roma/keys/

# Regenerate host key
ssh-keygen -t rsa -b 4096 -f /usr/local/roma/keys/id_rsa
```

---

## Security Checklist

- [ ] Change default password
- [ ] Configure strong JWT secret and encryption key
- [ ] Enable HTTPS/TLS
- [ ] Configure firewall rules
- [ ] Enable rate limiting
- [ ] Enable IP blacklist
- [ ] Configure regular backups
- [ ] Set up log monitoring
- [ ] Regular version updates
- [ ] Review user permissions

---

## Related Documentation

- [Security Guide](SECURITY.md)
- [Development Guide](DEVELOPMENT.md)
- [API Documentation](API.md)
