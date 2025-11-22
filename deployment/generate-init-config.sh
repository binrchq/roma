#!/bin/bash

# ROMA 初始化配置生成工具
# 生成必需的初始化参数，供 docker-compose 直接使用
# 支持使用现有私钥或自动生成

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
CREDENTIALS_FILE="${SCRIPT_DIR}/credentials.txt"

# 生成随机字符串
generate_random_string() {
    local length=${1:-32}
    openssl rand -hex $((length / 2)) | tr -d '\n'
}

# 生成随机密码
generate_password() {
    openssl rand -base64 24 | tr -d "=+/" | cut -c1-16
}

# 生成 SSH 密钥对
generate_ssh_key() {
    local key_type=${1:-rsa}
    local key_size=${2:-2048}
    local temp_dir=$(mktemp -d)
    local key_file="${temp_dir}/roma_key"
    
    ssh-keygen -t "${key_type}" -b "${key_size}" -f "${key_file}" -N "" -q
    
    echo "${temp_dir}"
}

# 生成 API Key
generate_api_key() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-48
}

# 读取私钥文件
read_private_key() {
    local key_file="$1"
    if [ ! -f "$key_file" ]; then
        echo "错误: 私钥文件不存在: $key_file" >&2
        return 1
    fi
    
    # 检查是否为有效的私钥文件
    if ! ssh-keygen -y -f "$key_file" > /dev/null 2>&1; then
        echo "错误: 无效的私钥文件: $key_file" >&2
        return 1
    fi
    
    cat "$key_file"
}

# 从私钥文件提取公钥
extract_public_key() {
    local key_file="$1"
    ssh-keygen -y -f "$key_file"
}

echo "🔧 生成 ROMA 初始化配置..."
echo ""

# 询问是否使用现有私钥
read -p "是否使用现有 SSH 私钥？(y/n，默认 n): " USE_EXISTING_KEY
USE_EXISTING_KEY=${USE_EXISTING_KEY:-n}

if [ "$USE_EXISTING_KEY" = "y" ] || [ "$USE_EXISTING_KEY" = "Y" ]; then
    read -p "请输入私钥文件路径: " PRIVATE_KEY_PATH
    PRIVATE_KEY_PATH=$(eval echo "$PRIVATE_KEY_PATH")  # 展开 ~ 等路径
    
    if [ ! -f "$PRIVATE_KEY_PATH" ]; then
        echo "❌ 错误: 私钥文件不存在: $PRIVATE_KEY_PATH"
        exit 1
    fi
    
    echo "📖 读取私钥文件: $PRIVATE_KEY_PATH"
    CONTROL_PRIVATE_KEY=$(read_private_key "$PRIVATE_KEY_PATH")
    CONTROL_PUBLIC_KEY=$(extract_public_key "$PRIVATE_KEY_PATH")
    echo "✅ 私钥读取成功"
else
    echo "🔑 生成新的 SSH 密钥对..."
    SSH_KEY_DIR=$(generate_ssh_key rsa 2048)
    CONTROL_PUBLIC_KEY=$(cat "${SSH_KEY_DIR}/roma_key.pub")
    CONTROL_PRIVATE_KEY=$(cat "${SSH_KEY_DIR}/roma_key")
    # 清理临时目录
    trap "rm -rf ${SSH_KEY_DIR}" EXIT
    echo "✅ SSH 密钥对生成成功"
fi

# 将私钥转换为单行格式（用 \n 替换换行符）
CONTROL_PRIVATE_KEY_ESCAPED=$(echo "$CONTROL_PRIVATE_KEY" | awk '{printf "%s\\n", $0}' | sed 's/\\n$//')

# 生成数据库密码
MYSQL_ROOT_PASSWORD=$(generate_password)
MYSQL_PASSWORD=$(generate_password)
POSTGRES_PASSWORD=$(generate_password)

# 生成应用配置
read -p "管理员用户名 (默认: admin_随机): " ADMIN_USERNAME_INPUT
if [ -z "$ADMIN_USERNAME_INPUT" ]; then
    ADMIN_USERNAME="admin_$(generate_random_string 8 | cut -c1-8)"
else
    ADMIN_USERNAME="$ADMIN_USERNAME_INPUT"
fi

read -p "管理员邮箱 (默认: ${ADMIN_USERNAME}@roma.local): " ADMIN_EMAIL_INPUT
if [ -z "$ADMIN_EMAIL_INPUT" ]; then
    ADMIN_EMAIL="${ADMIN_USERNAME}@roma.local"
else
    ADMIN_EMAIL="$ADMIN_EMAIL_INPUT"
fi

ADMIN_PASSWORD=$(generate_password)
ADMIN_NAME="系统管理员"
ADMIN_NICKNAME="Admin"
ADMIN_ROLES="super,system,ops,ordinary,trial"

API_KEY_PREFIX="apikey."
API_KEY=$(generate_api_key)

