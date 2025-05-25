package repository_test

import (
	"context"
	"testing"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/internal/repository"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateUser(t *testing.T) {
	ctx := context.Background()

	ctx, _ = logger.New(ctx)

	mock, err := pgxmock.NewPool()

	require.NoError(t, err)

	defer mock.Close()

	repo := repository.NewRepository(mock) // всё огонь

	user := &domain.User{
		Guid:       "u1",
		Name:       "Ivan",
		Surname:    "Ivanov",
		Avatar_url: "avatar.jpg",
		Age:        25,
		Country:    "RU",
	}

	mock.ExpectExec(`INSERT INTO users`).
		WithArgs(user.Guid, user.Name, user.Surname, user.Avatar_url, user.Age, user.Country).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateUser(ctx, user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_UpdateUser(t *testing.T) {
	ctx := context.Background()
	ctx, _ = logger.New(ctx)
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	defer mock.Close()

	repo := repository.NewRepository(mock)

	user := &domain.User{
		Guid:       "u1",
		Name:       "Ivan",
		Surname:    "Petrov",
		Avatar_url: "avatar2.jpg",
		Age:        30,
		Country:    "BY",
	}

	mock.ExpectExec(`UPDATE users`).
		WithArgs(user.Name, user.Surname, user.Avatar_url, user.Age, user.Country, user.Guid).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err = repo.UpdateUser(ctx, user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_DeleteUser(t *testing.T) {
	ctx := context.Background()
	ctx, _ = logger.New(ctx)
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	defer mock.Close()

	repo := repository.NewRepository(mock)

	user := &domain.User{Guid: "u1"}

	mock.ExpectExec(`DELETE FROM users`).
		WithArgs(user.Guid).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.DeleteUser(ctx, user)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_CreateBuisness(t *testing.T) {
	ctx := context.Background()
	ctx, _ = logger.New(ctx)
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	defer mock.Close()

	repo := repository.NewRepository(mock)

	bis := &domain.Business{Guid: "b1", Name: "My Biz"}

	mock.ExpectExec(`INSERT INTO buisness`).
		WithArgs(bis.Guid, bis.Name).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = repo.CreateBuisness(ctx, bis)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_GetBuisness(t *testing.T) {
	ctx := context.Background()
	ctx, _ = logger.New(ctx)
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)

	defer mock.Close()

	repo := repository.NewRepository(mock)

	bis := &domain.Business{Guid: "b1"}
	expectedName := "My Biz"

	rows := pgxmock.NewRows([]string{"id", "name"}).
		AddRow(bis.Guid, expectedName)

	mock.ExpectQuery(`SELECT id, name FROM buisness`).
		WithArgs(bis.Guid).
		WillReturnRows(rows)

	name, err := repo.GetBuisness(ctx, bis)
	require.NoError(t, err)
	require.Equal(t, expectedName, name)
	require.NoError(t, mock.ExpectationsWereMet())
}
