# ROMA 安全指南

本文档详细说明ROMA的安全机制和最佳实践。

---

## 🛡️ 安全架构

ROMA采用多层防御架构，确保系统安全：

```
┌─────────────────────────────────────┐
│  网络层 (Firewall/VPN/Nginx)        │
├─────────────────────────────────────┤
│  应用层 (Rate Limit/IP Blacklist)   │
├─────────────────────────────────────┤
│  认证层 (SSH Key/API Key/JWT)       │
├─────────────────────────────────────┤
│  授权层 (RBAC/Space Isolation)      │
├─────────────────────────────────────┤
│  数据层 (Encryption/Audit Log)      │
└─────────────────────────────────────┘
```

---

## 🔐 认证与授权

### SSH密钥认证

ROMA使用SSH密钥进行用户认证，禁用密码登录：

**生成SSH密钥:**

```bash
# 生成RSA密钥
ssh-keygen -t rsa -b 4096 -f ~/.ssh/roma_key -C "user@example.com"

# 生成ED25519密钥 (推荐)
ssh-keygen -t ed25519 -f ~/.ssh/roma_key -C "user@example.com"
```

**上传公钥:**

方式1: 通过Web UI上传
- 登录Web界面
- 进入 Settings -> SSH Keys
- 点击 Upload Public Key
- 粘贴 `~/.ssh/roma_key.pub` 内容

方式2: 通过API上传
```bash
curl -X POST http://roma-server:6999/api/v1/users/me/ssh-keys \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "my-laptop",
    "PUBLIC_KEY": "ssh-rsa AAAAB3..."
  }'
```

**连接:**

```bash
ssh user@roma-server -p 2200 -i ~/.ssh/roma_key
```

### API密钥授权

API访问使用API密钥进行授权：

**生成API密钥:**

通过Web UI:
- Settings -> API Keys -> Generate New Key

通过配置文件:
```toml
[apikey]
prefix = 'apikey.'
key = 'your-secure-random-api-key-here'
```

**使用API密钥:**

```bash
curl -H "apikey: apikey.your-key" http://roma-server:6999/api/v1/resources
```

**API密钥最佳实践:**
- ✅ 使用长度 ≥ 32字符的随机密钥
- ✅ 定期轮换API密钥
- ✅ 不同环境使用不同密钥
- ✅ 不要在代码中硬编码密钥
- ✅ 使用环境变量或密钥管理工具

### JWT令牌

Web UI和API使用JWT令牌进行会话管理：

**配置JWT:**

```toml
[security]
jwt_secret = 'your-jwt-secret-at-least-32-bytes-long-change-me'
jwt_expire_hours = 24
```

**JWT最佳实践:**
- ✅ 使用强随机字符串作为secret (≥ 32字节)
- ✅ 设置合理的过期时间 (1-24小时)
- ✅ 生产环境定期更换secret
- ✅ 使用HTTPS传输JWT

---

## 🔒 基于角色的访问控制 (RBAC)

### 内置角色

| 角色 | 权限 | 使用场景 |
|------|------|----------|
| `super` | 超级管理员，所有权限 | 系统管理员 |
| `system` | 系统资源管理 | 运维工程师 |
| `ops` | 操作和监控权限 | 运维人员 |
| `ordinary` | 基本资源访问 | 普通用户 |

### 自定义角色

通过API创建自定义角色：

```bash
curl -X POST http://roma-server:6999/api/v1/roles \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "db-admin",
    "DESCRIPTION": "数据库管理员",
    "PERMISSIONS": ["database.read", "database.execute"]
  }'
```

### 资源级权限

为资源指定特定角色：

```bash
# 创建资源时指定角色
curl -X POST http://roma-server:6999/api/v1/resources \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "TYPE": "database",
    "NAME": "prod-mysql",
    "ROLES": ["db-admin", "ops"]
  }'
```

---

## 🧩 空间隔离

空间(Space)提供多租户级别的资源隔离：

### 创建空间

```bash
curl -X POST http://roma-server:6999/api/v1/spaces \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "production",
    "DESCRIPTION": "生产环境"
  }'
```

### 分配资源到空间

```bash
# 创建资源时指定空间
curl -X POST http://roma-server:6999/api/v1/resources \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "web-server-01",
    "SPACE_ID": "production",
    "TYPE": "linux"
  }'
```

### 用户加入空间

```bash
curl -X POST http://roma-server:6999/api/v1/spaces/production/members \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "USER_ID": "user123",
    "ROLES": ["ops"]
  }'
```

---

## 🛡️ 防护机制

### IP黑名单

**功能:**
- 全局IP封禁
- 地理位置查询 (ipseek.cc)
- 自动封禁暴力破解IP
- API和SSH层统一防护

