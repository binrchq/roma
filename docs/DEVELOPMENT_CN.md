# ROMA 开发指南

本文档介绍如何参与ROMA的开发。

---

## 🏗️ 项目架构

### 整体架构

```
┌─────────────────────────────────────────┐
│           用户/AI助手                    │
└────────┬──────────────────┬─────────────┘
         │ SSH (2200)       │ HTTPS
         ▼                  ▼
┌─────────────────┐  ┌──────────────────┐
│   SSH Gateway   │  │    Web UI        │
│   (TUI)         │  │    (React)       │
└────────┬────────┘  └────────┬─────────┘
         │                    │
         └─────────┬──────────┘
                   ▼
         ┌──────────────────┐
         │   ROMA Backend   │
         │   (Go)           │
         ├──────────────────┤
         │  • API Service   │
         │  • Auth/RBAC     │
         │  • Resource Mgmt │
         │  • Audit Log     │
         └─────────┬────────┘
                   │
         ┌─────────┴────────┐
         ▼                  ▼
    ┌─────────┐      ┌──────────────┐
    │Database │      │   Target     │
    │(SQLite/ │      │   Resources  │
    │MySQL/   │      │  (Servers/   │
    │PgSQL)   │      │   Databases) │
    └─────────┘      └──────────────┘
```

### 目录结构

```
roma/
├── cmd/roma/              # 主程序入口
│   └── main.go
├── core/                  # 核心功能
│   ├── api/              # API控制器
│   ├── model/            # 数据模型
│   ├── operation/        # 业务逻辑
│   ├── connect/          # 连接处理器
│   ├── tui/              # SSH命令行界面
│   └── constants/        # 常量定义
├── mcp/                  # MCP集成
│   └── bridge/           # MCP Bridge
├── web/                  # Web组件
│   ├── frontend/         # React前端
│   └── vscode-extension/ # VSCode扩展
├── configs/              # 配置文件
├── deployment/           # 部署配置
└── docs/                 # 文档
```

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+ (开发Web UI)
- Git
- Docker (可选)

### 克隆仓库

```bash
git clone https://github.com/binrchq/roma.git
cd roma
```

### 安装依赖

```bash
# 安装Go依赖
go mod download

# 安装Web UI依赖 (可选)
cd web/frontend
npm install
cd ../..
```

### 配置开发环境

```bash
# 复制示例配置
cp configs/config.ex.toml configs/config.dev.toml

# 编辑配置
vim configs/config.dev.toml
```

**开发配置示例:**

```toml
[api]
host = '0.0.0.0'
port = '6999'

[common]
port = '2200'
prompt = 'roma-dev'

[database]
type = 'sqlite'
cdb_url = './dev.db'

[log]
level = 'debug'
format = 'text'

[user_1st]
username = 'dev'
password = 'dev123456'
email = 'dev@example.com'
roles = "super,system,ops"
```

### 启动开发服务器

```bash
# 启动后端
go run cmd/roma/main.go -c configs/config.dev.toml

# 或使用热重载 (air)
air

# 启动前端 (另一个终端)
cd web/frontend
npm run dev
```

---

## 📝 代码规范

### Go代码规范

遵循标准Go代码风格：

```bash
# 格式化代码
go fmt ./...

# 检查代码
go vet ./...

# 静态分析
golangci-lint run
```

**命名规范:**

```go
// 包名: 小写单词
package operation

// 导出函数: PascalCase
func CreateResource() {}

// 私有函数: camelCase
func validateInput() {}

// 常量: PascalCase或UPPER_SNAKE_CASE
const DefaultTimeout = 30
const MAX_RETRY_COUNT = 3

// 接口: 动词 + er
type ResourceManager interface {}
type CommandExecutor interface {}
```

**注释规范:**

```go
// CreateResource 创建新资源
// 功能: 在数据库中创建一个新资源记录
// 输入: resource - 资源对象
// 输出: error - 错误信息，成功返回nil
// 必要性: 提供统一的资源创建接口，确保数据一致性
func CreateResource(resource *model.Resource) error {
    // 简单逻辑不需要注释
    if resource.Name == "" {
        return errors.New("资源名称不能为空")
    }
    
    // 复杂逻辑需要说明
    if err := validateResourceConfig(resource); err != nil {
        return fmt.Errorf("资源配置验证失败: %w", err)
    }
    
    return db.Create(resource).Error
}

// validateResourceConfig 验证资源配置
// 功能: 验证资源配置的完整性和正确性
// 输入: resource - 待验证的资源对象
// 输出: error - 验证错误，通过返回nil
// 必要性: 确保资源配置符合规范，避免运行时错误
func validateResourceConfig(resource *model.Resource) error {
    // 实现细节...
}
```

### 数据模型规范

所有GORM模型必须指定表名和列名：

