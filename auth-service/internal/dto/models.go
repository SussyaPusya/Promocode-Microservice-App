package dto

import "time"

type AccountReqs struct {
	Name string `json:"name"`

	Surname    string `json:"surname"`
	Avatar_url string `json:"avatar_url"`

	Age     int32  `json:"age"`
	Country string `json:"country"`

	Email string `json:"email"`

	Password string `json:"password"`
}

const (
	AccesTimeExpr   = 5 * time.Minute
	RefreshTimeExpr = 24 * time.Hour
	RedisTTL        = 15 * time.Minute
)

type AuthWithAccountReq struct {
	ID    string
	Email string `json:"email"`

	Password string `json:"password"`
}

type GetProfileID struct {
	ID string
}

type PromoReq struct {
	Description string `json:"description"`

	ImageUrl string `json:"image_url"`

	Target struct {
		Age_from   int64    `json:"age_from"`
		Age_until  int64    `json:"age_until"`
		Country    string   `json:"country"`
		Categories []string `json:"categories"`
	} `json:"target"`

	MaxCount int64 `json:"max_count"`

	Active_from  time.Time `json:"active_from"`
	Active_until time.Time `json:"active_until"`
	Mode         string    `json:"mode"`

	Promo_common string   `json:"promo_common"`
	Promo_unique []string `json:"promo_unique"`
}
