# ROMA 资源类型支持说明

## 📦 支持的资源类型概览

ROMA 跳板机完整支持 **6 种资源类型**，每种资源都支持完整的 CRUD 操作和 MCP 自动化管理。

| 资源类型 | 状态 | CRUD | MCP | SSH执行 | 特色功能 |
|---------|------|------|-----|---------|---------|
| 🐧 Linux | ✅ 完整支持 | ✅ | ✅ | ✅ | 系统监控、批量操作 |
| 🪟 Windows | ✅ 完整支持 | ✅ | ✅ | ✅ | PowerShell执行、服务管理 |
| 🐳 Docker | ✅ 完整支持 | ✅ | ✅ | ✅ | 容器管理、镜像操作 |
| 🗄️ Database | ✅ 完整支持 | ✅ | ✅ | ✅ | SQL查询（MySQL/PG） |
| 🌐 Router | ✅ 完整支持 | ✅ | ✅ | ✅ | 路由表、接口配置 |
| 🔌 Switch | ✅ 完整支持 | ✅ | ✅ | ✅ | VLAN、端口管理 |

**图例说明**：
- ✅ 完整支持
- ⏳ 规划中
- ❌ 不支持

---

## 1️⃣ Linux 服务器

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateLinuxResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateLinuxResource()`
- 删除：`DeleteLinuxResource()`

✅ **MCP 工具支持**
- `list_resources` - 列出所有 Linux 服务器
- `get_resource` - 获取服务器详情
- `add_resource` - 添加新服务器
- `update_resource` - 更新服务器配置
- `delete_resource` - 删除服务器

✅ **SSH 远程执行**
- `execute_command` - 执行任意命令
- `get_system_info_ssh` - 获取系统信息（CPU、内存、磁盘、网络、进程）
- `check_resource_health` - 健康检查
- `batch_execute_command` - 批量执行
- `get_disk_usage` - 磁盘使用情况
- `get_process_list` - 进程列表

### 资源配置示例

```json
{
  "hostname": "web-01",
  "ip": "192.168.1.100",
  "port": 22,
  "tags": ["web", "production"],
  "description": "Web 服务器"
}
```

### 使用示例

```bash
# TUI 命令
roma> use linux
roma> ls linux
roma> ln web-01

# MCP / AI 命令
AI: 列出所有 Linux 服务器
AI: 在 web-01 上执行 df -h
AI: 获取 web-01 的系统负载
AI: 批量检查所有 web 服务器的磁盘空间
```

---

## 2️⃣ Windows 服务器

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateWindowsResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateWindowsResource()`
- 删除：`DeleteWindowsResource()`

✅ **MCP 工具支持**
- 基础资源管理（list/get/add/update/delete）

⏳ **待支持功能**
- PowerShell 远程执行
- WinRM 连接
- RDP 会话管理
- Windows 系统监控

### 资源配置示例

```json
{
  "hostname": "win-server-01",
  "ip": "192.168.1.200",
  "port": 3389,
  "username": "administrator",
  "tags": ["windows", "production"]
}
```

---

## 3️⃣ Docker 容器

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateDockerResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateDockerResource()`
- 删除：`DeleteDockerResource()`

✅ **MCP 工具支持**
- 基础资源管理（list/get/add/update/delete）

⏳ **待支持功能**
- Docker 容器启动/停止
- 容器日志查看
- 容器 exec 命令执行
- 镜像管理
- Docker Compose 支持

### 资源配置示例

```json
{
  "container_name": "nginx-web",
  "image": "nginx:latest",
  "host_ip": "192.168.1.100",
  "port": 2375,
  "tags": ["docker", "web"]
}
```

---

## 4️⃣ 数据库

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateDatabaseResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateDatabaseResource()`
- 删除：`DeleteDatabaseResource()`

✅ **MCP 工具支持**
- 基础资源管理（list/get/add/update/delete）

✅ **支持的数据库类型**
- MySQL
- PostgreSQL
- MongoDB
- Redis
- Oracle
- SQL Server

⏳ **待支持功能**
- SQL 查询执行
- 数据库备份
- 慢查询分析
- 连接池管理

### 资源配置示例

```json
{
  "database_nick": "prod-mysql",
  "database_type": "mysql",
  "host": "192.168.1.50",
  "port": 3306,
  "database": "myapp",
  "username": "admin",
  "tags": ["database", "mysql", "production"]
}
```

---

## 5️⃣ 路由器

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateRouterResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateRouterResource()`
- 删除：`DeleteRouterResource()`

✅ **MCP 工具支持**
- 基础资源管理（list/get/add/update/delete）

⏳ **待支持功能**
- 路由表查看
- 路由配置管理
- 接口状态监控
- SNMP 监控

### 资源配置示例

```json
{
  "router_name": "core-router-01",
  "ip": "192.168.1.1",
  "port": 22,
  "model": "Cisco ISR 4000",
  "tags": ["router", "core", "network"]
}
```

---

## 6️⃣ 交换机

### 支持功能

✅ **完整 CRUD 操作**
- 创建：`CreateSwitchResource()`
- 读取：`GetResourceListByRoleId()`
- 更新：`UpdateSwitchResource()`
- 删除：`DeleteSwitchResource()`

✅ **MCP 工具支持**
- 基础资源管理（list/get/add/update/delete）

⏳ **待支持功能**
- 端口配置管理
- VLAN 管理
- 端口状态监控
- MAC 地址表查询

### 资源配置示例

```json
{
  "switch_name": "access-switch-01",
  "ip": "192.168.1.2",
  "port": 22,
  "model": "Cisco Catalyst 3850",
  "ports": 48,
  "tags": ["switch", "access", "network"]
}
```

---

## 🔄 通用 CRUD API

所有资源类型都支持统一的 CRUD 接口：

### 1. 创建资源

**API**: `POST /api/resource/add`

```json
{
  "type": "linux",
  "data": [
    {
      "hostname": "web-01",
      "ip": "192.168.1.100",
      "port": 22
    }
  ]
}
```

### 2. 查询资源

**MCP 工具**: `list_resources`

```json
{
  "resource_type": "linux",
  "role_name": "ops"
}
```

### 3. 更新资源

**API**: `POST /api/resource/update`

```json
{
  "type": "linux",
  "data": [
    {
      "hostname": "web-01",
      "ip": "192.168.1.101"
    }
  ]
}
```

### 4. 删除资源

**MCP 工具**: `delete_resource`

```json
{
  "resource_type": "linux",
  "identifier": "web-01"
}
```

---

## 🚀 路线图

### 近期计划（Q1 2025）

- [ ] Windows PowerShell 远程执行
- [ ] Docker 容器管理命令
- [ ] 数据库查询执行工具

### 中期计划（Q2 2025）

- [ ] 路由器配置管理
- [ ] 交换机端口管理
- [ ] 统一监控面板

### 长期计划

- [ ] Kubernetes 集群管理
- [ ] 云平台资源集成（AWS/Azure/阿里云）
- [ ] 自动化编排和工作流

---

## 📚 相关文档

- [资源模型定义](../core/model/)
- [资源操作实现](../core/operation/resource_operation.go)
- [MCP 工具说明](../mcp/FEATURES.md)
- [API 文档](../core/api/resource_control.go)

---

## 💬 反馈与贡献

如果你需要其他资源类型支持，欢迎：
1. 提交 Issue
2. 贡献代码
3. 参与讨论

**项目地址**: https://github.com/binrchq/roma

