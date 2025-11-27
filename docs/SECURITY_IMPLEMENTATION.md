# 安全功能实施指南

## 快速开始

### 1. 在 `main.go` 中初始化安全组件

```go
import (
    "binrc.com/roma/core/middleware"
)

func main() {
    // ... 其他初始化代码 ...

    // 初始化速率限制器
    // 参数：每个IP最大并发连接数，每秒最大连接数
    middleware.InitRateLimiter(10, 3) // 每个IP最多10个并发连接，每秒最多3个新连接

    // 初始化IP黑名单
    middleware.InitIPBlacklist()

    // 初始化认证失败追踪器
    // 参数：最大失败次数，封禁时长，失败计数窗口，是否启用指数退避
    middleware.InitAuthFailureTracker(
        5,                      // 5次失败后封禁
        15*time.Minute,         // 封禁15分钟
        5*time.Minute,          // 5分钟内的失败计数
        true,                   // 启用指数退避
    )

    // ... 其他代码 ...
}
```

### 2. 在 `router.go` 中添加中间件

```go
func SetupRouter() *gin.Engine {
    r := gin.Default()
    
    // 基础中间件
    r.Use(gin.Recovery())
    r.Use(middleware.CORSMiddleware())
    
    // 安全中间件（按顺序添加）
    r.Use(middleware.IPBlacklistMiddleware())      // 1. IP黑名单检查（最先执行）
    r.Use(middleware.RateLimitMiddleware())        // 2. 速率限制
    r.Use(middleware.AuthFailureMiddleware())      // 3. 认证失败追踪（仅对登录接口生效）

    // ... 其他路由 ...
}
```

## 配置建议

### 生产环境推荐配置

```go
// 速率限制
middleware.InitRateLimiter(
    5,   // 每个IP最多5个并发连接（根据实际需求调整）
    2,   // 每秒最多2个新连接
)

// 认证失败追踪
middleware.InitAuthFailureTracker(
    3,                      // 3次失败后封禁（更严格）
    30*time.Minute,         // 封禁30分钟
    10*time.Minute,         // 10分钟内的失败计数
    true,                   // 启用指数退避
)
```

### 开发环境配置

```go
// 更宽松的限制，便于开发测试
middleware.InitRateLimiter(20, 10)
middleware.InitAuthFailureTracker(10, 5*time.Minute, 5*time.Minute, false)
```

## 功能说明

### 1. 速率限制 (Rate Limiting)

**功能**：
- 限制每个IP的最大并发连接数
- 限制每个IP的连接建立速率

**防护效果**：
- 防止单个IP占用过多资源
- 防止连接洪水攻击

**配置参数**：
- `maxConnectionsPerIP`: 每个IP的最大并发连接数
- `maxConnectionsPerSecond`: 每个IP每秒最大连接数

### 2. IP黑名单 (IP Blacklist)

**功能**：
- 永久黑名单：手动添加的恶意IP
- 临时黑名单：自动封禁的IP（带过期时间）

**防护效果**：
- 阻止已知恶意IP访问
- 自动封禁频繁认证失败的IP

**使用方式**：
```go
// 手动添加IP到黑名单
middleware.AddToBlacklist("192.168.1.100", 0) // 永久封禁
middleware.AddToBlacklist("192.168.1.101", 1*time.Hour) // 封禁1小时

// 从黑名单移除
middleware.RemoveFromBlacklist("192.168.1.100")
```

### 3. 认证失败追踪 (Auth Failure Tracking)

**功能**：
- 记录每个IP的认证失败次数
- 达到阈值后自动封禁
- 支持指数退避（封禁时长递增）

**防护效果**：
- 防止暴力破解攻击
- 自动封禁恶意IP

**工作流程**：
1. 记录认证失败
2. 检查失败次数是否达到阈值
3. 达到阈值后自动封禁并添加到黑名单
4. 认证成功时清除失败记录

**指数退避示例**：
- 第1次封禁：15分钟
- 第2次封禁：30分钟
- 第3次封禁：1小时
- 第4次封禁：2小时
- 最大封禁：24小时

## 监控和告警

### 建议监控指标

1. **连接数监控**
   - 总连接数
   - 每个IP的连接数
   - 连接数趋势

2. **认证失败监控**
   - 认证失败率
   - 被封禁的IP数量
   - 失败IP的地理位置分布

3. **性能监控**
   - API响应时间
   - 错误率
   - 资源使用率（CPU、内存、带宽）

### 告警阈值建议

- **连接数告警**：总连接数 > 1000
- **认证失败告警**：失败率 > 10%
- **封禁IP告警**：1小时内封禁 > 50个IP
- **错误率告警**：5xx错误率 > 1%

## 高级功能（待实现）

### 1. IP白名单
- 仅允许白名单IP访问（适用于内网环境）
- 白名单IP不受速率限制

### 2. 地理位置限制
- 基于IP地理位置限制访问
- 仅允许特定国家/地区访问

### 3. 用户行为分析
- 建立用户行为基线
- 检测异常行为（异常登录时间、异常IP等）

### 4. 威胁情报集成
- 集成威胁情报API
- 自动封禁已知恶意IP

### 5. 会话管理
- 限制每个用户的最大并发会话数
- 会话超时自动断开
- 会话审计

## 注意事项

1. **性能影响**：中间件会增加少量延迟，建议在负载测试中验证
2. **误封问题**：如果合法用户被误封，需要提供手动解封机制
3. **日志记录**：建议记录所有封禁操作，便于审计和问题排查
4. **配置调优**：根据实际流量调整参数，避免过于严格影响正常用户

## 测试建议

1. **压力测试**：使用工具（如Apache Bench）测试速率限制是否生效
2. **暴力破解测试**：模拟多次认证失败，验证自动封禁功能
3. **正常用户测试**：确保正常用户不受影响
4. **性能测试**：验证安全功能对性能的影响