```go
// Resource 资源模型
type Resource struct {
    ID        uint      `gorm:"column:ID;primaryKey" json:"ID"`
    NAME      string    `gorm:"column:NAME;size:100;not null" json:"NAME"`
    TYPE      string    `gorm:"column:TYPE;size:50;not null" json:"TYPE"`
    HOST      string    `gorm:"column:HOST;size:255" json:"HOST"`
    PORT      int       `gorm:"column:PORT" json:"PORT"`
    USERNAME  string    `gorm:"column:USERNAME;size:100" json:"USERNAME"`
    PASSWORD  string    `gorm:"column:PASSWORD;size:255" json:"PASSWORD"`
    SPACE_ID  uint      `gorm:"column:SPACE_ID" json:"SPACE_ID"`
    CREATED_AT time.Time `gorm:"column:CREATED_AT" json:"CREATED_AT"`
    UPDATED_AT time.Time `gorm:"column:UPDATED_AT" json:"UPDATED_AT"`
}

// TableName 指定表名
func (Resource) TableName() string {
    return "RESOURCES"
}
```

**JSON字段命名:** 使用大写

```go
// ✅ 正确
type Response struct {
    CODE    int    `json:"CODE"`
    MESSAGE string `json:"MESSAGE"`
    DATA    any    `json:"DATA"`
}

// ❌ 错误
type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"`
}
```

### 分层架构

遵循分层架构原则：

```
┌─────────────────────────────────┐
│        API Layer (api/)          │  HTTP路由和请求处理
├─────────────────────────────────┤
│      Service Layer (operation/)  │  业务逻辑处理
├─────────────────────────────────┤
│        DAO Layer (model/)        │  数据库操作
├─────────────────────────────────┤
│       Util Layer (util/)         │  通用工具函数
└─────────────────────────────────┘
```

**示例:**

```go
// DAO层 - model/resource_dao.go
package model

// CreateResource 在数据库中创建资源
func CreateResource(resource *Resource) error {
    return db.Create(resource).Error
}

// GetResourceByID 根据ID获取资源
func GetResourceByID(id uint) (*Resource, error) {
    var resource Resource
    err := db.First(&resource, id).Error
    return &resource, err
}

// Service层 - operation/resource_service.go
package operation

// CreateResource 创建资源服务
// 功能: 处理资源创建的业务逻辑
// 输入: req - 资源创建请求
// 输出: resource - 创建的资源, error - 错误信息
// 必要性: 封装资源创建的业务规则，包括验证、加密等
func CreateResource(req *CreateResourceRequest) (*model.Resource, error) {
    // 1. 验证输入
    if err := validateCreateRequest(req); err != nil {
        return nil, err
    }
    
    // 2. 加密敏感信息
    encryptedPassword, err := encryptPassword(req.Password)
    if err != nil {
        return nil, err
    }
    
    // 3. 构建资源对象
    resource := buildResourceFromRequest(req, encryptedPassword)
    
    // 4. 保存到数据库
    if err := model.CreateResource(resource); err != nil {
        return nil, err
    }
    
    return resource, nil
}

// validateCreateRequest 验证创建请求
func validateCreateRequest(req *CreateResourceRequest) error {
    // 验证逻辑...
}

// encryptPassword 加密密码
func encryptPassword(password string) (string, error) {
    // 加密逻辑...
}

// buildResourceFromRequest 构建资源对象
func buildResourceFromRequest(req *CreateResourceRequest, encryptedPassword string) *model.Resource {
    // 构建逻辑...
}

// API层 - api/resource_api.go
package api

// CreateResourceHandler 创建资源接口
func CreateResourceHandler(c *gin.Context) {
    var req operation.CreateResourceRequest
    
    // 绑定请求参数
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"CODE": 400, "MESSAGE": "参数错误"})
        return
    }
    
    // 调用服务层
    resource, err := operation.CreateResource(&req)
    if err != nil {
        c.JSON(500, gin.H{"CODE": 500, "MESSAGE": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"CODE": 200, "MESSAGE": "成功", "DATA": resource})
}
```

---

## 🧪 测试

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./core/operation

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**测试示例:**

```go
// operation/resource_service_test.go
package operation

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCreateResource(t *testing.T) {
    // 准备测试数据
    req := &CreateResourceRequest{
        NAME: "test-server",
        TYPE: "linux",
        HOST: "192.168.1.100",
    }
    
    // 执行测试
    resource, err := CreateResource(req)
    
    // 断言结果
    assert.NoError(t, err)
    assert.NotNil(t, resource)
    assert.Equal(t, "test-server", resource.NAME)
}
```

### 集成测试

```bash
# 运行集成测试
go test -tags=integration ./...
```

### API测试

使用Postman或curl测试API:

```bash
# 测试创建资源
curl -X POST http://localhost:6999/api/v1/resources \
  -H "apikey: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "NAME": "test-server",
    "TYPE": "linux",
    "HOST": "192.168.1.100"
  }'
```

---

## 🐛 调试

### 启用调试模式

```toml
[log]
level = 'debug'
format = 'text'  # 开发环境使用text，生产环境使用json
```

### 使用Delve调试器

```bash
# 安装Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 启动调试
dlv debug cmd/roma/main.go -- -c configs/config.dev.toml

