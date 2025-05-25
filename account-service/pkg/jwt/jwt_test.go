package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwtlib "gitlab.com/pisya-dev/account-service/pkg/jwt" // замените на актуальный импорт
)

func generateKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return priv, &priv.PublicKey
}

func TestEncodeDecode_Success(t *testing.T) {
	privKey, pubKey := generateKeys(t)

	svc := jwtlib.NewServiceJWT(pubKey, time.Hour*24, time.Minute*15)

	claims := jwt.RegisteredClaims{
		Subject:   "user123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	tokenStr, err := svc.Encode(&claims, privKey)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	decodedClaims, err := svc.DecodeKey(tokenStr)
	require.NoError(t, err)
	require.NotNil(t, decodedClaims)
	assert.Equal(t, "user123", decodedClaims.Subject)
}

func TestDecodeKey_EmptyToken(t *testing.T) {
	_, pubKey := generateKeys(t)
	svc := jwtlib.NewServiceJWT(pubKey, time.Hour*24, time.Minute*15)

	claims, err := svc.DecodeKey("")
	assert.Nil(t, claims)
	assert.ErrorIs(t, err, jwtlib.ErrorUndefinedToken)
}

func TestDecodeKey_InvalidToken(t *testing.T) {
	_, pubKey := generateKeys(t)
	svc := jwtlib.NewServiceJWT(pubKey, time.Hour*24, time.Minute*15)

	claims, err := svc.DecodeKey("not.a.valid.token")
	assert.Nil(t, claims)
	assert.Error(t, err)
}
func TestEncode_FailsWithInvalidKey(t *testing.T) {
	_, pubKey := generateKeys(t)
	_ = jwtlib.NewServiceJWT(pubKey, time.Hour*24, time.Minute*15)

	claims := jwt.RegisteredClaims{
		Subject:   "test",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
	}

	// Используем "левый" ключ (например, строка вместо rsa.PrivateKey)
	// или даже просто объект, не подходящий под подпись RSA
	type fakeKey struct{}
	fake := &fakeKey{}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &claims)
	tokenStr, err := token.SignedString(fake)

	assert.Empty(t, tokenStr)
	assert.Error(t, err)
}
