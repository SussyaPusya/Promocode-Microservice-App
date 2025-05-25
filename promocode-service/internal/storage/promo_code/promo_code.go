package promo_code

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/model"
)

type Repository struct {
	db *sqlx.DB
}

var (
	ErrNotFound      = errors.New("not found")
	ErrNoActivations = errors.New("no activations")
)

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, promoCodeModel *model.PromoCode) (id string, err error) {

	const op = "storage.promo_code.Create"

	query := `
		INSERT INTO promo_code(id, promo_id, code, activations, max_count) 
		VALUES (:id, :promo_id, :code, :activations, :max_count) 
		RETURNING id
		`

	stmt, err := r.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRowxContext(
		ctx, promoCodeModel,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil

}

func (r *Repository) Activate(ctx context.Context, promoId string) (code string, err error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	query := `
		SELECT id, code 
		FROM promo_code 
		WHERE promo_id = :promo_id AND activations < max_count 
		LIMIT 1 
		FOR UPDATE
	`

	params := map[string]interface{}{
		"promo_id": promoId,
	}

	namedQuery, args, err := sqlx.Named(query, params)
	if err != nil {
		return "", fmt.Errorf("named query prepare: %w", err)
	}

	namedQuery = tx.Rebind(namedQuery)

	var promoCodeId string

	row := tx.QueryRowxContext(ctx, namedQuery, args...)
	if err = row.Scan(&promoCodeId, &code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoActivations
		}
		return "", fmt.Errorf("scan code: %w", err)
	}

	updateQuery := `
		UPDATE promo_code 
		SET activations = activations + 1 
		WHERE code = :code
	`

	_, err = tx.NamedExecContext(ctx, updateQuery, map[string]interface{}{
		"code": code,
	})
	if err != nil {
		return "", fmt.Errorf("update activations: %w", err)
	}

	return code, nil
}
