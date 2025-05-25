package logger_test

import (
	"context"
	"testing"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestNewLogger(t *testing.T) {
	ctx := context.Background()

	ctxWithLogger, err := logger.New(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, ctxWithLogger)

	l := logger.GetLoggerFromCtx(ctxWithLogger)
	assert.NotNil(t, l)
}

func TestLogger_Info(t *testing.T) {
	t.Parallel()

	// создаём zap logger с выводом в тест
	zapTestLogger := zaptest.NewLogger(t)

	// Вставляем наш кастомный логгер в контекст
	ctx := context.WithValue(context.Background(), domain.Logger, &logger.Logger{L: zapTestLogger})
	ctx = context.WithValue(ctx, domain.RequestID, "123-req-id")

	l := logger.GetLoggerFromCtx(ctx)
	assert.NotNil(t, l)

	l.Info(ctx, "testing info message")
}

func TestLogger_Fatal_Skip(t *testing.T) {
	t.Skip("Logger.Fatal вызывает os.Exit — не тестим напрямую")
}
