package main


import (
	"context"
	"log"
	"net"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
		// 需要 agentID 的路由组
		taskHandler := &handler.TaskHandler{DB: pool}
		protected.Route("/v1/agents/{agentID}", func(agent chi.Router) {
			agent.Use(handler.AgentCtx)
			agent.Post("/tasks", taskHandler.Create)
			// 未来可添加更多与 agentID 相关的路由
		})
		// gRPC-Gateway mux
		gwMux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		grpcGatewayAddr := "localhost" + cfg.Server.GrpcPort
		err = api.RegisterDataServiceHandlerFromEndpoint(ctx, gwMux, grpcGatewayAddr, opts)
		if err != nil {
			slog.Error("failed to register DataService handler", "error", err)
			panic(err)
		}
		err = api.RegisterAgentServiceHandlerFromEndpoint(ctx, gwMux, grpcGatewayAddr, opts)
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
