package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/pisya-dev/account-service/pkg/jwt" // замените на реальный импорт
)

func writePEMFile(t *testing.T, keyBytes []byte, filename string, pemType string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), filename)
	file, err := os.Create(tmpFile)
	require.NoError(t, err)

	block := &pem.Block{
		Type:  pemType,
		Bytes: keyBytes,
	}
	err = pem.Encode(file, block)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	return tmpFile
}

func TestLoadPublicKey_Success(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubASN1, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	pubFile := writePEMFile(t, pubASN1, "public.pem", "PUBLIC KEY")

	loadedKey, err := jwt.LoadPublicKey(pubFile)
	require.NoError(t, err)
	require.Equal(t, &privateKey.PublicKey, loadedKey)
}

func TestLoadPrivateKey_PKCS1_Success(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs1 := x509.MarshalPKCS1PrivateKey(privateKey)
	privFile := writePEMFile(t, pkcs1, "private_pkcs1.pem", "RSA PRIVATE KEY")

	loadedKey, err := jwt.LoadPrivateKey(privFile)
	require.NoError(t, err)
	require.Equal(t, privateKey, loadedKey)
}

func TestLoadPrivateKey_PKCS8_Success(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pkcs8, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	privFile := writePEMFile(t, pkcs8, "private_pkcs8.pem", "PRIVATE KEY")

	loadedKey, err := jwt.LoadPrivateKey(privFile)
	require.NoError(t, err)
	require.Equal(t, privateKey, loadedKey)
}

func TestLoadPublicKey_InvalidPath(t *testing.T) {
	_, err := jwt.LoadPublicKey("/non/existent/path.pem")
	require.Error(t, err)
}

func TestLoadPrivateKey_InvalidPath(t *testing.T) {
	_, err := jwt.LoadPrivateKey("/non/existent/path.pem")
	require.Error(t, err)
}

func TestLoadPublicKey_InvalidPEM(t *testing.T) {
	file := filepath.Join(t.TempDir(), "invalid.pem")
	require.NoError(t, os.WriteFile(file, []byte("invalid pem"), 0600))

	_, err := jwt.LoadPublicKey(file)
	require.Error(t, err)
}

func TestLoadPrivateKey_InvalidPEM(t *testing.T) {
	file := filepath.Join(t.TempDir(), "invalid.pem")
	require.NoError(t, os.WriteFile(file, []byte("invalid pem"), 0600))

	_, err := jwt.LoadPrivateKey(file)
	require.Error(t, err)
}

func TestLoadPrivateKey_NotRSA(t *testing.T) {
	dummyBytes := []byte("not a real key")
	file := writePEMFile(t, dummyBytes, "notrsa.pem", "PRIVATE KEY")

	_, err := jwt.LoadPrivateKey(file)
	require.Error(t, err)
}

func TestLoadPublicKey_NotRSA(t *testing.T) {
	dummyBytes := []byte("not a real key")
	file := writePEMFile(t, dummyBytes, "notrsa_pub.pem", "PUBLIC KEY")

	_, err := jwt.LoadPublicKey(file)
	require.Error(t, err)
}
