## 🔐 ROMA 权限系统详解

## 权限系统设计

ROMA 采用**基于角色的细粒度权限控制**（RBAC），通过角色描述中嵌入的权限规则实现灵活的权限管理。

### 权限规则格式

```
operation:目标类型-范围过滤.(操作1|操作2|...)
```

**组成部分：**
- `operation:` - 固定前缀
- `目标类型` - `user`（用户）或 `resource`（资源）
- `范围过滤` - 可选，用于限制资源范围（如 `*peripheral`, `*trial`）
- `操作列表` - 允许的操作，用 `|` 分隔

**范围过滤规则：**
- `*scope` - 排除规则，表示**不包括**该范围的资源
- `scope` - 包含规则，表示**只包括**该范围的资源
- 无范围 - 应用于所有资源

## 五级角色体系

### 1. super - 超级管理员

```toml
[[roles]]
name = "super"
desc = "all permissions [operation:user.(add|delete|update|get|list)]"
```

**权限范围：**
- ✅ **用户管理**：增删改查所有用户
- ✅ **资源管理**：所有资源的所有操作
- ✅ **系统配置**：修改系统配置
- ✅ **审计日志**：查看所有日志

**操作权限：**
| 目标 | 操作 |
|------|------|
| 用户 | add, delete, update, get, list |
| 资源 | 所有操作（继承） |

**适用人员：** 系统管理员、CTO

---

### 2. system - 系统管理员

```toml
[[roles]]
name = "system"
desc = "system administrator [operation:resource.(add|delete|update|get|list)]"
```

**权限范围：**
- ✅ **资源管理**：增删改查所有资源
- ✅ **资源配置**：配置资源参数
- ✅ **资源分配**：分配资源给用户
- ❌ **用户管理**：不能管理用户

**操作权限：**
| 目标 | 操作 |
|------|------|
| 资源 | add, delete, update, get, list |

**适用人员：** 系统运维负责人、资源管理员

---

### 3. ops - 运维人员

```toml
[[roles]]
name = "ops"
desc = "system operations personnel [operation:resource.(get|list|use)]"
```

**权限范围：**
- ✅ **查看资源**：查看已分配的资源信息
- ✅ **使用资源**：连接、执行命令
- ❌ **修改资源**：不能修改资源配置
- ❌ **删除资源**：不能删除资源

**操作权限：**
| 目标 | 操作 | 说明 |
|------|------|------|
| 资源 | get | 查看资源详情 |
| 资源 | list | 列出资源列表 |
| 资源 | use | 连接资源、执行命令 |

**适用人员：** 运维工程师、DevOps

---

### 4. ordinary - 普通用户

```toml
[[roles]]
name = "ordinary"
desc = "system ordinary [operation:resource-(*peripheral).(get|list)]"
```

**权限范围：**
- ✅ **查看资源**：查看已分配的资源（**排除外围设备**）
- ❌ **使用资源**：不能连接或执行命令
- ❌ **修改资源**：不能修改资源

**特殊限制：**
- 🚫 不能访问标记为 `peripheral`（外围设备）的资源
- 只读权限

**操作权限：**
| 目标 | 操作 | 范围过滤 |
|------|------|----------|
| 资源 | get, list | 排除 `peripheral` |

**适用人员：** 开发人员、测试人员

---

### 5. trial - 试用用户

```toml
[[roles]]
name = "trial"
desc = "system trial [operation:resource-(*trial).(get|list|use)]"
```

**权限范围：**
- ✅ **试用资源**：只能访问标记为 `trial` 的资源
- ✅ **使用资源**：可以连接和执行命令
- ❌ **生产资源**：不能访问生产环境资源

**特殊限制：**
- 🔒 只能访问 `trial` 范围的资源
- 不能访问其他任何资源

**操作权限：**
| 目标 | 操作 | 范围过滤 |
|------|------|----------|
| 资源 | get, list, use | 只包括 `trial` |

**适用人员：** 试用用户、临时访客、实习生

---

## 权限检查流程

```
MCP 请求
    ↓
【1. 身份验证】
├─ 验证令牌
├─ 查询用户信息
└─ 加载用户角色和权限规则
    ↓
【2. 解析权限规则】
├─ 从角色描述中提取权限规则
├─ 解析目标类型 (user/resource)
├─ 解析范围过滤 (*peripheral, *trial)
└─ 解析操作列表 (add|delete|...)
    ↓
【3. 权限检查】
├─ 检查操作类型是否允许
├─ 检查资源范围是否匹配
└─ 检查是否被分配了该资源
    ↓
【4. 返回结果】
├─ 允许：执行操作
└─ 拒绝：返回 PERMISSION_DENIED
```

