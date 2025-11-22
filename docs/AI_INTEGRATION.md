# 运维客户端 AI 集成指南

本文档详细说明如何在 ROMA 运维客户端中集成大语言模型。

## 🎯 集成方案对比

| 方案 | 实现难度 | 功能完整度 | 成本 | 适用场景 |
|-----|---------|----------|------|---------|
| **直接集成 LLM API** | ⭐⭐ | ⭐⭐⭐⭐⭐ | 💰💰 | 最推荐，功能完整 |
| **MCP 桥接** | ⭐⭐⭐ | ⭐⭐⭐⭐ | 💰 | 配合 Cursor/Claude |
| **本地模型 (Ollama)** | ⭐⭐⭐⭐ | ⭐⭐⭐ | 免费 | 内网/安全要求高 |
| **自建 LLM 服务** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | 💰💰💰 | 企业级部署 |

## 方案1: 直接集成 LLM API（已实现）✅

### 架构

```
运维客户端
├── ai-assistant.js    # AI 核心逻辑
├── chat-ui.js         # 聊天界面
└── ROMA API           # 工具执行层
    └── Function Calling
        ├── list_resources
        ├── ssh_execute
        ├── get_system_info
        ├── batch_execute
        └── get_logs
```

### 支持的 AI 提供商

#### 1. OpenAI

```javascript
// 配置
{
  "provider": "openai",
  "apiKey": "sk-...",
  "model": "gpt-4",
  "baseUrl": "https://api.openai.com/v1"  // 可选
}
```

**推荐模型：**
- `gpt-4` - 最强能力，适合复杂任务
- `gpt-4-turbo` - 性价比高
- `gpt-3.5-turbo` - 速度快，适合简单任务

**特点：**
- ✅ Function Calling 支持完善
- ✅ 响应速度快
- ✅ 稳定性好
- ❌ 需要科学上网（国内）
- ❌ 相对较贵

#### 2. Anthropic (Claude)

```javascript
{
  "provider": "anthropic",
  "apiKey": "sk-ant-...",
  "model": "claude-3-5-sonnet-20241022",
  "baseUrl": "https://api.anthropic.com/v1"
}
```

**推荐模型：**
- `claude-3-5-sonnet-20241022` - 最新，最强
- `claude-3-opus-20240229` - 旗舰模型
- `claude-3-sonnet-20240229` - 均衡

**特点：**
- ✅ 理解能力强
- ✅ 长上下文（200K tokens）
- ✅ 安全性好
- ❌ 国内访问困难
- ❌ 价格较高

#### 3. DeepSeek（国产，推荐）

```javascript
{
  "provider": "deepseek",
  "apiKey": "sk-...",
  "model": "deepseek-chat",
  "baseUrl": "https://api.deepseek.com/v1"
}
```

**推荐模型：**
- `deepseek-chat` - 通用对话
- `deepseek-coder` - 代码专用

**特点：**
- ✅ 国内可直接访问
- ✅ 价格便宜（0.001元/千tokens）
- ✅ 中文理解好
- ✅ Function Calling 支持
- ⭐ **最推荐国内用户使用**

#### 4. Ollama（本地）

```javascript
{
  "provider": "ollama",
  "apiKey": "",  // 不需要
  "model": "llama3",
  "baseUrl": "http://localhost:11434"
}
```

**推荐模型：**
- `llama3` - Meta 开源，通用能力强
- `qwen` - 阿里通义千问，中文好
- `deepseek-coder` - 代码专用
- `mistral` - 欧洲开源，均衡

**特点：**
- ✅ 完全免费
- ✅ 数据私密
- ✅ 离线可用
- ❌ 需要本地安装
- ❌ 性能要求高（需要 GPU）

### 使用示例

#### 基础对话

```javascript
const ai = new AIAssistant()
ai.updateConfig({
    provider: 'deepseek',
    apiKey: 'your-api-key',
    model: 'deepseek-chat'
})

const response = await ai.chat('列出所有 Linux 服务器')
console.log(response.message)
```

#### 带工具调用

