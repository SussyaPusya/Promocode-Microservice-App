package promo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"
)

func validDTO() *DTO {
	return &DTO{
		PromoId:     "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		CompanyId:   "b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		CompanyName: "Valid Company Name Ltd.",
		Mode:        COMMON,
		Description: "This is a valid promotion description with more than 10 characters",
		ImageURL:    "https://example.com/image.jpg",
		Target: &target.DTO{
			AgeFrom:    0,
			AgeUntil:   100,
			Country:    "RU",
			Categories: nil,
		},
		Codes: []Code{
			{Code: "VALIDCODE123", Activations: 0, MaxCount: 100},
		},
		ActiveFrom:       time.Now().UTC(),
		ActiveUntil:      time.Now().Add(24 * time.Hour).UTC(),
		ActivationsCount: 0,
	}
}

func TestDTO_Validate_Success(t *testing.T) {
	dto := validDTO()
	err := dto.Validate()
	assert.NoError(t, err)
}

func TestDTO_Validate_InvalidPromoID(t *testing.T) {
	dto := validDTO()
	dto.PromoId = "invalid-uuid"
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Equal(t, "PromoId", verr.Field)
	assert.Contains(t, verr.Message, "UUID v4")
}

func TestDTO_Validate_ShortCompanyName(t *testing.T) {
	dto := validDTO()
	dto.CompanyName = "abc"
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "at least 5")
}

func TestDTO_Validate_InvalidMode(t *testing.T) {
	dto := validDTO()
	dto.Mode = "INVALID_MODE"
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "one of COMMON UNIQUE")
}

func TestDTO_Validate_ActiveUntilBeforeActiveFrom(t *testing.T) {
	dto := validDTO()
	dto.ActiveFrom = time.Now()
	dto.ActiveUntil = dto.ActiveFrom.Add(-1 * time.Hour)
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
}

func TestDTO_Validate_EmptyCodes(t *testing.T) {
	dto := validDTO()
	dto.Codes = []Code{}
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "at least one code")
}

func TestDTO_Validate_CodeActivationsExceedMax(t *testing.T) {
	dto := validDTO()
	dto.Codes = []Code{
		{Code: "TESTCODE", Activations: 10, MaxCount: 5},
	}
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "exceed max_count")
}

func TestDTO_Validate_InvalidImageURL(t *testing.T) {
	dto := validDTO()
	dto.ImageURL = "not-a-valid-url"
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "valid URL")
}

func TestDTO_Validate_MissingTarget(t *testing.T) {
	dto := validDTO()
	dto.Target = nil
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "required")
}

func TestDTO_Validate_ShortCode(t *testing.T) {
	dto := validDTO()
	dto.Codes = []Code{
		{Code: "AB", Activations: 0, MaxCount: 10},
	}
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.Message, "at least 3")
}

func TestDTO_Validate_InvalidTarget(t *testing.T) {
	dto := validDTO()
	dto.Target = &target.DTO{} // Assume this creates an invalid target
	err := dto.Validate()

	var verr domainerrors.ValidationError
	assert.ErrorAs(t, err, &verr)
}
