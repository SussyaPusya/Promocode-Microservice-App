package promo

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/model"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, promoModel *model.Promo) (id string, err error) {

	const op = "storage.promo.Create"

	query := `
		INSERT INTO promo(
			id, company_id, description, image_url, active_from, active_until,
			created_at, mode, target_age_from, target_age_until,
			target_country, target_categories
		) VALUES (
			:id, :company_id, :description, :image_url, :active_from, :active_until,
			:created_at, :mode, :target_age_from, :target_age_until,
			:target_country, :target_categories
		)
		RETURNING id
	`

	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("%s: prepare failed: %w", op, err)
	}

	err = stmt.QueryRowxContext(ctx, promoModel).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *Repository) List(ctx context.Context, companyId string, countries []string, offset int, sortBy promoenum.SortBy, limit int) (promoModels []PromoDetails, err error) {

	query := `select 
				p.id,
				p.company_id,
				p.description,
				p.image_url,
				p.active_from,
				p.active_until,
				p.created_at,
				p.mode,
				p.target_age_from,
				p.target_age_until,
				p.target_country,
				p.target_categories,
				COALESCE(json_agg(json_build_object(
					'code', pc.code,
					'activations', pc.activations,
					'max_count', pc.max_count
				)) filter (where pc.id is not null), '[]') as codes
			from promo p
			left join promo_code pc on pc.promo_id = p.id 
			where p.company_id = :company_id`

	sqlParams := map[string]interface{}{
		"company_id": companyId,
		"offset":     offset,
		"limit":      limit,
	}

	sqlCountries := ""

	if len(countries) > 0 {
		var buffer bytes.Buffer
		for idx, country := range countries {
			buffer.WriteString(fmt.Sprintf(":country_%d, ", idx))
			sqlParams[fmt.Sprintf("country_%d", idx)] = strings.ToLower(country)
		}

		sqlCountries = buffer.String()[:buffer.Len()-2]

		query += fmt.Sprintf(" and (lower(p.target_country) in (%s) or p.target_country is null)", sqlCountries)
	}

	query += fmt.Sprintf(" group by p.id, p.%s order by p.%s desc offset :offset limit :limit", sortBy, sortBy)

	rows, err := r.db.NamedQueryContext(ctx, query, sqlParams)

	if err != nil {
		return nil, fmt.Errorf("storage.promo.List: %w", err)
	}

	for rows.Next() {
		var promo PromoDetails
		if err = rows.StructScan(&promo); err != nil {
			return nil, fmt.Errorf("storage.promo.List: %w", err)
		}
		promoModels = append(promoModels, promo)
	}

	return promoModels, nil
}

func (r *Repository) Count(ctx context.Context, companyId string, countries []string) (count int, err error) {
	query := `
	select count(1) from promo p where p.company_id = :company_id
	`

	sqlCountries := ""
	sqlParams := map[string]interface{}{
		"company_id": companyId,
	}

	if len(countries) > 0 {
		var buffer bytes.Buffer
		for idx, country := range countries {
			buffer.WriteString(fmt.Sprintf(":country_%d, ", idx))
			sqlParams[fmt.Sprintf("country_%d", idx)] = strings.ToLower(country)
		}

		sqlCountries = buffer.String()[:buffer.Len()-2]

		query += fmt.Sprintf(" and (lower(target_country) in (%s) or target_country is null)", sqlCountries)
	}

	stmt, err := r.db.PrepareNamedContext(ctx, query)

	if err != nil {
		return 0, fmt.Errorf("db.PrepareNamedContext: prepare failed: %w", err)
	}

	err = stmt.QueryRowxContext(ctx, sqlParams).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("stmt.QueryRowxContext: %w", err)
	}

	return count, nil

}

func (r *Repository) GetById(ctx context.Context, promoId string) (promoModel *PromoDetails, err error) {

	query := `select 
				p.id,
				p.company_id,
				p.description,
				p.image_url,
				p.active_from,
				p.active_until,
				p.created_at,
				p.mode,
				p.target_age_from,
				p.target_age_until,
				p.target_country,
				p.target_categories,
				COALESCE(json_agg(json_build_object(
					'code', pc.code,
					'activations', pc.activations,
					'max_count', pc.max_count
				)) filter (where pc.id is not null), '[]') as codes
			from promo p
			left join promo_code pc on pc.promo_id = p.id 
			where p.id = :promo_id
			group by p.id`

	sqlParams := map[string]interface{}{
		"promo_id": promoId,
	}

	rows, err := r.db.NamedQueryContext(ctx, query, sqlParams)

	if err != nil {
		return nil, fmt.Errorf("storage.promo.GetById: %w", err)
	}

	for rows.Next() {
		var promo PromoDetails
		if err = rows.StructScan(&promo); err != nil {
			return nil, fmt.Errorf("storage.promo.GetById: %w", err)
		}
		promoModel = &promo
	}

	return promoModel, nil

}

func (r *Repository) Update(
	ctx context.Context,
	promoId string,
	description string,
	imageUrl string,
	targetAgeFrom int64,
	targetAgeUntil int64,
	targetCountry string,
	targetCategories []string,
	activeFrom time.Time,
	activeUntil time.Time,
) error {

	query := `
		update promo set
			description = :description,
			image_url = :image_url,
			target_age_from = :target_age_from,
			target_age_until = :target_age_until,
			target_country = :target_country,
			target_categories = :target_categories,
			active_from = :active_from,
			active_until = :active_until
		where id = :promo_id
	`

	sqlParams := map[string]interface{}{
		"promo_id":          promoId,
		"description":       description,
		"image_url":         imageUrl,
		"target_age_from":   targetAgeFrom,
		"target_age_until":  targetAgeUntil,
		"target_country":    targetCountry,
		"target_categories": pq.Array(targetCategories),
		"active_from":       activeFrom,
		"active_until":      activeUntil,
	}

	_, err := r.db.NamedExecContext(ctx, query, sqlParams)

	if err != nil {
		return fmt.Errorf("r.db.NamedExecContext: %w", err)
	}

	return nil

}

func (r *Repository) Delete(ctx context.Context, promoId string) error {
	query := `delete from promo_code p where p.promo_id = :promo_id`

	sqlParams := map[string]interface{}{
		"promo_id": promoId,
	}

	_, err := r.db.NamedExecContext(ctx, query, sqlParams)

	if err != nil {
		return fmt.Errorf("r.db.NamedExecContext: %w", err)
	}

	query = `delete from promo where id = :promo_id`

	_, err = r.db.NamedExecContext(ctx, query, sqlParams)

	if err != nil {
		return fmt.Errorf("r.db.NamedExecContext: %w", err)
	}

	return nil

}
