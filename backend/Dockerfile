
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
FROM gcr.io/distroless/static-debian11

WORKDIR /app

COPY --from=builder /app/out/guardian .
COPY --from=builder /app/config.yaml .

EXPOSE 8080
EXPOSE 50051

ENTRYPOINT ["/app/guardian"]
