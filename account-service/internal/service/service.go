package service

import (
	"context"
	"errors"
	"log"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Repository interface {
	CreateUser(context.Context, *domain.User) error
	UpdateUser(context.Context, *domain.User) error
	DeleteUser(context.Context, *domain.User) error

	CreateBuisness(context.Context, *domain.Business) error
	GetBuisness(context.Context, *domain.Business) (string, error)
	GetUser(context.Context, string) (*domain.User, error)
}

type Service struct {
	repo Repository

	redisClient *redis.Client
}

func NewService(repo Repository, client *redis.Client) *Service {
	return &Service{repo: repo, redisClient: client}
}

func (s *Service) CreateUser(ctx context.Context, user *domain.User) error {
	err := s.repo.CreateUser(ctx, user)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "filed create user error:", zap.Error(err))

		return err
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, user *domain.User) error {
	err := s.repo.UpdateUser(ctx, user)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "filed update user error:", zap.Error(err))

		return err
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, user *domain.User) error {
	err := s.repo.DeleteUser(ctx, user)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "filed delete user error:", zap.Error(err))

		return err
	}

	return nil
}

func (s *Service) CreateBuisness(ctx context.Context, bis *domain.Business) error {
	log.Println("repo create")
	err := s.repo.CreateBuisness(ctx, bis)
	log.Println("repo create was")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "filed create business error:", zap.Error(err))

		return err
	}

	log.Println("rпосле редиса")
	return nil
}

func (s *Service) GetBuisness(ctx context.Context, bis *domain.Business) (string, error) {
	val, err := s.redisClient.Get(ctx, bis.Guid).Result()

	if errors.Is(err, redis.Nil) {
		name, err := s.repo.GetBuisness(ctx, bis)

		if err != nil {
			logger.GetLoggerFromCtx(ctx).Info(ctx, "filed get business error:", zap.Error(err))

			return "", err
		}

		s.redisClient.Set(ctx, bis.Guid, name, domain.RedisTLl)

		return name, nil
	}

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "error:", zap.Error(err))

		return "", err
	}

	return val, nil
}

func (s *Service) GetUser(ctx context.Context, user *domain.User) (*domain.User, error) {

	userProfile, err := s.repo.GetUser(ctx, user.Guid)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "filed get user error:", zap.Error(err))

		return nil, err
	}
	return userProfile, nil

}
