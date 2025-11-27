# ROMA SCP File Transfer Guide

This document provides detailed instructions for file transfer using SCP through ROMA jump server.

[中文版本](./SCP_USAGE_CN.md)

---

## Table of Contents

- [How It Works](#how-it-works)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Common Scenarios](#common-scenarios)
- [Troubleshooting](#troubleshooting)
- [Limitations](#limitations)

---

## How It Works

ROMA implements standard SCP protocol proxy, allowing file transfer through the jump server:

```
┌──────────┐         ┌──────────────┐         ┌──────────────┐
│  Local   │  SCP    │  ROMA        │  SCP    │  Target      │
│  Client  ├────────►│  Jump Server ├────────►│  Server      │
│          │ Port    │  (Port 2200) │         │  (web-01)    │
└──────────┘  2200   └──────────────┘         └──────────────┘
```

**Transfer Flow:**

1. Client initiates SCP connection to ROMA jump server (port 2200)
2. ROMA parses special path format: `user@hostname:/path`
3. ROMA looks up target server configuration by hostname
4. ROMA establishes SCP connection to target server
5. File data is transferred through ROMA relay
6. All operations are recorded in audit logs

---

## Basic Usage

### Path Format

ROMA SCP uses a special path format to specify target server:

```
user@jumpserver:user@hostname:/remote/path
│              │ │          │ │
│              │ │          │ └─ Target server file path
│              │ │          └─── Target server hostname
│              │ └────────────── Target server user
│              └──────────────── ROMA jump server address
└───────────────────────────── ROMA username
```

**Key Points:**
- `hostname` must be registered in ROMA
- User needs access permissions to the resource
- Supports IP address (if registered in ROMA)

### Upload Files

**Basic Syntax:**

```bash
scp -P <roma_port> <local_file> <user>@<roma_host>:<user>@<hostname>:<remote_path>
```

**Examples:**

```bash
# Upload single file
scp -P 2200 /tmp/app.log user@roma.example.com:user@web-01:/var/log/

# Using SSH key
scp -P 2200 -i ~/.ssh/roma_key config.yaml user@roma.example.com:user@web-01:/etc/app/

# Preserve permissions
scp -P 2200 -p script.sh user@roma.example.com:user@web-01:/usr/local/bin/
```

### Download Files

**Basic Syntax:**

```bash
scp -P <roma_port> <user>@<roma_host>:<user>@<hostname>:<remote_path> <local_path>
```

**Examples:**

```bash
# Download single file
scp -P 2200 user@roma.example.com:user@web-01:/var/log/app.log ./

# Download to specific directory
scp -P 2200 user@roma.example.com:user@db-01:/backup/db.sql.gz ./backup/

# Rename downloaded file
scp -P 2200 user@roma.example.com:user@web-01:/etc/nginx/nginx.conf ./nginx.conf.backup
```

---

## Advanced Usage

### Using Configuration File

Create `~/.ssh/config` to simplify commands:

```
Host roma
    HostName roma.example.com
    Port 2200
    User your-username
    IdentityFile ~/.ssh/roma_key
    StrictHostKeyChecking no
    UserKnownHostsFile /dev/null
```

**Commands with config:**

```bash
# Upload
scp /tmp/file.txt roma:user@web-01:/tmp/

# Download
scp roma:user@web-01:/tmp/file.txt ./
```

### Batch Transfer (Compression)

Since ROMA doesn't support directory transfer, compress first:

```bash
# 1. Compress directory
tar -czf app.tar.gz /path/to/app/

# 2. Upload archive
scp -P 2200 app.tar.gz user@roma:user@web-01:/tmp/

# 3. SSH to server and extract
ssh user@roma -p 2200
roma> ln -t linux web-01
web-01$ cd /tmp && tar -xzf app.tar.gz
```

### Automation Script

**Upload script example:**

```bash
#!/bin/bash
# upload-to-server.sh

ROMA_HOST="roma.example.com"
ROMA_PORT="2200"
ROMA_USER="admin"
SSH_KEY="~/.ssh/roma_key"

# Configuration
TARGET_SERVER="web-01"
LOCAL_FILE="$1"
REMOTE_PATH="$2"

if [ -z "$LOCAL_FILE" ] || [ -z "$REMOTE_PATH" ]; then
    echo "Usage: $0 <local_file> <remote_path>"
    exit 1
fi

# Check if file exists
if [ ! -f "$LOCAL_FILE" ]; then
    echo "Error: File not found $LOCAL_FILE"
    exit 1
fi

# Execute upload
echo "Uploading $LOCAL_FILE to $TARGET_SERVER:$REMOTE_PATH ..."
scp -P "$ROMA_PORT" -i "$SSH_KEY" \
    -o StrictHostKeyChecking=no \
    -o UserKnownHostsFile=/dev/null \
    "$LOCAL_FILE" \
    "${ROMA_USER}@${ROMA_HOST}:${ROMA_USER}@${TARGET_SERVER}:${REMOTE_PATH}"

if [ $? -eq 0 ]; then
    echo "✓ Upload successful"
else
    echo "✗ Upload failed"
    exit 1
fi
```

---

## Common Scenarios

### Scenario 1: Application Deployment

```bash
# 1. Package application
tar -czf myapp-v1.0.tar.gz /path/to/myapp/

# 2. Upload to server
scp -P 2200 myapp-v1.0.tar.gz user@roma:user@prod-server:/opt/apps/

# 3. SSH login and deploy
ssh user@roma -p 2200
roma> ln -t linux prod-server
prod-server$ cd /opt/apps && tar -xzf myapp-v1.0.tar.gz
prod-server$ systemctl restart myapp
```

### Scenario 2: Database Backup

```bash
# 1. SSH to database server and create backup
ssh user@roma -p 2200
roma> ln -t linux db-01
db-01$ mysqldump -u root -p mydb > /tmp/mydb-backup.sql
db-01$ gzip /tmp/mydb-backup.sql
db-01$ exit

# 2. Download backup to local
scp -P 2200 user@roma:user@db-01:/tmp/mydb-backup.sql.gz ./backups/mydb-$(date +%Y%m%d).sql.gz
```

### Scenario 3: Batch Configuration Update

```bash
#!/bin/bash
# deploy-config.sh - Batch deploy configuration files

SERVERS=("web-01" "web-02" "web-03")
CONFIG_FILE="nginx.conf"
REMOTE_PATH="/etc/nginx/nginx.conf"

for server in "${SERVERS[@]}"; do
    echo "Deploying configuration to $server ..."
    scp -P 2200 "$CONFIG_FILE" "user@roma:user@$server:$REMOTE_PATH"
    
    if [ $? -eq 0 ]; then
        echo "✓ $server deployed successfully"
        
        # Restart nginx
        ssh user@roma -p 2200 "ln -t linux $server -- 'systemctl reload nginx'"
    else
        echo "✗ $server deployment failed"
    fi
done
```

### Scenario 4: Log Collection

```bash
#!/bin/bash
# collect-logs.sh - Collect logs from multiple servers

SERVERS=("web-01" "web-02" "api-01")
LOG_PATH="/var/log/app/error.log"
OUTPUT_DIR="./logs/$(date +%Y%m%d)"

mkdir -p "$OUTPUT_DIR"

for server in "${SERVERS[@]}"; do
    echo "Collecting logs from $server..."
    scp -P 2200 "user@roma:user@$server:$LOG_PATH" "$OUTPUT_DIR/${server}-error.log"
done

echo "✓ Logs collected to $OUTPUT_DIR"
```

### Scenario 5: Windows Server File Transfer

```bash
# Upload to Windows server (requires OpenSSH Server)
scp -P 2200 app.zip user@roma:user@win-server-01:/C:/temp/

# Download from Windows server
scp -P 2200 user@roma:user@win-server-01:/C:/logs/app.log ./
```

---

## Troubleshooting

### Issue 1: Connection Refused

**Error Message:**
```
ssh: connect to host roma.example.com port 2200: Connection refused
```

**Solutions:**
```bash
# 1. Check if ROMA is running
ssh user@roma-host -p 2200

# 2. Check firewall
sudo ufw status
sudo ufw allow 2200/tcp

# 3. Check ROMA configuration
grep "port.*2200" /path/to/roma/config.toml
```

### Issue 2: Permission Denied

**Error Message:**
```
Permission denied (publickey)
```

**Solutions:**
```bash
# 1. Check SSH key permissions
chmod 600 ~/.ssh/roma_key

# 2. Confirm public key is uploaded to ROMA
# Via Web UI: Settings -> SSH Keys

# 3. Use verbose mode to see errors
scp -v -P 2200 file.txt user@roma:user@web-01:/tmp/
```

### Issue 3: Resource Not Found

**Error Message:**
```
resource not found: hostname 'web-01'
```

**Solutions:**
```bash
# 1. Check hostname is correct
ssh user@roma -p 2200
roma> ls linux
roma> whoami  # View accessible resources

# 2. Confirm resource is registered in ROMA
# Via Web UI: Resources -> Linux

# 3. Check user permissions
# Via Web UI: Users -> [Your User] -> Permissions
```

### Issue 4: Transfer Interrupted

**Error Message:**
```
Connection closed by remote host
```

**Solutions:**
```bash
# 1. Check network connection
ping roma.example.com

# 2. Increase timeout
scp -o ConnectTimeout=30 -P 2200 file.txt user@roma:user@web-01:/tmp/

# 3. Check ROMA logs
tail -f /var/log/roma/roma.log
```

### Issue 5: Directory Transfer Failed

**Error Message:**
```
Folder transfer is not yet supported. You can try to compress the folder and upload it.
```

**Solution:**
```bash
# ROMA doesn't support directory transfer, compress first
tar -czf folder.tar.gz /path/to/folder/
scp -P 2200 folder.tar.gz user@roma:user@web-01:/tmp/

# Then SSH to server and extract
ssh user@roma -p 2200
roma> ln -t linux web-01
web-01$ tar -xzf /tmp/folder.tar.gz
```

---

## Limitations

### Current Limitations

1. **Directory Transfer Not Supported**
   - Cannot transfer directories directly
   - Compress before transfer

2. **Wildcards Not Supported**
   - Cannot use `*.txt`
   - Use script to loop through multiple files

3. **Recursive Transfer Not Supported**
   - Cannot use `scp -r`
   - Compress before transfer

### Supported Resource Types

| Resource Type | Upload | Download | Notes |
|--------------|--------|----------|-------|
| Linux Servers | ✓ | ✓ | Fully supported |
| Windows Servers | ✓ | ✓ | Requires OpenSSH Server |
| Docker Containers | ✗ | ✗ | Not supported (use host) |
| Databases | ✗ | ✗ | Not supported |

### Performance Recommendations

- **Small files** (< 10MB): Direct transfer
- **Large files** (> 100MB): Compress before transfer
- **Many small files**: Package as tar.gz
- **Unstable network**: Use rsync instead of scp

---

## Best Practices

1. **Use SSH Config File** - Simplify command-line parameters
2. **Compress Transfers** - Reduce transfer time and bandwidth
3. **Automation Scripts** - Use scripts for batch operations
4. **Audit Logs** - Regularly check file transfer records
5. **Access Control** - Follow least privilege principle
6. **Backup Verification** - Verify file integrity after transfer

```bash
# Verify file integrity
# 1. Calculate checksum before transfer
md5sum local_file.txt

# 2. Transfer file
scp -P 2200 local_file.txt user@roma:user@web-01:/tmp/

# 3. Verify after transfer
ssh user@roma -p 2200 "ln -t linux web-01 -- 'md5sum /tmp/local_file.txt'"
```

---

## Related Documentation

- [Deployment Guide](DEPLOYMENT.md)
- [Security Guide](SECURITY.md)
- [API Documentation](API.md)

---

**ROMA SCP** - Secure and efficient file transfer solution
