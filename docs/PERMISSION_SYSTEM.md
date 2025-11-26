# ROMA 权限系统文档

## 概述

ROMA 权限系统支持多维度权限控制，包括：

1. **用户角色权限**：基于角色的访问控制（RBAC）
2. **资源角色绑定**：资源可以指定哪些角色可以访问
3. **项目隔离**：资源可以属于项目，只有项目成员才能访问
4. **灵活的权限策略**：可配置的权限检查规则

## 权限模型

### 1. 用户角色（User Roles）

用户拥有一个或多个角色，每个角色定义了可执行的操作。

**配置示例：**
```toml
[[roles]]
name = "super"
description = "Super Administrator"
is_default_super = true
  [[roles.permissions]]
  target = "*"
  actions = ["*"]

[[roles]]
name = "ops"
description = "Ops engineer"
  [[roles.permissions]]
  target = "resource"
  actions = ["get", "list", "use"]
```

### 2. 资源角色（Resource Roles）

资源可以绑定一个或多个角色，只有拥有匹配角色的用户才能访问。

**使用场景：**
- 生产服务器只允许 `ops` 角色访问
- 敏感资源只允许 `super` 角色访问
- 开发环境资源允许 `ops` 和 `system` 角色访问

**API 示例：**
```bash
# 为资源分配角色
POST /api/v1/resources/{id}/roles
{
  "role_ids": [2, 3],  # ops 和 system 角色
  "project_id": 1      # 可选，指定项目
}
```

### 3. 项目隔离（Project Isolation）

资源可以属于项目，只有项目成员才能访问项目内的资源。

**配置示例：**
```toml
[[projects]]
name = "production"
description = "生产环境项目"
members = ["super", "system"]
default_role = "ops"
```

**权限检查流程：**
1. 检查用户是否是项目成员
2. 检查用户在项目中的角色权限
3. 检查资源是否属于该项目

### 4. 权限策略配置

```toml
[permission_policy]
# 是否启用资源角色检查
enable_resource_role = true
# 是否启用项目隔离
enable_project_isolation = true
# 是否要求用户角色和资源角色完全匹配
require_exact_role_match = false
# 是否允许 super 角色绕过所有限制
super_bypass_all = true
# 默认项目名称（全局资源所属的项目）
default_project = ""
```

## 权限检查流程

### 资源访问权限检查

当用户尝试访问资源时，系统按以下顺序检查：

1. **Super 角色检查**（如果 `super_bypass_all = true`）
   - 如果用户是 super 角色，直接允许访问

2. **资源角色检查**（如果 `enable_resource_role = true`）
   - 获取资源绑定的角色列表
   - 检查用户是否拥有匹配的角色
   - 如果 `require_exact_role_match = true`，必须完全匹配

3. **项目隔离检查**（如果 `enable_project_isolation = true`）
   - 获取资源所属的项目
   - 检查用户是否是项目成员
   - 检查用户在项目中的角色权限

4. **全局角色权限检查**
   - 检查用户的全局角色是否有权限执行该操作

## 配置示例

### 完整配置示例

```toml
# 权限策略
[permission_policy]
enable_resource_role = true
enable_project_isolation = true
require_exact_role_match = false
super_bypass_all = true
default_project = ""

# 角色定义
[[roles]]
name = "super"
description = "Super Administrator"
is_default_super = true
  [[roles.permissions]]
  target = "*"
  actions = ["*"]

[[roles]]
name = "ops"
description = "Ops engineer"
  [[roles.permissions]]
  target = "resource"
  actions = ["get", "list", "use"]

# 项目定义
[[projects]]
name = "production"
description = "生产环境项目"
members = ["super", "system"]
default_role = "ops"

[[projects]]
name = "development"
description = "开发环境项目"
members = ["super", "system", "ops"]
default_role = "ops"
```

## API 使用

### 1. 为资源分配角色

```bash
POST /api/v1/resources/{id}/roles
Content-Type: application/json

{
  "role_ids": [2, 3],
  "project_id": 1  # 可选
}
```

### 2. 将资源分配到项目

```bash
POST /api/v1/projects/{project_id}/resources
Content-Type: application/json

{
  "resource_id": 123,
  "resource_type": "linux"
}
```

### 3. 添加项目成员

```bash
POST /api/v1/projects/{project_id}/members
Content-Type: application/json

{
  "user_id": 5,
  "role_id": 2  # 项目内角色
}
```

## 最佳实践

1. **生产环境**：
   - 启用项目隔离和资源角色检查
   - 设置 `super_bypass_all = true` 以便紧急访问
   - 为生产资源创建独立项目

2. **开发环境**：
   - 可以放宽权限策略
   - 使用项目隔离区分不同开发团队

3. **敏感资源**：
   - 绑定特定角色（如 `super`）
   - 设置 `require_exact_role_match = true`

4. **权限调试**：
   - 查看权限检查日志
   - 使用 `GET /api/v1/users/me` 查看当前用户权限
   - 检查项目成员关系

## 迁移指南

### 从旧权限系统迁移

1. **更新配置文件**：
   - 将 `roles[].desc` 迁移到结构化 `roles[].permissions`
   - 添加 `permission_policy` 配置
   - 定义项目（如果需要）

2. **数据库迁移**：
   - 系统会自动创建新的表结构
   - 现有角色描述会被保留（向后兼容）

3. **逐步启用**：
   - 先启用 `enable_resource_role`
   - 测试通过后启用 `enable_project_isolation`
   - 最后调整 `require_exact_role_match`

## 故障排查

### 用户无法访问资源

1. 检查用户角色：
   ```bash
   GET /api/v1/users/{id}
   ```

2. 检查资源角色绑定：
   ```bash
   GET /api/v1/resources/{id}/roles
   ```

3. 检查项目成员关系：
   ```bash
   GET /api/v1/projects/{project_id}/members
   ```

4. 查看权限策略配置：
   - 检查 `permission_policy` 设置
   - 确认 `super_bypass_all` 是否启用

### 权限检查失败

- 查看后端日志中的权限检查详情
- 确认资源是否属于项目
- 确认用户是否在项目成员列表中
