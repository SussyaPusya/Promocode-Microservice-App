package model

type PromoCode struct {
	Id          string `db:"id"`
	PromoId     string `db:"promo_id"`
	Code        string `db:"code"`
	Activations int64  `db:"activations"`
	MaxCount    int64  `db:"max_count"`
}
