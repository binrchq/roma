# ROMA - AI-Powered Jump Server

<div align="left">
  <img src="./readme.res/logo.png" alt="ROMA Logo" width="100" />
</div>

![License](https://img.shields.io/badge/License-AGPL_v3-blue)
![Lightweight](https://img.shields.io/badge/lightweight-green)
![AI-Powered](https://img.shields.io/badge/AI-Powered-orange)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)

**ROMA** is an AI-powered, ultra-lightweight jump server (bastion host) built with Go. It provides secure and efficient remote access solutions with native AI integration through Model Context Protocol (MCP). 

**Live Demo:** https://roma-demo.binrc.com (demo/demo123456)

Language: English • [中文](./README_CN.md)

---

## Related Projects

| Project | Description | Repository |
|---------|-------------|------------|
| **roma** | Core jump server service (Go) | This project |
| **roma-web** | Web management UI (React) | [github.com/binrchq/roma-web](https://github.com/binrchq/roma-web) |
| **roma-mcp** | Standalone MCP service | [github.com/binrchq/roma-mcp](https://github.com/binrchq/roma-mcp) |
| **roma-vsc-ext** | VSCode extension | [github.com/binrchq/roma-vsc-ext](https://github.com/binrchq/roma-vsc-ext) |

**Official Website:** https://roma.binrc.com

---

<div align="left">
  <img src="./readme.res/face.png" alt="ROMA Interface" width="800" />
</div>

## Key Features

- **Jump Server** - Unified remote access gateway with centralized control
- **AI-Powered** - Native MCP support for AI-driven infrastructure management
- **Space Isolation** - Multi-tenant level resource isolation
- **Security Hardening** - SSH key authentication, API key authorization, multi-layer protection
- **Lightweight** - Single binary, minimal dependencies
- **Multi-Resource Support** - Linux/Windows/Docker/Database/Router/Switch
- **Modern Web UI** - React-based management interface
- **Adaptive Security** - Rate limiting, IP blacklist, auth-failure guardrails
- **MCP Bridge** - Lightweight AI integration bridge

---

## Quick Deployment

### Docker Deployment (Recommended)

```bash
# 1. Download quickstart configuration
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. Start services
docker compose -f quickstart.yaml up -d

# 3. Access Web UI
open http://localhost:7000
```

**Default Credentials:**
- Username: `demo`
- Password: `demo123456`

**Service Ports:**
- Web UI: `7000`
- API: `6999`
- SSH: `2200`

### Binary Deployment

```bash
# 1. Clone repository
git clone https://github.com/binrchq/roma.git
cd roma

# 2. Build
go build -o roma cmd/roma/main.go

# 3. Configure (refer to configs/config.ex.toml)
cp configs/config.ex.toml configs/config.toml
vim configs/config.toml

# 4. Start
./roma -c configs/config.toml
```

### Production Deployment

Support for MySQL/PostgreSQL databases:

```bash
# MySQL
docker compose -f deployment/quickstart.mysql.yaml up -d

# PostgreSQL
docker compose -f deployment/quickstart.pgsql.yaml up -d
```

**Deployment Guide:** [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

---

## Usage

### SSH Command Line

Connect to ROMA jump server and use kubectl-style commands to manage resources:

```bash
# Connect to jump server
ssh demo@localhost -p 2200

# List resources (similar to kubectl get)
roma> ls                    # List all resources of current type
roma> ls linux              # List all Linux servers
roma> ls database           # List all databases

# Switch context (similar to kubectl use-context)
roma> use linux             # Switch to Linux context
roma> use database          # Switch to database context

# Login to resource (ln = login)
roma> ln web-server-01                          # Interactive login
roma> ln -t linux web-01 -- 'df -h'            # Execute single command
roma> ln -t database mysql-prod -- 'SHOW databases;'  # Database query

# User information
roma> whoami                # Show current user and permissions

# Help
roma> help                  # Show all available commands
```

### File Transfer (SCP)

ROMA supports standard SCP protocol for file transfer with a special path format through the jump server:

**Path Format:** `user@jumpserver:user@hostname:/remote/path`

**Upload file to server:**

```bash
# Basic usage
scp -P 2200 /local/file.txt user@roma-server:user@web-server-01:/tmp/

# Using SSH key
scp -P 2200 -i ~/.ssh/roma_key /local/config.json user@roma-server:user@web-server-01:/etc/app/

# Example: Upload log file
scp -P 2200 -i ~/.ssh/id_rsa /var/log/app.log demo@localhost:demo@web-01:/tmp/app.log
```

**Download file from server:**

```bash
# Basic usage
scp -P 2200 user@roma-server:user@web-server-01:/tmp/file.txt /local/path/

# Download configuration file
scp -P 2200 -i ~/.ssh/roma_key user@roma-server:user@db-01:/etc/mysql/my.cnf ./backup/

# Example: Download database backup
scp -P 2200 -i ~/.ssh/id_rsa demo@localhost:demo@db-01:/backup/db.sql.gz ./
```

**Supported Resource Types:**
- Linux servers
- Windows servers (requires OpenSSH Server)
- Directory transfer not supported (compress first)

**Path Components:**
- `user@jumpserver` - ROMA jump server user and address
- `user@hostname` - Target server user and hostname (must be registered in ROMA)
- `/remote/path` - File path on target server

**File Transfer via MCP:**

AI assistants can use built-in file transfer tools:

Examples:
```
"Upload config.json to web-server-01 /etc/app/ directory"
"Download /backup/db.sql.gz from db-01 to local"
```

MCP Tools:
- `copy_file_to_resource` - Upload file
- `copy_file_from_resource` - Download file

### MCP Integration (AI Assistant)

ROMA provides lightweight MCP Bridge for AI assistants to directly manage infrastructure:

**1. Build MCP Bridge:**

```bash
cd mcp/bridge
go build -o roma-mcp-bridge
```

**2. Configure Claude Desktop:**

Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "roma": {
      "command": "/path/to/roma-mcp-bridge",
      "env": {
        "ROMA_SSH_HOST": "your-roma-server",
        "ROMA_SSH_PORT": "2200",
        "ROMA_SSH_USER": "your-username",
        "ROMA_SSH_KEY": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
      }
    }
  }
}
```

**3. Using AI Assistant:**

Example commands:
```
"List all Linux servers"
"Check disk usage on web-01"
"Query user table in production database"
"Upload configuration file to server"
"Show all container running status"
```

**MCP Tools List:**

| Category | Tool | Description |
|----------|------|-------------|
| Resource Query | `list_resources` | List resources |
| | `get_resource_info` | Get resource details |
| | `get_current_user` | Get current user |
| Command Execution | `execute_command` | Execute shell commands |
| | `execute_database_query` | Execute SQL queries |
| | `execute_commands` | Batch command execution |
| File Transfer | `copy_file_to_resource` | Upload file |
| | `copy_file_from_resource` | Download file |
| System Monitoring | `get_disk_usage` | Disk usage |
| | `get_memory_usage` | Memory usage |
| | `get_cpu_info` | CPU information |
| | `get_process_list` | Process list |
| | `get_network_info` | Network information |
| | `get_system_info` | System information |

**Detailed Documentation:** [mcp/bridge/README.md](mcp/bridge/README.md)

---

## Security

ROMA provides multi-layer security protection suitable for production and internet-facing deployments:

### Authentication & Authorization

- **SSH Key Authentication** - Password login disabled
- **API Key Authorization** - Secure API access control
- **Role-Based Access Control (RBAC)** - Fine-grained permission management
- **Space Isolation** - Multi-tenant level resource isolation

### Credential Security

- **Bcrypt Password Hashing** - Encrypted user password storage
- **AES-256-GCM Encryption** - Resource credential encryption
- **Key Rotation** - Support for periodic key rotation
- **JWT Tokens** - Secure session management

### Protection Mechanisms

- **IP Blacklist** - Global IP blocking (with geolocation lookup)
- **Rate Limiting** - Per-IP concurrent and QPS limits
- **Auth Failure Tracking** - Automatic brute-force blocking
- **Connection Throttling** - Unified SSH and API layer protection
- **Audit Logging** - All operations are traceable

### Network Security

- **Firewall Recommendations** - Expose only necessary ports
- **VPN Integration** - Support VPN backend access
- **TLS/SSL** - HTTPS and encrypted transmission
- **DDoS Protection** - Connection throttling and IP blocking

### Security Best Practices

1. **Change Default Credentials** - Immediately after deployment
2. **Use Strong Passwords** - Length ≥ 12, mixed case, numbers, special characters
3. **Regular Updates** - Keep ROMA and dependencies up to date
4. **Monitor Audit Logs** - Regularly check for anomalous access
5. **Least Privilege Principle** - Grant only necessary roles and permissions
6. **Network Isolation** - Deploy ROMA in isolated network with access restrictions

**Security Configuration Guide:** [docs/SECURITY.md](docs/SECURITY.md)

---

## Supported Resource Types

| Type | Protocol | Features |
|------|----------|----------|
| Linux | SSH | Shell commands, file transfer |
| Windows | WinRM | PowerShell commands |
| Docker | Docker CLI | Container management, log viewing |
| Database | Native | SQL queries (MySQL/PostgreSQL/Redis/MongoDB etc.) |
| Router | SSH | Router CLI commands |
| Switch | SSH | Switch CLI commands |

**Detailed Support:** [docs/RESOURCE_SUPPORT.md](docs/RESOURCE_SUPPORT.md)

---

## Documentation

| Document | Description |
|----------|-------------|
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | Deployment guide (Docker/K8s/Binary) |
| [DEVELOPMENT.md](docs/DEVELOPMENT.md) | Development guide (Architecture/Contributing/Debugging) |
| [SECURITY.md](docs/SECURITY.md) | Security configuration and best practices |
| [SCP_USAGE.md](docs/SCP_USAGE.md) | Detailed SCP file transfer guide |
| [API.md](docs/API.md) | RESTful API documentation |
| [RESOURCE_SUPPORT.md](docs/RESOURCE_SUPPORT.md) | Resource type details |
| [MCP_BRIDGE.md](mcp/bridge/README.md) | MCP Bridge usage guide |
| [MCP_ARCHITECTURE.md](mcp/bridge/ARCHITECTURE.md) | MCP architecture design |

---

## Use Cases

- **Secure Remote Access** - Unified entry point, centralized control, full audit trail
- **AI-Driven Operations** - AI assistants automate routine operational tasks
- **Multi-Resource Management** - One-stop management for servers, databases, network devices
- **Team Collaboration** - Centralized credential management, role-based access control

---

## Support

- Email: support@binrc.com
- Issues: [GitHub Issues](https://github.com/binrchq/roma/issues)
- Official Website: https://roma.binrc.com

---

## License

This project is licensed under dual licenses:
- **GNU Affero General Public License (AGPL) v3.0**
- **Commercial Software License Agreement**

**Important**: Any organization or individual that modifies ROMA code for providing **remote access services** must **open source their modified version**.

See [LICENSE](./LICENSE) for details.

---

## Contributing

Contributions are welcome! Please read [DEVELOPMENT.md](docs/DEVELOPMENT.md) for how to get involved.

---

## Organization Support

ROMA is supported and developed by:

<p align="left" style="">
  <a href="https://binrc.com" target="_blank" style="display: inline-block; vertical-align: middle; margin-right: 20px;">
    <img src="https://binrc.com/img/logo_lite.png" alt="Binrc" height="40" />
  </a>
  <a href="https://ai2o.binrc.com" target="_blank" style="display: inline-block; vertical-align: middle;">
    <img src="docs/AI2O_logo_white.png" alt="AI2O" height="80"  style="image-rendering: auto;"/>
  </a>
</p>

---

## Contributors

Thanks to all developers who have contributed to ROMA:

<a href="https://github.com/binrchq/roma/graphs/contributors">
  <img src="https://avatars.githubusercontent.com/u/37877444?v=4" alt="Contributor" width="60" height="60" style="border-radius: 50%;" />
</a>

---

## Related Products

### ROMC - AI-Driven Operations Automation Platform

ROMC is an operations automation tool developed by Binrc, integrating MCP protocol with intelligent terminal AI assistant.

**Key Features:**
- **Terminal AI Assistant** - Natural language interaction, intelligent understanding of operational intent
- **Native MCP Integration** - Seamless integration with ROMA and other infrastructure
- **Intelligent Decision Making** - AI-based fault diagnosis and auto-remediation
- **Visual Operations** - Intuitive operational data display and analysis
- **Workflow Automation** - Orchestrated operational process automation

Coming soon.

Learn more: https://binrc.com

---

**ROMA** - Secure and efficient remote access solution