**配置:**

```toml
[ip_blacklist]
enabled = true
auto_ban_threshold = 5      # 失败次数阈值
auto_ban_duration = 3600    # 封禁时长(秒)
```

**手动封禁IP:**

通过Web UI:
- Security -> IP Blacklist -> Add IP

通过API:
```bash
curl -X POST http://roma-server:6999/api/v1/security/blacklist \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "IP": "192.168.1.100",
    "REASON": "暴力破解尝试",
    "DURATION": 7200
  }'
```

**查看黑名单:**

```bash
curl -H "apikey: your-api-key" \
  http://roma-server:6999/api/v1/security/blacklist
```

### 速率限制

**功能:**
- 每IP并发连接限制
- 每IP QPS限制
- 防止DDoS攻击

**配置:**

```toml
[rate_limit]
enabled = true
requests_per_second = 100      # 每秒请求数
burst = 200                    # 突发请求数
per_ip_concurrent_limit = 10   # 每IP并发连接数
```

### 认证失败追踪

**功能:**
- 追踪认证失败次数
- 达到阈值自动封禁
- 记录失败日志

**流程:**

```
1. 认证失败 -> 记录IP和次数
2. 达到警告阈值 -> 记录警告日志
3. 达到封禁阈值 -> 自动加入黑名单
4. 封禁时长过期 -> 自动解除封禁
```

**查看失败记录:**

```bash
# 查看审计日志
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?event=auth_failed"
```

---

## 🔑 凭据管理

### 用户密码加密

**算法:** Bcrypt (cost=10)

**配置:**

```toml
[security]
bcrypt_cost = 10  # 加密强度 (4-31)
```

**密码策略:**
- 最小长度: 8字符
- 必须包含: 大写字母、小写字母、数字
- 推荐: 包含特殊字符

### 资源凭据加密

**算法:** AES-256-GCM

**配置:**

```toml
[security]
encryption_key = '12345678901234567890123456789012'  # 32字节
```

**生成安全密钥:**

```bash
# Linux/macOS
openssl rand -hex 32

# 或使用Python
python3 -c "import secrets; print(secrets.token_hex(32))"
```

**密钥轮转:**

```bash
# 1. 生成新密钥
NEW_KEY=$(openssl rand -hex 32)

# 2. 更新配置
vim configs/config.toml
# encryption_key = '$NEW_KEY'

# 3. 重启服务
systemctl restart roma

# 4. 重新加密凭据 (自动完成)
```

### 密钥存储

**推荐方式:**

1. **环境变量:**
```bash
export ROMA_ENCRYPTION_KEY="your-32-byte-key"
export ROMA_JWT_SECRET="your-jwt-secret"
```

2. **密钥管理工具:**
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Kubernetes Secrets

3. **文件权限:**
```bash
# 限制配置文件权限
chmod 600 /usr/local/roma/configs/config.toml
chown roma:roma /usr/local/roma/configs/config.toml
```

---

## 📝 审计日志

### 日志内容

ROMA记录所有关键操作：

- 用户登录/登出
- 资源访问
- 命令执行
- 配置修改
- 认证失败
- 权限拒绝

### 日志格式

```json
{
  "TIMESTAMP": "2025-11-27T10:30:00Z",
  "USER_ID": "user123",
  "USERNAME": "admin",
  "EVENT": "resource_access",
  "RESOURCE_TYPE": "linux",
  "RESOURCE_NAME": "web-server-01",
  "ACTION": "execute_command",
  "COMMAND": "df -h",
  "STATUS": "success",
  "IP_ADDRESS": "192.168.1.100",
  "USER_AGENT": "SSH-2.0-OpenSSH_8.2"
}
```

### 查看审计日志

**通过Web UI:**
- Audit -> Audit Logs

**通过API:**
```bash
# 查询最近的日志
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?limit=100"

# 按用户查询
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?user=admin"

# 按事件类型查询
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?event=auth_failed"
```

### 日志导出

```bash
# 导出CSV格式
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs/export?format=csv" \
  > audit-logs.csv

# 导出JSON格式
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs/export?format=json" \
  > audit-logs.json
```

---

## 🌐 网络安全

### 防火墙配置

**iptables:**

```bash
# 允许SSH (堡垒机)
iptables -A INPUT -p tcp --dport 2200 -j ACCEPT

# 允许API (仅内网)
iptables -A INPUT -p tcp --dport 6999 -s 10.0.0.0/8 -j ACCEPT

# 允许Web UI (HTTPS)
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# 拒绝其他入站连接
iptables -A INPUT -j DROP
```

**ufw:**

