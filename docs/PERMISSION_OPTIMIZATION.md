# 权限系统优化总结

## 已完成的优化

### 1. ✅ 移除硬编码角色名称
**问题**：代码中多处硬编码了角色名称（如 "super"、"system"），导致不够灵活。

**优化方案**：
- 创建了 `IsSuperRole(role *model.Role)` 函数：通过权限描述符的 `IsSuper` 标志判断
- 创建了 `HasAllPermissions(role *model.Role)` 函数：通过权限描述符判断是否拥有所有权限
- 所有权限检查都改为使用这些函数，不再硬编码角色名称

**修改的文件**：
- `core/permissions/descriptors.go`：新增辅助函数
- `core/permissions/policy.go`：使用 `IsSuperRole` 替代硬编码
- `core/api/middleware/permission.go`：使用 `IsSuperRole` 替代硬编码
- `core/api/resource_control.go`：使用 `IsSuperRole` 和 `HasAllPermissions` 替代硬编码
- `core/api/space_control.go`：使用权限描述符判断管理员权限
- `core/services/setup.go`：使用 `IsSuperRole` 替代硬编码

### 2. ✅ 性能优化
**问题**：资源列表 API 和 TUI 命令中，对每个资源都调用 `CheckResourceAccess`，会重复查询用户角色。

**优化方案**：
- 创建了 `CheckResourceAccessWithRoles` 函数：允许传入已获取的用户角色，避免重复查询
- `GetAllResource` API：传入已获取的用户角色
- `ls` 和 `ln` TUI 命令：传入已获取的用户角色

**性能提升**：
- 减少了数据库查询次数
- 对于有大量资源的场景，性能提升明显

### 3. ✅ 完善无空间归属资源的处理
**问题**：资源没有空间归属时，处理逻辑不完整。

**优化方案**：
- 完善了无空间归属资源的处理逻辑
- 如果启用了空间隔离，没有空间归属的资源会继续检查全局角色权限

### 4. ✅ 修复权限检查中间件重复检查
**问题**：`RequirePermission` 中间件中有重复的权限检查逻辑。

**优化方案**：
- 移除了重复的权限检查
- 简化了逻辑流程

### 5. ✅ 改进用户创建 API 错误处理
**问题**：用户创建时，如果添加角色失败，错误处理逻辑有问题。

**优化方案**：
- 先创建用户，如果失败直接返回
- 然后逐个添加角色，如果失败立即返回错误
- 改进了错误消息，更清晰地指出失败原因

## 保留的硬编码（业务逻辑，非权限检查）

### 1. 默认角色名称 "ops"
**位置**：`core/api/resource_control.go` 第 38、283 行

**说明**：这是业务逻辑的默认值，用于资源创建时如果没有指定角色，使用默认角色。这不是权限检查，而是业务规则。

**建议**：如果需要，可以从配置文件中读取默认角色名称。

### 2. 系统信息中的 "system" 键名
**位置**：`core/api/system_control.go`、`core/connector/router_connector.go`

**说明**：这是 API 返回的 JSON 键名，不是权限检查，可以保留。

## 权限检查优化效果

### 优化前
```go
// 硬编码角色名称
if role.Name == "super" || role.Name == "system" {
    // ...
}
```

### 优化后
```go
// 通过权限描述符判断
if permissions.IsSuperRole(role) || permissions.HasAllPermissions(role) {
    // ...
}
```

## 性能优化效果

### 优化前
```go
// 每个资源都查询用户角色
for _, res := range resList {
    allowed, _ := permissions.CheckResourceAccess(user, res.GetID(), resType, "list")
    // ...
}
```

### 优化后
```go
// 只查询一次用户角色，然后复用
roles, err := opUser.GetUserRoles(user.ID)
for _, res := range resList {
    allowed, _ := permissions.CheckResourceAccessWithRoles(user, roles, res.GetID(), resType, "list")
    // ...
}
```

## 配置灵活性

现在权限系统完全基于权限描述符，不依赖硬编码的角色名称：

1. **Super 角色判断**：通过 `IsSuper` 标志或 `target="*"` 且 `actions=["*"]` 的权限
2. **管理员权限判断**：通过 `user.add` 权限或 `IsSuper` 标志
3. **所有权限判断**：通过 `HasAllPermissions` 函数

## 总结

✅ **已移除所有权限检查中的硬编码角色名称**
✅ **性能优化：减少重复数据库查询**
✅ **代码更灵活：支持自定义角色名称**
✅ **向后兼容：保留业务逻辑的默认值**

权限系统现在完全基于权限描述符，不依赖硬编码的角色名称，更加灵活和可配置。

