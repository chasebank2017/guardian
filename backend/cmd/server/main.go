package main


import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"guardian/pkg/grpc/api"
	"guardian/backend/internal/service"
	"guardian/backend/internal/database"
	"guardian/backend/internal/handler"
	"guardian/backend/internal/config"
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

	// gRPC 服务器
	lis, err := net.Listen("tcp", cfg.Server.GrpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	agentSrv := &service.AgentServer{DB: pool}
	api.RegisterAgentServiceServer(grpcServer, agentSrv)
	dataSrv := &service.DataServer{DB: pool}
	api.RegisterDataServiceServer(grpcServer, dataSrv)
	go func() {
		log.Printf("gRPC server listening on %s", cfg.Server.GrpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// HTTP/REST 服务器 (chi + grpc-gateway)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)


	// 登录API
	// 实例化 AuthHandler
	authHandler := &handler.AuthHandler{JWTSecret: cfg.Auth.JWTSecret}
	r.Post("/login", authHandler.Login)

	// 受保护API
	r.Group(func(protected chi.Router) {
		protected.Use(handler.JWTAuth(cfg.Auth.JWTSecret))
		// 任务下发API
		taskHandler := &handler.TaskHandler{DB: pool}
		protected.Post("/v1/agents/{agentID}/tasks", taskHandler.Create)
		// gRPC-Gateway mux
		gwMux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err = api.RegisterDataServiceHandlerFromEndpoint(ctx, gwMux, "localhost:50051", opts)
		if err != nil {
			slog.Error("failed to register DataService handler", "error", err)
			panic(err)
		}
		err = api.RegisterAgentServiceHandlerFromEndpoint(ctx, gwMux, "localhost:50051", opts)
		if err != nil {
			slog.Error("failed to register AgentService handler", "error", err)
			panic(err)
		}
		protected.Mount("/", gwMux)
	})

   slog.Info("HTTP server listening", "port", cfg.Server.Port)
   if err := http.ListenAndServe(cfg.Server.Port, r); err != nil {
	   slog.Error("failed to serve HTTP", "error", err)
	   panic(err)
   }
}
