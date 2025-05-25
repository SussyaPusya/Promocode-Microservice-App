package grpc_server

import (
	"context"
	"strconv"
	"time"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func MiddlewareInterceptor(

	ctx context.Context,

	req interface{},

	info *grpc.UnaryServerInfo,

	next grpc.UnaryHandler,

) (any, error) {
	guid := uuid.New()
	str := strconv.FormatUint(uint64(guid.ID()), 10)

	ctx = context.WithValue(ctx, domain.RequestID, str)

	ctx, _ = logger.New(ctx)

	logger.GetLoggerFromCtx(ctx).Info(ctx,
		"request", zap.String("method", info.FullMethod),
		zap.Time("request time", time.Now()), zap.String("request id:", str))

	return next(ctx, req)
}
