package service

import (
    "context"
    "log/slog"

    api "guardian-backend/pkg/grpc/api/guardian/pkg/grpc/api"
    "github.com/jackc/pgx/v5/pgxpool"
    "guardian-backend/internal/database"
)

type DataServer struct {
	api.UnimplementedDataServiceServer
    DB *pgxpool.Pool
}

func (s *DataServer) UploadMessages(ctx context.Context, req *api.UploadMessagesRequest) (*api.UploadMessagesResponse, error) {
    if s.DB == nil {
		slog.Error("DB pool is nil")
		return &api.UploadMessagesResponse{Success: false}, nil
	}
    // TODO: 这里应通过数据库接口层保存消息；当前以日志代替，避免直接在 *pgxpool.Pool 上调用未定义方法
    err := (&database.DB{Pool: s.DB}).SaveMessages(ctx, int(req.AgentId), req.Messages)
	if err != nil {
		slog.Error("Failed to save messages", "error", err, "agent_id", req.AgentId)
		return &api.UploadMessagesResponse{Success: false}, err
	}
	slog.Info("Saved messages", "count", len(req.Messages), "agent_id", req.AgentId)
	return &api.UploadMessagesResponse{Success: true}, nil
}