## 资源范围标签

### 资源分类

| 标签 | 说明 | 示例资源 |
|------|------|----------|
| `trial` | 试用资源 | test-server-01, trial-db |
| `peripheral` | 外围设备 | edge-node-01, iot-device |
| `production` | 生产资源 | prod-web-01, prod-db |
| 无标签 | 普通资源 | dev-server-01 |

### 标签设置方式

#### 方式1: 从资源名称自动识别

```go
// 系统会自动识别包含关键词的资源
"trial-web-01"       → trial
"test-server"        → trial
"peripheral-device"  → peripheral
"edge-node"          → peripheral
"prod-db-01"         → (无标签，普通资源)
```

#### 方式2: 从资源标签字段

```json
{
  "name": "web-server-01",
  "tags": "trial,test,development"  // 包含 trial 标签
}
```

## 实际使用示例

### 示例 1: super 角色访问

```bash
# super 用户: alice
角色: super
权限: operation:user.(add|delete|update|get|list)

# 操作1: 列出所有资源
请求: list_resources(type="linux")
检查:
  - 目标: resource
  - 操作: list
  - super 角色可以访问所有资源
结果: ✅ 返回所有 Linux 资源

# 操作2: 添加用户
请求: add_user(username="bob")
检查:
  - 目标: user
  - 操作: add
  - super 角色有 operation:user.add 权限
结果: ✅ 成功添加用户

# 操作3: 删除资源
请求: delete_resource(id=123)
检查:
  - 目标: resource
  - 操作: delete
  - super 角色可以执行所有操作
结果: ✅ 成功删除资源
```

### 示例 2: system 角色访问

```bash
# system 用户: bob
角色: system
权限: operation:resource.(add|delete|update|get|list)

# 操作1: 添加资源
请求: add_resource(name="web-03", type="linux")
检查:
  - 目标: resource
  - 操作: add
  - system 角色有 operation:resource.add 权限
结果: ✅ 成功添加资源

# 操作2: 修改资源
请求: update_resource(id=123, address="192.168.1.100")
检查:
  - 目标: resource
  - 操作: update
  - system 角色有 operation:resource.update 权限
结果: ✅ 成功修改资源

# 操作3: 添加用户
请求: add_user(username="charlie")
检查:
  - 目标: user
  - 操作: add
  - system 角色没有用户管理权限
结果: ❌ PERMISSION_DENIED
```

### 示例 3: ops 角色访问

```bash
# ops 用户: charlie
角色: ops
权限: operation:resource.(get|list|use)
已分配资源: web-01, web-02, db-01

# 操作1: 列出资源
请求: list_resources(type="linux")
检查:
  - 目标: resource
  - 操作: list
  - ops 角色有 operation:resource.list 权限
  - 过滤：只返回已分配的资源
结果: ✅ 返回 [web-01, web-02, db-01]

# 操作2: 执行命令
请求: execute_command(resource_id=1, command="df -h")
检查:
  - 目标: resource
  - 操作: use
  - ops 角色有 operation:resource.use 权限
  - 资源 1 (web-01) 已分配给用户
结果: ✅ 执行成功

# 操作3: 删除资源
请求: delete_resource(id=1)
检查:
  - 目标: resource
  - 操作: delete
  - ops 角色没有 delete 权限
结果: ❌ PERMISSION_DENIED
```

### 示例 4: ordinary 角色访问

```bash
# ordinary 用户: david
角色: ordinary
权限: operation:resource-(*peripheral).(get|list)
已分配资源: web-01, edge-node-01(peripheral)

# 操作1: 列出资源
请求: list_resources(type="linux")
检查:
  - 目标: resource
  - 操作: list
  - ordinary 角色有 operation:resource.list 权限
  - 范围过滤: 排除 peripheral 资源
  - 过滤：只返回已分配且非 peripheral 的资源
结果: ✅ 返回 [web-01]（edge-node-01 被过滤）

# 操作2: 查看外围设备
请求: get_resource(resource_id=2)  // edge-node-01
检查:
  - 目标: resource
  - 操作: get
  - ordinary 角色有 operation:resource.get 权限
  - 但资源范围是 peripheral，被排除规则过滤
结果: ❌ PERMISSION_DENIED

# 操作3: 执行命令
请求: execute_command(resource_id=1, command="uptime")
检查:
  - 目标: resource
  - 操作: use
  - ordinary 角色没有 use 权限
结果: ❌ PERMISSION_DENIED
```

### 示例 5: trial 角色访问

