package adaptergrpc

import (
	"errors"

	promodto "gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"

	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
)

var InvalidPromoMode = errors.New("unknown promo mode")

func MapPbSortByToDomain(s promopb.PromoSortBy) promoenum.SortBy {
	switch s {
	case promopb.PromoSortBy_ACTIVE_FROM:
		return promoenum.SortByActiveFrom
	case promopb.PromoSortBy_ACTIVE_UNTIL:
		return promoenum.SortByActiveUntil
	default:
		return promoenum.SortByCreatedAt // или promo.SortByCreatedAt по умолчанию, если хочешь fallback
	}
}

func MapDomainModeToPb(m promodto.Mode) (promopb.Mode, error) {
	switch m {
	case promodto.COMMON:
		return promopb.Mode_COMMON, nil
	case promodto.UNIQUE:
		return promopb.Mode_UNIQUE, nil
	default:
		return 0, InvalidPromoMode
	}

}

func MapPbModeToDomain(m promopb.Mode) (promodto.Mode, error) {
	switch m {
	case promopb.Mode_COMMON:
		return promodto.COMMON, nil
	case promopb.Mode_UNIQUE:
		return promodto.UNIQUE, nil
	default:
		return "", InvalidPromoMode
	}
}
