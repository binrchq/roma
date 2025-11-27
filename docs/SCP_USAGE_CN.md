# ROMA SCP 文件传输指南

本文档详细说明如何通过ROMA堡垒机使用SCP进行文件传输。

---

## 📋 目录

- [工作原理](#工作原理)
- [基本用法](#基本用法)
- [高级用法](#高级用法)
- [常见场景](#常见场景)
- [故障排查](#故障排查)
- [限制说明](#限制说明)

---

## 🔧 工作原理

ROMA实现了标准SCP协议的代理功能，允许通过堡垒机中转文件传输：

```
┌──────────┐         ┌──────────────┐         ┌──────────────┐
│  本地    │  SCP    │  ROMA        │  SCP    │  目标服务器  │
│  客户端  ├────────►│  堡垒机      ├────────►│  (web-01)    │
│          │ Port    │  (Port 2200) │         │              │
└──────────┘  2200   └──────────────┘         └──────────────┘
```

**传输流程:**

1. 客户端发起SCP连接到ROMA堡垒机 (端口2200)
2. ROMA解析特殊路径格式: `user@hostname:/path`
3. ROMA根据hostname查找目标服务器配置
4. ROMA建立到目标服务器的SCP连接
5. 文件数据通过ROMA中转传输
6. 所有操作记录到审计日志

---

## 🚀 基本用法

### 路径格式

ROMA SCP使用特殊的路径格式来指定目标服务器：

```
user@jumpserver:user@hostname:/remote/path
│              │ │          │ │
│              │ │          │ └─ 目标服务器文件路径
│              │ │          └─── 目标服务器hostname
│              │ └────────────── 目标服务器用户
│              └──────────────── ROMA堡垒机地址
└───────────────────────────── ROMA用户名
```

**关键点:**
- `hostname` 必须是在ROMA中注册的服务器主机名
- 用户需要有访问该资源的权限
- 支持通过IP地址（如果在ROMA中以IP注册）

### 上传文件

**基本语法:**

```bash
scp -P <roma_port> <local_file> <user>@<roma_host>:<user>@<hostname>:<remote_path>
```

**示例:**

```bash
# 上传单个文件
scp -P 2200 /tmp/app.log user@roma.example.com:user@web-01:/var/log/

# 使用SSH密钥
scp -P 2200 -i ~/.ssh/roma_key config.yaml user@roma.example.com:user@web-01:/etc/app/

# 指定权限保留
scp -P 2200 -p script.sh user@roma.example.com:user@web-01:/usr/local/bin/
```

### 下载文件

**基本语法:**

```bash
scp -P <roma_port> <user>@<roma_host>:<user>@<hostname>:<remote_path> <local_path>
```

**示例:**

```bash
# 下载单个文件
scp -P 2200 user@roma.example.com:user@web-01:/var/log/app.log ./

# 下载到指定目录
scp -P 2200 user@roma.example.com:user@db-01:/backup/db.sql.gz ./backup/

# 重命名下载的文件
scp -P 2200 user@roma.example.com:user@web-01:/etc/nginx/nginx.conf ./nginx.conf.backup
```

---

## 🎯 高级用法

### 使用配置文件

创建 `~/.ssh/config` 简化命令：

```
Host roma
    HostName roma.example.com
    Port 2200
    User your-username
    IdentityFile ~/.ssh/roma_key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

**使用配置后的命令:**

```bash
# 上传
scp /tmp/file.txt roma:user@web-01:/tmp/

# 下载
scp roma:user@web-01:/tmp/file.txt ./
```

### 批量传输（压缩）

由于ROMA暂不支持文件夹传输，建议先压缩：

```bash
# 1. 压缩文件夹
tar -czf app.tar.gz /path/to/app/

# 2. 上传压缩包
scp -P 2200 app.tar.gz user@roma:user@web-01:/tmp/

# 3. SSH到服务器解压
ssh user@roma -p 2200
roma> ln -t linux web-01
web-01$ cd /tmp && tar -xzf app.tar.gz
```

### 自动化脚本

**上传脚本示例:**

```bash
#!/bin/bash
# upload-to-server.sh

ROMA_HOST="roma.example.com"
ROMA_PORT="2200"
ROMA_USER="admin"
SSH_KEY="~/.ssh/roma_key"

# 配置
TARGET_SERVER="web-01"
LOCAL_FILE="$1"
REMOTE_PATH="$2"

if [ -z "$LOCAL_FILE" ] || [ -z "$REMOTE_PATH" ]; then
    echo "用法: $0 <本地文件> <远程路径>"
    exit 1
fi

# 检查文件是否存在
if [ ! -f "$LOCAL_FILE" ]; then
    echo "错误: 文件不存在 $LOCAL_FILE"
    exit 1
fi

# 执行上传
echo "正在上传 $LOCAL_FILE 到 $TARGET_SERVER:$REMOTE_PATH ..."
scp -P "$ROMA_PORT" -i "$SSH_KEY" \
    -o StrictHostKeyChecking=no \
    -o UserKnownHostsFile=/dev/null \
    "$LOCAL_FILE" \
    "${ROMA_USER}@${ROMA_HOST}:${ROMA_USER}@${TARGET_SERVER}:${REMOTE_PATH}"

if [ $? -eq 0 ]; then
    echo "✓ 上传成功"
else
    echo "✗ 上传失败"
    exit 1
fi
```

**使用示例:**

```bash
chmod +x upload-to-server.sh
./upload-to-server.sh /tmp/app.log /var/log/
```

### 进度显示

使用 `rsync` 通过ROMA传输并显示进度：

```bash
# 注意: 需要目标服务器支持rsync
rsync -avz --progress -e "ssh -p 2200" \
    /local/file.txt \
    user@roma:user@web-01:/remote/path/
```

---

## 💼 常见场景

### 场景1: 部署应用

```bash
# 1. 打包应用
tar -czf myapp-v1.0.tar.gz /path/to/myapp/

# 2. 上传到服务器
scp -P 2200 myapp-v1.0.tar.gz user@roma:user@prod-server:/opt/apps/

# 3. SSH登录并部署
ssh user@roma -p 2200
roma> ln -t linux prod-server
prod-server$ cd /opt/apps && tar -xzf myapp-v1.0.tar.gz
prod-server$ systemctl restart myapp
```

### 场景2: 备份数据库

```bash
# 1. SSH到数据库服务器创建备份
ssh user@roma -p 2200
roma> ln -t linux db-01
db-01$ mysqldump -u root -p mydb > /tmp/mydb-backup.sql
db-01$ gzip /tmp/mydb-backup.sql
db-01$ exit

# 2. 下载备份到本地
scp -P 2200 user@roma:user@db-01:/tmp/mydb-backup.sql.gz ./backups/mydb-$(date +%Y%m%d).sql.gz
```

### 场景3: 批量更新配置

```bash
#!/bin/bash
# deploy-config.sh - 批量部署配置文件

SERVERS=("web-01" "web-02" "web-03")
CONFIG_FILE="nginx.conf"
REMOTE_PATH="/etc/nginx/nginx.conf"

for server in "${SERVERS[@]}"; do
    echo "正在部署配置到 $server ..."
    scp -P 2200 "$CONFIG_FILE" "user@roma:user@$server:$REMOTE_PATH"
    
    if [ $? -eq 0 ]; then
        echo "✓ $server 部署成功"
        
        # 重启nginx
        ssh user@roma -p 2200 "ln -t linux $server -- 'systemctl reload nginx'"
    else
        echo "✗ $server 部署失败"
    fi
done
```

### 场景4: 日志收集

```bash
#!/bin/bash
# collect-logs.sh - 从多台服务器收集日志

SERVERS=("web-01" "web-02" "api-01")
LOG_PATH="/var/log/app/error.log"
OUTPUT_DIR="./logs/$(date +%Y%m%d)"

mkdir -p "$OUTPUT_DIR"

for server in "${SERVERS[@]}"; do
    echo "正在从 $server 收集日志..."
    scp -P 2200 "user@roma:user@$server:$LOG_PATH" "$OUTPUT_DIR/${server}-error.log"
done

echo "✓ 日志已收集到 $OUTPUT_DIR"
```

### 场景5: Windows服务器文件传输

```bash
# 上传到Windows服务器（需要OpenSSH Server）
scp -P 2200 app.zip user@roma:user@win-server-01:/C:/temp/

# 下载Windows服务器文件
scp -P 2200 user@roma:user@win-server-01:/C:/logs/app.log ./
```

---

## 🐛 故障排查

### 问题1: 连接被拒绝

**错误信息:**
```
ssh: connect to host roma.example.com port 2200: Connection refused
```

**解决方法:**
```bash
# 1. 检查ROMA是否运行
ssh user@roma-host -p 2200

# 2. 检查防火墙
sudo ufw status
sudo ufw allow 2200/tcp

# 3. 检查ROMA配置
grep "port.*2200" /path/to/roma/config.toml
```

### 问题2: 权限被拒绝

**错误信息:**
```
Permission denied (publickey)
```

**解决方法:**
```bash
# 1. 检查SSH密钥权限
chmod 600 ~/.ssh/roma_key

# 2. 确认公钥已上传到ROMA
# 通过Web UI: Settings -> SSH Keys

# 3. 使用详细模式查看错误
scp -v -P 2200 file.txt user@roma:user@web-01:/tmp/
```

### 问题3: 找不到资源

**错误信息:**
```
resource not found: hostname 'web-01'
```

**解决方法:**
```bash
# 1. 检查hostname是否正确
ssh user@roma -p 2200
roma> ls linux
roma> whoami  # 查看有权限访问的资源

# 2. 确认资源已在ROMA中注册
# 通过Web UI: Resources -> Linux

# 3. 检查用户权限
# 通过Web UI: Users -> [Your User] -> Permissions
```

### 问题4: 传输中断

**错误信息:**
```
Connection closed by remote host
```

**解决方法:**
```bash
# 1. 检查网络连接
ping roma.example.com

# 2. 增加超时时间
scp -o ConnectTimeout=30 -P 2200 file.txt user@roma:user@web-01:/tmp/

# 3. 检查ROMA日志
tail -f /var/log/roma/roma.log
```

### 问题5: 文件夹传输失败

**错误信息:**
```
Folder transfer is not yet supported. You can try to compress the folder and upload it.
```

**解决方法:**
```bash
# ROMA暂不支持文件夹传输，需要先压缩
tar -czf folder.tar.gz /path/to/folder/
scp -P 2200 folder.tar.gz user@roma:user@web-01:/tmp/

# 然后SSH到服务器解压
ssh user@roma -p 2200
roma> ln -t linux web-01
web-01$ tar -xzf /tmp/folder.tar.gz
```

---

## ⚠️ 限制说明

### 当前限制

1. **不支持文件夹传输**
   - ❌ 直接传输文件夹
   - ✅ 压缩后传输单个文件

2. **不支持通配符**
   - ❌ `scp *.txt user@roma:...`
   - ✅ 使用脚本循环传输多个文件

3. **不支持递归传输**
   - ❌ `scp -r folder/ user@roma:...`
   - ✅ 先压缩再传输

### 支持的资源类型

| 资源类型 | 上传 | 下载 | 说明 |
|---------|------|------|------|
| Linux服务器 | ✅ | ✅ | 完全支持 |
| Windows服务器 | ✅ | ✅ | 需要OpenSSH Server |
| Docker容器 | ❌ | ❌ | 不支持（可通过宿主机） |
| 数据库 | ❌ | ❌ | 不支持 |

### 性能建议

- **小文件** (< 10MB): 直接传输
- **大文件** (> 100MB): 建议压缩后传输
- **大量小文件**: 打包成tar.gz后传输
- **网络不稳定**: 使用rsync替代scp

---

## 📚 相关文档

- [DEPLOYMENT.md](DEPLOYMENT.md) - 部署指南
- [SECURITY.md](SECURITY.md) - 安全配置
- [API.md](API.md) - API文档

---

## 💡 最佳实践

1. **使用SSH配置文件** - 简化命令行参数
2. **压缩传输** - 减少传输时间和流量
3. **自动化脚本** - 批量操作使用脚本
4. **审计日志** - 定期检查文件传输记录
5. **权限控制** - 遵循最小权限原则
6. **备份验证** - 传输后验证文件完整性

```bash
# 验证文件完整性
# 1. 传输前计算校验和
md5sum local_file.txt

# 2. 传输文件
scp -P 2200 local_file.txt user@roma:user@web-01:/tmp/

# 3. 传输后验证
ssh user@roma -p 2200 "ln -t linux web-01 -- 'md5sum /tmp/local_file.txt'"
```

---

**ROMA SCP** - 安全、高效的文件传输解决方案 🚀

