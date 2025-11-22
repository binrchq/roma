# ROMA - AI-Powered Jump Server

![Static Badge](https://img.shields.io/badge/License-AGPL_v3-blue)
![Static Badge](https://img.shields.io/badge/lightweight-green)
![Static Badge](https://img.shields.io/badge/AI-Powered-orange)

**ROMA** is an AI-powered, ultra-lightweight jump server (bastion host) built with Go. It provides secure and efficient remote access solutions with native AI integration through Model Context Protocol (MCP). 

Language: [中文](./README_CN.md)

---

<div align="left">
  <img src="./readme.res/logo.png" alt="ROMA Logo" width="100" />
</div>

## 🎯 What is ROMA?

ROMA is a **jump server (bastion host)** that acts as a secure gateway to access your infrastructure resources. Instead of directly connecting to servers, databases, and network devices, you connect to ROMA first, and ROMA manages all the connections, credentials, and access control.

<div align="left">
  <img src="./readme.res/face.png" alt="ROMA face"/>
</div>

### Key Features

- 🚀 **Jump Server** - Secure gateway for remote access
- 🤖 **AI-Powered** - Native MCP support for AI-driven operations
- 🔐 **Secure** - SSH key authentication, API key authorization
- 📦 **Lightweight** - Single binary, minimal dependencies
- 🌐 **Multi-Resource** - Support 6 types of resources
- 💻 **Web UI** - Modern React-based management interface
- 🔌 **MCP Bridge** - Lightweight MCP bridge for AI integration

---

## 🏗️ Jump Server Architecture

```
┌─────────────┐
│   User      │
│  (SSH/API)  │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      ROMA Jump Server           │
│  ┌──────────────────────────┐   │
│  │  SSH Service (Port 2200) │   │
│  │  API Service (Port 6999) │   │
│  │  MCP Bridge (Optional)   │   │
│  └──────────────────────────┘   │
│                                  │
│  • Authentication & Authorization│
│  • Credential Management         │
│  • Audit Logging                  │
│  • Resource Registry              │
└──────┬───────────────────────────┘
       │
       ├──► Linux Servers (SSH)
       ├──► Windows Servers (PowerShell)
       ├──► Docker Containers
       ├──► Databases (MySQL/PostgreSQL/Redis/MongoDB)
       ├──► Routers (Network Devices)
       └──► Switches (Network Devices)
```

### Why Use a Jump Server?

✅ **Security** - Centralized access control, no direct exposure of internal resources  
✅ **Audit** - All access is logged and traceable  
✅ **Credential Management** - Centralized credential storage, no need to distribute keys  
✅ **Access Control** - Role-based permissions, fine-grained access control  
✅ **Simplified Management** - One entry point to manage all resources  

---

## 📦 Supported Resource Types

ROMA supports **6 types of resources**, each with specialized connection and execution capabilities:

### 1. 🐧 Linux Servers

- **Connection**: SSH
- **Execution**: Shell commands (bash, sh, etc.)
- **Features**:
  - Interactive SSH terminal
  - Non-interactive command execution
  - File transfer (SCP/SFTP)
  - System monitoring commands

**Example Usage:**
```bash
# Interactive connection
ssh user@roma-jump-server -p 2200
> ln -t linux web-server-01
> df -h
> exit

# Non-interactive command execution
ssh user@roma-jump-server -p 2200 "ln -t linux web-server-01 -- 'df -h'"
```

### 2. 🪟 Windows Servers

- **Connection**: PowerShell Remoting (WinRM)
- **Execution**: PowerShell commands
- **Features**:
  - Remote PowerShell execution
  - Windows service management
  - Event log queries
  - Registry operations

**Example Usage:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t windows win-server-01 -- 'Get-Service | Where-Object {$_.Status -eq \"Running\"}'"
```

### 3. 🐳 Docker Containers

- **Connection**: SSH to host + Docker CLI
- **Execution**: Docker commands
- **Features**:
  - Container lifecycle management (start/stop/restart)
  - Container logs viewing
  - Exec into containers
  - Image management

**Example Usage:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t docker container-01 -- 'docker ps'"
ssh user@roma-jump-server -p 2200 "ln -t docker container-01 -- 'docker logs -f app'"
```

### 4. 🗄️ Databases

- **Connection**: Native database protocols
- **Execution**: SQL queries
- **Supported Databases**:
  - MySQL / MariaDB
  - PostgreSQL
  - Redis
  - MongoDB
  - Microsoft SQL Server
  - ClickHouse
  - Elasticsearch

**Features**:
  - Interactive database CLI
  - Non-interactive SQL query execution
  - Multiple statement support (semicolon-separated)
  - Formatted query results

**Example Usage:**
```bash
# Interactive mode
ssh user@roma-jump-server -p 2200
> ln -t database links-mysql
mysql [links]> SHOW databases;
mysql [links]> SELECT * FROM users LIMIT 10;
mysql [links]> exit

# Non-interactive mode
ssh user@roma-jump-server -p 2200 "ln -t database links-mysql -- 'SHOW databases;'"
ssh user@roma-jump-server -p 2200 "ln -t database links-mysql -- 'SHOW databases;SHOW tables;'"
```

### 5. 🛣️ Routers

- **Connection**: SSH (Cisco, Huawei, etc.)
- **Execution**: Router CLI commands
- **Features**:
  - Interface configuration
  - Routing table management
  - Network status queries
  - Configuration backup/restore

**Example Usage:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t router core-router-01 -- 'show ip route'"
ssh user@roma-jump-server -p 2200 "ln -t router core-router-01 -- 'show interfaces'"
```

### 6. 🔌 Switches

- **Connection**: SSH (Cisco, Huawei, etc.)
- **Execution**: Switch CLI commands
- **Features**:
  - Port management
  - VLAN configuration
  - MAC address table queries
  - Port status monitoring

**Example Usage:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t switch access-switch-01 -- 'show vlan'"
ssh user@roma-jump-server -p 2200 "ln -t switch access-switch-01 -- 'show mac address-table'"
```

---

## 🤖 AI MCP Integration

ROMA provides **native AI integration** through Model Context Protocol (MCP), allowing AI assistants (Claude, GPT, Cursor, etc.) to directly interact with your infrastructure.

### MCP Architecture

ROMA offers **two MCP integration modes**:

#### 1. MCP Bridge (Lightweight, Recommended)

A lightweight bridge that connects AI assistants to ROMA jump server via SSH.

```
AI Assistant (Claude Desktop/Cursor)
        ↓ stdio (JSON-RPC)
MCP Bridge (~5MB binary)
        ↓ SSH (Port 2200)
ROMA Jump Server
        ↓
Target Resources (Linux/Windows/Docker/Database/Router/Switch)
```

**Features:**
- ✅ Lightweight (~5MB binary)
- ✅ SSH-based connection (no HTTP API needed)
- ✅ Full ROMA command support (ln, ls, whoami, etc.)
- ✅ Automatic resource name matching
- ✅ Multi-step execution support
- ✅ Conversation history awareness

**Quick Start:**
```bash
# 1. Build MCP Bridge
cd mcp/bridge
go build -o roma-mcp-bridge

# 2. Configure Claude Desktop
# ~/.config/Claude/claude_desktop_config.json
{
  "mcpServers": {
    "roma": {
      "command": "/path/to/roma-mcp-bridge",
      "env": {
        "ROMA_SSH_HOST": "10.2.2.230",
        "ROMA_SSH_PORT": "2200",
        "ROMA_SSH_USER": "super",
        "ROMA_SSH_KEY": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
      }
    }
  }
}
```

**Documentation:** [mcp/bridge/README.md](mcp/bridge/README.md)

#### 2. MCP Server (Integrated)

MCP server integrated into ROMA main service (legacy, being phased out).

---

### MCP Tools Overview

ROMA MCP Bridge provides **20+ powerful tools** organized into categories:

#### 📋 ROMA Resource Info Tools

Query ROMA jump server's resource registry (not actual server data):

- `list_resources` - List all resources of a type
- `get_resource_info` - Get detailed resource configuration
- `get_current_user` - Get current user info and permissions
- `get_command_history` - Get ROMA command history

#### 💻 Actual Data Query Tools

Query actual server/database data:

- `execute_command` - Execute shell commands on resources
- `execute_database_query` - Execute SQL queries on databases
- `execute_commands` - Execute multiple commands
- `copy_file_to_resource` / `copy_file_from_resource` - File transfer (SCP)

#### 🔧 System Monitoring Tools

Convenience tools for common operations:

- `get_disk_usage` - Disk usage (df -h)
- `get_memory_usage` - Memory usage (free -h)
- `get_cpu_info` - CPU information (lscpu)
- `get_process_list` - Process list (ps aux)
- `get_network_info` - Network information (ip addr)
- `get_uptime` - System uptime (uptime)
- `get_system_info` - System details (uname, os-release)

### AI Usage Examples

With MCP integration, you can use natural language to control your infrastructure:

```
💬 "List all Linux servers"
💬 "What databases are in links-mysql?"
💬 "Check disk usage on web-server-01"
💬 "Show me the system logs on db-01"
💬 "Execute 'df -h' on all production servers"
💬 "Query the users table in links-mysql database"
💬 "Upload file config.json to server-01 at /tmp/"
```

The AI will automatically:
1. Choose the right tool
2. Execute commands/queries
3. Present results in a readable format
4. Handle errors and retry with correct resource names

---

## 🚀 Quick Start

### 1. Installation

```bash
git clone https://github.com/binrchq/roma.git
cd roma
go build -o roma cmd/roma/main.go
```

### 2. Configuration

Create `configs/config.toml`:

```toml
[api]
host = '0.0.0.0'
port = '6999'

[common]
port = '2200'  # SSH jump server port
prompt = 'roma'

[database]
cdb_url = '/usr/local/roma/roma.db'

[apikey]
prefix = 'apikey.'
key = 'your-api-key-here'

[user_1st]
username = 'super'
email = 'super@example.com'
password = 'super001.'
public_key = 'ssh-rsa AAAAB3...'  # Your SSH public key
roles = "super,system,ops"
```

### 3. Start ROMA

```bash
./roma -c configs/config.toml
```

ROMA will start:
- **SSH Service** on port 2200 (jump server)
- **API Service** on port 6999 (RESTful API)

### 4. Connect via SSH

```bash
ssh super@your-roma-server -p 2200 -i ~/.ssh/your_key
```

You'll see the ROMA TUI with commands:
- `ls` - List resources
- `ln` - Login to resources
- `use` - Switch resource type context
- `whoami` - User information
- `help` - Command help

### 5. Setup MCP Bridge (Optional)

```bash
# Build MCP Bridge
cd mcp/bridge
go build -o roma-mcp-bridge

# Configure environment variables
export ROMA_SSH_HOST="your-roma-server"
export ROMA_SSH_PORT="2200"
export ROMA_SSH_USER="super"
export ROMA_SSH_KEY="$(cat ~/.ssh/your_private_key)"

# Test
./roma-mcp-bridge
```

Then configure your AI assistant (Claude Desktop, Cursor, etc.) to use the bridge.

---

## 📚 Documentation

- **[MCP Bridge Guide](mcp/bridge/README.md)** - Complete MCP Bridge documentation
- **[MCP Bridge Architecture](mcp/bridge/ARCHITECTURE.md)** - Architecture details
- **[Resource Support](docs/RESOURCE_SUPPORT.md)** - Detailed resource type support
- **[Web Frontend](web/frontend/README.md)** - Web UI documentation
- **[VSCode Extension](web/vscode-extension/README.md)** - IDE integration

---

## 🎯 Use Cases

### 1. Secure Remote Access

Instead of exposing all servers directly:
- Deploy ROMA as a jump server
- Users connect to ROMA only
- ROMA manages connections to internal resources
- All access is logged and audited

### 2. AI-Driven Operations

Use AI assistants to:
- Automate routine operations
- Query infrastructure status
- Execute commands across multiple servers
- Generate reports and summaries

### 3. Multi-Resource Management

Manage diverse infrastructure from one place:
- Linux servers
- Windows servers
- Docker containers
- Databases (MySQL, PostgreSQL, Redis, etc.)
- Network devices (routers, switches)

### 4. Team Collaboration

- Centralized credential management
- Role-based access control
- Audit logging for compliance
- Web UI for non-technical users

---

## 🔐 Security Features

- ✅ **SSH Key Authentication** - No password authentication
- ✅ **API Key Authorization** - Secure API access
- ✅ **Role-Based Access Control (RBAC)** - Fine-grained permissions
- ✅ **Audit Logging** - All operations are logged
- ✅ **Credential Encryption** - Secure credential storage
- ✅ **Session Management** - Track and manage active sessions

---

## 🌐 Web Management Interface

ROMA includes a modern web UI built with React:

- 📊 Dashboard with resource statistics
- 🖥️ Resource management (CRUD operations)
- 👥 User and role management
- 💻 Web SSH terminal
- 📝 Audit log viewer

**Start Web UI:**
```bash
cd web/frontend
npm install
npm run dev
# Access at http://localhost:3000
```

---

## 📦 Project Structure

```
roma/
├── cmd/roma/              # Main program entry
├── core/                  # Core functionality
│   ├── api/              # API controllers
│   ├── model/            # Data models
│   ├── operation/        # Business logic
│   ├── connect/          # Connection handlers
│   ├── tui/              # Terminal UI (SSH commands)
│   └── constants/        # Constants (resource types, etc.)
├── mcp/                  # MCP integration
│   └── bridge/           # MCP Bridge (lightweight)
│       ├── mappings/     # Tool mappings (organized)
│       ├── main.go       # Bridge entry point
│       └── client.go     # ROMA SSH client
├── web/                  # Web components
│   ├── frontend/         # React frontend
│   ├── vscode-extension/ # VSCode extension
│   └── ops-client/       # Electron desktop app
├── configs/              # Configuration files
└── docs/                # Documentation
```

---

## 🔗 License

This project is licensed under Dual License ->  **GNU Affero General Public License (AGPL) v3.0** And **COMMERCIAL SOFTWARE LICENSE AGREEMENT**.

📢 **Important**: Any organization or individual that modifies ROMA code for providing **remote access services** must **open source their modified version**.

See [LICENSE](./LICENSE) for details.

---

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines and code of conduct.

---

## 📞 Support

- 📧 Email: support@binrc.com
- 🐛 Issues: [GitHub Issues](https://github.com/binrchq/roma/issues)
- 📖 Documentation: [docs/](docs/)

---

**ROMA** - A seamless solution for remote access, ensuring both efficiency and security. 🚀
