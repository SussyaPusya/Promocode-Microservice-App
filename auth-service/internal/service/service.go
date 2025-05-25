package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	"gitlab.com/pisya-dev/auth-service/internal/transport/grpc_client"
	"gitlab.com/pisya-dev/auth-service/pkg/api/account_service"
	"gitlab.com/pisya-dev/auth-service/pkg/api/promopb"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"gitlab.com/pisya-dev/auth-service/pkg/security"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Repository interface {
	CreateAccount(ctx context.Context, req *dto.AccountReqs, id string) error
	GetAccount(ctx context.Context, req *dto.AuthWithAccountReq) (string, error)
	GetProfile(ctx context.Context, id string) (*dto.AccountReqs, error)
}

type Service struct {
	redisClient *redis.Client
	repo        Repository

	account *grpc_client.AccountServiceClient

	promo *grpc_client.PromoSvcClient
}

func NewService(repo Repository, reds *redis.Client, accountService *grpc_client.AccountServiceClient, promoService *grpc_client.PromoSvcClient) *Service {
	return &Service{repo: repo, redisClient: reds, account: accountService, promo: promoService}
}

func (s *Service) CreateAccount(ctx context.Context, req *dto.AccountReqs, id string) error {
	const op = "service.CreateAccount"

	newPasswordHash, err := security.Encode(req.Password)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return err
	}
	req.Password = newPasswordHash

	if err := s.repo.CreateAccount(ctx, req, id); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return err
	}

	profile := &account_service.CreateUserRequest{AvatarUrl: req.Avatar_url,
		Age:     req.Age,
		Name:    req.Name,
		Surname: req.Surname,
		Country: req.Country,
		Id:      id,
	}

	if _, err := s.account.CreateUserAccount(ctx, profile); err != nil {
		return err
	}

	return nil
}

func (s *Service) AuthWithAccount(ctx context.Context, req *dto.AuthWithAccountReq) (string, error) {

	id, err := s.redisClient.Get(ctx, req.Email).Result()
	if errors.Is(err, redis.Nil) {

		newPasswordHash, err := security.Encode(req.Password)
		if err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "encode password error", zap.Error(err))
			return "", err
		}
		req.Password = newPasswordHash

		//redis nado
		id, err := s.repo.GetAccount(ctx, req)

		if err != nil {
			logger.GetLoggerFromCtx(ctx).Info(ctx, "error:", zap.Error(err))
			return "", err
		}
		s.redisClient.Set(ctx, req.Email, id, dto.RedisTTL)
		return id, nil
	}
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "error:", zap.Error(err))

		return "", err
	}

	return id, nil

}

func (s *Service) GetProfileFromDb(ctx context.Context, req *dto.GetProfileID) (*dto.AccountReqs, error) {
	const op = "service.GetProfileFromDb"
	profile, err := s.repo.GetProfile(ctx, req.ID)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return nil, err
	}

	//receiving profile data from grpc
	//...
	account := &account_service.GetUserProfileRequest{Uuid: req.ID}

	resp, err := s.account.GetUserAccount(ctx, account)

	if err != nil {
		return nil, err
	}

	profile.Age = resp.Age
	profile.Avatar_url = resp.AvatarUrl
	profile.Name = resp.Name
	profile.Surname = resp.Surname
	profile.Country = resp.Country

	return profile, nil
}

func (s *Service) CreateBuisnessAccount(ctx context.Context, req *dto.AccountReqs, id string) error {
	const op = "service.BuisnessCreateAccount"

	newPasswordHash, err := security.Encode(req.Password)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return err
	}
	req.Password = newPasswordHash

	if err := s.repo.CreateAccount(ctx, req, id); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return err
	}

	profile := &account_service.CreateBuisnessRequest{
		Uuid: id,
		Name: req.Name,
	}

	if _, err := s.account.CreateBuisnessAccount(ctx, profile); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : error: ", op), zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) CreatePromo(ctx context.Context, req *dto.PromoReq, id string) error {

	promo := &promopb.CreatePromoRequest{
		Mode:        promopb.Mode_COMMON,
		PromoCommon: &req.Promo_common,
		PromoUnique: req.Promo_unique,
		Description: req.Description,
		ImageUrl:    &req.ImageUrl,
		Target: &promopb.Target{
			AgeFrom:    &req.Target.Age_from,
			AgeUntil:   &req.Target.Age_until,
			Country:    &req.Target.Country,
			Categories: req.Target.Categories,
		},
		MaxCount:    req.MaxCount,
		ActiveFrom:  timestamppb.New(req.Active_from),
		ActiveUntil: timestamppb.New(req.Active_until),
		CompanyId:   &id,
	}

	_, err := s.promo.CreatePromo(ctx, promo)

	return err
}
