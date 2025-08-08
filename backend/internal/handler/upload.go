package handler

import (
	"encoding/json"
	"net/http"
	"log/slog"

	"guardian/backend/pkg/validator"
)

// UploadMessages 处理消息上传请求
func (h *TaskHandler) UploadMessages(w http.ResponseWriter, r *http.Request) {
	var payload UploadMessagesPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := validator.ValidateStruct(payload); err != nil {
		slog.Warn("Validation failed", "error", err)
		http.Error(w, "Invalid data: "+err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: 业务逻辑处理，如保存消息到数据库
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "messages uploaded"}`))
}
