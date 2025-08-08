package handler

import (
	"context"
	"strconv"
	"net/http"
	"strings"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// 定义一个唯一的key类型，用于在context中存取值，避免冲突
type contextKey string
const AgentIDKey contextKey = "agentID"

// AgentCtx 是一个中间件，负责从URL中解析agentID并存入请求的context中
func AgentCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agentIDStr := chi.URLParam(r, "agentID")
		if agentIDStr == "" {
			http.Error(w, "Agent ID is required", http.StatusBadRequest)
			return
		}
		agentID, err := strconv.Atoi(agentIDStr)
		if err != nil {
			http.Error(w, "Invalid Agent ID format", http.StatusBadRequest)
			return
		}
		// 将解析出的 agentID 存入 context
		ctx := context.WithValue(r.Context(), AgentIDKey, agentID)
		// 将带有新 context 的请求传递给下一个处理器
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
// ...existing code...

func JWTAuth(jwtSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, "missing token", http.StatusUnauthorized)
				return
			}
			tokenStr := header[7:]
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
