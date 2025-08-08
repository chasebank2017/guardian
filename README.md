
# Guardian

Guardian 是一套面向企业级微信数据采集与安全管控的分布式系统，致力于安全、自动化地采集、同步和管理多终端微信数据。系统由高性能后端、跨平台 Agent 以及现代化前端组成。

---

## 架构一览

```
┌──────────┐   mTLS(gRPC)   ┌──────────┐   REST(HTTP)   ┌──────────┐
│  Agent   │ <------------> │ Backend  │ <------------> │ Frontend │
└──────────┘                 └──────────┘                 └──────────┘
             └─────────────── PostgreSQL ────────────────┘
```

---

## 技术栈
- **后端**：Go（chi、pgx、JWT）、gRPC、Docker
- **Agent**：Rust（tokio/tonic）
- **前端**：React、TypeScript、Vite、Zustand、MUI
- **数据库**：PostgreSQL

---

## 快速开始（推荐：Docker Compose）

1) 准备环境
- 需要 Docker Desktop（或兼容环境）

2) 首次启动（自动建表）
- 本仓库已内置初始化 SQL：`backend/db/init/001_schema.sql`
- `docker-compose.yml` 已挂载到 `postgres` 的 `/docker-entrypoint-initdb.d`
- 首次启动或清空数据卷后，Postgres 会自动执行建表

3) 启动服务
```bash
docker compose up -d
```

4) 访问与验证
- 后端 HTTP：`http://localhost:8080`
- 登录（JWT）：`POST /login`，请求体：`{"username":"admin","password":"password"}`
- 拿到 token 后，携带 `Authorization: Bearer <token>` 访问受保护接口

5) 清理（可选）
```bash
docker compose down
# 如需重新初始化数据库（触发自动建表），需同时清空数据卷：
docker volume rm guardian_postgres_data
```

---

## 本地开发

### 后端（不通过 Docker）
1) 准备 Postgres（可本机或容器），并执行初始化 SQL：
```bash
PGPASSWORD=password psql -h localhost -p 5432 -U user -d guardian -f backend/db/init/001_schema.sql
```
2) 配置与运行
- 配置文件：`backend/config.yaml`（已提供示例；也可通过环境变量覆盖）
- 重要环境变量：`DATABASE_URL`（将覆盖配置中的 DSN），示例：
  `postgres://user:password@localhost:5432/guardian?sslmode=disable`
- 启动：
```bash
cd backend
go build ./...
go run ./cmd/server/main.go
```

3) gRPC（可选）
- 代码当前默认启用 mTLS gRPC 服务，需在后端工作目录放置证书：`ca.crt`、`server.crt`、`server.key`
- 本地自签发示例（OpenSSL）：
```bash
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt -subj "/CN=Guardian-CA"
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=localhost"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365 -sha256
```
- 如仅需 HTTP，可后续引入开关以禁用 gRPC（当前代码未提供开关）。

### 前端
- 现有可运行前端位于：`agent/frontend`
```bash
cd agent/frontend
npm ci
npm run dev    # 开发
npm run build  # 生产构建
```
- 注意：开发环境下如需联通后端，需要在 Vite 配置中加代理或将 API Base 指向后端地址。

---

## API 速览（HTTP）

- 登录获取 JWT
  - `POST /login`  body: `{ "username":"admin", "password":"password" }`
- 受保护接口（需 `Authorization: Bearer <token>`）
  - `GET /v1/agents`：获取 Agent 列表（当前返回字段：`id`, `name`）
  - `GET /v1/agents/{agentID}/messages`：获取指定 Agent 的消息（`content`, `timestamp(ms)`）
  - `POST /v1/agents/{agentID}/tasks`：为 Agent 下发任务（当前示例任务类型：`DUMP_WECHAT_DATA`）

健康与指标：
- 健康检查：`GET /healthz`、就绪检查：`GET /readyz`
- 指标（Prometheus）：`GET /metrics`

gRPC 接口定义：`backend/api/proto/guardian.proto`

---

## 配置与环境变量
- 文件：`backend/config.yaml`
  - `server.port`：HTTP 端口（默认 `:8080`）
  - `server.grpc_port`：gRPC 端口（默认 `:50051`）
  - `server.request_timeout_seconds`：请求超时时间（秒）
  - `server.cors_origins`：CORS 允许来源，支持 `*` 或具体域名数组
  - `server.rate_limit.login_rps` / `login_burst`：登录接口限流
  - `server.rate_limit.protected_rps` / `protected_burst`：受保护接口限流
  - `database.dsn`：数据库连接串
  - `auth.jwt_secret`：JWT 密钥
  - `auth.admin_username` / `auth.admin_password`：登录凭据
- 环境变量覆盖：`DATABASE_URL` 会覆盖 `database.dsn`；`ADMIN_USERNAME`、`ADMIN_PASSWORD` 覆盖登录凭据

---

## 监控（Prometheus）
- 暴露端点：`/metrics`
- 指标示例：
  - `guardian_http_requests_total{method,route,status}`：HTTP 请求总数
  - `guardian_http_request_duration_seconds{method,route,status}`：HTTP 请求时延
- Prometheus 抓取配置示例：
```yaml
scrape_configs:
  - job_name: 'guardian'
    static_configs:
      - targets: ['localhost:8080']
```

---

## 测试与构建
```bash
cd backend
go build ./...
go test ./... -count=1

cd ../agent/frontend
npm run build
```

---

## 故障排查
- 后端启动报证书相关错误：未提供 `ca.crt/server.crt/server.key`，请参考上文生成自签名证书或在后续版本启用 gRPC 开关。
- 访问受保护接口返回 401：未携带 `Authorization: Bearer <token>` 或 token 已过期。
- 数据为空：确认已建表并导入数据，可用 `INSERT INTO agents(hostname) VALUES ('agent-1');` 进行冒烟。

---

## 贡献
欢迎提交 Issue / PR 参与维护与完善。

---

© 2025 Guardian Project
