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
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	RefreshTimeExp time.Duration
	AccessTimeExp  time.Duration
}

func NewServiceJWT(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey,
	refreshTimeExp time.Duration, accessTimeExp time.Duration) *ServiceJWT {
	return &ServiceJWT{
		privateKey:     privateKey,
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

func (j *ServiceJWT) Encode(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(j.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed create token: %v", err)
	}

	return tokenString, nil
}
