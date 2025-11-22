# 构建阶段
FROM registry.cn-hongkong.aliyuncs.com/binrcbase/golang:1.23.1-alpine AS build
WORKDIR /app

# 设置 Go 代理和模块模式
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
ENV GO111MODULE=on

# 先复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o links ./cmd/links/main.go

# 运行阶段
FROM registry.cn-hongkong.aliyuncs.com/binrcbase/alpine:latest
WORKDIR /app
COPY --from=build /app/links .
COPY --from=build /app/cmd/links/configs/config.yaml ./cmd/links/configs/
COPY --from=build /app/cmd/links/configs/etcd-ssl ./cmd/links/configs/etcd-ssl
COPY --from=build /app/.env.dev ./.env.dev
ENTRYPOINT ["./links", "serve", "-c", "./cmd/links/configs/config.yaml"]