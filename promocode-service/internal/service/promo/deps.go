//go:generate mockgen -source deps.go -package $GOPACKAGE -typed -destination mock_deps_test.go
package promo

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
	"gitlab.com/pisya-dev/promo-code-service/internal/storage/model"
	promoStorage "gitlab.com/pisya-dev/promo-code-service/internal/storage/promo"
)

type promoRepository interface {
	Create(ctx context.Context, promoModel *model.Promo) (id string, err error)
	List(ctx context.Context, companyId string, countries []string, offset int, sortBy promo.SortBy, limit int) (promoModels []promoStorage.PromoDetails, err error)
	Count(ctx context.Context, companyId string, countries []string) (count int, err error)
	GetById(ctx context.Context, promoId string) (promoModel *promoStorage.PromoDetails, err error)
	Update(
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
	) error
	Delete(ctx context.Context, promoId string) error
}

type promoCodeRepository interface {
	Create(ctx context.Context, promoCodeModel *model.PromoCode) (id string, err error)
	Activate(ctx context.Context, promoId string) (code string, err error)
}

type accountServiceClient interface {
	GetCompanyNameByCompanyID(ctx context.Context, companyID string) (companyName string, err error)
}

type redisDb interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}
