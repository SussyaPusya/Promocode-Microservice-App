package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/pisya-dev/auth-service/internal/dto"
	"gitlab.com/pisya-dev/auth-service/pkg/logger"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Repository struct {
	pg *pgxpool.Pool
}

func NewRepository(pg *pgxpool.Pool) *Repository {

	return &Repository{pg: pg}

}

func (r *Repository) CreateAccount(ctx context.Context, req *dto.AccountReqs, id string) error {

	const op = "repository.CreateAccount"

	query := sq.Insert("platform_user").
		Columns("id", "email", "password").
		Values(id, req.Email, req.Password).
		PlaceholderFormat(sq.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%sFailed to build SQL:", op), zap.Error(err))

		return err
	}

	// --- Выполняем запрос ---

	_, err = r.pg.Exec(ctx, sql, args...)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s :INSERT failed:", op), zap.Error(err))

		return err
	}

	return nil

}

func (r *Repository) GetAccount(ctx context.Context, req *dto.AuthWithAccountReq) (string, error) {

	const op = "repository.GetAccount"

	query := sq.Select("id", "email", "password").
		From("platform_user").
		Where(sq.Eq{"email": req.Email}, sq.Eq{"password": req.Password}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : Failed to build SQL", op), zap.Error(err))

		return "", err
	}

	var parseId dto.AuthWithAccountReq

	err = r.pg.QueryRow(ctx, sqlStr, args...).Scan(
		&parseId.ID,
		&parseId.Email,
		&parseId.Password,
	)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : Failed to execute SELECT:", op), zap.Error(err))

		return "", err
	}

	return parseId.ID, nil

}

func (r *Repository) GetProfile(ctx context.Context, id string) (*dto.AccountReqs, error) {
	const op = "repository.GetProfile"

	query := sq.Select("email").
		From("platform_user").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : Failed to build SQL", op), zap.Error(err))

		return nil, err
	}

	var user dto.AccountReqs
	err = r.pg.QueryRow(ctx, sqlStr, args...).Scan(
		&user.Email,
	)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("%s : Failed to execute SELECT:", op), zap.Error(err))

		return nil, err
	}

	return &user, nil

}
