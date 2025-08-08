package handler

import (
    "context"
    "net/http"
    "strconv"
    "strings"
    "time"
    "log/slog"

    "github.com/go-chi/chi/v5"
    "github.com/golang-jwt/jwt/v5"
    chimid "github.com/go-chi/chi/v5/middleware"
)

// 定义一个唯一的key类型，用于在context中存取值，避免冲突
type contextKey string
const AgentIDKey contextKey = "agentID"
const RequestIDKey contextKey = "requestID"

// TODO: 统一错误响应结构可通过包装 http.Error，附带 requestID 与 code

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

// WithRequestTimeout 为请求设置超时
func WithRequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.TimeoutHandler(next, timeout, "request timeout")
    }
}

// SimpleCORS 允许开发阶段的跨域（生产应收敛域名）
func CORSWithOrigins(origins []string) func(http.Handler) http.Handler {
    allowAll := false
    m := map[string]struct{}{}
    for _, o := range origins {
        if o == "*" { allowAll = true }
        m[o] = struct{}{}
    }
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            if allowAll {
                w.Header().Set("Access-Control-Allow-Origin", "*")
            } else if origin != "" {
                if _, ok := m[origin]; ok {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                }
            }
            w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// TokenBucketLimiter 简易令牌桶限流中间件
func TokenBucketLimiter(rps int, burst int) func(http.Handler) http.Handler {
    if rps <= 0 { rps = 1 }
    if burst < 0 { burst = 0 }
    ch := make(chan struct{}, burst)
    ticker := time.NewTicker(time.Second / time.Duration(rps))
    go func() {
        for range ticker.C {
            select { case ch <- struct{}{}: default: }
        }
    }()
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            select {
            case <-r.Context().Done():
                http.Error(w, "canceled", http.StatusRequestTimeout)
                return
            case <-ch:
                next.ServeHTTP(w, r)
            default:
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            }
        })
    }
}

// RequestLogger 记录每个请求的开始/结束日志，包含 request_id
func RequestLogger(next http.Handler) http.Handler {
    type rw struct { http.ResponseWriter; status int }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rid := ""
        if v := r.Context().Value(chimid.RequestIDKey); v != nil {
            if s, ok := v.(string); ok { rid = s }
        }
        start := time.Now()
        wrapper := &rw{ResponseWriter: w, status: http.StatusOK}
        slog.Info("request start", "request_id", rid, "method", r.Method, "path", r.URL.Path)
        next.ServeHTTP(wrapper, r)
        slog.Info("request end", "request_id", rid, "method", r.Method, "path", r.URL.Path, "status", wrapper.status, "duration_ms", time.Since(start).Milliseconds())
    })
}
