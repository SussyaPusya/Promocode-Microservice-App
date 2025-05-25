package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	absolutePathError = "failed get absolute path: %v"
	fileNotFoundError = "file is not found: %v"
	readFileError     = "failed read file: %v"
	readPemBlockError = "failed to read pem block: %v"
	parseError        = "failed to parse, invalid algorithm: %v"
	rsaKeyError       = "is not rsa key, please check: %v"
)

func LoadPublicKey(filePath string) (*rsa.PublicKey, error) {
	fullPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf(absolutePathError, err)
	}

	if _, err := os.Stat(fullPath); err != nil {
		return nil, fmt.Errorf(fileNotFoundError, err)
	}

	fileData, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf(readFileError, err)
	}

	block, _ := pem.Decode(fileData)
	if block == nil {
		return nil, fmt.Errorf(readPemBlockError, err)
	}

	keyPub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf(parseError, err)
	}

	resKeyPub, ok := keyPub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf(rsaKeyError, err)
	}

	return resKeyPub, nil
}

func LoadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	fullFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf(absolutePathError, err)
	}

	if _, err := os.Stat(fullFilePath); err != nil {
		return nil, fmt.Errorf(fileNotFoundError, err)
	}

	fileData, err := os.ReadFile(fullFilePath)
	if err != nil {
		return nil, fmt.Errorf(readFileError, err)
	}

	block, _ := pem.Decode(fileData)
	if block == nil {
		return nil, fmt.Errorf(readPemBlockError, err)
	}

	keyPKCS1, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return keyPKCS1, nil
	}

	keyPKCS8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf(parseError, err)
	}

	key, ok := keyPKCS8.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf(rsaKeyError, err)
	}

	return key, err
}
