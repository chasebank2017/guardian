package service

import (
	"context"
	"log"
	"time"

	"guardian/pkg/grpc/api"
	"github.com/jackc/pgx/v5/pgxpool"
	"guardian/backend/internal/database"
)

type DataServer struct {
	api.UnimplementedDataServiceServer
	DB *pgxpool.Pool
}

func (s *DataServer) UploadMessages(ctx context.Context, req *api.UploadMessagesRequest) (*api.UploadMessagesResponse, error) {
	if s.DB == nil {
		log.Printf("DB pool is nil")
		return &api.UploadMessagesResponse{Success: false}, nil
	}
	err := s.DB.SaveMessages(ctx, int(req.AgentId), req.Messages)
	if err != nil {
		log.Printf("Failed to save messages: %v", err)
		return &api.UploadMessagesResponse{Success: false}, err
	}
	log.Printf("Saved %d messages for agent %d", len(req.Messages), req.AgentId)
	return &api.UploadMessagesResponse{Success: true}, nil
}
