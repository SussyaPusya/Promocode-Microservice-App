package service_test

import (
	"context"
	"testing"
	"time"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/internal/service"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Мок репозиторий.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *MockRepository) DeleteUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *MockRepository) CreateBuisness(ctx context.Context, bis *domain.Business) error {
	args := m.Called(ctx, bis)

	return args.Error(0)
}

func (m *MockRepository) GetBuisness(ctx context.Context, bis *domain.Business) (string, error) {
	args := m.Called(ctx, bis)

	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetUser(ctx context.Context, user string) (*domain.User, error) {
	args := m.Called(ctx, user)

	return nil, args.Error(1)
}

// Тест CreateUser.
func TestService_CreateUser(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, _ := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	user := &domain.User{Guid: "1", Name: "Test"}

	repo.On("CreateUser", ctx, user).Return(nil)

	err := svc.CreateUser(ctx, user)
	require.NoError(t, err)

	repo.AssertExpectations(t)
}

// Тест UpdateUser.
func TestService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, _ := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	user := &domain.User{Guid: "1", Name: "Test"}

	repo.On("UpdateUser", ctx, user).Return(nil)

	err := svc.UpdateUser(ctx, user)
	require.NoError(t, err)
}

// Тест DeleteUser.
func TestService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, _ := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	user := &domain.User{Guid: "1"}

	repo.On("DeleteUser", ctx, user).Return(nil)

	err := svc.DeleteUser(ctx, user)
	require.NoError(t, err)
}

// Тест CreateBuisness.
func TestService_CreateBuisness(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, mock := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	bis := &domain.Business{Guid: "123", Name: "BizName"}

	repo.On("CreateBuisness", ctx, bis).Return(nil)
	mock.ExpectSet(bis.Guid, bis.Name, time.Minute*5).SetVal("OK")

	err := svc.CreateBuisness(ctx, bis)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

// Тест GetBuisness (из Redis).
func TestService_GetBuisness_CacheHit(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, mock := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	bis := &domain.Business{Guid: "123"}

	mock.ExpectGet(bis.Guid).SetVal("CachedName")

	name, err := svc.GetBuisness(ctx, bis)
	require.NoError(t, err)
	require.Equal(t, "CachedName", name)

	require.NoError(t, mock.ExpectationsWereMet())
}

// Тест GetBuisness (из репозитория, если Redis пустой).
func TestService_GetBuisness_CacheMiss(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	rdb, mock := redismock.NewClientMock()
	svc := service.NewService(repo, rdb)

	bis := &domain.Business{Guid: "123"}

	mock.ExpectGet(bis.Guid).RedisNil()
	repo.On("GetBuisness", ctx, bis).Return("DBName", nil)
	mock.ExpectSet(bis.Guid, "DBName", time.Minute*5).SetVal("OK")

	name, err := svc.GetBuisness(ctx, bis)
	require.NoError(t, err)
	require.Equal(t, "DBName", name)

	require.NoError(t, mock.ExpectationsWereMet())
}
