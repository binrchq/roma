# ROMA Security Guide

This document details ROMA's security mechanisms and best practices.

[中文版本](./SECURITY_CN.md)

---

## Security Architecture

ROMA employs a multi-layer defense architecture:

```
┌─────────────────────────────────────┐
│  Network Layer (Firewall/VPN/Nginx) │
├─────────────────────────────────────┤
│  Application Layer (Rate Limit/IP)  │
├─────────────────────────────────────┤
│  Authentication (SSH Key/API/JWT)   │
├─────────────────────────────────────┤
│  Authorization (RBAC/Space)         │
├─────────────────────────────────────┤
│  Data Layer (Encryption/Audit Log)  │
└─────────────────────────────────────┘
```

---

## Authentication & Authorization

### SSH Key Authentication

ROMA uses SSH key authentication, password login is disabled:

**Generate SSH Key:**

```bash
# Generate RSA key
ssh-keygen -t rsa -b 4096 -f ~/.ssh/roma_key -C "user@example.com"

# Generate ED25519 key (recommended)
ssh-keygen -t ed25519 -f ~/.ssh/roma_key -C "user@example.com"
```

**Upload Public Key:**

Method 1: Via Web UI
- Login to Web interface
- Go to Settings -> SSH Keys
- Click Upload Public Key
- Paste `~/.ssh/roma_key.pub` content

Method 2: Via API
```bash
curl -X POST http://roma-server:6999/api/v1/users/me/ssh-keys \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "my-laptop",
    "PUBLIC_KEY": "ssh-rsa AAAAB3..."
  }'
```

### API Key Authorization

**Generate API Key:**

Via Web UI:
- Settings -> API Keys -> Generate New Key

Via configuration file:
```toml
[apikey]
prefix = 'apikey.'
key = 'your-secure-random-api-key-here'
```

**Use API Key:**

```bash
curl -H "apikey: apikey.your-key" http://roma-server:6999/api/v1/resources
```

**API Key Best Practices:**
- Use random keys with length ≥ 32 characters
- Rotate API keys periodically
- Use different keys for different environments
- Never hardcode keys in code
- Use environment variables or secret management tools

### JWT Tokens

Web UI and API use JWT tokens for session management:

```toml
[security]
jwt_secret = 'your-jwt-secret-at-least-32-bytes-long-change-me'
jwt_expire_hours = 24
```

**JWT Best Practices:**
- Use strong random string as secret (≥ 32 bytes)
- Set reasonable expiration time (1-24 hours)
- Rotate secret periodically in production
- Always use HTTPS to transmit JWT

---

## Role-Based Access Control (RBAC)

### Built-in Roles

| Role | Permissions | Use Case |
|------|------------|----------|
| `super` | Super admin, all permissions | System administrators |
| `system` | System resource management | DevOps engineers |
| `ops` | Operation and monitoring | Operations personnel |
| `ordinary` | Basic resource access | Regular users |

### Custom Roles

Create custom roles via API:

```bash
curl -X POST http://roma-server:6999/api/v1/roles \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "db-admin",
    "DESCRIPTION": "Database Administrator",
    "PERMISSIONS": ["database.read", "database.execute"]
  }'
```

---

## Space Isolation

Spaces provide multi-tenant level resource isolation:

### Create Space

```bash
curl -X POST http://roma-server:6999/api/v1/spaces \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "production",
    "DESCRIPTION": "Production Environment"
  }'
```

---

## Protection Mechanisms

### IP Blacklist

**Features:**
- Global IP blocking
- Geolocation lookup (ipseek.cc)
- Automatic brute-force IP blocking
- Unified API and SSH layer protection

**Configuration:**

```toml
[ip_blacklist]
enabled = true
auto_ban_threshold = 5      # Failure threshold
auto_ban_duration = 3600    # Ban duration (seconds)
```

**Manual IP Blocking:**

Via API:
```bash
curl -X POST http://roma-server:6999/api/v1/security/blacklist \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "IP": "192.168.1.100",
    "REASON": "Brute force attempt",
    "DURATION": 7200
  }'
```

### Rate Limiting

**Features:**
- Per-IP concurrent connection limit
- Per-IP QPS limit
- DDoS attack prevention

**Configuration:**

```toml
[rate_limit]
enabled = true
requests_per_second = 100      # Requests per second
burst = 200                    # Burst requests
per_ip_concurrent_limit = 10   # Concurrent connections per IP
```

---

## Credential Management

### User Password Encryption

**Algorithm:** Bcrypt (cost=10)

**Configuration:**

```toml
[security]
bcrypt_cost = 10  # Encryption strength (4-31)
```

