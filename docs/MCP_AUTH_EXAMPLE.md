# MCP 身份验证使用示例

## 快速开始

### 1. 生成访问令牌

```bash
# 登录获取 JWT
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "password123"
  }'

# 响应
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... }
}

# 生成 MCP 令牌（24小时有效）
curl -X POST http://localhost:8080/api/v1/mcp/tokens \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "expires_in": "24h",
    "description": "我的运维客户端"
  }'

# 响应
{
  "success": true,
  "data": {
    "token": "mcp_1234567890_abcdef",
    "user_id": 5,
    "username": "alice",
    "expires_at": "2024-01-16T10:00:00Z"
  },
  "message": "令牌创建成功，请妥善保存（只显示一次）"
}
```

### 2. 使用令牌调用 MCP 工具

```bash
# 列出有权限的资源
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "list_resources",
    "params": {
      "resource_type": "linux"
    },
    "auth": {
      "token": "mcp_1234567890_abcdef"
    }
  }'

# 响应（只返回 alice 有权限的资源）
{
  "success": true,
  "data": {
    "user": "alice",
    "resource_type": "linux",
    "count": 3,
    "resources": [
      {
        "id": 1,
        "name": "web-01",
        "address": "192.168.1.10",
        "status": "online"
      },
      {
        "id": 2,
        "name": "web-02",
        "address": "192.168.1.11",
        "status": "online"
      },
      {
        "id": 3,
        "name": "db-01",
        "address": "192.168.1.20",
        "status": "online"
      }
    ]
  }
}
```

### 3. 执行命令（需要执行权限）

```bash
# 在资源上执行命令
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "execute_command",
    "params": {
      "resource_id": 1,
      "command": "df -h"
    },
    "auth": {
      "token": "mcp_1234567890_abcdef"
    }
  }'

# 成功响应
{
  "success": true,
  "data": {
    "user": "alice",
    "resource_id": 1,
    "command": "df -h",
    "output": "Filesystem      Size  Used Avail Use%\n/dev/sda1       100G   45G   55G  45%",
    "exit_code": 0
  }
}

# 权限不足时
{
  "success": false,
  "error": {
    "code": "PERMISSION_DENIED",
    "message": "权限不足: 该用户没有执行权限"
  }
}
```

## 客户端集成

### JavaScript (Electron/Browser)

```javascript
// mcp-client-with-auth.js
class SecureMCPClient {
    constructor(apiUrl, token) {
        this.apiUrl = apiUrl
        this.token = token
    }

    async callTool(method, params) {
        const response = await fetch(`${this.apiUrl}/mcp`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                method,
                params,
                auth: {
                    token: this.token
                }
            })
        })

        const result = await response.json()

        if (!result.success) {
            throw new Error(result.error.message)
        }

        return result.data
    }

    async listResources(type = 'all') {
        return await this.callTool('list_resources', {
            resource_type: type
        })
    }

    async executeCommand(resourceId, command) {
        return await this.callTool('execute_command', {
            resource_id: resourceId,
            command: command
        })
    }

    async getSystemInfo(resourceId) {
        return await this.callTool('get_system_info', {
            resource_id: resourceId
        })
    }
}

// 使用
const client = new SecureMCPClient(
    'http://localhost:8080/api/v1',
    'mcp_1234567890_abcdef'
)

try {
    const resources = await client.listResources('linux')
    console.log('我的资源:', resources)
} catch (error) {
    if (error.message.includes('PERMISSION_DENIED')) {
        console.error('权限不足')
    } else if (error.message.includes('INVALID_TOKEN')) {
        console.error('令牌无效，请重新登录')
    } else {
        console.error('错误:', error.message)
    }
}
```

### Python

```python
# mcp_client.py
import requests
import json

class SecureMCPClient:
    def __init__(self, api_url, token):
        self.api_url = api_url
        self.token = token
    
    def call_tool(self, method, params):
        response = requests.post(
            f"{self.api_url}/mcp",
            json={
                "method": method,
                "params": params,
                "auth": {
                    "token": self.token
                }
            }
        )
        
        result = response.json()
        
        if not result.get('success'):
            error = result.get('error', {})
            raise Exception(f"{error.get('code')}: {error.get('message')}")
        
        return result.get('data')
    
    def list_resources(self, resource_type='all'):
        return self.call_tool('list_resources', {
            'resource_type': resource_type
        })
    
    def execute_command(self, resource_id, command):
        return self.call_tool('execute_command', {
            'resource_id': resource_id,
            'command': command
        })

# 使用
client = SecureMCPClient(
    'http://localhost:8080/api/v1',
    'mcp_1234567890_abcdef'
)

try:
    resources = client.list_resources('linux')
    print(f"我的资源: {resources}")
    
    # 执行命令
    result = client.execute_command(1, 'uptime')
    print(f"命令输出: {result['output']}")
    
except Exception as e:
    print(f"错误: {e}")
```

### Go

