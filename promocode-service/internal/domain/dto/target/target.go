package target

import "github.com/go-playground/validator/v10"

type DTO struct {
	AgeFrom    int64    `validate:"min=0,max=100"`
	AgeUntil   int64    `validate:"min=0,max=100,gtfield=AgeFrom"`
	Country    string   `validate:"required,iso3166_1_alpha2"`
	Categories []string `validate:"dive,min=0,max=20"`
}

func (d *DTO) Validate() error {
	validate := validator.New()
	return validate.Struct(d)
}