```bash
# trial 用户: eve
角色: trial
权限: operation:resource-(*trial).(get|list|use)
已分配资源: trial-web-01, prod-db-01

# 操作1: 列出资源
请求: list_resources(type="linux")
检查:
  - 目标: resource
  - 操作: list
  - trial 角色有 operation:resource.list 权限
  - 范围过滤: 只包括 trial 资源
  - 过滤：只返回 trial 范围的资源
结果: ✅ 返回 [trial-web-01]（prod-db-01 被过滤）

# 操作2: 访问生产资源
请求: get_resource(resource_id=2)  // prod-db-01
检查:
  - 目标: resource
  - 操作: get
  - trial 角色有 operation:resource.get 权限
  - 但资源范围不是 trial，被包含规则过滤
结果: ❌ PERMISSION_DENIED

# 操作3: 在试用资源上执行命令
请求: execute_command(resource_id=1, command="hostname")
检查:
  - 目标: resource
  - 操作: use
  - trial 角色有 operation:resource.use 权限
  - 资源 1 (trial-web-01) 是 trial 范围
结果: ✅ 执行成功
```

## MCP 集成示例

### JavaScript 客户端

```javascript
// 不同角色的使用场景

// 1. super 角色 - 所有操作
const superClient = new SecureMCPClient(apiUrl, 'super_token')
await superClient.listResources('all')       // ✅ 所有资源
await superClient.addResource({...})         // ✅ 添加资源
await superClient.deleteResource(123)        // ✅ 删除资源
await superClient.addUser({...})             // ✅ 添加用户

// 2. system 角色 - 资源管理
const systemClient = new SecureMCPClient(apiUrl, 'system_token')
await systemClient.listResources('all')      // ✅ 所有资源
await systemClient.addResource({...})        // ✅ 添加资源
await systemClient.executeCommand(1, 'ls')   // ✅ 执行命令
await systemClient.addUser({...})            // ❌ PERMISSION_DENIED

// 3. ops 角色 - 运维操作
const opsClient = new SecureMCPClient(apiUrl, 'ops_token')
await opsClient.listResources('linux')       // ✅ 已分配的资源
await opsClient.executeCommand(1, 'uptime')  // ✅ 执行命令
await opsClient.deleteResource(1)            // ❌ PERMISSION_DENIED

// 4. ordinary 角色 - 只读访问
const ordinaryClient = new SecureMCPClient(apiUrl, 'ordinary_token')
await ordinaryClient.listResources('linux')  // ✅ 非外围资源
await ordinaryClient.getResource(1)          // ✅ 查看资源
await ordinaryClient.executeCommand(1, 'ls') // ❌ PERMISSION_DENIED

// 5. trial 角色 - 试用资源
const trialClient = new SecureMCPClient(apiUrl, 'trial_token')
await trialClient.listResources('linux')     // ✅ 只返回 trial 资源
await trialClient.executeCommand(1, 'pwd')   // ✅ (如果是 trial 资源)
await trialClient.getResource(2)             // ❌ (如果不是 trial 资源)
```

## 权限配置最佳实践

### 1. 最小权限原则

```
✅ DO:
- 新用户默认分配 ordinary 或 trial 角色
- 根据实际需要逐步提升权限
- 定期审查用户权限

❌ DON'T:
- 给所有人 super 或 system 权限
- 长期使用高权限账号进行日常操作
```

### 2. 资源标签管理

```
✅ DO:
- 为试用环境的资源添加 trial 标签
- 为边缘设备添加 peripheral 标签
- 统一命名规范（如 trial-xxx, prod-xxx）

❌ DON'T:
- 混用生产和试用资源而不做区分
- 频繁修改资源标签
```

### 3. 角色分配策略

```
角色分配建议：
- 实习生/试用人员    → trial
- 开发人员           → ordinary
- 测试人员           → ordinary
- 运维工程师         → ops
- 运维负责人         → system
- CTO/技术总监       → super
```

## 审计和监控

### 权限违规告警

```sql
-- 查询被拒绝的操作
SELECT user_id, username, action, COUNT(*) as denied_count
FROM access_logs
WHERE status = 'failed' 
  AND details LIKE '%denied%'
  AND accessed_at > NOW() - INTERVAL 1 DAY
GROUP BY user_id, action
HAVING denied_count > 5;

-- 查询试用用户尝试访问生产资源
SELECT * FROM access_logs
WHERE user_id IN (SELECT user_id FROM user_roles WHERE role_id = (SELECT id FROM roles WHERE name = 'trial'))
  AND status = 'failed'
  AND details NOT LIKE '%trial%';
```

---

**精细的权限控制是安全运维的基石！** 🔐🛡️


