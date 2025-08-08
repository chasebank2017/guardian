package database

import (
    "context"
    api "guardian-backend/pkg/grpc/api/guardian/pkg/grpc/api"
)

type DBOperations interface {
	CreateTaskForAgent(ctx context.Context, agentID int, taskType string) error
    SaveMessages(ctx context.Context, agentID int, messages []*api.ChatMessage) error
    // 未来可以添加更多方法，如 GetAgentByID 等
}
