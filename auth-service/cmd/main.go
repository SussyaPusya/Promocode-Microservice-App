package main

import (
	"context"

	"gitlab.com/pisya-dev/auth-service/internal/config"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	"gitlab.com/pisya-dev/auth-service/internal/repository"
	"gitlab.com/pisya-dev/auth-service/internal/service"
	"gitlab.com/pisya-dev/auth-service/internal/transport/grpc_client"
	"gitlab.com/pisya-dev/auth-service/internal/transport/rest"
	"gitlab.com/pisya-dev/auth-service/internal/transport/rest/middleware"
	"gitlab.com/pisya-dev/auth-service/pkg/jwt"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"gitlab.com/pisya-dev/auth-service/pkg/postgres"
	"gitlab.com/pisya-dev/auth-service/pkg/redis"
	"go.uber.org/zap"
)

func main() {

	ctx := context.Background()

	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "error", zap.Error(err))
	}

	config, err := config.NewConfig()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "error", zap.Error(err))
	}

	privateKey, err := jwt.LoadPrivateKey("./certs/private.pem")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "load private key error", zap.Error(err))
	}
	publicKey, err := jwt.LoadPublicKey("./certs/public.pem")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "load public key error", zap.Error(err))

	}

	jwtService := jwt.NewServiceJWT(privateKey, publicKey, dto.RefreshTimeExpr, dto.AccesTimeExpr)

	middlware := middleware.NewMiddlware(jwtService)

	db, err := postgres.NewPostgres(ctx, &config.Postgres)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "filed to connect pstgres", zap.Error(err))
	}

	authRedis, err := redis.NewRedis(&config.Redis, ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "filed to connect redis", zap.Error(err))
	}

	authRepository := repository.NewRepository(db)

	accountClient, err := grpc_client.NewAccountServiceClient(config.GrpcConfig)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "FILED TO CONNECT ACCOUNT CLIENT STUPID NIGGA SON OF A BITch")
	}

	promoService, err := grpc_client.NewPromoServiceClient(config.GrpcConfig)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "FILED TO CONNECT PROMO CLIENT STUPID NIGGA SON OF A BITch")
	}

	authService := service.NewService(authRepository, authRedis, accountClient, promoService)

	authHandlers := rest.NewHandlers(authService, jwtService)

	authRouter := rest.NewRouter(config.Rest, authHandlers, ctx, middlware)

	authRouter.Run(ctx)

}
