package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type mode string

const (
	RefreshTokenMode       mode = "refresh"
	AccessTokenMode        mode = "access"
	RefreshTokenCookieName      = "refresh-token"
	AccessTokenCookieName       = "access-token"
)

func (j *ServiceJWT) GetClaims(id string, tokenMode mode) *jwt.RegisteredClaims {
	var expiration *jwt.NumericDate

	if tokenMode == RefreshTokenMode {
		expiration = jwt.NewNumericDate(time.Now().Add(j.RefreshTimeExp))
	} else if tokenMode == AccessTokenMode {
		expiration = jwt.NewNumericDate(time.Now().Add(j.AccessTimeExp))
	} else {
		panic("invalid type")
	}

	return &jwt.RegisteredClaims{
		Subject:   id,
		ExpiresAt: expiration,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
}


