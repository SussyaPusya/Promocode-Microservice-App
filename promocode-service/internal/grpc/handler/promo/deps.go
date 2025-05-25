//go:generate mockgen -source deps.go -package $GOPACKAGE -typed -destination mock_deps_test.go
package promo

import (
	"context"
	"time"

	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
)

type promoService interface {
	Create(ctx context.Context, promoDto *promo.CreatePromoDTO) (id string, err error)
	List(ctx context.Context, companyId string, countries []string, limit int, offset int, sortBy promoenum.SortBy) (promoDTOs []promo.DTO, err error)
	Count(ctx context.Context, companyId string, countries []string) (count int, err error)
	GetById(ctx context.Context, promoId string, companyId string) (promoDTO *promo.DTO, err error)
	Update(
		ctx context.Context,
		promoId string,
		companyId string,
		description string,
		imageUrl string,
		targetAgeFrom int64,
		targetAgeUntil int64,
		targetCountry string,
		targetCategories []string,
		activeFrom time.Time,
		activeUntil time.Time) error
	Delete(ctx context.Context, promoId string, companyId string) error
	Activate(ctx context.Context, promoId string) (code string, err error)
}
