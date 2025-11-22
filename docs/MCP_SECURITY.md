# MCP 安全机制文档

## 🔒 概述

ROMA MCP 实现了完整的身份验证和权限控制机制，确保所有 MCP 操作都经过严格的安全检查。

## 核心安全特性

### 1. 身份验证

所有 MCP 请求都必须包含有效的认证令牌。

#### 支持的令牌类型

| 令牌类型 | 格式 | 用途 | 有效期 |
|---------|-----|------|--------|
| **API Key** | `roma_xxxxx` | 长期使用 | 永久（可撤销） |
| **MCP Token** | `mcp_xxxxx` | 临时访问 | 可配置（如 24h） |

#### 令牌验证流程

```
客户端请求
    ↓
提取令牌（从 auth 字段或 params）
    ↓
验证令牌有效性
    ↓
查询用户信息
    ↓
检查用户状态（是否禁用）
    ↓
加载用户角色
    ↓
创建 MCPContext
```

### 2. 权限控制

#### 权限模型

```
用户 (User)
  ↓
角色 (Roles)  [admin, operator, developer, viewer]
  ↓
资源权限 (Resource Permissions)
  ↓
操作类型 (Actions)  [read, execute, write]
```

#### 权限矩阵

| 角色 | 查看资源 | 执行命令 | 修改资源 | 管理用户 |
|-----|---------|----------|---------|---------|
| **admin** | ✅ | ✅ | ✅ | ✅ |
| **operator** | ✅ | ✅ | ❌ | ❌ |
| **developer** | ✅ | ✅ | ❌ | ❌ |
| **viewer** | ✅ | ❌ | ❌ | ❌ |

### 3. 资源过滤

用户只能看到自己有权限访问的资源。

```go
// 示例：普通用户列出资源
请求：ListResources
响应：只返回该用户有权限的资源

// 管理员列出资源
请求：ListResources
响应：返回所有资源
```

### 4. 审计日志

所有 MCP 操作都会记录到 `access_logs` 表。

#### 日志字段

```sql
CREATE TABLE access_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,           -- 操作用户
    resource_id INT,                -- 目标资源
    action VARCHAR(100),            -- 操作类型
    status VARCHAR(20),             -- 成功/失败
    details TEXT,                   -- 详细信息
    source_ip VARCHAR(50),          -- 来源（mcp）
    accessed_at TIMESTAMP           -- 访问时间
);
```

#### 审计日志示例

```json
{
  "user_id": 5,
  "username": "alice",
  "resource_id": 123,
  "action": "execute_command",
  "status": "success",
  "details": "command=df -h",
  "source_ip": "mcp",
  "accessed_at": "2024-01-15T10:30:00Z"
}
```

## 🚀 使用指南

### 1. 生成 MCP 令牌

#### 方法 A：使用现有 API Key

用户的 API Key 可以直接用作 MCP 令牌：

```bash
# 查看用户 API Key
curl http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <jwt_token>"

# 响应包含 api_key 字段
{
  "username": "alice",
  "api_key": "roma_abc123...",
  ...
}
```

#### 方法 B：生成专用 MCP Token

```bash
# 生成 24 小时有效的 MCP Token
curl -X POST http://localhost:8080/api/v1/mcp/tokens \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "expires_in": "24h",
    "description": "我的运维客户端"
  }'

# 响应
{
  "token": "mcp_xyz789...",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

**⚠️ 重要：令牌只在创建时返回一次，请妥善保存！**

### 2. MCP 请求格式

#### 标准格式（推荐）

```json
{
  "method": "list_resources",
  "params": {
    "resource_type": "linux"
  },
  "auth": {
    "type": "bearer",
    "token": "roma_abc123..."
  }
}
```

#### 简化格式（兼容）

```json
{
  "method": "list_resources",
  "params": {
    "resource_type": "linux",
    "_auth": "roma_abc123..."
  }
}
```

或

```json
{
  "method": "list_resources",
  "params": {
    "resource_type": "linux",
    "_token": "roma_abc123..."
  }
}
```

### 3. 权限不足时的响应

```json
{
  "success": false,
  "error": {
    "code": "PERMISSION_DENIED",
    "message": "权限不足"
  }
}
```

### 4. 令牌无效时的响应

```json
{
  "success": false,
  "error": {
    "code": "INVALID_TOKEN",
    "message": "无效的认证令牌"
  }
}
```

## 🛡️ 安全最佳实践

### 1. 令牌管理

```bash
✅ DO:
- 使用短期令牌（24小时）进行日常操作
- 定期轮换 API Key
- 为不同用途创建不同的令牌
- 妥善保存令牌，不要提交到代码库

❌ DON'T:
- 将令牌硬编码在代码中
- 使用同一个令牌在多个地方
- 令牌泄露后不及时撤销
```

### 2. 权限分配

```bash
✅ DO:
- 遵循最小权限原则
- 定期审查用户权限
- 使用角色而不是直接分配资源
- 为临时任务创建临时用户

❌ DON'T:
- 给所有人管理员权限
- 长期使用高权限账号
- 跨团队共享账号
```

### 3. 审计监控

```bash
✅ DO:
- 定期检查审计日志
- 监控失败的访问尝试
- 设置异常告警
- 保留足够长的日志历史

❌ DON'T:
- 忽略安全告警
- 删除审计日志
- 允许用户修改自己的日志
```

## 📊 示例场景

### 场景 1：开发人员查询资源

```bash
# 开发人员 Bob 尝试列出 Linux 服务器
请求：
{
  "method": "list_resources",
  "params": { "resource_type": "linux" },
  "auth": { "token": "roma_bob123" }
}

# 系统处理：
1. 验证 token "roma_bob123" → 用户 Bob
2. 查询 Bob 的角色 → [developer]
3. 查询 Bob 有权限的资源 → [web-01, web-02]
4. 过滤资源列表
5. 记录审计日志

