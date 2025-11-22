#!/bin/bash
# 配置中国镜像源脚本

echo "🚀 开始配置中国镜像源..."

# ==================== NPM 镜像 ====================
echo ""
echo "📦 配置 NPM 镜像（淘宝）..."
npm config set registry https://registry.npmmirror.com
echo "✅ NPM 镜像已设置为: $(npm config get registry)"

# ==================== Go 镜像 ====================
echo ""
echo "🐹 配置 Go 镜像（阿里云）..."
go env -w GO111MODULE=on
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
echo "✅ Go 代理已设置为: $(go env GOPROXY)"

# ==================== Docker 镜像 ====================
echo ""
echo "🐳 配置 Docker 镜像加速器..."
if [ -f /etc/docker/daemon.json ]; then
    echo "⚠️  /etc/docker/daemon.json 已存在，请手动添加以下内容："
    echo '{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://registry.docker-cn.com"
  ]
}'
else
    echo "ℹ️  Docker 配置文件不存在，跳过（如需配置请手动创建 /etc/docker/daemon.json）"
fi

# ==================== Python pip 镜像 ====================
echo ""
echo "🐍 配置 pip 镜像（阿里云）..."
if command -v pip &> /dev/null; then
    pip config set global.index-url https://mirrors.aliyun.com/pypi/simple/
    echo "✅ pip 镜像已设置"
else
    echo "ℹ️  未检测到 pip，跳过"
fi

# ==================== APT 镜像 (Ubuntu/Debian) ====================
echo ""
echo "📦 APT 镜像配置..."
if command -v apt &> /dev/null; then
    echo "ℹ️  检测到 APT 包管理器"
    echo "⚠️  APT 源配置需要 root 权限，请手动修改 /etc/apt/sources.list"
    echo "推荐使用阿里云镜像: https://developer.aliyun.com/mirror/ubuntu"
else
    echo "ℹ️  未检测到 APT，跳过"
fi

# ==================== 显示配置摘要 ====================
echo ""
echo "================================"
echo "📋 配置摘要"
echo "================================"
echo "NPM:    $(npm config get registry)"
echo "Go:     $(go env GOPROXY)"
echo ""
echo "✅ 所有镜像源配置完成！"
echo ""
echo "💡 提示："
echo "  - 恢复 NPM 官方源: npm config set registry https://registry.npmjs.org/"
echo "  - 恢复 Go 官方源:  go env -w GOPROXY=https://proxy.golang.org,direct"
echo ""


