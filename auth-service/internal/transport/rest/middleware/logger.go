package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"go.uber.org/zap"
)

func (m *Middleware) Logger(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {
		// Генерация request ID
		guid := uuid.New()
		reqID := strconv.FormatUint(uint64(guid.ID()), 10)

		// Получение оригинального контекста
		ctx := c.Request().Context()

		// Создание нового контекста с request ID
		ctx = context.WithValue(ctx, dto.RequestID, reqID)

		// Обновление контекста в echo.Context
		req := c.Request().WithContext(ctx)
		c.SetRequest(req)

		// Теперь можешь использовать этот контекст
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			"request", zap.String("method", c.Request().Method),
			zap.Time("request_time", time.Now()), zap.String("request_id", reqID))

		// Передача дальше
		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil
	}

}
