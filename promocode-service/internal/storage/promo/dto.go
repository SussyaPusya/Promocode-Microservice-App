package promo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
)

type CodeDTO struct {
	Code        string `json:"code"`
	Activations int64  `json:"activations"`
	MaxCount    int64  `json:"max_count"`
}

type CodeDTOs []CodeDTO

func (c *CodeDTOs) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to []byte", src)
	}
	return json.Unmarshal(bytes, c)
}

type PromoDetails struct {
	Id               string         `db:"id"`
	CompanyId        string         `db:"company_id"`
	Description      string         `db:"description"`
	ImageUrl         string         `db:"image_url"`
	ActiveFrom       time.Time      `db:"active_from"`
	ActiveUntil      time.Time      `db:"active_until"`
	CreatedAt        time.Time      `db:"created_at"`
	Mode             promo.Mode     `db:"mode"`
	TargetAgeFrom    int            `db:"target_age_from"`
	TargetAgeUntil   int            `db:"target_age_until"`
	TargetCountry    string         `db:"target_country"`
	TargetCategories pq.StringArray `db:"target_categories"`
	Codes            CodeDTOs       `db:"codes"`
}
