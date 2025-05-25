package promo

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
)

type CreatePromoDTO struct {
	CompanyId   string     `validate:"required,uuid4"`
	Mode        Mode       `validate:"required,oneof=COMMON UNIQUE"`
	PromoCommon string     `validate:"required_if=Mode COMMON,omitempty,min=5,max=30"`
	PromoUnique []string   `validate:"required_if=Mode UNIQUE,omitempty,dive,min=3,max=30"`
	Description string     `validate:"required,min=10,max=300"`
	ImageUrl    string     `validate:"omitempty,url,max=350"`
	Target      target.DTO `validate:"required"`
	MaxCount    int64      `validate:"required"`
	ActiveFrom  time.Time  `validate:"omitempty"`
	ActiveUntil time.Time  `validate:"omitempty,gtfield=ActiveFrom"`
}

func (dto *CreatePromoDTO) Validate() error {
	if err := validator.New().Struct(dto); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fe := ve[0]
			return domainerrors.ValidationError{
				Field:   fe.Field(),
				Message: validationMessage(fe.Tag(), fe.Param()),
			}
		}
		return domainerrors.ValidationError{
			Field:   "",
			Message: "invalid request data",
		}
	}

	// Дополнительные проверки
	if err := dto.validateMode(); err != nil {
		return err
	}

	if err := dto.validateDates(); err != nil {
		return err
	}

	// Валидация Target
	if err := dto.Target.Validate(); err != nil {
		var verr domainerrors.ValidationError
		if errors.As(err, &verr) {
			return verr
		}
		return domainerrors.ValidationError{
			Field:   "target",
			Message: err.Error(),
		}
	}

	return nil
}

func (dto *CreatePromoDTO) validateMode() error {
	if dto.Mode == COMMON && dto.PromoCommon == "" {
		return domainerrors.ValidationError{
			Field:   "promo_common",
			Message: "promo code is required for COMMON mode",
		}
	}

	if dto.Mode == UNIQUE && len(dto.PromoUnique) == 0 {
		return domainerrors.ValidationError{
			Field:   "promo_unique",
			Message: "at least one promo code is required for UNIQUE mode",
		}
	}

	return nil
}

func (dto *CreatePromoDTO) validateDates() error {
	if !dto.ActiveFrom.IsZero() && !dto.ActiveUntil.IsZero() && dto.ActiveUntil.Before(dto.ActiveFrom) {
		return domainerrors.ValidationError{
			Field:   "active_until",
			Message: "must be after active_from",
		}
	}
	return nil
}
