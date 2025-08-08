package main


import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    "log"
    "log/slog"
    "net"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    // "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"

    api "guardian-backend/pkg/grpc/api/guardian/pkg/grpc/api"
    "guardian-backend/internal/service"
    "guardian-backend/internal/database"
    "guardian-backend/internal/handler"
    "guardian-backend/internal/config"
    m "guardian-backend/pkg/metrics"
    promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)



func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	pool, err := database.NewConnectionWithDSN(ctx, cfg.Database.DSN)
	if err != nil {
		log.Panicf("failed to connect to database: %v", err)
	}

	// gRPC 服务器 (mTLS)
	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// 加载 CA 证书
	caCert, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA cert: %v", err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)
	// 加载服务器证书和私钥
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("failed to load server cert/key: %v", err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
	creds := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(grpc.Creds(creds))
    agentSrv := &service.AgentServer{DB: pool.Pool}
	api.RegisterAgentServiceServer(grpcServer, agentSrv)
    dataSrv := &service.DataServer{DB: pool.Pool}
	api.RegisterDataServiceServer(grpcServer, dataSrv)
	go func() {
		log.Printf("gRPC server listening on %s", cfg.Server.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// HTTP/REST 服务器 (chi + grpc-gateway)
	r := chi.NewRouter()
    r.Use(middleware.RequestID)
    r.Use(handler.RequestLogger)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    r.Use(handler.CORSWithOrigins(cfg.Server.CORSOrigins))
    r.Use(m.HTTPMetrics)
    if cfg.Server.RequestTimeoutSeconds <= 0 { cfg.Server.RequestTimeoutSeconds = 15 }
    r.Use(handler.WithRequestTimeout(time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second))


	// 登录API
	// 实例化 AuthHandler
    authHandler := &handler.AuthHandler{JWTSecret: cfg.Auth.JWTSecret, AdminUsername: cfg.Auth.AdminUsername, AdminPassword: cfg.Auth.AdminPassword}
    // 登录独立限流（每秒 5 次，突发 10）
    loginRPS := cfg.Server.RateLimit.LoginRPS; if loginRPS <= 0 { loginRPS = 5 }
    loginBurst := cfg.Server.RateLimit.LoginBurst; if loginBurst <= 0 { loginBurst = 10 }
    r.With(handler.TokenBucketLimiter(loginRPS, loginBurst)).Post("/login", authHandler.Login)

	// 受保护API
    r.Group(func(protected chi.Router) {
		protected.Use(handler.JWTAuth(cfg.Auth.JWTSecret))
        // 健康/就绪探针（无需鉴权也可考虑暴露在 /healthz）
        r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
        r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ready")) })
        r.Handle("/metrics", promhttp.Handler())
        taskHandler := &handler.TaskHandler{DB: pool}
        // 列表：GET /v1/agents
        // 对受保护接口设置较高阈值（每秒 50 次，突发 100）
        prps := cfg.Server.RateLimit.ProtectedRPS; if prps <= 0 { prps = 50 }
        pburst := cfg.Server.RateLimit.ProtectedBurst; if pburst <= 0 { pburst = 100 }
        protected.With(handler.TokenBucketLimiter(prps, pburst)).Get("/v1/agents", taskHandler.Agents)
        // 需要 agentID 的路由组
        protected.Route("/v1/agents/{agentID}", func(agent chi.Router) {
            agent.Use(handler.AgentCtx)
            agent.Post("/tasks", taskHandler.Create)
            agent.Get("/messages", taskHandler.MessagesByAgent) // GET /v1/agents/{agentID}/messages
        })
        // TODO: 如需启用 HTTP 转码，请生成 *.gw.go 并在此注册 gRPC-Gateway
	})

   slog.Info("HTTP server listening", "port", cfg.Server.Port)
   if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
	   slog.Error("failed to serve HTTP", "error", err)
	   panic(err)
   }
}