CONTROL_SERVICE_USER="root"
CONTROL_PASSWORD=$(generate_password)
CONTROL_RESOURCE_TYPE="linux"
CONTROL_DESCRIPTION="Default control passport for ops use"

# 生成 .env 文件（供 docker-compose 使用）
cat > "${ENV_FILE}" <<EOF
# ROMA 初始化配置
# 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
# 警告: 请妥善保管此文件！

# MySQL 配置
MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
MYSQL_DATABASE=roma
MYSQL_USER=roma
MYSQL_PASSWORD=${MYSQL_PASSWORD}

# PostgreSQL 配置
POSTGRES_DB=roma
POSTGRES_USER=roma
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

# ROMA 数据库连接 URL
ROMA_DATABASE_CDB_URL_MYSQL=roma:${MYSQL_PASSWORD}@tcp(mysql:3306)/roma?charset=utf8mb4&parseTime=True&loc=Local
ROMA_DATABASE_CDB_URL_PGSQL=postgres://roma:${POSTGRES_PASSWORD}@postgres:5432/roma?sslmode=disable

# ROMA API Key
ROMA_APIKEY_PREFIX=${API_KEY_PREFIX}
ROMA_APIKEY_KEY=${API_KEY}

# ROMA 第一个用户
ROMA_USER_1ST_EMAIL=${ADMIN_EMAIL}
ROMA_USER_1ST_NAME=${ADMIN_NAME}
ROMA_USER_1ST_NICKNAME=${ADMIN_NICKNAME}
ROMA_USER_1ST_PASSWORD=${ADMIN_PASSWORD}
ROMA_USER_1ST_USERNAME=${ADMIN_USERNAME}
ROMA_USER_1ST_ROLES=${ADMIN_ROLES}
ROMA_USER_1ST_PUBLIC_KEY=${CONTROL_PUBLIC_KEY}

# ROMA 控制通行证
ROMA_CONTROL_PASSPORT_SERVICE_USER=${CONTROL_SERVICE_USER}
ROMA_CONTROL_PASSPORT_PASSWORD=${CONTROL_PASSWORD}
ROMA_CONTROL_PASSPORT_RESOURCE_TYPE=${CONTROL_RESOURCE_TYPE}
ROMA_CONTROL_PASSPORT_PASSPORT_PUB=${CONTROL_PUBLIC_KEY}
ROMA_CONTROL_PASSPORT_DESCRIPTION=${CONTROL_DESCRIPTION}
ROMA_CONTROL_PASSPORT_PASSPORT=${CONTROL_PRIVATE_KEY_ESCAPED}
EOF

# 生成凭据摘要文件
cat > "${CREDENTIALS_FILE}" <<EOF
========================================
ROMA 初始化凭据
========================================
生成时间: $(date '+%Y-%m-%d %H:%M:%S')

⚠️  警告: 请妥善保管这些凭据！

----------------------------------------
MySQL 数据库
----------------------------------------
Root 密码: ${MYSQL_ROOT_PASSWORD}
数据库名: roma
用户名: roma
密码: ${MYSQL_PASSWORD}

----------------------------------------
PostgreSQL 数据库
----------------------------------------
数据库名: roma
用户名: roma
密码: ${POSTGRES_PASSWORD}

----------------------------------------
管理员账户
----------------------------------------
用户名: ${ADMIN_USERNAME}
邮箱: ${ADMIN_EMAIL}
密码: ${ADMIN_PASSWORD}
角色: ${ADMIN_ROLES}

----------------------------------------
API Key
----------------------------------------
前缀: ${API_KEY_PREFIX}
密钥: ${API_KEY}

----------------------------------------
控制通行证
----------------------------------------
服务用户: ${CONTROL_SERVICE_USER}
密码: ${CONTROL_PASSWORD}
资源类型: ${CONTROL_RESOURCE_TYPE}

----------------------------------------
SSH 公钥
----------------------------------------
${CONTROL_PUBLIC_KEY}

----------------------------------------
配置文件
----------------------------------------
环境变量文件: ${ENV_FILE}
凭据文件: ${CREDENTIALS_FILE}

========================================
EOF

echo ""
echo "✅ 初始化配置生成完成！"
echo ""
echo "📁 环境变量文件: ${ENV_FILE}"
echo "🔐 凭据文件: ${CREDENTIALS_FILE}"
echo ""
echo "⚠️  请查看 ${CREDENTIALS_FILE} 获取生成的凭据"
echo ""
echo "💡 使用方法:"
echo "   # 方式1: 使用 --env-file（推荐）"
echo "   docker-compose -f docker-compose.mysql.yml --env-file .env up -d"
echo "   docker-compose -f docker-compose.pgsql.yml --env-file .env up -d"
echo ""
echo "   # 方式2: docker-compose 会自动读取 .env 文件"
echo "   docker-compose -f docker-compose.mysql.yml up -d"
echo "   docker-compose -f docker-compose.pgsql.yml up -d"
echo ""
