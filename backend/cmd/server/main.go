package main


import (
	"context"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"guardian/pkg/grpc/api"
	"guardian/backend/internal/service"
	"guardian/backend/internal/database"
	"guardian/backend/internal/handler"
)



func main() {
	ctx := context.Background()
	pool, err := database.NewConnection(ctx)
	if err != nil {
		log.Panicf("failed to connect to database: %v", err)
	}

	// gRPC 服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	agentSrv := &service.AgentServer{DB: pool}
	api.RegisterAgentServiceServer(grpcServer, agentSrv)
	dataSrv := &service.DataServer{DB: pool}
	api.RegisterDataServiceServer(grpcServer, dataSrv)
	go func() {
		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// HTTP/REST 服务器 (chi + grpc-gateway)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)


	// 登录API
	r.Post("/login", handler.Login)

	// 受保护API
	r.Group(func(protected chi.Router) {
		protected.Use(handler.JWTAuth)
		// 任务下发API
		taskHandler := &handler.TaskHandler{DB: pool}
		protected.Post("/v1/agents/{agentID}/tasks", taskHandler.Create)
		// gRPC-Gateway mux
		gwMux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err = api.RegisterDataServiceHandlerFromEndpoint(ctx, gwMux, "localhost:50051", opts)
		if err != nil {
			log.Fatalf("failed to register DataService handler: %v", err)
		}
		err = api.RegisterAgentServiceHandlerFromEndpoint(ctx, gwMux, "localhost:50051", opts)
		if err != nil {
			log.Fatalf("failed to register AgentService handler: %v", err)
		}
		protected.Mount("/", gwMux)
	})

	log.Println("HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("failed to serve HTTP: %v", err)
	}
}
