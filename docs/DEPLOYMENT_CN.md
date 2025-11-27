# ROMA 部署指南

本文档详细介绍ROMA的各种部署方式。

---

## 📦 部署方式

- [Docker Compose (推荐)](#docker-compose-部署)
- [Kubernetes](#kubernetes-部署)
- [二进制部署](#二进制部署)
- [生产环境配置](#生产环境配置)

---

## Docker Compose 部署

### 快速启动 (SQLite)

最简单的部署方式，适合测试和小型环境：

```bash
# 1. 下载配置文件
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. 启动服务
docker compose -f quickstart.yaml up -d

# 3. 查看日志
docker compose -f quickstart.yaml logs -f

# 4. 访问服务
# Web UI: http://localhost:7000
# API: http://localhost:6999
# SSH: localhost:2200
```

### MySQL 部署

适合中型环境和生产环境：

```bash
# 1. 克隆仓库
git clone https://github.com/binrchq/roma.git
cd roma/deployment

# 2. 启动服务
docker compose -f quickstart.mysql.yaml up -d

# 3. 检查服务状态
docker compose -f quickstart.mysql.yaml ps
```

**配置文件位置:** `deployment/quickstart.mysql.yaml`

### PostgreSQL 部署

```bash
docker compose -f quickstart.pgsql.yaml up -d
```

### 自定义配置

创建 `.env` 文件覆盖默认配置：

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

启动时会自动加载环境变量：

```bash
docker compose -f quickstart.yaml up -d
```

---

## Kubernetes 部署

### 前置要求

- Kubernetes 集群 (版本 ≥ 1.20)
- kubectl 工具
- Helm 3 (可选)

### 使用 Helm (推荐)

```bash
# 1. 添加Helm仓库
helm repo add roma https://charts.binrc.com
helm repo update

# 2. 安装
helm install roma roma/roma \
  --namespace roma \
  --create-namespace \
  --set image.tag=latest \
  --set database.type=mysql

# 3. 检查部署
kubectl get pods -n roma
```

### 使用 YAML 清单

```bash
# 1. 克隆仓库
git clone https://github.com/binrchq/roma.git
cd roma/deployment/k8s

# 2. 修改配置
vim roma-configmap.yaml
vim roma-secret.yaml

# 3. 部署
kubectl apply -f namespace.yaml
kubectl apply -f roma-configmap.yaml
kubectl apply -f roma-secret.yaml
kubectl apply -f roma-deployment.yaml
kubectl apply -f roma-service.yaml

# 4. 检查状态
kubectl get all -n roma
```

### Ingress 配置

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

## 二进制部署

### 编译

```bash
# 1. 克隆仓库
git clone https://github.com/binrchq/roma.git
cd roma

# 2. 安装依赖
go mod download

# 3. 编译
go build -o roma cmd/roma/main.go

# 或使用 Makefile
make build
```

### 配置

创建配置文件 `configs/config.toml`:

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

### 启动服务

```bash
# 直接运行
./roma -c configs/config.toml

# 使用 systemd (推荐生产环境)
sudo cp roma /usr/local/bin/
sudo cp deployment/roma.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable roma
sudo systemctl start roma
sudo systemctl status roma
```

### Systemd 服务文件

创建 `/etc/systemd/system/roma.service`:

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

## 生产环境配置

### 数据库

#### MySQL 配置

```toml
[database]
type = 'mysql'
cdb_url = 'roma:SecurePassword@tcp(mysql-host:3306)/roma?charset=utf8mb4&parseTime=True&loc=Local'
max_idle_conns = 10
max_open_conns = 100
conn_max_lifetime = 3600
```

#### PostgreSQL 配置

```toml
[database]
type = 'postgresql'
cdb_url = 'host=postgres-host port=5432 user=roma password=SecurePassword dbname=roma sslmode=require'
```

### 安全配置

```toml
[security]
# JWT密钥 (32字节以上)
jwt_secret = 'change-this-to-a-secure-random-string-32-bytes-or-more'

# AES加密密钥 (32字节)
encryption_key = '12345678901234567890123456789012'

# SSH主机密钥
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

### HTTPS/TLS 配置

使用Nginx作为反向代理：

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

# SSH端口转发 (stream模块)
stream {
    server {
        listen 2200;
        proxy_pass localhost:2200;
    }
}
```

### 备份策略

#### 数据库备份

```bash
#!/bin/bash
# backup-roma.sh

BACKUP_DIR="/backup/roma"
DATE=$(date +%Y%m%d_%H%M%S)

# MySQL备份
mysqldump -u roma -p roma > "$BACKUP_DIR/roma_$DATE.sql"

# 压缩备份
gzip "$BACKUP_DIR/roma_$DATE.sql"

# 保留最近7天的备份
find "$BACKUP_DIR" -name "roma_*.sql.gz" -mtime +7 -delete
```

添加到crontab:

```bash
# 每天凌晨2点备份
0 2 * * * /usr/local/bin/backup-roma.sh
```

#### 配置文件备份

```bash
# 备份配置和密钥
tar -czf roma-config-backup.tar.gz \
  /usr/local/roma/configs/ \
  /usr/local/roma/keys/
```

### 监控和日志

#### 日志配置

```toml
[log]
level = 'info'  # debug, info, warn, error
format = 'json'
output = '/var/log/roma/roma.log'
max_size = 100  # MB
max_backups = 10
max_age = 30  # days
compress = true
```

#### Prometheus监控

ROMA暴露Prometheus指标：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'roma'
    static_configs:
      - targets: ['localhost:6999']
    metrics_path: '/metrics'
```

---

## 升级

### Docker升级

```bash
# 1. 拉取最新镜像
docker compose -f quickstart.yaml pull

# 2. 重启服务
docker compose -f quickstart.yaml up -d

# 3. 清理旧镜像
docker image prune -f
```

### 二进制升级

```bash
# 1. 备份当前版本
cp /usr/local/bin/roma /usr/local/bin/roma.backup

# 2. 下载新版本
wget https://github.com/binrchq/roma/releases/download/v1.x.x/roma-linux-amd64
chmod +x roma-linux-amd64

# 3. 替换二进制
sudo mv roma-linux-amd64 /usr/local/bin/roma

# 4. 重启服务
sudo systemctl restart roma

# 5. 检查版本
roma --version
```

---

## 故障排查

### 服务无法启动

```bash
# 检查日志
docker compose -f quickstart.yaml logs roma

# 或 systemd
sudo journalctl -u roma -f

# 常见问题:
# 1. 端口被占用
sudo netstat -tlnp | grep -E '6999|2200|7000'

# 2. 数据库连接失败
# 检查数据库配置和连接字符串

# 3. 权限问题
sudo chown -R roma:roma /usr/local/roma
```

### SSH连接失败

```bash
# 测试SSH连接
ssh -vvv user@roma-server -p 2200

# 检查SSH主机密钥
ls -la /usr/local/roma/keys/

# 重新生成主机密钥
ssh-keygen -t rsa -b 4096 -f /usr/local/roma/keys/id_rsa
```

### 性能问题

```bash
# 检查资源使用
docker stats

# 或系统资源
top
htop

# 数据库连接池
# 增加 max_open_conns 和 max_idle_conns
```

---

## 安全检查清单

- [ ] 修改默认密码
- [ ] 配置强JWT密钥和加密密钥
- [ ] 启用HTTPS/TLS
- [ ] 配置防火墙规则
- [ ] 启用速率限制
- [ ] 启用IP黑名单
- [ ] 配置定期备份
- [ ] 设置日志监控
- [ ] 定期更新版本
- [ ] 审查用户权限

---

## 相关文档

- [安全指南](SECURITY.md)
- [开发指南](DEVELOPMENT.md)
- [API文档](API.md)

