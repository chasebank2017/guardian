package httpx

import (
    "encoding/json"
    "net/http"

    chimid "github.com/go-chi/chi/v5/middleware"
)

type ErrorResponse struct {
    Code      string `json:"code"`
    Message   string `json:"message"`
    RequestID string `json:"requestId,omitempty"`
}

func requestIDFrom(r *http.Request) string {
    if r == nil {
        return ""
    }
    if v := r.Header.Get("X-Request-Id"); v != "" {
        return v
    }
    if v := r.Context().Value(chimid.RequestIDKey); v != nil {
        if s, ok := v.(string); ok { return s }
    }
    return ""
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
    resp := ErrorResponse{Code: code, Message: message, RequestID: requestIDFrom(r)}
    WriteJSON(w, status, resp)
}


