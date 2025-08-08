package service

	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"guardian/pkg/grpc/api"
)

type AgentServer struct {
	api.UnimplementedAgentServiceServer
	DB *pgxpool.Pool
}

func (s *AgentServer) RegisterAgent(ctx context.Context, req *api.RegisterAgentRequest) (*api.RegisterAgentResponse, error) {
	slog.Info("RegisterAgent request received", "hostname", req.Hostname, "os_version", req.OsVersion)
	if s.DB == nil {
		slog.Error("DB pool is nil")
		return &api.RegisterAgentResponse{Status: "error"}, nil
	}

	agentID, err := database.CreateAgent(ctx, s.DB, req.Hostname, req.OsVersion)
	if err != nil {
		slog.Error("Failed to create agent", "error", err)
		return &api.RegisterAgentResponse{Status: "error"}, err
	}

	slog.Info("New agent registered", "agent_id", agentID)
	return &api.RegisterAgentResponse{AgentId: int32(agentID), Status: "ok"}, nil
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *api.HeartbeatRequest) (*api.HeartbeatResponse, error) {
	slog.Info("Heartbeat received", "agent_id", req.AgentId)
	var taskType api.TaskType = api.TaskType_NONE
	if s.DB != nil {
		// TODO: Implement database logic for GetAndDispatchPendingTaskForAgent
		// t, err := s.DB.GetAndDispatchPendingTaskForAgent(ctx, int(req.AgentId))
		// if err != nil {
		// 	slog.Error("db error", "error", err)
		// } else if t == "DUMP_WECHAT_DATA" {
		// 	taskType = api.TaskType_DUMP_WECHAT_DATA
		// }
	}
	return &api.HeartbeatResponse{
		TaskType: taskType,
	}, nil
}