# 响应：只返回 web-01 和 web-02
{
  "user": "bob",
  "resource_type": "linux",
  "count": 2,
  "resources": [
    { "id": 1, "name": "web-01", ... },
    { "id": 2, "name": "web-02", ... }
  ]
}
```

### 场景 2：运维人员执行命令

```bash
# 运维人员 Alice 执行命令
请求：
{
  "method": "execute_command",
  "params": {
    "resource_id": 123,
    "command": "systemctl restart nginx"
  },
  "auth": { "token": "roma_alice456" }
}

# 系统处理：
1. 验证 token → 用户 Alice
2. 查询 Alice 的角色 → [operator]
3. 检查 Alice 是否有资源 123 的权限 → ✅
4. 检查 Alice 是否有执行权限 → ✅ (operator 可执行)
5. 执行命令
6. 记录审计日志（包含命令内容）

# 响应：
{
  "user": "alice",
  "resource_id": 123,
  "command": "systemctl restart nginx",
  "output": "...",
  "exit_code": 0
}
```

### 场景 3：查看者尝试执行命令（被拒绝）

```bash
# 查看者 Charlie 尝试执行命令
请求：
{
  "method": "execute_command",
  "params": {
    "resource_id": 123,
    "command": "rm -rf /"
  },
  "auth": { "token": "roma_charlie789" }
}

# 系统处理：
1. 验证 token → 用户 Charlie
2. 查询 Charlie 的角色 → [viewer]
3. 检查执行权限 → ❌ (viewer 不能执行)
4. 记录失败的访问尝试

# 响应：
{
  "success": false,
  "error": {
    "code": "PERMISSION_DENIED",
    "message": "权限不足"
  }
}

# 审计日志：
{
  "user_id": 7,
  "username": "charlie",
  "resource_id": 123,
  "action": "execute_command",
  "status": "failed",
  "details": "command=rm -rf /, denied",
  "source_ip": "mcp"
}
```

## 🔧 配置

### ROMA 配置文件

```toml
# config.toml

[mcp]
enable = true
# 是否强制身份验证
require_auth = true
# 令牌默认过期时间
token_expiry = "24h"
# 是否记录所有 MCP 操作
audit_all = true
```

### 环境变量

```bash
# 强制启用 MCP 认证
export ROMA_MCP_REQUIRE_AUTH=true

# MCP 令牌过期时间
export ROMA_MCP_TOKEN_EXPIRY=24h
```

## 📈 监控和告警

### 关键指标

```sql
-- 失败的认证尝试
SELECT COUNT(*) 
FROM access_logs 
WHERE source_ip = 'mcp' 
  AND status = 'failed' 
  AND accessed_at > NOW() - INTERVAL 1 HOUR;

-- 高危操作统计
SELECT user_id, COUNT(*) as count
FROM access_logs
WHERE action IN ('execute_command', 'delete_resource', 'update_resource')
  AND accessed_at > NOW() - INTERVAL 1 DAY
GROUP BY user_id
ORDER BY count DESC;

-- 权限拒绝统计
SELECT user_id, resource_id, COUNT(*) as denied_count
FROM access_logs
WHERE status = 'failed' 
  AND details LIKE '%denied%'
  AND accessed_at > NOW() - INTERVAL 1 DAY
GROUP BY user_id, resource_id
HAVING COUNT(*) > 10;  -- 10次以上可能是攻击
```

### 告警规则

```bash
# 1. 认证失败次数过多
IF failed_auth_count > 10 IN 5 minutes
THEN alert "可能的暴力破解攻击"

# 2. 权限拒绝次数过多
IF permission_denied > 20 IN 10 minutes
THEN alert "用户尝试非授权访问"

# 3. 高危命令执行
IF command CONTAINS "rm -rf" OR "DROP TABLE"
THEN alert "高危命令执行" AND require_approval

# 4. 异常访问时间
IF access_time BETWEEN 02:00 AND 06:00
THEN alert "非工作时间访问"
```

## ��� 常见问题

### Q1: 如何为 AI 客户端配置认证？

```javascript
// web/ops-client/js/ai-assistant.js
const mcpClient = {
  auth: {
    token: localStorage.getItem('mcp_token') || 'roma_default'
  },
  
  async callTool(toolName, params) {
    return await fetch('/mcp', {
      method: 'POST',
      body: JSON.stringify({
        method: toolName,
        params: params,
        auth: this.auth
      })
    })
  }
}
```

### Q2: 令牌过期了怎么办？

```bash
# 自动刷新令牌
if (error.code === 'TOKEN_EXPIRED') {
  const newToken = await refreshToken()
  localStorage.setItem('mcp_token', newToken)
  // 重试请求
}
```

### Q3: 如何撤销泄露的令牌？

```bash
# 方法 1：通过 API 撤销
curl -X DELETE http://localhost:8080/api/v1/mcp/tokens/mcp_xyz789

# 方法 2：数据库直接删除
DELETE FROM mcp_tokens WHERE token = '<hashed_token>';

# 方法 3：禁用用户
UPDATE users SET status = 'disabled' WHERE id = <user_id>;
```

### Q4: 如何批量授权资源？

```bash
# 给用户授权多个资源
curl -X POST http://localhost:8080/api/v1/users/5/resources \
  -d '{
    "resource_ids": [1, 2, 3, 4, 5]
  }'
```

## 🎯 总结

ROMA MCP 安全机制确保：

✅ **所有请求都经过身份验证**
✅ **用户只能访问授权的资源**
✅ **所有操作都有审计日志**
✅ **权限基于角色管理**
✅ **支持令牌撤销和过期**

---

**安全是运维的生命线！** 🔒


