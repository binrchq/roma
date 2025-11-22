#!/bin/bash

# 快速启动 PostgreSQL 模式

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 检查 .env 文件是否存在
if [ ! -f ".env" ]; then
    echo "❌ 错误: .env 文件不存在"
    echo ""
    echo "请先运行生成脚本:"
    echo "  ./generate-init-config.sh"
    exit 1
fi

echo "🚀 启动 ROMA (PostgreSQL 模式)..."
docker-compose -f docker-compose.pgsql.yml up -d

echo ""
echo "✅ 服务已启动"
echo ""
echo "📊 查看服务状态:"
echo "  docker-compose -f docker-compose.pgsql.yml ps"
echo ""
echo "📋 查看日志:"
echo "  docker-compose -f docker-compose.pgsql.yml logs -f"
echo ""
echo "🛑 停止服务:"
echo "  docker-compose -f docker-compose.pgsql.yml down"
echo ""

