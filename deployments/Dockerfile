# 使用一个更小的基础镜像运行项目
FROM golang:1.21.4-alpine

# 安装 SQLite3 的依赖
RUN apk add --no-cache sqlite-dev

WORKDIR /roma

# 复制go.mod和go.sum并下载依赖
COPY ../go.mod .
COPY ../go.sum .
RUN go mod download

# 复制整个项目到容器中
COPY .. .

# 构建应用程序
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o roma .

# 设置环境变量
ENV APP_ENV=production
ENV ROMA_BASE_ROOT=/roma

# 暴露应用程序端口
EXPOSE 2200
EXPOSE 7999

# 运行应用程序
ENTRYPOINT ["./roma"]

# 构建完成后清理复制的文件
RUN rm -rf ./*
