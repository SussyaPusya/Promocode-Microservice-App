package jwt_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtlib "gitlab.com/pisya-dev/account-service/pkg/jwt" // замените на путь к вашему пакету
)

type mockJWTService struct {
	AccessTimeExp  time.Duration
	RefreshTimeExp time.Duration
}

// Явно приводим mockJWTService к типу *jwtlib.ServiceJWT.
// Это работает, если у вас структура ServiceJWT имеет такие же поля.
func (s *mockJWTService) asServiceJWT() *jwtlib.ServiceJWT {
	return &jwtlib.ServiceJWT{
		AccessTimeExp:  s.AccessTimeExp,
		RefreshTimeExp: s.RefreshTimeExp,
	}
}

func TestGetClaims_AccessToken(t *testing.T) {
	service := &mockJWTService{
		AccessTimeExp:  time.Minute * 15,
		RefreshTimeExp: time.Hour * 24,
	}

	id := "user123"
	claims := service.asServiceJWT().GetClaims(id, jwtlib.AccessTokenMode)

	require.NotNil(t, claims)
	assert.Equal(t, id, claims.Subject)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	now := time.Now()
	assert.WithinDuration(t, now.Add(15*time.Minute), claims.ExpiresAt.Time, time.Second*2)
	assert.WithinDuration(t, now, claims.IssuedAt.Time, time.Second*2)
}

func TestGetClaims_RefreshToken(t *testing.T) {
	service := &mockJWTService{
		AccessTimeExp:  time.Minute * 15,
		RefreshTimeExp: time.Hour * 24,
	}

	id := "user123"
	claims := service.asServiceJWT().GetClaims(id, jwtlib.RefreshTokenMode)

	require.NotNil(t, claims)
	assert.Equal(t, id, claims.Subject)
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)

	now := time.Now()
	assert.WithinDuration(t, now.Add(24*time.Hour), claims.ExpiresAt.Time, time.Second*2)
	assert.WithinDuration(t, now, claims.IssuedAt.Time, time.Second*2)
}

func TestGetClaims_InvalidMode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid mode, but did not panic")
		}
	}()

	service := &mockJWTService{
		AccessTimeExp:  time.Minute * 15,
		RefreshTimeExp: time.Hour * 24,
	}

	_ = service.asServiceJWT().GetClaims("user123", "invalid")
}
