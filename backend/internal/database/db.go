import (
	"context"
	"fmt"
	"time"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TimeoutStaleTasks 将状态为 'sent' 且在指定时间内未更新的任务标记为 'timeout'
func (p *pgxpool.Pool) TimeoutStaleTasks(ctx context.Context, timeoutDuration time.Duration) (int64, error) {
	// SQL 语句的逻辑是：
	// UPDATE tasks
	// SET status = 'timeout', updated_at = NOW()
	// WHERE status = 'sent' AND updated_at < NOW() - $1::interval;
	// $1 的值应该是类似 "1 hour" 这样的字符串
	timeoutStr := fmt.Sprintf("%f seconds", timeoutDuration.Seconds())
	tag, err := p.Exec(ctx, "UPDATE tasks SET status = 'timeout', updated_at = NOW() WHERE status = 'sent' AND updated_at < NOW() - $1::interval", timeoutStr)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
	"guardian/pkg/grpc/api"
	"time"
)

// ErrAgentNotFound 用于 agent_id 不存在时返回
var ErrAgentNotFound = errors.New("agent not found")

// GetAndDispatchPendingTaskForAgent 查询指定 agent 是否有待执行任务，有则返回类型并将其状态置为 sent
func (p *pgxpool.Pool) GetAndDispatchPendingTaskForAgent(ctx context.Context, agentID int) (string, error) {
	var taskType string
	tx, err := p.Begin(ctx)
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
func (p *pgxpool.Pool) CreateTaskForAgent(ctx context.Context, agentID int, taskType string) error {
	// 先检查 agent 是否存在
	var count int
	err := p.QueryRow(ctx, `SELECT COUNT(*) FROM agents WHERE id=$1`, agentID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrAgentNotFound
	}
	_, err = p.Exec(ctx, `
		INSERT INTO tasks (agent_id, task_type, status, created_at, updated_at)
		VALUES ($1, $2, 'pending', NOW(), NOW())
	`, agentID, taskType)
	return err
}
import (
	"time"
	"github.com/jackc/pgx/v5"
	"guardian/pkg/grpc/api"
)
// SaveMessages 批量写入 wechat_messages 表
func (p *pgxpool.Pool) SaveMessages(ctx context.Context, agentID int, messages []*api.ChatMessage) error {
	rows := make([][]interface{}, 0, len(messages))
	for _, m := range messages {
		rows = append(rows, []interface{}{
			agentID,
			m.Content,
			m.Timestamp.AsTime(),
		})
	}
	_, err := p.CopyFrom(ctx,
		pgx.Identifier{"wechat_messages"},
		[]string{"agent_id", "content", "timestamp"},
		pgx.CopyFromRows(rows),
	)
	return err
}
package database

import (
	"context"
	"os"
	"github.com/jackc/pgx/v5/pgxpool"
	"fmt"
)

// NewConnection 读取环境变量 DATABASE_DSN 并返回 pgxpool.Pool
func NewConnection(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_DSN env not set")
	}
	return NewConnectionWithDSN(ctx, dsn)
}

// NewConnectionWithDSN 通过参数传递 DSN，便于 config 驱动
func NewConnectionWithDSN(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN is empty")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