# 设置断点
(dlv) break operation.CreateResource
(dlv) continue
```

### VSCode调试配置

`.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug ROMA",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/roma",
      "args": ["-c", "configs/config.dev.toml"],
      "env": {},
      "showLog": true
    }
  ]
}
```

---

## 📦 构建和发布

### 本地构建

```bash
# 构建当前平台
go build -o roma cmd/roma/main.go

# 构建特定平台
GOOS=linux GOARCH=amd64 go build -o roma-linux-amd64 cmd/roma/main.go
GOOS=windows GOARCH=amd64 go build -o roma-windows-amd64.exe cmd/roma/main.go
GOOS=darwin GOARCH=amd64 go build -o roma-darwin-amd64 cmd/roma/main.go
GOOS=darwin GOARCH=arm64 go build -o roma-darwin-arm64 cmd/roma/main.go
```

### Docker构建

```bash
# 构建镜像
docker build -t roma:latest .

# 多平台构建
docker buildx build --platform linux/amd64,linux/arm64 -t roma:latest .
```

### 版本发布

```bash
# 1. 更新版本号
vim version.go

# 2. 提交代码
git add .
git commit -m "chore: bump version to v1.2.0"

# 3. 打标签
git tag -a v1.2.0 -m "Release v1.2.0"

# 4. 推送
git push origin main --tags

# 5. GitHub Actions自动构建和发布
```

---

## 🤝 贡献指南

### 分支策略

- `main` - 稳定版本
- `develop` - 开发版本
- `feature/*` - 新功能
- `bugfix/*` - Bug修复
- `hotfix/*` - 紧急修复

### 提交规范

遵循Conventional Commits规范：

```bash
# 功能: feat
git commit -m "feat: 添加资源标签功能"

# 修复: fix
git commit -m "fix: 修复SSH连接超时问题"

# 文档: docs
git commit -m "docs: 更新部署文档"

# 样式: style
git commit -m "style: 格式化代码"

# 重构: refactor
git commit -m "refactor: 重构资源管理模块"

# 性能: perf
git commit -m "perf: 优化数据库查询性能"

# 测试: test
git commit -m "test: 添加资源服务单元测试"

# 构建: build
git commit -m "build: 更新Docker镜像构建流程"

# CI: ci
git commit -m "ci: 添加GitHub Actions工作流"

# 杂项: chore
git commit -m "chore: 更新依赖版本"
```

### Pull Request流程

1. **Fork仓库**
```bash
# Fork到自己的账号
# 克隆Fork后的仓库
git clone https://github.com/your-username/roma.git
cd roma
```

2. **创建分支**
```bash
git checkout -b feature/my-feature
```

3. **开发和测试**
```bash
# 开发代码
# 运行测试
go test ./...
# 格式化代码
go fmt ./...
```

4. **提交代码**
```bash
git add .
git commit -m "feat: 添加新功能"
git push origin feature/my-feature
```

5. **创建Pull Request**
- 访问GitHub仓库
- 点击"New Pull Request"
- 选择你的分支
- 填写PR描述
- 等待Code Review

### Code Review检查项

- [ ] 代码符合规范
- [ ] 有充分的测试覆盖
- [ ] 文档已更新
- [ ] 没有引入新的linter错误
- [ ] 提交信息符合规范
- [ ] 功能完整且可用

---

## 📚 开发资源

### 技术栈

- **后端:** Go 1.21+, Gin, GORM
- **前端:** React 18, TypeScript, Ant Design
- **数据库:** SQLite, MySQL, PostgreSQL
- **协议:** SSH, MCP (Model Context Protocol)

### 依赖库

```go
require (
    github.com/gin-gonic/gin v1.9.1
    gorm.io/gorm v1.25.5
    gorm.io/driver/sqlite v1.5.4
    gorm.io/driver/mysql v1.5.2
    gorm.io/driver/postgres v1.5.4
    golang.org/x/crypto v0.17.0
    github.com/golang-jwt/jwt/v5 v5.2.0
)
```

### 学习资源

- [Go官方文档](https://go.dev/doc/)
- [Gin框架文档](https://gin-gonic.com/docs/)
- [GORM文档](https://gorm.io/docs/)
- [MCP协议规范](https://modelcontextprotocol.io/)

---

## 🔧 常见问题

### 编译错误

**问题:** `cannot find package`

**解决:**
```bash
go mod download
go mod tidy
```

### 数据库错误

**问题:** `database locked`

**解决:**
```toml
# SQLite配置
[database]
cdb_url = 'file:roma.db?cache=shared&mode=rwc'
```

### SSH连接失败

**问题:** 无法连接到堡垒机

**解决:**
```bash
# 检查SSH服务是否启动
netstat -tlnp | grep 2200

# 检查SSH主机密钥
ls -la /path/to/ssh/keys/

# 重新生成主机密钥
ssh-keygen -t rsa -b 4096 -f /path/to/ssh/keys/id_rsa
```

---

## 📞 获取帮助

- 📖 文档: [docs/](.)
- 💬 讨论: [GitHub Discussions](https://github.com/binrchq/roma/discussions)
- 🐛 报告Bug: [GitHub Issues](https://github.com/binrchq/roma/issues)
- 📧 Email: dev@binrc.com

---

**感谢你对ROMA的贡献！** 🎉

