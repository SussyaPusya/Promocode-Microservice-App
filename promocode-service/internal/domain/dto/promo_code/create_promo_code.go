package promo_code

type CreatePromoCodeDTO struct {
	PromoId     string `validate:"required,uuid4"`
	Code        string `validate:"required,min=3,max=30"`
	IsActivated bool
	MaxCount    int64 `validate:"required,min=0,max=100000000"`
}
