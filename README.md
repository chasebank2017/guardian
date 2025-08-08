
# Guardian

Guardian 是一套面向企业级微信数据采集与安全管控的分布式系统，致力于安全、自动化地采集、同步和管理多终端微信数据。系统由高性能后端、可热更新的跨平台 Agent 以及现代化前端组成，支持多版本微信协议适配和企业级安全合规。

---

## 架构图

```
┌──────────┐   mTLS   ┌──────────┐   REST/gRPC   ┌──────────┐
│  Agent   │<------->│ Backend  │<------------->│ Frontend │
└──────────┘          └──────────┘               └──────────┘
      │                    │                          │
      │<---------------- PostgreSQL ---------------->│
```

（如需图片版可放置于 docs/architecture.png）

---

## 技术栈
- **后端**：Go 1.20+、gRPC、chi、pgx、Docker、mTLS
- **Agent**：Rust 1.70+、tonic、tokio、sysinfo、shred、Windows Service
- **前端**：React 18+、TypeScript、Zustand、MUI、react-hot-toast、react-window、Vite
- **数据库**：PostgreSQL
- **协议/接口**：gRPC（.proto）、RESTful API

---

## 本地开发环境搭建指南

### 后端启动
1. 复制 `backend/config.yaml.example` 为 `backend/config.yaml`，根据实际环境填写数据库、证书等配置。
2. 启动后端服务：
   ```bash
   cd backend
   go run ./cmd/server/main.go
   ```

### Agent 编译
1. 安装 Rust 环境（推荐使用 rustup）。
2. 编译 Agent：
   ```bash
   cd agent
   cargo build --release
   ```
   生成的可执行文件位于 `agent/target/release/`。

### 前端启动
1. 安装 Node.js 16+。
2. 安装依赖并启动开发服务器：
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   默认端口为 5173。

---

## 部署指南

### 后端 Docker 部署
1. 构建镜像：
   ```bash
   cd backend
   docker build -t guardian-backend .
   ```
2. 运行容器（示例）：
   ```bash
   docker run -d -p 50051:50051 -v $(pwd)/config.yaml:/app/config.yaml guardian-backend
   ```

### Agent 安装为 Windows 服务
1. 以管理员身份运行命令行。
2. 安装服务：
   ```powershell
   guardian_agent.exe install
   guardian_agent.exe start
   ```
3. 卸载服务：
   ```powershell
   guardian_agent.exe stop
   guardian_agent.exe uninstall
   ```

---

## API 文档

- gRPC 协议定义见： [`backend/api/proto/guardian.proto`](backend/api/proto/guardian.proto)
- 主要 RESTful API 示例：
  - `GET /api/v1/agents`：获取 Agent 列表
  - `GET /api/v1/messages/{agentId}`：获取指定 Agent 的消息
  - `POST /api/v1/agents/{agentId}/tasks`：为 Agent 下发任务

详细字段和消息体请参考 proto 文件和前端接口实现。

---

## 贡献与支持
如需贡献代码、反馈问题或获取支持，请提交 Issue 或 PR。

---

© 2025 Guardian Project
