package service

import (
    "context"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"
    api "guardian-backend/pkg/grpc/api/guardian/pkg/grpc/api"
)

type AgentServer struct {
	api.UnimplementedAgentServiceServer
	DB *pgxpool.Pool
}

// RegisterAgent 在当前 proto 中不存在，移除实现以避免未定义类型错误

func (s *AgentServer) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	slog.Info("Heartbeat received", "agent_id", req.AgentId)
	// TODO: 未来可加入任务分发逻辑
	return &api.HeartbeatResponse{TaskId: ""}, nil
}
