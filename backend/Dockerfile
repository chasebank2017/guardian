
# --- 构建阶段 (Build Stage) ---
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件，并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

# 执行编译，生成一个静态链接的、在 Linux 环境下运行的二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -o ./out/guardian ./cmd/server

# --- 最终阶段 (Final Stage) ---
FROM alpine:3.20

WORKDIR /app

# 安装依赖（包含下载 migrate 所需工具）
RUN apk add --no-cache ca-certificates curl bash netcat-openbsd

# 拷贝二进制与配置
COPY --from=builder /app/out/guardian /app/guardian
COPY --from=builder /app/config.yaml /app/config.yaml

# 复制迁移脚本和入口脚本（相对于构建上下文 ./backend）
COPY ./db/migrations /app/migrations
COPY ./entrypoint.sh /app/entrypoint.sh

# 下载 migrate CLI（静态二进制）
RUN curl -sSL -o /app/migrate.tar.gz https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz \
    && tar -xzf /app/migrate.tar.gz -C /app \
    && rm -f /app/migrate.tar.gz \
    && chmod +x /app/migrate /app/guardian /app/entrypoint.sh

EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["/app/entrypoint.sh"]
