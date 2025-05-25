package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	jw "gitlab.com/pisya-dev/auth-service/pkg/jwt"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"go.uber.org/zap"
)

type Middleware struct {
	jwtService *jw.ServiceJWT
}

func NewMiddlware(jwtServ *jw.ServiceJWT) *Middleware {
	return &Middleware{jwtService: jwtServ}
}

func (m *Middleware) Auth(next echo.HandlerFunc) echo.HandlerFunc {

	return func(c echo.Context) error {
		ctx := c.Request().Context()

		if c.Request().URL.Path == "/user/auth/sign-up" || c.Request().URL.Path == "/user/auth/sign-in" || c.Request().URL.Path == "/business/auth/sign-up" || c.Request().URL.Path == "/business/auth/sign-in" {
			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}

		authHeader := c.Request().Header.Get("Authorization")

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		_, err := m.jwtService.DecodeKey(tokenString)

		if err != nil {
			refresh, err := c.Cookie("refresh_token")
			if err != nil {
				fmt.Println(refresh)
				logger.GetLoggerFromCtx(ctx).Info(ctx, "iinvalid access and none refresh jwt :", zap.Error(err))

				return echo.NewHTTPError(http.StatusBadRequest, "Invalid access and none refresh jwt")
			}

			claims, err := m.jwtService.DecodeKey(refresh.Value)
			if errors.Is(err, jw.ErrorInvalidToken) {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid access and refresh jwt")
			}

			guid, err := claims.GetSubject()
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID in JWT")
			}

			newAccesClaims := jwt.MapClaims{"sub": guid, "exp": dto.AccesTimeExpr}

			newRefreshClaims := jwt.MapClaims{"sub": guid, "exp": dto.RefreshTimeExpr}

			accessToken, err := m.jwtService.Encode(newAccesClaims)
			if err != nil {
				return echo.ErrBadGateway
			}

			refreshToken, err := m.jwtService.Encode(newRefreshClaims)
			if err != nil {
				return echo.ErrBadGateway
			}

			newCokie := &http.Cookie{
				Name:     "refresh_token",
				Value:    refreshToken,
				HttpOnly: true,
				Path:     "/",
			}
			c.SetCookie(newCokie)

			return c.JSON(http.StatusOK, map[string]string{"token": accessToken})

		}

		if err := next(c); err != nil {
			c.Error(err)
		}
		return nil

	}
}
