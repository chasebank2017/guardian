package metrics

import (
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "guardian_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "route", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "guardian_http_request_duration_seconds",
            Help:    "Duration of HTTP requests in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "route", "status"},
    )
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(code int) {
    w.status = code
    w.ResponseWriter.WriteHeader(code)
}

// HTTPMetrics returns a middleware that records basic HTTP metrics.
func HTTPMetrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        next.ServeHTTP(rw, r)

        route := chi.RouteContext(r.Context())
        pattern := "unknown"
        if route != nil {
            pattern = route.RoutePattern()
            if pattern == "" {
                pattern = "unknown"
            }
        }

        labels := prometheus.Labels{
            "method": r.Method,
            "route":  pattern,
            "status": http.StatusText(rw.status),
        }
        httpRequestsTotal.With(labels).Inc()
        httpRequestDuration.With(labels).Observe(time.Since(start).Seconds())
    })
}


