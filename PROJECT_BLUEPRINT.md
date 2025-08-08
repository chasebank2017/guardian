# PROJECT_BLUEPRINT.md

## 1. 项目概述 (Overview)

**项目名称:** Guardian - 微信远程取证与审计平台

**目标:** 构建一个C/S架构的工具，用于审计监察部门。Agent端能无感推送到员工Windows PC，解密微信数据并安全上传。Console端供审计人员管理、查看和分析数据。

---

## 2. 技术选型 (Tech Stack)

- **终端 Agent:** Rust
- **后端服务:** Go (Golang)
- **前端控制台:** TypeScript + React
- **数据库:** PostgreSQL
- **通信协议:** gRPC (基于 Protocol Buffers)

---

## 3. 核心架构 (Architecture)

- **Agent:** 运行在Windows上的后台服务，负责执行任务和通信。
- **Backend:** Go语言实现的gRPC服务，包含Agent管理、数据接入和查询API。
- **Database:** PostgreSQL，存储所有数据。

---

## 4. 数据库 Schema (PostgreSQL)

```sql
-- 审计员账户
CREATE TABLE audit_users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL, -- 'admin', 'auditor'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 终端Agent信息
CREATE TABLE agents (
    id SERIAL PRIMARY KEY,
    hostname VARCHAR(255) NOT NULL,
    os_version VARCHAR(100),
    status VARCHAR(50) NOT NULL, -- 'online', 'offline'
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 聊天记录 (示例，后续可扩展)
CREATE TABLE wechat_messages (
    id BIGSERIAL PRIMARY KEY,
    agent_id INTEGER REFERENCES agents(id),
    conversation_id VARCHAR(255) NOT NULL, -- 联系人或群组的wxid
    sender_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(50) NOT NULL, -- 'text', 'image', etc.
    content TEXT,
    timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 审计日志
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES audit_users(id),
    action TEXT NOT NULL, -- e.g., 'dispatched_task_to_agent_5'
    ip_address VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

 -- 案件表
 CREATE TABLE cases (
     id SERIAL PRIMARY KEY,
     name VARCHAR(255) NOT NULL,
     description TEXT,
     status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'closed', 'archived'
     created_by INTEGER REFERENCES audit_users(id),
     created_at TIMESTAMPTZ DEFAULT NOW(),
     updated_at TIMESTAMPTZ DEFAULT NOW()
 );
 -- 案件与Agent的关联表 (多对多关系)
 CREATE TABLE case_agents (
     case_id INTEGER REFERENCES cases(id) ON DELETE CASCADE,
     agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
     PRIMARY KEY (case_id, agent_id)
 );
```

---

## 5. API 契约 (gRPC Proto)

请参考 `guardian.proto` 文件：

```proto
syntax = "proto3";

package guardian;

service AgentService {
    // Agent注册
    rpc RegisterAgent(RegisterAgentRequest) returns (RegisterAgentResponse);
    // 心跳
    rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
    // 下发任务
    rpc DispatchTask(DispatchTaskRequest) returns (DispatchTaskResponse);
    // 上传数据
    rpc UploadData(UploadDataRequest) returns (UploadDataResponse);
}

service DataService {
    // 查询消息
    rpc QueryMessages(QueryMessagesRequest) returns (QueryMessagesResponse);
    // 查询Agent状态
    rpc ListAgents(ListAgentsRequest) returns (ListAgentsResponse);
}

// AgentService Messages
message RegisterAgentRequest {
    string hostname = 1;
    string os_version = 2;
}
message RegisterAgentResponse {
    int32 agent_id = 1;
    string status = 2;
}

message HeartbeatRequest {
    int32 agent_id = 1;
}
message HeartbeatResponse {
    string status = 1;
}

message DispatchTaskRequest {
    int32 agent_id = 1;
    string task_type = 2; // e.g., 'fetch_wechat_data'
    string params = 3;
}
message DispatchTaskResponse {
    string status = 1;
    string message = 2;
}

message UploadDataRequest {
    int32 agent_id = 1;
    bytes data = 2;
    string data_type = 3; // e.g., 'wechat_message'
}
message UploadDataResponse {
    string status = 1;
}

// DataService Messages
message QueryMessagesRequest {
    int32 agent_id = 1;
    string conversation_id = 2;
    int32 limit = 3;
    int32 offset = 4;
}
message QueryMessagesResponse {
    repeated WechatMessage messages = 1;
}

message ListAgentsRequest {}
message ListAgentsResponse {
    repeated AgentInfo agents = 1;
}

message WechatMessage {
    int64 id = 1;
    string conversation_id = 2;
    string sender_id = 3;
    string message_type = 4;
    string content = 5;
    string timestamp = 6;
}

message AgentInfo {
    int32 id = 1;
    string hostname = 2;
    string os_version = 3;
    string status = 4;
    string last_seen_at = 5;
}
```
