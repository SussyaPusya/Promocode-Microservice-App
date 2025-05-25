package adaptergrpc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	adaptergrpc "gitlab.com/pisya-dev/promo-code-service/internal/adapter/grpc"
	promodto "gitlab.com/pisya-dev/promo-code-service/internal/domain/dto/promo"
	promoenum "gitlab.com/pisya-dev/promo-code-service/internal/domain/enum/promo"
	promopb "gitlab.com/pisya-dev/promo-code-service/pkg/api/pb"
)

func TestMapPbSortByToDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    promopb.PromoSortBy
		expected promoenum.SortBy
	}{
		{
			name:     "ActiveFrom",
			input:    promopb.PromoSortBy_ACTIVE_FROM,
			expected: promoenum.SortByActiveFrom,
		},
		{
			name:     "ActiveUntil",
			input:    promopb.PromoSortBy_ACTIVE_UNTIL,
			expected: promoenum.SortByActiveUntil,
		},
		{
			name:     "UnknownValue",
			input:    promopb.PromoSortBy(999),
			expected: promoenum.SortByCreatedAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adaptergrpc.MapPbSortByToDomain(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapDomainModeToPb(t *testing.T) {
	tests := []struct {
		name        string
		input       promodto.Mode
		expectedPb  promopb.Mode
		expectedErr error
	}{
		{
			name:        "CommonMode",
			input:       promodto.COMMON,
			expectedPb:  promopb.Mode_COMMON,
			expectedErr: nil,
		},
		{
			name:        "UniqueMode",
			input:       promodto.UNIQUE,
			expectedPb:  promopb.Mode_UNIQUE,
			expectedErr: nil,
		},
		{
			name:        "InvalidMode",
			input:       promodto.Mode("invalid"),
			expectedPb:  0,
			expectedErr: adaptergrpc.InvalidPromoMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adaptergrpc.MapDomainModeToPb(tt.input)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expectedPb, result)
		})
	}
}

func TestMapPbModeToDomain(t *testing.T) {
	tests := []struct {
		name        string
		input       promopb.Mode
		expected    promodto.Mode
		expectedErr error
	}{
		{
			name:        "CommonMode",
			input:       promopb.Mode_COMMON,
			expected:    promodto.COMMON,
			expectedErr: nil,
		},
		{
			name:        "UniqueMode",
			input:       promopb.Mode_UNIQUE,
			expected:    promodto.UNIQUE,
			expectedErr: nil,
		},
		{
			name:        "InvalidMode",
			input:       promopb.Mode(999),
			expected:    "",
			expectedErr: adaptergrpc.InvalidPromoMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adaptergrpc.MapPbModeToDomain(tt.input)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
