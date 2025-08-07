package service

import (
	"context"
	"log"

	"guardian/pkg/grpc/api"
)

type AgentServer struct {
	api.UnimplementedAgentServiceServer
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	log.Printf("Heartbeat received: agent_id=%d, hostname=%s", req.AgentId, req.Hostname)
	var taskType api.TaskType = api.TaskType_NONE
	if s.DB != nil {
		t, err := s.DB.GetAndDispatchPendingTaskForAgent(ctx, int(req.AgentId))
		if err != nil {
			log.Printf("db error: %v", err)
		} else if t == "DUMP_WECHAT_DATA" {
			taskType = api.TaskType_DUMP_WECHAT_DATA
		}
	}
	return &api.HeartbeatResponse{
		TaskType: taskType,
	}, nil
}
