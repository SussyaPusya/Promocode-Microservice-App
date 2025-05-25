package interceptor

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LogRequestUnaryInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {

		log.Info("Handling request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)

		resp, err = handler(ctx, req)

		if err != nil {
			log.Info("Error handling request",
				zap.String("method", info.FullMethod),
				zap.Error(err),
			)
			return nil, err
		}

		log.Info("Successfully handled request",
			zap.String("method", info.FullMethod),
			zap.Any("response", resp),
		)
		return resp, err
	}
}
