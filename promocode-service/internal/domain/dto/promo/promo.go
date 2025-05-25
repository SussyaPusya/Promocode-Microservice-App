package promo

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
)

type Code struct {
	Code        string `validate:"required,min=3,max=30"`
	Activations int64  `validate:"min=0"`
	MaxCount    int64  `validate:"min=0"`
}

type DTO struct {
	PromoId          string      `validate:"required,uuid4"`
	CompanyId        string      `validate:"required,uuid4"`
	CompanyName      string      `validate:"required,min=5,max=50"`
	Mode             Mode        `validate:"required,oneof=COMMON UNIQUE"`
	Description      string      `validate:"required,min=10,max=300"`
	ImageURL         string      `validate:"omitempty,url,max=350"`
	Target           *target.DTO `validate:"required"`
	ActiveFrom       time.Time   `validate:"omitempty"`
	ActiveUntil      time.Time   `validate:"omitempty,gtfield=ActiveFrom"`
	Codes            []Code      `validate:"required,dive"`
	ActivationsCount int64       `validate:"min=0"`
}

func (dto *DTO) Validate() error {
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

	if err := dto.validateDates(); err != nil {
		return err
	}

	if dto.Target != nil {
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
	}

	if err := dto.validateCodes(); err != nil {
		return err
	}

	return nil
}

func (dto *DTO) validateDates() error {
	if !dto.ActiveFrom.IsZero() && !dto.ActiveUntil.IsZero() && dto.ActiveUntil.Before(dto.ActiveFrom) {
		return domainerrors.ValidationError{
			Field:   "active_until",
			Message: "must be after active_from",
		}
	}
	return nil
}

func (dto *DTO) validateCodes() error {
	if len(dto.Codes) == 0 {
		return domainerrors.ValidationError{
			Field:   "codes",
			Message: "at least one code is required",
		}
	}

	for i, code := range dto.Codes {
		if code.MaxCount > 0 && code.Activations > code.MaxCount {
			return domainerrors.ValidationError{
				Field:   fmt.Sprintf("codes[%d].activations", i),
				Message: "cannot exceed max_count",
			}
		}
	}

	return nil
}

// validationMessage возвращает понятное сообщение об ошибке для каждого тега валидации
func validationMessage(tag string, param string) string {
	switch tag {
	case "required":
		return "is required"
	case "uuid4":
		return "must be a valid UUID v4"
	case "oneof":
		return fmt.Sprintf("must be one of %s", param)
	case "min":
		return fmt.Sprintf("must be at least %s", param)
	case "max":
		return fmt.Sprintf("must be at most %s", param)
	case "url":
		return "must be a valid URL"
	case "gtfield":
		return fmt.Sprintf("must be greater than %s", param)
	case "dive":
		return "contains invalid items"
	default:
		return "failed validation"
	}
}
