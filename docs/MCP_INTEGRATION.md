# MCP 集成指南

本文档详细说明如何在不同环境中集成和使用 ROMA 的 MCP 功能。

## 目录

- [什么是 MCP](#什么是-mcp)
- [ROMA MCP 架构](#roma-mcp-架构)
- [集成方式](#集成方式)
- [AI IDE 配置](#ai-ide-配置)
- [自定义客户端开发](#自定义客户端开发)
- [最佳实践](#最佳实践)

## 什么是 MCP

**MCP (Model Context Protocol)** 是 Anthropic 推出的开放标准协议，用于 AI 助手与外部工具之间的通信。

### 核心概念

- **Server**: 提供工具和资源的服务端（ROMA）
- **Client**: 调用工具的客户端（AI 助手、IDE）
- **Transport**: 通信方式（stdio、socket）
- **Tool**: 可被调用的功能单元
- **Resource**: 可被访问的数据源

### MCP 协议优势

1. **标准化**: 统一的接口，兼容多种 AI 助手
2. **安全性**: 基于能力的权限模型
3. **可扩展**: 支持自定义工具和资源
4. **双向通信**: 支持流式数据和实时更新

## ROMA MCP 架构

```
┌─────────────────────────────────────────────────────┐
│                   AI 助手 / IDE                      │
│  (Cursor, Claude Desktop, VSCode, 自定义客户端)     │
└────────────────┬────────────────────────────────────┘
                 │ MCP Protocol
                 │
┌────────────────┴────────────────────────────────────┐
│              ROMA MCP Server                         │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ Resource │  │   SSH    │  │ Enhanced │         │
│  │  Tools   │  │  Tools   │  │  Tools   │         │
│  └──────────┘  └──────────┘  └──────────┘         │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │   User   │  │   Log    │  │  System  │         │
│  │  Tools   │  │  Tools   │  │  Tools   │         │
│  └──────────┘  └──────────┘  └──────────┘         │
└────────────────┬────────────────────────────────────┘
                 │
┌────────────────┴────────────────────────────────────┐
│             ROMA Core Services                       │
│  (资源管理、用户管理、SSH、API、数据库)             │
└─────────────────────────────────────────────────────┘
```

## 集成方式

ROMA 支持两种 MCP 传输方式：

### 1. Stdio 传输（本地模式）

#### 工作原理

```
AI Client → 启动 roma mcp 进程
         ↓
    stdin/stdout 通信
         ↓
      MCP Server
```

#### 配置 ROMA

```toml
# config.toml
[mcp]
enable = true
transport = "stdio"
```

#### 启动方式

```bash
# 作为独立进程
roma mcp

# 或集成在主程序中
roma start  # MCP 会自动启动
```

#### 适用场景

- ✅ 本地开发和测试
- ✅ Cursor/VSCode 插件
- ✅ Claude Desktop
- ✅ 单用户使用
- ❌ 远程访问
- ❌ 多用户共享

### 2. Socket 传输（网络模式）

#### 工作原理

```
AI Client → WebSocket 连接
         ↓
    ws://host:port/mcp
         ↓
      MCP Server
```

#### 配置 ROMA

```toml
# config.toml
[mcp]
enable = true
transport = "socket"
host = "0.0.0.0"
port = 8081
```

#### 客户端连接

```javascript
const ws = new WebSocket('ws://roma-server:8081/mcp')

ws.onopen = () => {
  ws.send(JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'initialize',
    params: {
      protocolVersion: '1.0.0',
      clientInfo: { name: 'MyClient', version: '1.0.0' }
    }
  }))
}
```

#### 适用场景

- ✅ 远程访问
- ✅ 多用户共享
- ✅ Web 应用集成
- ✅ 独立运维客户端
- ❌ 需要额外的网络配置
- ❌ 需要考虑认证和安全

## AI IDE 配置

### Cursor IDE

#### 1. 创建配置文件

```bash
mkdir -p ~/.cursor
cat > ~/.cursor/mcp-servers.json << 'EOF'
{
  "roma": {
    "command": "roma",
    "args": ["mcp"],
    "env": {
      "ROMA_CONFIG": "/path/to/config.toml"
    }
  }
}
EOF
```

#### 2. 使用示例

在 Cursor 中直接使用自然语言：

```
"列出所有 Linux 服务器"
"在 prod-web-01 上检查 Nginx 状态"
"批量更新所有 Web 服务器的配置"
```

Cursor AI 会自动识别并调用 ROMA MCP 工具。

### Claude Desktop

#### 1. 配置文件位置

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

#### 2. 配置内容

```json
{
  "mcpServers": {
    "roma": {
      "command": "roma",
      "args": ["mcp"],
      "env": {
        "ROMA_CONFIG": "/path/to/config.toml"
      }
    }
  }
}
```

#### 3. 使用示例

```
User: "ROMA，帮我查看所有服务器状态"

Claude: 我来为你查询...
[调用 MCP: list_resources(type="linux")]
[调用 MCP: ssh_check_health(...)]

这是所有服务器的状态：
- web-01: ✅ 正常 (CPU: 15%, 内存: 40%)
- web-02: ✅ 正常 (CPU: 20%, 内存: 35%)
- db-01: ⚠️  内存使用率偏高 (80%)
```

### VSCode

#### 使用 ROMA 扩展（推荐）

ROMA VSCode 扩展已内置 MCP 支持，无需手动配置。

#### 手动配置 MCP

```json
// settings.json
{
  "mcp.servers": {
    "roma": {
      "command": "roma",
      "args": ["mcp"]
    }
  }
}
```

### Continue.dev

```json
// ~/.continue/config.json
{
  "mcpServers": [
    {
      "name": "roma",
      "command": "roma",
      "args": ["mcp"]
    }
  ]
}
```

## 自定义客户端开发

### Node.js 示例

```javascript
const { spawn } = require('child_process')

class ROMCMCPClient {
  constructor() {
    this.process = null
    this.requestId = 0
    this.pendingRequests = new Map()
  }

  async connect() {
    this.process = spawn('roma', ['mcp'], {
      stdio: ['pipe', 'pipe', 'pipe']
    })

    this.process.stdout.on('data', (data) => {
      this.handleResponse(data.toString())
    })

    await this.initialize()
  }

  async initialize() {
    return this.sendRequest('initialize', {
      protocolVersion: '1.0.0',
      clientInfo: {
        name: 'MyClient',
        version: '1.0.0'
      }
    })
  }

  sendRequest(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = ++this.requestId
      const request = {
        jsonrpc: '2.0',
        id,
        method,
        params
      }

      this.pendingRequests.set(id, { resolve, reject })
      this.process.stdin.write(JSON.stringify(request) + '\n')

      setTimeout(() => {
        if (this.pendingRequests.has(id)) {
          this.pendingRequests.delete(id)
          reject(new Error('Request timeout'))
        }
      }, 30000)
    })
  }

  handleResponse(data) {
    const lines = data.trim().split('\n')
    for (const line of lines) {
      try {
        const response = JSON.parse(line)
        if (this.pendingRequests.has(response.id)) {
          const { resolve, reject } = this.pendingRequests.get(response.id)
          this.pendingRequests.delete(response.id)
          
          if (response.error) {
            reject(new Error(response.error.message))
          } else {
            resolve(response.result)
          }
        }
      } catch (error) {
        console.error('Failed to parse response:', error)
      }
    }
  }

  async callTool(name, args = {}) {
    return this.sendRequest('tools/call', {
      name,
      arguments: args
    })
  }

  disconnect() {
    if (this.process) {
      this.process.kill()
    }
  }
}

// 使用示例
async function main() {
  const client = new ROMCMCPClient()
  await client.connect()

  // 列出资源
  const resources = await client.callTool('list_resources', {
    resource_type: 'linux'
  })
  console.log('Resources:', resources)

  // 执行命令
  const result = await client.callTool('ssh_execute_command', {
    identifier: 'web-server-01',
    resource_type: 'linux',
    command: 'uptime'
  })
  console.log('Command result:', result)

  client.disconnect()
}

main().catch(console.error)
```

### Python 示例

```python
import json
import subprocess
from typing import Dict, Any

class ROMCMCPClient:
    def __init__(self):
        self.process = None
        self.request_id = 0
        self.pending_requests = {}

    def connect(self):
        self.process = subprocess.Popen(
            ['roma', 'mcp'],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
            bufsize=1
        )
        self.initialize()

    def send_request(self, method: str, params: Dict[str, Any] = None) -> Dict:
        self.request_id += 1
        request = {
            'jsonrpc': '2.0',
            'id': self.request_id,
            'method': method,
            'params': params or {}
        }
        
        self.process.stdin.write(json.dumps(request) + '\n')
        self.process.stdin.flush()
        
        # 读取响应
        response_line = self.process.stdout.readline()
        response = json.loads(response_line)
        
        if 'error' in response:
            raise Exception(response['error']['message'])
        
        return response.get('result', {})

    def initialize(self):
        return self.send_request('initialize', {
            'protocolVersion': '1.0.0',
            'clientInfo': {
                'name': 'PyClient',
                'version': '1.0.0'
            }
        })

    def call_tool(self, name: str, args: Dict[str, Any] = None):
        return self.send_request('tools/call', {
            'name': name,
            'arguments': args or {}
        })

    def disconnect(self):
        if self.process:
            self.process.terminate()

# 使用示例
if __name__ == '__main__':
    client = ROMCMCPClient()
    client.connect()
    
    # 列出资源
    resources = client.call_tool('list_resources', {
        'resource_type': 'linux'
    })
    print('Resources:', resources)
    
    client.disconnect()
```

## 最佳实践

### 1. 错误处理

始终捕获和处理 MCP 调用可能的错误：

```javascript
try {
  const result = await mcpClient.callTool('ssh_execute_command', args)
  console.log('Success:', result)
} catch (error) {
  console.error('MCP Error:', error.message)
  // 友好的错误提示
  if (error.message.includes('connection')) {
    alert('无法连接到服务器，请检查网络')
  }
}
```

### 2. 超时处理

设置合理的超时时间：

```javascript
const timeout = (ms) => new Promise((_, reject) => 
  setTimeout(() => reject(new Error('Timeout')), ms)
)

const result = await Promise.race([
  mcpClient.callTool('ssh_execute_command', args),
  timeout(30000)  // 30秒超时
])
```

### 3. 连接管理

保持连接活跃，实现自动重连：

```javascript
class MCPManager {
  async ensureConnected() {
    if (!this.isConnected()) {
      await this.reconnect()
    }
  }

  async reconnect() {
    for (let i = 0; i < 3; i++) {
      try {
        await this.connect()
        return
      } catch (error) {
        if (i === 2) throw error
        await new Promise(r => setTimeout(r, 1000 * (i + 1)))
      }
    }
  }
}
```

### 4. 权限控制

在调用工具前验证权限：

```javascript
async function safeCallTool(toolName, args) {
  // 检查是否有权限
  if (!hasPermission(toolName)) {
    throw new Error('没有权限执行此操作')
  }
  
  // 敏感操作二次确认
  if (DANGEROUS_TOOLS.includes(toolName)) {
    const confirmed = await confirm(`确定要执行 ${toolName}？`)
    if (!confirmed) return
  }
  
  return mcpClient.callTool(toolName, args)
}
```

### 5. 日志记录

记录所有 MCP 调用：

```javascript
async function loggedCallTool(toolName, args) {
  console.log(`[MCP] Calling ${toolName}`, args)
  const startTime = Date.now()
  
  try {
    const result = await mcpClient.callTool(toolName, args)
    const duration = Date.now() - startTime
    console.log(`[MCP] ${toolName} succeeded in ${duration}ms`)
    return result
  } catch (error) {
    console.error(`[MCP] ${toolName} failed:`, error)
    throw error
  }
}
```

## 常见问题

### Q: stdio 和 socket 模式该如何选择？

**A**: 
- 本地使用、单用户：选择 stdio
- 远程访问、多用户：选择 socket
- Web 应用集成：必须使用 socket

### Q: 如何调试 MCP 通信？

**A**: 
```bash
# 启用调试日志
export ROMA_LOG_LEVEL=debug
roma mcp

# 或查看通信内容
roma mcp --log-mcp-messages
```

### Q: MCP 连接失败怎么办？

**A**: 
1. 检查 ROMA 是否正常运行
2. 验证配置文件路径正确
3. 确认端口没有被占用（socket 模式）
4. 查看 ROMA 日志

### Q: 如何添加自定义 MCP 工具？

**A**: 参考 `/usr/sourcecode/roma/mcp/server/tools/` 下的示例代码。

## 参考资源

- [MCP 官方文档](https://modelcontextprotocol.io)
- [ROMA MCP 工具列表](./MCP_TOOLS_COMPLETE.md)
- [运维客户端开发指南](../web/ops-client/README.md)


