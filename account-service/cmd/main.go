package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/pisya-dev/account-service/internal/config"
	"gitlab.com/pisya-dev/account-service/internal/repository"
	"gitlab.com/pisya-dev/account-service/internal/service"
	"gitlab.com/pisya-dev/account-service/internal/transport/grpc_server"
	"gitlab.com/pisya-dev/account-service/pkg/jwt"
	"gitlab.com/pisya-dev/account-service/pkg/logger"
	"gitlab.com/pisya-dev/account-service/pkg/postgres"
	"gitlab.com/pisya-dev/account-service/pkg/redis"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	ctx, err := logger.New(ctx)
	if err != nil {

	}

	config, err := config.NewConfig()

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "error  reading config:", zap.Error(err))
	}

	pg, err := postgres.NewPostgres(ctx, &config.Postgres)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "error open postgres", zap.Error(err))
	}

	rds, err := redis.NewRedis(&config.Redis, ctx)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "error connect redis", zap.Error(err))
	}

	publicKey, err := jwt.LoadPublicKey("./certs/public.pem")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "load public key error", zap.Error(err))

	}

	jwtService := jwt.NewServiceJWT(publicKey, time.Hour*6, time.Minute*15)

	repo := repository.NewRepository(pg)

	servise := service.NewService(repo, rds)

	accountServer := grpc_server.NewServer(servise, jwtService)

	server := grpc_server.NewGRPCServer(config, accountServer)

	go server.Run(ctx)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	defer stop()

	<-ctx.Done()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Shutdowning server...")

	server.ShutDown()
	pg.Close()

	rds.ShutdownSave(ctx)

	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server stoped")
}