```go
// mcp_client.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MCPClient struct {
	APIURL string
	Token  string
}

type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	Auth   map[string]string      `json:"auth"`
}

type MCPResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   *MCPError              `json:"error,omitempty"`
}

type MCPError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewMCPClient(apiURL, token string) *MCPClient {
	return &MCPClient{
		APIURL: apiURL,
		Token:  token,
	}
}

func (c *MCPClient) CallTool(method string, params map[string]interface{}) (map[string]interface{}, error) {
	request := MCPRequest{
		Method: method,
		Params: params,
		Auth: map[string]string{
			"token": c.Token,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/mcp", c.APIURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result MCPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Success {
		return nil, fmt.Errorf("%s: %s", result.Error.Code, result.Error.Message)
	}

	return result.Data, nil
}

func (c *MCPClient) ListResources(resourceType string) (map[string]interface{}, error) {
	return c.CallTool("list_resources", map[string]interface{}{
		"resource_type": resourceType,
	})
}

func (c *MCPClient) ExecuteCommand(resourceID int, command string) (map[string]interface{}, error) {
	return c.CallTool("execute_command", map[string]interface{}{
		"resource_id": resourceID,
		"command":     command,
	})
}

func main() {
	client := NewMCPClient(
		"http://localhost:8080/api/v1",
		"mcp_1234567890_abcdef",
	)

	// 列出资源
	resources, err := client.ListResources("linux")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("我的资源: %+v\n", resources)

	// 执行命令
	result, err := client.ExecuteCommand(1, "hostname")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}
	fmt.Printf("命令输出: %s\n", result["output"])
}
```

## 令牌管理

### 列出所有令牌

```bash
curl http://localhost:8080/api/v1/mcp/tokens \
  -H "Authorization: Bearer <jwt_token>"

# 响应
{
  "success": true,
  "data": {
    "count": 2,
    "tokens": [
      {
        "token": "mcp_...cdef",  // 已脱敏
        "user_id": 5,
        "username": "alice",
        "expires_at": "2024-01-16T10:00:00Z",
        "created_at": "2024-01-15T10:00:00Z"
      },
      {
        "token": "mcp_...xyz9",
        "user_id": 5,
        "username": "alice",
        "expires_at": "2024-01-20T10:00:00Z",
        "created_at": "2024-01-14T10:00:00Z"
      }
    ]
  }
}
```

### 撤销令牌

```bash
curl -X DELETE http://localhost:8080/api/v1/mcp/tokens/mcp_1234567890_abcdef \
  -H "Authorization: Bearer <jwt_token>"

# 响应
{
  "success": true,
  "message": "令牌已撤销"
}
```

### 验证令牌

```bash
curl "http://localhost:8080/api/v1/mcp/tokens/validate?token=mcp_1234567890_abcdef"

# 有效
{
  "success": true,
  "valid": true,
  "data": {
    "user_id": 5,
    "username": "alice",
    "roles": ["operator", "developer"]
  }
}

# 无效
{
  "success": false,
  "valid": false,
  "error": "无效的认证令牌"
}
```

## 错误处理

### 常见错误代码

| 错误代码 | 说明 | 处理方式 |
|---------|------|---------|
| `INVALID_TOKEN` | 令牌无效 | 重新登录获取新令牌 |
| `TOKEN_EXPIRED` | 令牌已过期 | 刷新或重新生成令牌 |
| `PERMISSION_DENIED` | 权限不足 | 联系管理员授权 |
| `USER_NOT_FOUND` | 用户不存在 | 检查用户状态 |
| `USER_DISABLED` | 用户已禁用 | 联系管理员启用账号 |
| `INTERNAL_ERROR` | 内部错误 | 检查服务器日志 |

### 错误处理示例

```javascript
async function callMCPWithRetry(client, method, params, maxRetries = 3) {
    for (let i = 0; i < maxRetries; i++) {
        try {
            return await client.callTool(method, params)
        } catch (error) {
            if (error.message.includes('TOKEN_EXPIRED')) {
                // 令牌过期，刷新后重试
                await refreshToken()
                continue
            } else if (error.message.includes('PERMISSION_DENIED')) {
                // 权限不足，不重试
                throw new Error('权限不足，请联系管理员')
            } else if (i === maxRetries - 1) {
                // 最后一次重试失败
                throw error
            }
            
            // 指数退避
            await sleep(Math.pow(2, i) * 1000)
        }
    }
}
```

## 审计日志查询

```bash
# 查看我的操作记录
curl http://localhost:8080/api/v1/logs/access?user_id=5 \
  -H "Authorization: Bearer <jwt_token>"

# 查看失败的访问尝试
curl http://localhost:8080/api/v1/logs/access?status=failed&limit=100 \
  -H "Authorization: Bearer <jwt_token>"

# 查看高危操作
curl http://localhost:8080/api/v1/logs/access?action=execute_command&start_time=2024-01-15 \
  -H "Authorization: Bearer <jwt_token>"
```

---

**完整的身份验证让 MCP 更安全！** 🔒


