package grpc_server_test

import (
	"context"
	"testing"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/internal/transport/grpc_server"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMiddlewareInterceptor(t *testing.T) {
	// Входные значения
	req := "test request"
	info := &grpc.UnaryServerInfo{
		FullMethod: "/test.Service/Method",
	}

	// Флаг, проверим, вызвался ли next
	called := false

	// Фейковый next хендлер
	next := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true

		// Проверим, что RequestID есть в контексте
		val := ctx.Value(domain.RequestID)
		assert.NotNil(t, val, "RequestID должен быть установлен в контексте")

		return "response", nil
	}

	// Вызов middleware
	resp, err := grpc_server.MiddlewareInterceptor(context.Background(), req, info, next)

	// Проверки
	assert.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.True(t, called, "next handler должен быть вызван")
}
