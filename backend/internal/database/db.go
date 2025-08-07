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
	_, err := p.Exec(ctx, `
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
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