```bash
# 允许SSH堡垒机
ufw allow 2200/tcp

# 允许API (仅内网)
ufw allow from 10.0.0.0/8 to any port 6999

# 允许HTTPS
ufw allow 443/tcp

# 启用防火墙
ufw enable
```

### VPN集成

**推荐配置:**

1. ROMA部署在VPN内网
2. 用户通过VPN连接到内网
3. 只暴露必要端口

**WireGuard示例:**

```bash
# 服务器端
wg-quick up wg0

# 客户端连接
wg-quick up wg-client

# 然后连接ROMA
ssh user@10.0.0.10 -p 2200
```

### TLS/HTTPS

**使用Nginx反向代理:**

```nginx
server {
    listen 443 ssl http2;
    server_name roma.example.com;

    ssl_certificate /etc/letsencrypt/live/roma.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/roma.example.com/privkey.pem;
    
    # SSL安全配置
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;
    
    # 代理配置
    location / {
        proxy_pass http://localhost:7000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 📋 安全最佳实践

### 部署阶段

- [ ] 修改所有默认密码和密钥
- [ ] 使用强随机字符串作为JWT secret和加密密钥
- [ ] 配置HTTPS/TLS
- [ ] 限制网络访问 (防火墙/VPN)
- [ ] 启用速率限制和IP黑名单
- [ ] 配置定期备份
- [ ] 设置日志监控和告警

### 运维阶段

- [ ] 定期更新ROMA版本
- [ ] 定期审查用户权限
- [ ] 定期检查审计日志
- [ ] 定期轮换密钥和密码
- [ ] 定期备份数据库
- [ ] 监控系统资源使用
- [ ] 检查安全漏洞和补丁

### 用户管理

- [ ] 遵循最小权限原则
- [ ] 定期审查用户账户
- [ ] 禁用不活跃用户
- [ ] 强制使用SSH密钥认证
- [ ] 禁止共享账户
- [ ] 定期培训安全意识

### 密码策略

- [ ] 最小长度 ≥ 12字符
- [ ] 包含大小写字母、数字、特殊字符
- [ ] 禁止常见密码
- [ ] 定期更换密码 (90天)
- [ ] 禁止重复使用历史密码

---

## 🚨 安全事件响应

### 检测到暴力破解

1. **识别攻击源:**
```bash
# 查看认证失败日志
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?event=auth_failed"
```

2. **封禁IP:**
```bash
curl -X POST http://roma-server:6999/api/v1/security/blacklist \
  -H "apikey: your-api-key" \
  -d '{"IP": "attacker-ip", "REASON": "brute force"}'
```

3. **加强防护:**
```toml
[ip_blacklist]
auto_ban_threshold = 3  # 降低阈值
auto_ban_duration = 7200  # 增加封禁时长
```

### 检测到可疑命令

1. **查看审计日志:**
```bash
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?action=execute_command"
```

2. **禁用用户:**
```bash
curl -X PATCH http://roma-server:6999/api/v1/users/{user_id} \
  -H "apikey: your-api-key" \
  -d '{"ENABLED": false}'
```

3. **通知管理员:**
- 发送告警邮件
- 记录事件报告
- 更新安全策略

### 密钥泄露

1. **立即轮换密钥:**
```bash
# 生成新密钥
NEW_KEY=$(openssl rand -hex 32)

# 更新配置并重启
vim configs/config.toml
systemctl restart roma
```

2. **撤销受影响的凭据:**
- 删除泄露的SSH密钥
- 重置API密钥
- 强制用户重新登录

3. **审查访问记录:**
- 检查泄露期间的所有访问
- 识别可疑活动
- 生成事件报告

---

## 🔍 安全检查清单

### 每日检查

- [ ] 查看认证失败日志
- [ ] 检查IP黑名单
- [ ] 查看活跃会话

### 每周检查

- [ ] 审查用户权限
- [ ] 检查异常操作
- [ ] 验证备份完整性

### 每月检查

- [ ] 更新ROMA版本
- [ ] 轮换API密钥
- [ ] 审查审计日志
- [ ] 检查安全配置

### 每季度检查

- [ ] 轮换加密密钥
- [ ] 安全审计
- [ ] 渗透测试
- [ ] 更新安全策略

---

## 📞 安全支持

发现安全漏洞请联系:

- 📧 Email: security@binrc.com
- 🔒 GPG Key: [public key](https://binrc.com/security/pgp-key.asc)

**漏洞披露政策:**
- 不公开披露漏洞，直至修复版本发布
- 90天负责任披露期限
- 提供漏洞赏金计划

---

## 相关文档

- [部署指南](DEPLOYMENT.md)
- [开发指南](DEVELOPMENT.md)
- [API文档](API.md)

