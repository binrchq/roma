.PHONY: help run run-roma test build clean docker docker-backend docker-monolithic docker-up docker-down docker-logs

# 变量定义
APP_NAME := roma
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
IMAGE_REGISTRY := binrc
IMAGE_BACKEND := $(IMAGE_REGISTRY)/roma-backend
IMAGE_MONOLITHIC := $(IMAGE_REGISTRY)/roma
IMAGE_TAG := latest

# 目录定义
ROOT_DIR := $(shell pwd)
DEPLOYMENT_DIR := $(ROOT_DIR)/deployment
CMD_DIR := $(ROOT_DIR)/cmd/roma

# Docker 相关
DOCKER_COMPOSE := docker-compose
DOCKER_COMPOSE_FILE := $(DEPLOYMENT_DIR)/docker-compose.yml
DOCKER_COMPOSE_MONO := $(DEPLOYMENT_DIR)/docker-compose.monolithic.yml
DOCKER_COMPOSE_MYSQL := $(DEPLOYMENT_DIR)/docker-compose.mysql.yml
DOCKER_COMPOSE_PGSQL := $(DEPLOYMENT_DIR)/docker-compose.pgsql.yml

# 默认目标
.DEFAULT_GOAL := help

##@ 帮助信息

help: ## 显示此帮助信息
	@echo "ROMA 项目 Makefile"
	@echo ""
	@echo "用法: make [target]"
	@echo ""
	@echo "可用目标:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ 开发

run: ## 运行应用（开发模式）
	@echo "🚀 启动应用..."
	@go run $(CMD_DIR)/main.go

run-roma: ## 运行应用（带 -f 参数）
	@echo "🚀 启动应用..."
	@go run $(CMD_DIR)/main.go -f

test: ## 运行测试
	@echo "🧪 运行测试..."
	@go run cmd/test/main.go

build: ## 构建应用
	@echo "🔨 构建应用..."
	@go build -ldflags="-s -w" -o $(APP_NAME) $(CMD_DIR)/main.go
	@echo "✅ 构建完成: $(APP_NAME)"

##@ 清理

clean: ## 清理构建产物
	@echo "🧹 清理构建产物..."
	@rm -f $(APP_NAME)
	@echo "✅ 清理完成"

clean-ports: ## 清理端口进程（2222, 6999）
	@echo "🧹 清理端口进程..."
	@-lsof -ti:2222 | xargs -r kill -9 2>/dev/null || true
	@-lsof -ti:6999 | xargs -r kill -9 2>/dev/null || true
	@echo "✅ 端口清理完成"

clean-db: ## 备份数据库文件
	@echo "💾 备份数据库文件..."
	@if [ -f /usr/local/roma/c.db ]; then \
		mv /usr/local/roma/c.db /usr/local/roma/c.db.bk; \
		echo "✅ 数据库已备份"; \
	else \
		echo "⚠️  数据库文件不存在"; \
	fi

clean-all: clean clean-ports clean-db ## 清理所有（构建产物、端口、数据库）

##@ Docker 构建

docker: docker-backend ## 构建后端 Docker 镜像
	@echo "✅ 后端镜像构建完成"

docker-backend: ## 构建后端镜像
	@echo "🔨 构建后端镜像..."
	@docker build -f $(DEPLOYMENT_DIR)/Dockerfile.backend \
		-t $(IMAGE_BACKEND):$(IMAGE_TAG) \
		-t $(IMAGE_BACKEND):$(VERSION) \
		.
	@echo "✅ 后端镜像构建完成: $(IMAGE_BACKEND):$(IMAGE_TAG)"

docker-monolithic: ## 构建单体镜像
	@echo "🔨 构建单体镜像..."
	@docker build -f $(DEPLOYMENT_DIR)/Dockerfile.monolithic \
		-t $(IMAGE_MONOLITHIC):$(IMAGE_TAG) \
		-t $(IMAGE_MONOLITHIC):$(VERSION) \
		.
	@echo "✅ 单体镜像构建完成: $(IMAGE_MONOLITHIC):$(IMAGE_TAG)"

##@ Docker Compose

docker-up: ## 启动 Docker Compose（分离模式）
	@echo "🚀 启动 Docker Compose 服务..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) up -d
	@echo "✅ 服务已启动"

docker-up-mono: ## 启动 Docker Compose（单体模式）
	@echo "🚀 启动 Docker Compose 服务（单体模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MONO) up -d
	@echo "✅ 服务已启动"

docker-down: ## 停止 Docker Compose（分离模式）
	@echo "🛑 停止 Docker Compose 服务..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) down
	@echo "✅ 服务已停止"

docker-down-mono: ## 停止 Docker Compose（单体模式）
	@echo "🛑 停止 Docker Compose 服务（单体模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MONO) down
	@echo "✅ 服务已停止"

docker-logs: ## 查看 Docker Compose 日志（分离模式）
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) logs -f

docker-logs-mono: ## 查看 Docker Compose 日志（单体模式）
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MONO) logs -f

docker-ps: ## 查看 Docker Compose 服务状态
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) ps

docker-rebuild: docker-down docker-backend docker-up ## 重新构建并启动（分离模式）

docker-rebuild-mono: docker-down-mono docker-monolithic docker-up-mono ## 重新构建并启动（单体模式）

docker-up-mysql: ## 启动 Docker Compose（MySQL 模式）
	@echo "🚀 启动 Docker Compose 服务（MySQL 模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MYSQL) up -d
	@echo "✅ 服务已启动（MySQL）"

docker-down-mysql: ## 停止 Docker Compose（MySQL 模式）
	@echo "🛑 停止 Docker Compose 服务（MySQL 模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MYSQL) down
	@echo "✅ 服务已停止（MySQL）"

docker-logs-mysql: ## 查看 Docker Compose 日志（MySQL 模式）
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_MYSQL) logs -f

docker-up-pgsql: ## 启动 Docker Compose（PostgreSQL 模式）
	@echo "🚀 启动 Docker Compose 服务（PostgreSQL 模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_PGSQL) up -d
	@echo "✅ 服务已启动（PostgreSQL）"

docker-down-pgsql: ## 停止 Docker Compose（PostgreSQL 模式）
	@echo "🛑 停止 Docker Compose 服务（PostgreSQL 模式）..."
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_PGSQL) down
	@echo "✅ 服务已停止（PostgreSQL）"

docker-logs-pgsql: ## 查看 Docker Compose 日志（PostgreSQL 模式）
	@cd $(DEPLOYMENT_DIR) && $(DOCKER_COMPOSE) -f $(DOCKER_COMPOSE_PGSQL) logs -f

docker-rebuild-mysql: docker-down-mysql docker-backend docker-up-mysql ## 重新构建并启动（MySQL 模式）

docker-rebuild-pgsql: docker-down-pgsql docker-backend docker-up-pgsql ## 重新构建并启动（PostgreSQL 模式）

##@ 工具

fmt: ## 格式化代码
	@echo "📝 格式化代码..."
	@go fmt ./...
	@echo "✅ 格式化完成"

lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@go vet ./...
	@echo "✅ 代码检查完成"

mod: ## 整理 Go 模块依赖
	@echo "📦 整理 Go 模块依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖整理完成"
