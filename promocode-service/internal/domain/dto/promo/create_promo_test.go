package promo

import (
	"testing"
	"time"

	"gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/target"
	domainerrors "gitlab.com/pisya-dev/promo-code-service/internal/domain/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreatePromoDTO_Validate(t *testing.T) {
	validUUID := uuid.New().String()
	validTime := time.Now().Add(24 * time.Hour)
	validTarget := target.DTO{} // assume valid target implementation

	tests := []struct {
		name        string
		dto         CreatePromoDTO
		wantErr     bool
		expectedErr domainerrors.ValidationError
	}{
		{
			name: "valid common promo",
			dto: CreatePromoDTO{
				CompanyId:   validUUID,
				Mode:        COMMON,
				PromoCommon: "SUMMER25",
				Description: "Valid summer promotion",
				Target:      validTarget,
				MaxCount:    1000,
				ActiveFrom:  time.Now(),
				ActiveUntil: validTime,
			},
			wantErr: false,
		},
		{
			name: "invalid company ID format",
			dto: CreatePromoDTO{
				CompanyId:   "invalid-uuid",
				Mode:        COMMON,
				PromoCommon: "SUMMER25",
				Description: "Test description",
				Target:      validTarget,
				MaxCount:    100,
			},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "CompanyId",
				Message: "invalid UUID format",
			},
		},
		{
			name:    "missing required fields",
			dto:     CreatePromoDTO{},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "CompanyId",
				Message: "required",
			},
		},
		{
			name: "invalid mode type",
			dto: CreatePromoDTO{
				CompanyId:   validUUID,
				Mode:        "INVALID_MODE",
				PromoCommon: "SUMMER25",
				Description: "Test description",
				Target:      validTarget,
				MaxCount:    100,
			},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "Mode",
				Message: "must be one of COMMON UNIQUE",
			},
		},
		{
			name: "common mode without promo code",
			dto: CreatePromoDTO{
				CompanyId:   validUUID,
				Mode:        COMMON,
				Description: "Test description",
				Target:      validTarget,
				MaxCount:    100,
			},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "PromoCommon",
				Message: "promo code is required for COMMON mode",
			},
		},
		{
			name: "invalid dates order",
			dto: CreatePromoDTO{
				CompanyId:   validUUID,
				Mode:        COMMON,
				PromoCommon: "SUMMER25",
				Description: "Test description",
				Target:      validTarget,
				MaxCount:    100,
				ActiveFrom:  validTime,
				ActiveUntil: time.Now(),
			},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "ActiveUntil",
				Message: "must be after active_from",
			},
		},
		{
			name: "invalid target",
			dto: CreatePromoDTO{
				CompanyId:   validUUID,
				Mode:        COMMON,
				PromoCommon: "SUMMER25",
				Description: "Test description",
				Target:      target.DTO{}, // assume invalid target
				MaxCount:    100,
			},
			wantErr: true,
			expectedErr: domainerrors.ValidationError{
				Field:   "target",
				Message: "target validation error", // adjust based on actual target validation
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// Helper function to create valid target (implement according to your target validation logic)
func createValidTarget() target.DTO {
	return target.DTO{
		// populate with valid target data
	}
}
