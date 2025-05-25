package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	jw "gitlab.com/pisya-dev/auth-service/pkg/jwt"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"go.uber.org/zap"
)

type Service interface {
	CreateAccount(ctx context.Context, req *dto.AccountReqs, id string) error

	AuthWithAccount(ctx context.Context, req *dto.AuthWithAccountReq) (string, error)
	GetProfileFromDb(ctx context.Context, req *dto.GetProfileID) (*dto.AccountReqs, error)

	CreateBuisnessAccount(ctx context.Context, req *dto.AccountReqs, id string) error

	CreatePromo(ctx context.Context, req *dto.PromoReq, id string) error
}

type Handlers struct {
	jwtService *jw.ServiceJWT

	service Service
}

func NewHandlers(service Service, jwtService *jw.ServiceJWT) *Handlers {
	return &Handlers{service: service, jwtService: jwtService}
}

func (h *Handlers) Ping(c echo.Context) error {
	return c.String(http.StatusOK, "POONG!")
}

func (h *Handlers) SingUp(c echo.Context) error {
	const op = "transport.rest.SingUp"
	ctx := c.Request().Context()
	var reqSturct dto.AccountReqs

	if err := c.Bind(&reqSturct); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return err
	}

	id := uuid.New()

	err := h.service.CreateAccount(ctx, &reqSturct, id.String())
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			fmt.Sprintf("%s error: ", op),
			zap.Error(err),
		)

		return c.JSON(http.StatusBadRequest, `{"status": "error","message": "Ошибка в данных запроса."}`)

	}

	accessToken, err := h.generateNewToken(c, ctx, id.String())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "idi nahui")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": accessToken})

}

func (h *Handlers) SignIn(c echo.Context) error {
	const op = "transport.rest.SingIn"
	ctx := c.Request().Context()
	var reqSturct dto.AuthWithAccountReq

	if err := c.Bind(&reqSturct); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return err
	}

	id, err := h.service.AuthWithAccount(ctx, &reqSturct)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return c.JSON(http.StatusBadRequest, "Not found account")
	}

	accessToken, err := h.generateNewToken(c, ctx, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "idi nahui")
	}
	return c.JSON(http.StatusOK, map[string]string{"token": accessToken})

}

func (h *Handlers) Profile(c echo.Context) error {
	const op = "transport.rest.Profile"
	ctx := c.Request().Context()

	req := dto.GetProfileID{}
	id, err := h.getIdFromSubject(c)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return c.JSON(http.StatusBadGateway, "die please!")
	}

	req.ID = id

	response, err := h.service.GetProfileFromDb(ctx, &req)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, "kill your self!!!")
	}

	return c.JSON(http.StatusOK, response)

}

func (h *Handlers) SingUpBuisness(c echo.Context) error {
	const op = "transport.rest.SingUpBuisness"
	ctx := c.Request().Context()

	var req dto.AccountReqs

	if err := c.Bind(&req); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return err
	}

	id := uuid.New()

	err := h.service.CreateBuisnessAccount(ctx, &req, id.String())
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx,
			fmt.Sprintf("%s error: ", op),
			zap.Error(err),
		)

		return c.JSON(http.StatusBadRequest, `{"status": "error","message": "Ошибка в данных запроса."}`)

	}

	accessToken, err := h.generateNewToken(c, ctx, id.String())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "idi nahui")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": accessToken})

}

func (h *Handlers) CreatePromo(c echo.Context) error {
	const op = "transport.rest.CreatePromo"
	ctx := c.Request().Context()

	var req dto.PromoReq

	if err := c.Bind(&req); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return err
	}
	id, err := h.getIdFromSubject(c)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprint("%w error:", op), zap.Error(err))
		return c.JSON(http.StatusBadRequest, "invalid id  in jwt")
	}

	err = h.service.CreatePromo(ctx, &req, id)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprint("%w error:", op), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ERORRR"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Succesful"})

}

func (h *Handlers) getIdFromSubject(c echo.Context) (string, error) {

	authHeader := c.Request().Header.Get("Authorization")

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := h.jwtService.DecodeKey(tokenString)
	if err != nil {

		return "", err
	}

	id, err := token.GetSubject()
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *Handlers) generateNewToken(c echo.Context, ctx context.Context, id string) (string, error) {
	const op = "transport.rest.GenerateNEwToken"

	newAccesClaims := jwt.MapClaims{"sub": id, "exp": dto.AccesTimeExpr}

	newRefreshClaims := jwt.MapClaims{"sub": id, "exp": dto.RefreshTimeExpr}

	accessToken, err := h.jwtService.Encode(newAccesClaims)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return "", err
	}

	refreshToken, err := h.jwtService.Encode(newRefreshClaims)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s error:", op), zap.Error(err))
		return "", err
	}

	newCokie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
	}
	c.SetCookie(newCokie)

	return accessToken, nil
}
