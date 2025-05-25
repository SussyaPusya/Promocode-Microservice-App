package repository

import (
	"context"
	"sync"

	"gitlab.com/pisya-dev/account-service/internal/domain"
	"gitlab.com/pisya-dev/account-service/pkg/logger"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Repository struct {
	pg PgxIface

	mu sync.Mutex
}

type PgxIface interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func NewRepository(pg PgxIface) *Repository {
	return &Repository{pg: pg, mu: sync.Mutex{}}
}

func (r *Repository) CreateUser(ctx context.Context, user *domain.User) error {
	query := sq.Insert("users").
		Columns("id", "name", "surname", "avatar_url", "age", "country").
		Values(user.Guid, user.Name, user.Surname, user.Avatar_url, user.Age, user.Country).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL:", zap.Error(err))

		return err
	}

	// --- Выполняем запрос ---

	_, err = r.pg.Exec(ctx, sql, args...)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "INSERT failed:", zap.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *domain.User) error {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Update("users").
		Set("name", user.Name).
		Set("surname", user.Surname).
		Set("avatar_url", user.Avatar_url).
		Set("age", user.Age).
		Set("country", user.Country).
		Where(sq.Eq{"id": user.Guid})

	sql, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL:", zap.Error(err))

		return err
	}

	// --- Выполняем запрос ---

	_, err = r.pg.Exec(ctx, sql, args...)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Update failed:", zap.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, user *domain.User) error {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Delete("users").
		Where(sq.Eq{"id": user.Guid})

	sql, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL:", zap.Error(err))

		return err
	}

	_, err = r.pg.Exec(ctx, sql, args...)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Delete failed:", zap.Error(err))

		return err
	}

	return nil
}

func (r *Repository) CreateBuisness(ctx context.Context, business *domain.Business) error {
	query := sq.Insert("buisness").
		Columns("id", "name").
		Values(business.Guid, business.Name).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL:", zap.Error(err))

		return err
	}

	// --- Выполняем запрос ---

	_, err = r.pg.Exec(ctx, sql, args...)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "INSERT failed:", zap.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetBuisness(ctx context.Context, business *domain.Business) (string, error) {
	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("id", "name").
		From("buisness").
		Where(sq.Eq{"id": business.Guid})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL", zap.Error(err))

		return "", err
	}

	// Структура для результата
	var bis domain.Business

	// Выполняем запрос
	err = r.pg.QueryRow(ctx, sqlStr, args...).Scan(
		&bis.Guid,
		&bis.Name,
	)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to execute SELECT", zap.Error(err))

		return "", err
	}

	return bis.Name, nil
}

func (r *Repository) GetUser(ctx context.Context, userId string) (*domain.User, error) {

	query := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("name", "surname", "avatar_url", "age", "country").
		From("users").
		Where(sq.Eq{"id": userId})

	sqlStr, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to build SQL", zap.Error(err))

		return nil, err
	}

	// Структура для результата
	var userResult domain.User

	err = r.pg.QueryRow(ctx, sqlStr, args...).Scan(
		&userResult.Name,
		&userResult.Surname,
		&userResult.Avatar_url,
		&userResult.Age,
		&userResult.Country,
	)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed to execute SELECT", zap.Error(err))

		return nil, err
	}

	return &userResult, nil

}
