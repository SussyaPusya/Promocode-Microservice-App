package logger

import (
	"context"

	"gitlab.com/pisya-dev/auth-service/internal/dto"

	"go.uber.org/zap"
)

type Logger struct {
	L *zap.Logger
}

func New(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, dto.Logger, &Logger{logger})

	return ctx, nil
}

func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(dto.Logger).(*Logger)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(dto.RequestID) != nil {
		fields = append(fields, zap.String(string(dto.RequestID), ctx.Value(dto.RequestID).(string)))
	}

	l.L.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(dto.RequestID) != nil {
		fields = append(fields, zap.String(string(dto.RequestID), ctx.Value(dto.RequestID).(string)))
	}

	l.L.Fatal(msg, fields...)
}
