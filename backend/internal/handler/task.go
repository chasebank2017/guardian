package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskHandler struct {
	DB *pgxpool.Pool
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
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "task created"}`))
}
