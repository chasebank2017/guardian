package service

	"context"
	"log/slog"

	"guardian/pkg/grpc/api"
)

type AgentServer struct {
	api.UnimplementedAgentServiceServer
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	slog.Info("Heartbeat received", "agent_id", req.AgentId, "hostname", req.Hostname)
	var taskType api.TaskType = api.TaskType_NONE
	if s.DB != nil {
		t, err := s.DB.GetAndDispatchPendingTaskForAgent(ctx, int(req.AgentId))
		if err != nil {
			slog.Error("db error", "error", err)
		} else if t == "DUMP_WECHAT_DATA" {
			taskType = api.TaskType_DUMP_WECHAT_DATA
		}
	}
	return &api.HeartbeatResponse{
		TaskType: taskType,
	}, nil
}