**Password Policy:**
- Minimum length: 8 characters
- Must include: uppercase, lowercase, numbers
- Recommended: include special characters

### Resource Credential Encryption

**Algorithm:** AES-256-GCM

**Configuration:**

```toml
[security]
encryption_key = '12345678901234567890123456789012'  # 32 bytes
```

**Generate Secure Key:**

```bash
# Linux/macOS
openssl rand -hex 32

# Or use Python
python3 -c "import secrets; print(secrets.token_hex(32))"
```

---

## Audit Logging

### Log Content

ROMA records all critical operations:

- User login/logout
- Resource access
- Command execution
- Configuration changes
- Authentication failures
- Permission denials

### Log Format

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
  "IP_ADDRESS": "192.168.1.100"
}
```

### View Audit Logs

**Via API:**
```bash
# Query recent logs
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?limit=100"

# Query by user
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?user=admin"

# Query by event type
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?event=auth_failed"
```

---

## Network Security

### Firewall Configuration

**iptables:**

```bash
# Allow SSH (jump server)
iptables -A INPUT -p tcp --dport 2200 -j ACCEPT

# Allow API (internal network only)
iptables -A INPUT -p tcp --dport 6999 -s 10.0.0.0/8 -j ACCEPT

# Allow Web UI (HTTPS)
iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Deny other inbound connections
iptables -A INPUT -j DROP
```

### TLS/HTTPS

**Using Nginx Reverse Proxy:**

```nginx
server {
    listen 443 ssl http2;
    server_name roma.example.com;

    ssl_certificate /etc/letsencrypt/live/roma.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/roma.example.com/privkey.pem;
    
    # SSL security configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=63072000" always;
    
    # Proxy configuration
    location / {
        proxy_pass http://localhost:7000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## Security Best Practices

### Deployment Phase

- [ ] Change all default passwords and keys
- [ ] Use strong random strings for JWT secret and encryption key
- [ ] Configure HTTPS/TLS
- [ ] Limit network access (firewall/VPN)
- [ ] Enable rate limiting and IP blacklist
- [ ] Configure regular backups
- [ ] Set up log monitoring and alerts

### Operations Phase

- [ ] Regularly update ROMA version
- [ ] Periodically review user permissions
- [ ] Regularly check audit logs
- [ ] Rotate keys and passwords periodically
- [ ] Back up database regularly
- [ ] Monitor system resource usage
- [ ] Check for security vulnerabilities and patches

### User Management

- [ ] Follow least privilege principle
- [ ] Periodically review user accounts
- [ ] Disable inactive users
- [ ] Enforce SSH key authentication
- [ ] Prohibit shared accounts
- [ ] Regular security awareness training

---

## Security Incident Response

### Detected Brute Force Attack

1. **Identify attack source:**
```bash
# View authentication failure logs
curl -H "apikey: your-api-key" \
  "http://roma-server:6999/api/v1/audit-logs?event=auth_failed"
```

2. **Block IP:**
```bash
curl -X POST http://roma-server:6999/api/v1/security/blacklist \
  -H "apikey: your-api-key" \
  -d '{"IP": "attacker-ip", "REASON": "brute force"}'
```

3. **Strengthen protection:**
```toml
[ip_blacklist]
auto_ban_threshold = 3  # Lower threshold
auto_ban_duration = 7200  # Increase ban duration
```

### Key Leakage

1. **Immediately rotate keys:**
```bash
# Generate new key
NEW_KEY=$(openssl rand -hex 32)

# Update configuration and restart
vim configs/config.toml
systemctl restart roma
```

2. **Revoke affected credentials:**
- Delete leaked SSH keys
- Reset API keys
- Force user re-login

---

## Security Checklist

### Daily Checks

- [ ] Review authentication failure logs
- [ ] Check IP blacklist
- [ ] View active sessions

### Weekly Checks

- [ ] Review user permissions
- [ ] Check abnormal operations
- [ ] Verify backup integrity

### Monthly Checks

- [ ] Update ROMA version
- [ ] Rotate API keys
- [ ] Review audit logs
- [ ] Check security configuration

### Quarterly Checks

- [ ] Rotate encryption keys
- [ ] Security audit
- [ ] Penetration testing
- [ ] Update security policies

---

## Security Support

For security vulnerabilities, please contact:

- Email: security@binrc.com
- GPG Key: [public key](https://binrc.com/security/pgp-key.asc)

**Vulnerability Disclosure Policy:**
- Do not publicly disclose vulnerabilities until fix is released
- 90-day responsible disclosure period
- Bug bounty program available

---

## Related Documentation

- [Deployment Guide](DEPLOYMENT.md)
- [Development Guide](DEVELOPMENT.md)
- [API Documentation](API.md)
