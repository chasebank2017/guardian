-- Guardian minimal schema

-- agents
CREATE TABLE IF NOT EXISTS agents (
  id SERIAL PRIMARY KEY,
  hostname VARCHAR(255) NOT NULL,
  os_version VARCHAR(255),
  status VARCHAR(32) NOT NULL DEFAULT 'online',
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- wechat_messages
CREATE TABLE IF NOT EXISTS wechat_messages (
  id BIGSERIAL PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wechat_messages_agent_time
  ON wechat_messages(agent_id, timestamp DESC);

-- tasks (used by POST /v1/agents/{agentID}/tasks)
CREATE TABLE IF NOT EXISTS tasks (
  id BIGSERIAL PRIMARY KEY,
  agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
  task_type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_agent_status
  ON tasks(agent_id, status);


