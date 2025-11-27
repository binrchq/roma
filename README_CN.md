# ROMA - AI 驱动的跳板机

<div align="left">
  <img src="./readme.res/logo.png" alt="ROMA Logo" width="100" />
</div>

![Static Badge](https://img.shields.io/badge/License-AGPL_v3-blue)
![Static Badge](https://img.shields.io/badge/lightweight-green)
![Static Badge](https://img.shields.io/badge/AI-Powered-orange)
![Static Badge](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)
![Static Badge](https://img.shields.io/badge/Demo-在线体验-success)

**ROMA** 是一个 AI 驱动的、使用 Go 语言开发的超轻量级跳板机（堡垒机）服务，提供安全高效的远程访问解决方案，并通过 Model Context Protocol (MCP) 提供原生 AI 集成。

**相关项目：** [Web 界面](https://github.com/binrchq/roma-web) • [MCP 服务器](https://github.com/binrchq/roma-mcp) • [VSCode 扩展](https://github.com/binrchq/roma-vsc-ext) • [官方网站](https://roma.binrc.com)

---

Language: [English](./README.md)

## 🚀 立即体验 ROMA！

### Docker 快速启动（< 2 分钟）

```bash
# 1. 下载快速启动配置文件
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. 启动 ROMA
docker compose -f quickstart.yaml up -d

# 3. 访问 Web UI
open http://localhost:7000
```

**演示账号：**
- 用户名：`demo`
- 密码：`demo123456`

### 在线演示（无需安装）

🌐 **https://roma-demo.binrc.com**
- 凭证：***demo/demo123456***

---

## 🎯 什么是 ROMA？

ROMA 是一个**跳板机（堡垒机）**，作为访问基础设施资源的安全网关。您不需要直接连接到服务器、数据库和网络设备，而是先连接到 ROMA，由 ROMA 管理所有连接、凭证和访问控制。


<div align="left">
  <img src="./readme.res/face.png" alt="ROMA face"/>
</div>


### 核心特性

- 🚀 **跳板机** - 安全的远程访问网关
- 🤖 **AI 驱动** - 原生 MCP 支持，实现 AI 驱动的运维
- 🧩 **空间隔离** - 以「空间」为维度组织资源，实现多租户隔离
- 🔐 **安全** - SSH 密钥认证、API Key 授权
- 📦 **轻量级** - 单二进制文件，最小依赖
- 🌐 **多资源支持** - 支持 6 种资源类型
- 💻 **Web 界面** - 基于 React 的现代化管理界面
- 🛡️ **主动防护** - 速率限制、IP 黑名单、认证失败熔断
- 🔌 **MCP Bridge** - 轻量级 MCP 桥接器，用于 AI 集成

---

## 🏗️ 跳板机架构

```
┌─────────────┐
│   用户      │
│  (SSH/API)  │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      ROMA 跳板机                │
│  ┌──────────────────────────┐   │
│  │  SSH 服务 (端口 2200)     │   │
│  │  API 服务 (端口 6999)     │   │
│  │  MCP Bridge (可选)        │   │
│  └──────────────────────────┘   │
│                                  │
│  • 身份认证与授权                 │
│  • 凭证管理                       │
│  • 审计日志                       │
│  • 资源注册表                     │
└──────┬───────────────────────────┘
       │
       ├──► Linux 服务器 (SSH)
       ├──► Windows 服务器 (PowerShell)
       ├──► Docker 容器
       ├──► 数据库 (MySQL/PostgreSQL/Redis/MongoDB)
       ├──► 路由器 (网络设备)
       └──► 交换机 (网络设备)
```

### 为什么使用跳板机？

✅ **安全性** - 集中式访问控制，内部资源不直接暴露  
✅ **审计** - 所有访问都被记录和可追溯  
✅ **凭证管理** - 集中式凭证存储，无需分发密钥  
✅ **访问控制** - 基于角色的权限，细粒度访问控制  
✅ **简化管理** - 一个入口点管理所有资源  

---

## 📦 支持的资源类型

ROMA 支持 **6 种资源类型**，每种资源都有专门的连接和执行能力：

### 1. 🐧 Linux 服务器

- **连接方式**: SSH
- **执行方式**: Shell 命令 (bash, sh 等)
- **功能特性**:
  - 交互式 SSH 终端
  - 非交互式命令执行
  - 文件传输 (SCP/SFTP)
  - 系统监控命令

**使用示例:**
```bash
# 交互式连接
ssh user@roma-jump-server -p 2200
> ln -t linux web-server-01
> df -h
> exit

# 非交互式命令执行
ssh user@roma-jump-server -p 2200 "ln -t linux web-server-01 -- 'df -h'"
```

### 2. 🪟 Windows 服务器

- **连接方式**: PowerShell Remoting (WinRM)
- **执行方式**: PowerShell 命令
- **功能特性**:
  - 远程 PowerShell 执行
  - Windows 服务管理
  - 事件日志查询
  - 注册表操作

**使用示例:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t windows win-server-01 -- 'Get-Service | Where-Object {$_.Status -eq \"Running\"}'"
```

### 3. 🐳 Docker 容器

- **连接方式**: SSH 到主机 + Docker CLI
- **执行方式**: Docker 命令
- **功能特性**:
  - 容器生命周期管理 (启动/停止/重启)
  - 容器日志查看
  - 进入容器执行命令
  - 镜像管理

**使用示例:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t docker container-01 -- 'docker ps'"
ssh user@roma-jump-server -p 2200 "ln -t docker container-01 -- 'docker logs -f app'"
```

### 4. 🗄️ 数据库

- **连接方式**: 原生数据库协议
- **执行方式**: SQL 查询
- **支持的数据库**:
  - MySQL / MariaDB
  - PostgreSQL
  - Redis
  - MongoDB
  - Microsoft SQL Server
  - ClickHouse
  - Elasticsearch

**功能特性**:
  - 交互式数据库 CLI
  - 非交互式 SQL 查询执行
  - 支持多个语句（用分号分隔）
  - 格式化的查询结果

**使用示例:**
```bash
# 交互式模式
ssh user@roma-jump-server -p 2200
> ln -t database links-mysql
mysql [links]> SHOW databases;
mysql [links]> SELECT * FROM users LIMIT 10;
mysql [links]> exit

# 非交互式模式
ssh user@roma-jump-server -p 2200 "ln -t database links-mysql -- 'SHOW databases;'"
ssh user@roma-jump-server -p 2200 "ln -t database links-mysql -- 'SHOW databases;SHOW tables;'"
```

### 5. 🛣️ 路由器

- **连接方式**: SSH (Cisco, Huawei 等)
- **执行方式**: 路由器 CLI 命令
- **功能特性**:
  - 接口配置
  - 路由表管理
  - 网络状态查询
  - 配置备份/恢复

**使用示例:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t router core-router-01 -- 'show ip route'"
ssh user@roma-jump-server -p 2200 "ln -t router core-router-01 -- 'show interfaces'"
```

### 6. 🔌 交换机

- **连接方式**: SSH (Cisco, Huawei 等)
- **执行方式**: 交换机 CLI 命令
- **功能特性**:
  - 端口管理
  - VLAN 配置
  - MAC 地址表查询
  - 端口状态监控

**使用示例:**
```bash
ssh user@roma-jump-server -p 2200 "ln -t switch access-switch-01 -- 'show vlan'"
ssh user@roma-jump-server -p 2200 "ln -t switch access-switch-01 -- 'show mac address-table'"
```

---

## 🤖 AI MCP 集成

ROMA 通过 Model Context Protocol (MCP) 提供**原生 AI 集成**，允许 AI 助手（Claude、GPT、Cursor 等）直接与您的基础设施交互。

### MCP 架构

ROMA 提供**两种 MCP 集成模式**：

#### 1. MCP Bridge（轻量级，推荐）

一个轻量级桥接器，通过 SSH 将 AI 助手连接到 ROMA 跳板机。

```
AI 助手 (Claude Desktop/Cursor)
        ↓ stdio (JSON-RPC)
MCP Bridge (~5MB 二进制文件)
        ↓ SSH (端口 2200)
ROMA 跳板机
        ↓
目标资源 (Linux/Windows/Docker/数据库/路由器/交换机)
```

**特性:**
- ✅ 轻量级 (~5MB 二进制文件)
- ✅ 基于 SSH 连接（无需 HTTP API）
- ✅ 完整的 ROMA 命令支持 (ln, ls, whoami 等)
- ✅ 自动资源名称匹配
- ✅ 多步执行支持
- ✅ 对话历史感知

**快速开始:**
```bash
# 1. 编译 MCP Bridge
cd mcp/bridge
go build -o roma-mcp-bridge

# 2. 配置 Claude Desktop
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

**文档:** [mcp/bridge/README.md](mcp/bridge/README.md)

#### 2. MCP Server（集成模式）

集成到 ROMA 主服务中的 MCP 服务器（旧版，正在逐步淘汰）。

---

### MCP 工具概览

ROMA MCP Bridge 提供 **20+ 强大的工具**，按类别组织：

#### 📋 ROMA 资源信息查询工具

查询 ROMA 跳板机的资源注册表（不是实际服务器数据）：

- `list_resources` - 列出指定类型的所有资源
- `get_resource_info` - 获取详细的资源配置信息
- `get_current_user` - 获取当前用户信息和权限
- `get_command_history` - 获取 ROMA 命令历史

#### 💻 实际数据查询工具

查询实际服务器/数据库的数据：

- `execute_command` - 在资源上执行 Shell 命令
- `execute_database_query` - 在数据库上执行 SQL 查询
- `execute_commands` - 执行多个命令
- `copy_file_to_resource` / `copy_file_from_resource` - 文件传输 (SCP)

#### 🔧 系统监控工具

常用操作的便捷工具：

- `get_disk_usage` - 磁盘使用情况 (df -h)
- `get_memory_usage` - 内存使用情况 (free -h)
- `get_cpu_info` - CPU 信息 (lscpu)
- `get_process_list` - 进程列表 (ps aux)
- `get_network_info` - 网络信息 (ip addr)
- `get_uptime` - 系统运行时间 (uptime)
- `get_system_info` - 系统详细信息 (uname, os-release)

### AI 使用示例

通过 MCP 集成，您可以使用自然语言控制您的基础设施：

```
💬 "列出所有 Linux 服务器"
💬 "links-mysql 数据库里有哪些数据库？"
💬 "检查 web-server-01 的磁盘使用情况"
💬 "查看 db-01 上的系统日志"
💬 "在所有生产服务器上执行 'df -h'"
💬 "查询 links-mysql 数据库中的 users 表"
💬 "上传文件 config.json 到 server-01 的 /tmp/ 目录"
```

AI 会自动：
1. 选择合适的工具
2. 执行命令/查询
3. 以可读格式呈现结果
4. 处理错误并使用正确的资源名称重试

---

## 🧩 空间隔离与灵活 RBAC

ROMA 使用「空间（Space）」概念取代传统“项目”，让多租户隔离更直观、可视、可审计：

- **默认空间**：内置 `default` 空间，升级旧版本时无需额外迁移即可直接使用。
- **空间成员**：用户必须加入空间并拥有相应角色才能访问其中资源。
- **资源绑定**：所有资源都归属单一空间，super/system 角色可在前后端自由调整归属。
- **资源角色**：可将角色直接挂载到资源实现审批流，如 “ops”、“db-admin” 等。
- **前端体验**：资源管理页面提供空间列、筛选器以及创建/编辑时的空间选择器。

借助空间隔离，你可以在同一套 ROMA 集群里托管多业务单元或不同环境，并保持清晰的边界。

---

## 🛡️ 安全增强亮点

ROMA 现在默认内置多层安全防护，适合面向公网部署：

1. **API 防护**
   - 数据库持久化 IP 黑名单，并自动调用 ipseek.cc 填充地理/运营商信息。
   - 速率限制器：每 IP 并发与 QPS 双限流，抵御 DDoS。
   - 认证失败追踪：根据失败次数自动封禁，且支持指数回退与日志追踪。
2. **SSH 网关保护**
   - 连接速率与并发限制防止暴力破解。
   - API/SSH 共用黑名单，简化安全运营。
   - 详细的安全审计日志，便于排查。
3. **凭证与配置安全**
   - 用户密码使用 Bcrypt，资源密码使用 AES-256-GCM。
   - `config.ex.toml`、Kubernetes、Drone Secrets 中统一注入加解密密钥与 JWT 秘钥。
4. **前端治理**
   - 新增「IP 黑名单」页面，可搜索、添加、解禁并查看 IP 地理信息。
   - 所有资源/空间界面突出访问范围与审计脉络。

这些能力让 ROMA 能够主动抵御常见的暴力破解、撞库、扫段拉爆等攻击手法。

---

## 🚀 快速开始

### 方案 A：Docker 快速启动（推荐）

最快的体验方式 - 无需 git clone！

```bash
# 1. 下载快速启动配置
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. 启动所有服务
docker compose -f quickstart.yaml up -d

# 3. 访问服务
# - Web UI: http://localhost:7000
# - API: http://localhost:6999
# - SSH: localhost:2200
```

**演示账号：**
```
用户名：demo
密码：demo123456
邮箱：test@roma.binrc.com
```

> ⚠️ **安全提示**：生产环境使用前请修改默认密码！

**包含的服务：**
- ✅ ROMA 后端（API + SSH 服务）
- ✅ ROMA Web UI（React 前端）
- ✅ SQLite 数据库（轻量级，无需外部数据库）
- ✅ 预配置的演示账号

**验证安装：**
```bash
# 检查容器状态
docker compose -f quickstart.yaml ps

# 查看日志
docker compose -f quickstart.yaml logs -f

# SSH 连接到 ROMA 跳板机
ssh demo@localhost -p 2200
# 密码：demo123456

# 在 ROMA TUI 中：
roma> ls
roma> whoami
roma> help
```

**自定义配置：**
```bash
# 创建自定义环境文件
cat > .env << EOF
TAG=latest
WEB_PORT=8080
ROMA_SSH_PORT=2200
ROMA_API_PORT=6999
ROMA_USER_1ST_USERNAME=admin
ROMA_USER_1ST_PASSWORD=你的强密码123!
EOF

# 使用自定义配置启动
docker compose -f quickstart.yaml up -d
```

**停止和清理：**
```bash
# 停止服务
docker compose -f quickstart.yaml down

# 删除所有数据（包括数据库）
docker compose -f quickstart.yaml down -v
```

---

### 方案 B：手动安装

```bash
git clone https://github.com/binrchq/roma.git
cd roma
go build -o roma cmd/roma/main.go
```

### 配置

创建 `configs/config.toml`:

```toml
[api]
host = '0.0.0.0'
port = '6999'

[common]
port = '2200'  # SSH 跳板机端口
prompt = 'roma'

[database]
cdb_url = '/usr/local/roma/roma.db'

[apikey]
prefix = 'apikey.'
key = 'your-api-key-here'

[user_1st]
username = 'admin'
email = 'admin@example.com'
password = 'ChangeMe123!'  # ⚠️ 请修改此密码！
public_key = 'ssh-rsa AAAAB3...'  # 您的 SSH 公钥
roles = "super,system,ops"
```

### 启动 ROMA

```bash
./roma -c configs/config.toml
```

ROMA 将启动：
- **SSH 服务** 在端口 2200（跳板机）
- **API 服务** 在端口 6999（RESTful API）

### 通过 SSH 连接

```bash
ssh admin@your-roma-server -p 2200 -i ~/.ssh/your_key
```

您将看到 ROMA TUI，包含以下命令：
- `ls` - 列出资源
- `ln` - 登录到资源
- `use` - 切换资源类型上下文
- `whoami` - 用户信息
- `help` - 命令帮助

### 设置 MCP Bridge（可选）

```bash
# 编译 MCP Bridge
cd mcp/bridge
go build -o roma-mcp-bridge

# 配置环境变量
export ROMA_SSH_HOST="your-roma-server"
export ROMA_SSH_PORT="2200"
export ROMA_SSH_USER="admin"
export ROMA_SSH_KEY="$(cat ~/.ssh/your_private_key)"

# 测试
./roma-mcp-bridge
```

然后配置您的 AI 助手（Claude Desktop、Cursor 等）使用该桥接器。

---

## 🎮 演示与测试

### 在线演示

无需安装即可试用 ROMA：

🌐 **演示地址**：https://roma-demo.binrc.com

**演示凭证：**
- 凭证：***demo/demo123456***
- 只读操作以保证安全
- ⚠️ 演示数据每 24 小时重置一次

---

### 本地演示环境

快速搭建本地演示环境：

```bash
# 1. 下载并启动 ROMA
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml
docker compose -f quickstart.yaml up -d

# 2. 访问 Web UI
open http://localhost:7000

# 3. 使用演示账号登录
# 用户名：demo
# 密码：demo123456
```

**演示账号详情：**

凭证配置在 `deployment/config.toml` 的 `[user_1st]` 部分。

**默认角色：**
- `super` - 完整的管理员访问权限
- `system` - 系统资源管理
- `ops` - 运维和监控
- `ordinary` - 基本资源访问

**可以测试的功能：**

1. **Web UI 功能：**
   - 带资源统计的仪表盘
   - 资源管理（Linux、Windows、Docker、数据库、路由器、交换机）
   - 用户和角色管理（仅超级管理员）
   - 审计日志查看器
   - SSH 密钥管理

2. **SSH 跳板机：**
   ```bash
   # 生成 SSH 密钥（如果还没有）
   ssh-keygen -t rsa -b 4096 -f ~/.ssh/roma_demo_key
   
   # 通过 Web UI 上传公钥：
   # 设置 -> SSH 密钥 -> 上传公钥
   
   # 连接到 ROMA
   ssh demo@localhost -p 2200 -i ~/.ssh/roma_demo_key
   
   # 尝试 ROMA 命令
   roma> ls              # 列出资源
   roma> use linux       # 切换到 Linux 上下文
   roma> ls              # 列出 Linux 资源
   roma> whoami          # 显示用户信息
   roma> help            # 显示所有命令
   ```

3. **API 测试：**
   ```bash
   # 从 Web UI 获取 API 密钥：设置 -> API 密钥
   
   # 测试 API
   curl -H "apikey: your-api-key" http://localhost:6999/api/v1/resources
   ```

4. **MCP Bridge（AI 集成）：**
   ```bash
   # 编译 MCP bridge
   cd mcp/bridge
   go build -o roma-mcp-bridge
   
   # 配置本地演示环境
   export ROMA_SSH_HOST="localhost"
   export ROMA_SSH_PORT="2200"
   export ROMA_SSH_USER="demo"
   export ROMA_SSH_KEY="$(cat ~/.ssh/roma_demo_key)"
   
   # 测试
   ./roma-mcp-bridge
   ```

**示例资源（演示环境预配置）：**

演示环境包含用于测试的示例资源：
- 📦 Linux 服务器（web-01、db-01）
- 🐳 Docker 容器
- 🗄️ MySQL 数据库（demo-db）
- 🛣️ 网络设备（router-01、switch-01）

**清理演示环境：**
```bash
# 停止并删除容器
docker compose -f quickstart.yaml down

# 删除卷（可选，删除所有数据）
docker compose -f quickstart.yaml down -v
```

---

### 高级选项：使用 MySQL/PostgreSQL 的 Docker 部署

用于生产环境的外部数据库部署：

```bash
# 克隆仓库
git clone https://github.com/binrchq/roma.git
cd roma/deployment

# 选项 1：MySQL
docker compose -f quickstart.mysql.yaml up -d

# 选项 2：PostgreSQL
docker compose -f quickstart.pgsql.yaml up -d
```

更多配置选项请查看 [deployment/](deployment/) 目录。

---

## 📚 文档

- **[MCP Bridge 指南](mcp/bridge/README.md)** - 完整的 MCP Bridge 文档
- **[MCP Bridge 架构](mcp/bridge/ARCHITECTURE.md)** - 架构详情
- **[资源支持说明](docs/RESOURCE_SUPPORT.md)** - 详细的资源类型支持
- **[API 文档](docs/API.md)** - RESTful API 参考

---

## 🔗 相关项目

ROMA 生态系统包含多个项目，适用于不同场景：

### 🌐 [roma-web](https://github.com/binrchq/roma-web)
基于 React 的现代化 Web 管理界面。

**功能特性：**
- 📊 实时统计的资源仪表盘
- 🖥️ 基于 Web 的 SSH 终端
- 👥 用户和角色管理
- 🔑 SSH 密钥管理
- 📝 审计日志查看器
- 🎨 现代化、响应式设计

**快速开始：**
```bash
docker pull binrc/roma-web:latest
# 或访问：https://github.com/binrchq/roma-web
```

---

### 🤖 [roma-mcp](https://github.com/binrchq/roma-mcp)
独立的 MCP 服务器，用于 AI 集成（MCP Bridge 的替代方案）。

**功能特性：**
- 🔌 完整的 MCP 协议支持
- 🚀 独立部署
- 🛠️ 20+ AI 工具用于基础设施管理
- 💡 兼容 Claude Desktop、Cursor 等 MCP 客户端

**使用场景：**
- 需要独立的 MCP 服务器
- 希望在不同机器上运行 MCP 服务器
- 需要自定义 MCP 配置

**快速开始：**
```bash
git clone https://github.com/binrchq/roma-mcp.git
cd roma-mcp
go build -o roma-mcp-server
./roma-mcp-server
```

### 📊 项目对比

| 项目 | 用途 | 技术栈 | 部署方式 |
|------|------|--------|---------|
| **roma** | 核心跳板机 | Go | 二进制/Docker |
| **roma-web** | Web 管理界面 | React | Docker/Nginx |
| **roma-mcp** | 独立 MCP 服务器 | Go | 二进制/Docker |

---

## 🎯 使用场景

### 1. 安全远程访问

无需直接暴露所有服务器：
- 部署 ROMA 作为跳板机
- 用户只连接到 ROMA
- ROMA 管理到内部资源的连接
- 所有访问都被记录和审计

### 2. AI 驱动运维

使用 AI 助手来：
- 自动化日常运维操作
- 查询基础设施状态
- 在多台服务器上执行命令
- 生成报告和摘要

### 3. 多资源管理

从一个地方管理多样化的基础设施：
- Linux 服务器
- Windows 服务器
- Docker 容器
- 数据库 (MySQL, PostgreSQL, Redis 等)
- 网络设备 (路由器, 交换机)

### 4. 团队协作

- 集中式凭证管理
- 基于角色的访问控制
- 审计日志用于合规
- Web UI 供非技术用户使用

---

## 🔐 安全特性

- ✅ **SSH 密钥认证** - 无密码认证
- ✅ **API Key 授权** - 安全的 API 访问
- ✅ **基于角色的访问控制 (RBAC)** - 细粒度权限
- ✅ **审计日志** - 所有操作都被记录
- ✅ **凭证加密** - 安全的凭证存储
- ✅ **会话管理** - 跟踪和管理活动会话

---

## 🌐 Web 管理界面

ROMA 包含一个使用 React 构建的现代化 Web UI：

- 📊 带有资源统计的仪表盘
- 🖥️ 资源管理（CRUD 操作）
- 👥 用户和角色管理
- 💻 Web SSH 终端
- 📝 审计日志查看器

**启动 Web UI:**
```bash
cd web/frontend
npm install
npm run dev
# 访问 http://localhost:3000
```

---

## 📦 项目结构

```
roma/
├── cmd/roma/              # 主程序入口
├── core/                  # 核心功能
│   ├── api/              # API 控制器
│   ├── model/            # 数据模型
│   ├── operation/        # 业务逻辑
│   ├── connect/          # 连接处理器
│   ├── tui/              # 终端 UI (SSH 命令)
│   └── constants/        # 常量 (资源类型等)
├── mcp/                  # MCP 集成
│   └── bridge/           # MCP Bridge (轻量级)
│       ├── mappings/     # 工具映射 (已组织)
│       ├── main.go       # Bridge 入口点
│       └── client.go      # ROMA SSH 客户端
├── web/                  # Web 组件
│   ├── frontend/         # React 前端
│   ├── vscode-extension/ # VSCode 扩展
│   └── ops-client/       # Electron 桌面应用
├── configs/              # 配置文件
└── docs/                # 文档
```

---

## 🔗 许可证

本项目基于 **GNU Affero General Public License (AGPL) v3.0** 开源发布。

📢 **重要**: 任何组织或个人修改 ROMA 代码用于提供**远程访问服务**时，必须**开源其修改版本**。

详情请查看 [LICENSE](./LICENSE) 文件。

---

## 🤝 贡献

欢迎贡献！请阅读我们的贡献指南和行为准则。

---

## 📞 支持

- 📧 邮箱: support@binrc.com
- 🐛 问题: [GitHub Issues](https://github.com/binrchq/roma/issues)
- 📖 文档: [docs/](docs/)

---

**ROMA** - 为远程访问提供无缝解决方案，确保效率和安全性。 🚀
