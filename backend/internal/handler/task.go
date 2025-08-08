package handler

import (
	"net/http"
	"strconv"
	"errors"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"guardian/backend/internal/database"
)

	DB database.DBOperations
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	agentID, ok := r.Context().Value(AgentIDKey).(int)
	if !ok {
		slog.Error("Could not retrieve agentID from context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	taskType := "DUMP_WECHAT_DATA"
	err := h.DB.CreateTaskForAgent(r.Context(), agentID, taskType)
	if err != nil {
		if errors.Is(err, database.ErrAgentNotFound) {
			http.Error(w, "Agent not found", http.StatusNotFound)
		} else {
			slog.Error("Failed to create task", "error", err, "agent_id", agentID)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "task created"}`))
}
