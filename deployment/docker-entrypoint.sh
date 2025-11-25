#!/bin/sh

set -e

CONFIG_DIR="/app/configs"
DEFAULT_CONFIG="$CONFIG_DIR/config.ex.toml"
TARGET_CONFIG="$CONFIG_DIR/config.toml"

# 如果目标配置文件不存在，从默认配置复制
if [ ! -f "$TARGET_CONFIG" ]; then
    echo "配置文件 $TARGET_CONFIG 不存在，从默认配置 $DEFAULT_CONFIG 生成..."
    if [ -f "$DEFAULT_CONFIG" ]; then
        cp "$DEFAULT_CONFIG" "$TARGET_CONFIG"
        echo "已生成默认配置文件: $TARGET_CONFIG"
    else
        echo "错误: 默认配置文件 $DEFAULT_CONFIG 不存在"
        exit 1
    fi
fi

# 设置 TTY 相关环境变量，确保符合 TTY 使用习惯
# 如果 TERM 未设置，使用默认值
export TERM=${TERM:-xterm-256color}

# 强制启用颜色输出（在 Docker 容器中也需要颜色）
export FORCE_COLOR=1
export NO_COLOR=

# 确保输出流是交互式的（对于 SSH 会话很重要）
# 这些环境变量有助于程序正确检测 TTY 特性
export COLORTERM=${COLORTERM:-truecolor}

# 执行原始命令
exec "$@"

