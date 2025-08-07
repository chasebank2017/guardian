package database

import "context"

type DBOperations interface {
	CreateTaskForAgent(ctx context.Context, agentID int, taskType string) error
	// 未来可以添加更多方法，如 GetAgentByID, SaveMessages 等
}
