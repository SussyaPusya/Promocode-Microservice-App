package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ErrorInvalidTokenString = "token is invalid"
)

var (
	ErrorUndefinedToken = fmt.Errorf("undefined token")
	ErrorInvalidToken   = fmt.Errorf(ErrorInvalidTokenString)
)

type ServiceJWT struct {
	publicKey      *rsa.PublicKey
	RefreshTimeExp time.Duration
	AccessTimeExp  time.Duration
}

func NewServiceJWT(publicKey *rsa.PublicKey,
	refreshTimeExp time.Duration, accessTimeExp time.Duration) *ServiceJWT {
	return &ServiceJWT{
		publicKey:      publicKey,
		RefreshTimeExp: refreshTimeExp,
		AccessTimeExp:  accessTimeExp,
	}
}

func (j *ServiceJWT) DecodeKey(tokenString string) (*jwt.RegisteredClaims, error) {
	if tokenString == "" {
		return nil, ErrorUndefinedToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return j.publicKey, nil
		})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, ErrorInvalidToken
}

func (j *ServiceJWT) Encode(claims jwt.Claims, privetToken *rsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privetToken)
	if err != nil {
		return "", fmt.Errorf("failed create token: %v", err)
	}

	return tokenString, nil
}
