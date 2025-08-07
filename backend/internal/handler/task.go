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
	agentIDStr := chi.URLParam(r, "agentID")
	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		http.Error(w, "invalid agentID", http.StatusBadRequest)
		return
	}
	taskType := "DUMP_WECHAT_DATA"
	err = h.DB.CreateTaskForAgent(r.Context(), agentID, taskType)
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
