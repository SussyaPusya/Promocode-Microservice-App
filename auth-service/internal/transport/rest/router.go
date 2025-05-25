package rest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
	"gitlab.com/pisya-dev/auth-service/internal/config"
	"gitlab.com/pisya-dev/auth-service/internal/transport/rest/middleware"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"go.uber.org/zap"
)

type Router struct {
	router *echo.Echo

	handlers *Handlers

	config config.Rest
}

func NewRouter(cfg config.Rest, handlers *Handlers, ctx context.Context, middleware *middleware.Middleware) *Router {
	e := echo.New()

	e.Server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	//hasndlers
	e.Use(middleware.Logger)
	e.Use(middleware.Auth)

	e.POST("/user/auth/sign-up", handlers.SingUp)
	e.POST("/user/auth/sign-in", handlers.SignIn)
	e.POST("/business/auth/sign-up", handlers.SingUpBuisness)
	e.POST("/business/auth/sign-in", handlers.SignIn)
	e.POST("/business/promo", handlers.CreatePromo)

	e.GET(("/user/profile"), handlers.Profile)
	e.GET("/ping", handlers.Ping)
	//e.GET("/", h.asdasd)

	return &Router{router: e, handlers: handlers, config: cfg}

}

func (r *Router) Run(ctx context.Context) {

	restAddr := fmt.Sprintf(":%d", r.config.Port)

	if err := r.router.Start(restAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "failed to start server", zap.Error(err))
	}

}
