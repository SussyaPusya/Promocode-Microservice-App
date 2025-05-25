package grpc_server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"gitlab.com/pisya-dev/account-service/internal/config"
	pb "gitlab.com/pisya-dev/account-service/pkg/api/account_service"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	cfg    *config.Config
	server *grpc.Server

	accountSrvc *Server
}

func (g *GrpcServer) Run(ctx context.Context) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", g.cfg.GRPCPort))
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "filed to serve", zap.Error(err))
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server start on:", zap.String("url", lis.Addr().String()), zap.String("cfg port:", strconv.Itoa(g.cfg.GRPCPort)))

	g.server = grpc.NewServer(
		grpc.UnaryInterceptor(MiddlewareInterceptor),
	)

	pb.RegisterAccount_ServiceServer(g.server, g.accountSrvc)

	if err := g.server.Serve(lis); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "filed to serve", zap.Error(err))
	}
}

func NewGRPCServer(cfg *config.Config, accountServer *Server) *GrpcServer {
	return &GrpcServer{cfg: cfg, accountSrvc: accountServer}
}

func (g *GrpcServer) ShutDown() {
	g.server.GracefulStop()
}
