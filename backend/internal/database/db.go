package database

import (
    "context"
    "errors"
    "fmt"
    "os"
    "time"

    api "guardian-backend/pkg/grpc/api/guardian/pkg/grpc/api"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

// DB 封装 pgxpool.Pool，便于在本地定义方法
type DB struct {
    Pool *pgxpool.Pool
}

// TimeoutStaleTasks 将状态为 'sent' 且在指定时间内未更新的任务标记为 'timeout'
func (p *DB) TimeoutStaleTasks(ctx context.Context, timeoutDuration time.Duration) (int64, error) {
	// SQL 语句的逻辑是：
	// UPDATE tasks
	// SET status = 'timeout', updated_at = NOW()
	// WHERE status = 'sent' AND updated_at < NOW() - $1::interval;
	// $1 的值应该是类似 "1 hour" 这样的字符串
	timeoutStr := fmt.Sprintf("%f seconds", timeoutDuration.Seconds())
    tag, err := p.Pool.Exec(ctx, "UPDATE tasks SET status = 'timeout', updated_at = NOW() WHERE status = 'sent' AND updated_at < NOW() - $1::interval", timeoutStr)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
// ...existing code...

// ErrAgentNotFound 用于 agent_id 不存在时返回
var ErrAgentNotFound = errors.New("agent not found")

// GetAndDispatchPendingTaskForAgent 查询指定 agent 是否有待执行任务，有则返回类型并将其状态置为 sent
func (p *DB) GetAndDispatchPendingTaskForAgent(ctx context.Context, agentID int) (string, error) {
	var taskType string
    tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT id, task_type FROM tasks WHERE agent_id=$1 AND status='pending' ORDER BY created_at LIMIT 1`, agentID).Scan(new(int), &taskType)
	if err != nil {
		return "", nil // 没有待执行任务
	}
	_, err = tx.Exec(ctx, `UPDATE tasks SET status='sent', updated_at=NOW() WHERE agent_id=$1 AND status='pending' AND task_type=$2`, agentID, taskType)
	if err != nil {
		return "", err
	}
	tx.Commit(ctx)
	return taskType, nil
}
// CreateTaskForAgent 在数据库中为指定的 agent 创建一个新任务
func (p *DB) CreateTaskForAgent(ctx context.Context, agentID int, taskType string) error {
	// 先检查 agent 是否存在
	var count int
    err := p.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM agents WHERE id=$1`, agentID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrAgentNotFound
	}
    _, err = p.Pool.Exec(ctx, `
		INSERT INTO tasks (agent_id, task_type, status, created_at, updated_at)
		VALUES ($1, $2, 'pending', NOW(), NOW())
	`, agentID, taskType)
	return err
}
// SaveMessages 批量写入 wechat_messages 表
func (p *DB) SaveMessages(ctx context.Context, agentID int, messages []*api.ChatMessage) error {
	rows := make([][]interface{}, 0, len(messages))
	for _, m := range messages {
		rows = append(rows, []interface{}{
			agentID,
			m.Content,
			m.Timestamp.AsTime(),
		})
	}
    _, err := p.Pool.CopyFrom(ctx,
		pgx.Identifier{"wechat_messages"},
		[]string{"agent_id", "content", "timestamp"},
		pgx.CopyFromRows(rows),
	)
	return err
}
// ...existing code...

// NewConnection 读取环境变量 DATABASE_DSN 并返回 pgxpool.Pool
func NewConnection(ctx context.Context) (*DB, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN env not set")
	}
    return NewConnectionWithDSN(ctx, dsn)
}

// NewConnectionWithDSN 通过参数传递 DSN，便于 config 驱动
func NewConnectionWithDSN(ctx context.Context, dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN is empty")
	}
    pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
    return &DB{Pool: pool}, nil
}

// AgentInfo 是前端所需的基础 Agent 视图
type AgentInfo struct {
    ID       int
    Hostname string
}

// ListAgents 查询所有 agents 的基础信息
func (p *DB) ListAgents(ctx context.Context) ([]AgentInfo, error) {
    rows, err := p.Pool.Query(ctx, `SELECT id, hostname FROM agents ORDER BY id ASC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []AgentInfo
    for rows.Next() {
        var it AgentInfo
        if err := rows.Scan(&it.ID, &it.Hostname); err != nil {
            return nil, err
        }
        result = append(result, it)
    }
    return result, rows.Err()
}

// WechatMessageRecord 用于查询出的消息记录
type WechatMessageRecord struct {
    Content   string
    Timestamp time.Time
}

// ListMessagesByAgent 查询指定 agent 的消息
func (p *DB) ListMessagesByAgent(ctx context.Context, agentID int) ([]WechatMessageRecord, error) {
    rows, err := p.Pool.Query(ctx, `SELECT content, timestamp FROM wechat_messages WHERE agent_id=$1 ORDER BY timestamp DESC LIMIT 500`, agentID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var result []WechatMessageRecord
    for rows.Next() {
        var it WechatMessageRecord
        if err := rows.Scan(&it.Content, &it.Timestamp); err != nil {
            return nil, err
        }
        result = append(result, it)
    }
    return result, rows.Err()
}