```javascript
const tools = ai.getToolDefinitions()
const response = await ai.chat('检查 web-01 的磁盘使用情况', tools)

if (response.needsToolExecution) {
    // 自动执行工具并获取结果
    const finalResponse = await ai.executeToolAndContinue(response.toolCalls)
    console.log(finalResponse.message)
}
```

#### 完整示例（客户端使用）

```javascript
// 用户输入："批量检查所有 Web 服务器的 Nginx 状态"

// 1. AI 理解意图并调用工具
toolCalls: [
    {
        name: 'list_resources',
        arguments: { type: 'linux' }
    }
]

// 2. 执行工具，获取服务器列表
[
    { name: 'web-01', host: '192.168.1.10' },
    { name: 'web-02', host: '192.168.1.11' }
]

// 3. AI 继续调用
toolCalls: [
    {
        name: 'batch_execute',
        arguments: {
            resource_type: 'linux',
            identifiers: ['web-01', 'web-02'],
            command: 'systemctl status nginx'
        }
    }
]

// 4. 执行并返回结果
// 5. AI 总结："两台 Web 服务器的 Nginx 都在正常运行..."
```

## 方案2: MCP 桥接模式

如果你想让客户端通过 MCP 与外部 AI（Cursor/Claude）协作：

### 架构

```
AI IDE (Cursor)
    ↓ MCP Protocol
ROMA MCP Server (运维客户端内部)
    ↓ API Calls
ROMA 后端 API
    ↓
边缘服务器
```

### 实现步骤

1. **在客户端内启动 MCP Server**

```javascript
// 已有的 mcp-client.js
const mcpClient = new MCPClient()
await mcpClient.connect({ transport: 'stdio' })
```

2. **Cursor 配置**

```json
{
  "roma-ops-client": {
    "command": "/path/to/ops-client",
    "args": ["--mcp-mode"],
    "env": {
      "ROMA_API_URL": "http://localhost:8080/api/v1",
      "ROMA_API_KEY": "your-key"
    }
  }
}
```

3. **使用**

在 Cursor 中直接对话：
```
"ROMA，列出所有服务器"
"在 web-01 上检查磁盘"
```

### 优缺点

**优点：**
- ✅ 利用 Cursor 的强大 AI
- ✅ 统一的开发体验

**缺点：**
- ❌ 依赖外部 AI IDE
- ❌ 不能独立使用
- ❌ 运维人员可能不熟悉 IDE

## 方案3: 本地模型（Ollama）

### 安装 Ollama

```bash
# macOS/Linux
curl -fsSL https://ollama.com/install.sh | sh

# Windows
# 下载安装包：https://ollama.com/download

# 拉取模型
ollama pull llama3
ollama pull qwen  # 中文更好
```

### 客户端配置

```javascript
{
  "provider": "ollama",
  "model": "llama3",
  "baseUrl": "http://localhost:11434"
}
```

### 优化建议

1. **模型选择**
   - 内存 8GB: `llama3:8b`
   - 内存 16GB: `llama3:13b` 或 `qwen:14b`
   - 内存 32GB+: `llama3:70b`

2. **性能优化**
   ```bash
   # 启用 GPU 加速
   ollama run llama3 --gpu
   
   # 调整上下文长度
   ollama run llama3 --context-length 8192
   ```

3. **Function Calling**
   
   Ollama 的 Function Calling 支持有限，可以通过 Prompt Engineering：

```javascript
const systemPrompt = `你是运维助手。当用户需要执行操作时，返回 JSON 格式：
{
  "action": "工具名称",
  "params": { "参数": "值" }
}

可用工具：
- list_resources: 列出资源
- ssh_execute: 执行命令
...`
```

## 方案4: 自建 LLM 服务

适合大型企业，完全自主可控。

### 架构

```
运维客户端
    ↓ HTTP/gRPC
企业 LLM 网关
    ├── 负载均衡
    ├── 鉴权限流
    └── 审计日志
        ↓
多个 LLM 实例
    ├── GPU 服务器 1
    ├── GPU 服务器 2
    └── GPU 服务器 N
```

### 技术选型

1. **推理框架**
   - vLLM - 高性能推理
   - Text-Generation-Inference - Hugging Face 官方
   - TensorRT-LLM - NVIDIA 优化

