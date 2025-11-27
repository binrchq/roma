# ROMA Development Guide

This document provides information on how to contribute to ROMA development.

[中文版本](./DEVELOPMENT_CN.md)

---

## Project Architecture

### Overall Architecture

```
┌─────────────────────────────────────────┐
│           User/AI Assistant              │
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

### Directory Structure

```
roma/
├── cmd/roma/              # Main entry point
│   └── main.go
├── core/                  # Core functionality
│   ├── api/              # API controllers
│   ├── model/            # Data models
│   ├── operation/        # Business logic
│   ├── connect/          # Connection handlers
│   ├── tui/              # SSH terminal UI
│   └── constants/        # Constants
├── mcp/                  # MCP integration
│   └── bridge/           # MCP Bridge
├── web/                  # Web components
│   ├── frontend/         # React frontend
│   └── vscode-extension/ # VSCode extension
├── configs/              # Configuration files
├── deployment/           # Deployment configs
└── docs/                 # Documentation
```

---

## Quick Start

### Requirements

- Go 1.21+
- Node.js 18+ (for Web UI development)
- Git
- Docker (optional)

### Clone Repository

```bash
git clone https://github.com/binrchq/roma.git
cd roma
```

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Web UI dependencies (optional)
cd web/frontend
npm install
cd ../..
```

### Development Configuration

```bash
# Copy example configuration
cp configs/config.ex.toml configs/config.dev.toml

# Edit configuration
vim configs/config.dev.toml
```

**Development configuration example:**

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

### Start Development Server

```bash
# Start backend
go run cmd/roma/main.go -c configs/config.dev.toml

# Or use hot reload (air)
air

# Start frontend (another terminal)
cd web/frontend
npm run dev
```

---

## Code Standards

### Go Code Standards

Follow standard Go code style:

```bash
# Format code
go fmt ./...

# Check code
go vet ./...

# Static analysis
golangci-lint run
```

**Naming Conventions:**

```go
// Package: lowercase
package operation

// Exported functions: PascalCase
func CreateResource() {}

// Private functions: camelCase
func validateInput() {}

// Constants: PascalCase or UPPER_SNAKE_CASE
const DefaultTimeout = 30
const MAX_RETRY_COUNT = 3

// Interfaces: Verb + er
type ResourceManager interface {}
type CommandExecutor interface {}
```

### Data Model Standards

All GORM models must specify table and column names:

```go
// Resource model
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

// Specify table name
func (Resource) TableName() string {
    return "RESOURCES"
}
```

**JSON field naming:** Use uppercase

```go
// Correct
type Response struct {
    CODE    int    `json:"CODE"`
    MESSAGE string `json:"MESSAGE"`
    DATA    any    `json:"DATA"`
}
```

### Layered Architecture

Follow layered architecture principles:

```
┌─────────────────────────────────┐
│        API Layer (api/)          │  HTTP routing and request handling
├─────────────────────────────────┤
│      Service Layer (operation/)  │  Business logic
├─────────────────────────────────┤
│        DAO Layer (model/)        │  Database operations
├─────────────────────────────────┤
│       Util Layer (util/)         │  Utility functions
└─────────────────────────────────┘
```

---

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./core/operation

# View test coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### API Testing

Use Postman or curl to test APIs:

```bash
# Test create resource
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

## Debugging

### Enable Debug Mode

```toml
[log]
level = 'debug'
format = 'text'  # Use text for dev, json for production
```

### Using Delve Debugger

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugging
dlv debug cmd/roma/main.go -- -c configs/config.dev.toml

# Set breakpoint
(dlv) break operation.CreateResource
(dlv) continue
```

### VSCode Debug Configuration

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

## Build and Release

### Local Build

```bash
# Build for current platform
go build -o roma cmd/roma/main.go

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o roma-linux-amd64 cmd/roma/main.go
GOOS=windows GOARCH=amd64 go build -o roma-windows-amd64.exe cmd/roma/main.go
GOOS=darwin GOARCH=amd64 go build -o roma-darwin-amd64 cmd/roma/main.go
GOOS=darwin GOARCH=arm64 go build -o roma-darwin-arm64 cmd/roma/main.go
```

### Docker Build

```bash
# Build image
docker build -t roma:latest .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t roma:latest .
```

---

## Contributing Guidelines

### Branch Strategy

- `main` - Stable release
- `develop` - Development
- `feature/*` - New features
- `bugfix/*` - Bug fixes
- `hotfix/*` - Urgent fixes

### Commit Standards

Follow Conventional Commits:

```bash
# Feature
git commit -m "feat: add resource tagging"

# Fix
git commit -m "fix: resolve SSH connection timeout"

# Documentation
git commit -m "docs: update deployment guide"

# Style
git commit -m "style: format code"

# Refactor
git commit -m "refactor: refactor resource management module"

# Performance
git commit -m "perf: optimize database query performance"

# Test
git commit -m "test: add resource service unit tests"

# Build
git commit -m "build: update Docker image build process"

# CI
git commit -m "ci: add GitHub Actions workflow"

# Misc
git commit -m "chore: update dependency versions"
```

### Pull Request Process

1. **Fork Repository**
```bash
# Fork to your account
# Clone forked repository
git clone https://github.com/your-username/roma.git
cd roma
```

2. **Create Branch**
```bash
git checkout -b feature/my-feature
```

3. **Develop and Test**
```bash
# Develop code
# Run tests
go test ./...
# Format code
go fmt ./...
```

4. **Commit Code**
```bash
git add .
git commit -m "feat: add new feature"
git push origin feature/my-feature
```

5. **Create Pull Request**
- Visit GitHub repository
- Click "New Pull Request"
- Select your branch
- Fill in PR description
- Wait for Code Review

---

## Getting Help

- Documentation: [docs/](.)
- Discussions: [GitHub Discussions](https://github.com/binrchq/roma/discussions)
- Report Bugs: [GitHub Issues](https://github.com/binrchq/roma/issues)
- Email: dev@binrc.com

---

**Thank you for contributing to ROMA!**
