# ROMA - AI驱动的堡垒机

<div align="left">
  <img src="./readme.res/logo.png" alt="ROMA Logo" width="100" />
</div>

![License](https://img.shields.io/badge/License-AGPL_v3-blue)
![Lightweight](https://img.shields.io/badge/lightweight-green)
![AI-Powered](https://img.shields.io/badge/AI-Powered-orange)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)

**ROMA** 是一个基于Go开发的AI驱动、超轻量级堡垒机（跳板机）系统，通过Model Context Protocol (MCP) 原生集成AI能力，为您的基础设施提供安全高效的远程访问解决方案。

**在线演示:** https://roma-demo.binrc.com (demo/demo123456)

语言: [English](./README.md) • 中文

---

## 关联项目

| 项目 | 说明 | 仓库 |
|------|------|------|
| **roma** | 核心堡垒机服务 (Go) | 本项目 |
| **roma-web** | Web管理界面 (React) | [github.com/binrchq/roma-web](https://github.com/binrchq/roma-web) |
| **roma-mcp** | 独立MCP服务 | [github.com/binrchq/roma-mcp](https://github.com/binrchq/roma-mcp) |
| **roma-vsc-ext** | VSCode扩展 | [github.com/binrchq/roma-vsc-ext](https://github.com/binrchq/roma-vsc-ext) |

**官方网站:** https://roma.binrc.com

---

<div align="left">
  <img src="./readme.res/face.png" alt="ROMA Interface" width="800" />
</div>

## 核心特性

- **堡垒机** - 统一远程访问入口，集中管控
- **AI驱动** - 原生MCP支持，AI助手直接管理基础设施
- **空间隔离** - 多租户级别的资源隔离
- **安全强化** - SSH密钥认证、API密钥授权、多层防护
- **轻量级** - 单二进制文件，最小依赖
- **多资源支持** - Linux/Windows/Docker/数据库/路由器/交换机
- **现代化Web UI** - React构建的管理界面
- **自适应安全** - 速率限制、IP黑名单、认证失败防护
- **MCP Bridge** - 轻量级AI集成桥接

---

## 快速部署

### Docker部署 (推荐)

```bash
# 1. 下载快速启动配置
curl -O https://raw.githubusercontent.com/binrchq/roma/main/deployment/quickstart.yaml

# 2. 启动服务
docker compose -f quickstart.yaml up -d

# 3. 访问Web界面
open http://localhost:7000
```

**默认凭据:**
- 用户名: `demo`
- 密码: `demo123456`

**服务端口:**
- Web UI: `7000`
- API: `6999`
- SSH: `2200`

### 二进制部署

```bash
# 1. 克隆仓库
git clone https://github.com/binrchq/roma.git
cd roma

# 2. 编译
go build -o roma cmd/roma/main.go

# 3. 配置 (参考 configs/config.ex.toml)
cp configs/config.ex.toml configs/config.toml
vim configs/config.toml

# 4. 启动
./roma -c configs/config.toml
```

### 生产环境部署

支持MySQL/PostgreSQL数据库：

```bash
# MySQL
docker compose -f deployment/quickstart.mysql.yaml up -d

# PostgreSQL
docker compose -f deployment/quickstart.pgsql.yaml up -d
```

**部署说明详见:** [docs/DEPLOYMENT_CN.md](docs/DEPLOYMENT_CN.md)

---

## 功能使用

### SSH命令行

连接到ROMA堡垒机后，可使用类似kubectl的命令管理资源：

```bash
# 连接到堡垒机
ssh demo@localhost -p 2200

# 列出资源 (类似 kubectl get)
roma> ls                    # 列出当前类型的所有资源
roma> ls linux              # 列出所有Linux服务器
roma> ls database           # 列出所有数据库

# 切换上下文 (类似 kubectl use-context)
roma> use linux             # 切换到Linux上下文
roma> use database          # 切换到数据库上下文

# 登录资源 (ln = login)
roma> ln web-server-01                          # 交互式登录
roma> ln -t linux web-01 -- 'df -h'            # 执行单条命令
roma> ln -t database mysql-prod -- 'SHOW databases;'  # 数据库查询

# 用户信息
roma> whoami                # 显示当前用户和权限

# 帮助
roma> help                  # 显示所有可用命令
```

### 文件传输 (SCP)

ROMA支持标准SCP协议进行文件传输，使用特殊的路径格式通过堡垒机中转：

**路径格式:** `user@jumpserver:user@hostname:/remote/path`

**上传文件到服务器:**

```bash
# 基本用法
scp -P 2200 /local/file.txt user@roma-server:user@web-server-01:/tmp/

# 使用SSH密钥
scp -P 2200 -i ~/.ssh/roma_key /local/config.json user@roma-server:user@web-server-01:/etc/app/

# 示例：上传日志文件
scp -P 2200 -i ~/.ssh/id_rsa /var/log/app.log demo@localhost:demo@web-01:/tmp/app.log
```

**从服务器下载文件:**

```bash
# 基本用法
scp -P 2200 user@roma-server:user@web-server-01:/tmp/file.txt /local/path/

# 下载配置文件
scp -P 2200 -i ~/.ssh/roma_key user@roma-server:user@db-01:/etc/mysql/my.cnf ./backup/

# 示例：下载数据库备份
scp -P 2200 -i ~/.ssh/id_rsa demo@localhost:demo@db-01:/backup/db.sql.gz ./
```

**支持的资源类型:**
- Linux服务器
- Windows服务器（需要OpenSSH Server）
- 暂不支持文件夹传输（可以先压缩再传输）

**路径解析说明:**
- `user@jumpserver` - ROMA堡垒机的用户和地址
- `user@hostname` - 目标服务器的用户和主机名（hostname需要在ROMA中注册）
- `/remote/path` - 目标服务器上的文件路径

**通过MCP进行文件传输:**

AI助手可以使用内置的文件传输工具：

示例：
```
"上传配置文件 config.json 到 web-server-01 的 /etc/app/ 目录"
"从 db-01 下载 /backup/db.sql.gz 到本地"
```

MCP工具：
- `copy_file_to_resource` - 上传文件
- `copy_file_from_resource` - 下载文件

### MCP集成 (AI助手)

ROMA提供轻量级MCP Bridge，让AI助手直接管理基础设施：

**1. 构建MCP Bridge:**

```bash
cd mcp/bridge
go build -o roma-mcp-bridge
```

**2. 配置Claude Desktop:**

编辑 `~/.config/Claude/claude_desktop_config.json`:

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

**3. 使用AI助手:**

示例命令：
```
"列出所有Linux服务器"
"检查web-01的磁盘使用情况"
"在生产环境数据库中查询用户表"
"上传配置文件到服务器"
"显示所有容器的运行状态"
```

**MCP工具清单:**

| 分类 | 工具 | 说明 |
|------|------|------|
| 资源查询 | `list_resources` | 列出资源 |
| | `get_resource_info` | 获取资源详情 |
| | `get_current_user` | 获取当前用户 |
| 命令执行 | `execute_command` | 执行Shell命令 |
| | `execute_database_query` | 执行SQL查询 |
| | `execute_commands` | 批量执行命令 |
| 文件传输 | `copy_file_to_resource` | 上传文件 |
| | `copy_file_from_resource` | 下载文件 |
| 系统监控 | `get_disk_usage` | 磁盘使用 |
| | `get_memory_usage` | 内存使用 |
| | `get_cpu_info` | CPU信息 |
| | `get_process_list` | 进程列表 |
| | `get_network_info` | 网络信息 |
| | `get_system_info` | 系统信息 |

**详细文档:** [mcp/bridge/README.md](mcp/bridge/README.md)

---

## 安全说明

ROMA提供多层安全防护，适合生产环境和互联网部署：

### 认证与授权

- **SSH密钥认证** - 禁用密码登录
- **API密钥授权** - 安全的API访问控制
- **基于角色的访问控制 (RBAC)** - 细粒度权限管理
- **空间隔离** - 多租户级别资源隔离

### 凭据安全

- **Bcrypt密码哈希** - 用户密码加密存储
- **AES-256-GCM加密** - 资源凭据加密
- **密钥轮转** - 支持定期更换加密密钥
- **JWT令牌** - 安全的会话管理

### 防护机制

- **IP黑名单** - 全局IP封禁（支持地理位置查询）
- **速率限制** - 每IP并发和QPS限制
- **认证失败追踪** - 自动封禁暴力破解
- **连接限流** - SSH和API层统一防护
- **审计日志** - 所有操作可追溯

### 网络安全

- **防火墙建议** - 仅暴露必要端口
- **VPN集成** - 支持VPN后端访问
- **TLS/SSL** - HTTPS和加密传输
- **DDoS防护** - 连接限流和IP封禁

### 安全最佳实践

1. **修改默认凭据** - 部署后立即修改默认密码
2. **使用强密码** - 密码长度 ≥ 12位，包含大小写字母、数字和特殊字符
3. **定期更新** - 保持ROMA和依赖组件最新版本
4. **监控审计日志** - 定期检查异常访问行为
5. **最小权限原则** - 仅授予必要的角色和权限
6. **网络隔离** - ROMA部署在隔离网络，限制访问来源

**安全配置指南:** [docs/SECURITY_CN.md](docs/SECURITY_CN.md)

---

## 支持的资源类型

| 类型 | 协议 | 功能 |
|------|------|------|
| Linux | SSH | Shell命令、文件传输 |
| Windows | WinRM | PowerShell命令 |
| Docker | Docker CLI | 容器管理、日志查看 |
| 数据库 | Native | SQL查询 (MySQL/PostgreSQL/Redis/MongoDB等) |
| 路由器 | SSH | 路由器CLI命令 |
| 交换机 | SSH | 交换机CLI命令 |

**详细支持说明:** [docs/RESOURCE_SUPPORT_CN.md](docs/RESOURCE_SUPPORT_CN.md)

---

## 文档

| 文档 | 说明 |
|------|------|
| [DEPLOYMENT_CN.md](docs/DEPLOYMENT_CN.md) | 部署指南（Docker/K8s/二进制） |
| [DEVELOPMENT_CN.md](docs/DEVELOPMENT_CN.md) | 开发指南（架构/贡献/调试） |
| [SECURITY_CN.md](docs/SECURITY_CN.md) | 安全配置和最佳实践 |
| [SCP_USAGE_CN.md](docs/SCP_USAGE_CN.md) | SCP文件传输详细指南 |
| [API_CN.md](docs/API_CN.md) | RESTful API文档 |
| [RESOURCE_SUPPORT_CN.md](docs/RESOURCE_SUPPORT_CN.md) | 资源类型详细说明 |
| [MCP_BRIDGE_CN.md](mcp/bridge/README_CN.md) | MCP Bridge使用指南 |
| [MCP_ARCHITECTURE_CN.md](mcp/bridge/ARCHITECTURE_CN.md) | MCP架构设计 |

---

## 使用场景

- **安全远程访问** - 统一入口，集中管控，全程审计
- **AI驱动运维** - AI助手自动化日常运维任务
- **多资源管理** - 一站式管理服务器、数据库、网络设备
- **团队协作** - 集中凭据管理，基于角色的权限控制

---

## 支持

- Email: support@binrc.com
- Issues: [GitHub Issues](https://github.com/binrchq/roma/issues)
- 官方网站: https://roma.binrc.com

---

## 开源协议

本项目采用双重许可证:
- **GNU Affero General Public License (AGPL) v3.0**
- **商业软件许可协议**

**重要**: 任何组织或个人修改ROMA代码并提供**远程访问服务**时，必须**开源其修改版本**。

详见 [LICENSE](./LICENSE)

---

## 贡献

欢迎贡献！请阅读 [DEVELOPMENT_CN.md](docs/DEVELOPMENT_CN.md) 了解如何参与开发。

---

## 组织支持

ROMA由以下组织支持开发：

<p align="left" style="">
  <a href="https://binrc.com" target="_blank" style="display: inline-block; vertical-align: middle; margin-right: 20px;">
    <img src="https://binrc.com/img/logo_lite.png" alt="Binrc" height="40" />
  </a>
  <a href="https://ai2o.binrc.com" target="_blank" style="display: inline-block; vertical-align: middle;">
    <img src="docs/AI2O_logo_long.png" alt="AI2O" height="40"  style="image-rendering: auto;"/>
  </a>
</p>

---

## 贡献者

感谢所有为ROMA做出贡献的开发者：

<a href="https://github.com/binrchq/roma/graphs/contributors">
  <img src="https://avatars.githubusercontent.com/u/37877444?v=4" alt="Contributor" width="60" height="60" style="border-radius: 50%;" />
</a>

---

## 相关产品

### ROMC - AI驱动的运维自动化平台

ROMC 是 Binrc 开发的运维自动化工具，集成MCP协议的智能终端AI助手。

**核心特性:**
- **终端AI助手** - 自然语言交互，智能理解运维意图
- **MCP原生集成** - 无缝对接ROMA和其他基础设施
- **智能决策** - 基于AI的故障诊断和自动修复
- **可视化运维** - 直观的运维数据展示和分析
- **工作流自动化** - 可编排的运维流程自动化

即将发布。

了解更多: https://binrc.com

---

**ROMA** - 安全高效的远程访问解决方案