2. **模型管理**
   - Ray Serve - 分布式部署
   - KServe - Kubernetes 原生
   - Triton - NVIDIA 推理服务器

3. **网关**
   - FastAPI + Nginx
   - Kong + LLM 插件
   - 自研网关

### 示例部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  llm-server:
    image: vllm/vllm-openai:latest
    command: >
      --model deepseek-ai/deepseek-coder-6.7b-instruct
      --gpu-memory-utilization 0.9
      --max-num-seqs 256
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    ports:
      - "8000:8000"
```

## 成本分析

### OpenAI GPT-4

- 输入：$0.03 / 1K tokens
- 输出：$0.06 / 1K tokens
- 月成本（中等使用）：$50-200

### Claude 3.5 Sonnet

- 输入：$0.003 / 1K tokens  
- 输出：$0.015 / 1K tokens
- 月成本：$30-150

### DeepSeek（推荐）

- 输入：¥0.001 / 1K tokens
- 输出：¥0.002 / 1K tokens
- 月成本：¥10-50（约 $1.5-7）
- **性价比最高！**

### Ollama（本地）

- 一次性成本：GPU 服务器
- 运行成本：电费
- 月成本：基本免费

### 自建服务

- 一次性：GPU 服务器 + 部署
- 运行成本：电费 + 运维
- 月成本：$500-5000+

## 推荐方案

### 个人/小团队（1-10人）

**首选：DeepSeek**
```javascript
{
  "provider": "deepseek",
  "apiKey": "sk-...",
  "model": "deepseek-chat"
}
```

理由：
- ✅ 价格便宜（月成本 < ¥50）
- ✅ 国内直连
- ✅ 功能完整
- ✅ 无需部署

### 中型团队（10-50人）

**首选：DeepSeek + Ollama 备用**

主用 DeepSeek API，内网服务器部署 Ollama 作为备用：

```javascript
// 自动切换
const provider = isInternalNetwork ? 'ollama' : 'deepseek'
```

### 大型企业（50+人）

**首选：自建 LLM 服务**

- 部署 vLLM + DeepSeek 开源模型
- 统一网关管理
- 完整审计日志
- 数据不出内网

## 安全建议

1. **API Key 管理**
   ```javascript
   // 不要硬编码
   ❌ const apiKey = 'sk-xxxx'
   
   // 使用环境变量
   ✅ const apiKey = process.env.AI_API_KEY
   
   // 或加密存储
   ✅ const apiKey = decrypt(localStorage.getItem('encrypted_key'))
   ```

2. **请求审计**
   ```javascript
   function logAIRequest(prompt, response) {
       console.log({
           timestamp: new Date(),
           user: getCurrentUser(),
           prompt: maskSensitiveData(prompt),
           success: !!response
       })
   }
   ```

3. **数据脱敏**
   ```javascript
   function maskSensitiveData(text) {
       return text
           .replace(/password[=:]\s*\S+/gi, 'password=***')
           .replace(/\d{15,}/g, '***')  // 身份证号
           .replace(/\b\d{3}-\d{4}-\d{4}\b/g, '***')  // 电话
   }
   ```

4. **成本控制**
   ```javascript
   class RateLimiter {
       constructor(maxRequestsPerMinute = 20) {
           this.max = maxRequestsPerMinute
           this.requests = []
       }
       
       async checkLimit() {
           const now = Date.now()
           this.requests = this.requests.filter(t => now - t < 60000)
           
           if (this.requests.length >= this.max) {
               throw new Error('请求过于频繁，请稍后再试')
           }
           
           this.requests.push(now)
       }
   }
   ```

## 总结

| 场景 | 推荐方案 | 月成本 |
|-----|---------|-------|
| **个人用户** | DeepSeek | ¥10-30 |
| **小团队** | DeepSeek | ¥30-100 |
| **中型企业** | DeepSeek + Ollama | ¥100-500 |
| **大型企业** | 自建服务 | ¥5000+ |
| **内网环境** | Ollama | 免费 |
| **开发者** | MCP 桥接（Cursor） | IDE 订阅费 |

✅ 已在运维客户端中实现完整的 AI 集成，支持多种提供商，开箱即用！


