package model

import (
	"time"

	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
)

type Promo struct {
	Id               string     `db:"id"`
	CompanyId        string     `db:"company_id"`
	Description      string     `db:"description"`
	ImageUrl         string     `db:"image_url"`
	ActiveFrom       time.Time  `db:"active_from"`
	ActiveUntil      time.Time  `db:"active_until"`
	CreatedAt        time.Time  `db:"created_at"`
	Mode             promo.Mode `db:"mode"`
	TargetAgeFrom    int        `db:"target_age_from"`
	TargetAgeUntil   int        `db:"target_age_until"`
	TargetCountry    string     `db:"target_country"`
	TargetCategories []string   `db:"target_categories"`
	PromoCommon      string     `db:"promo_common"`
	PromoUnique      []string   `db:"promo_unique"`
	MaxCount         int64      `db:"max_count"`
}
