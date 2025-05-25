package security

import (
	"golang.org/x/crypto/bcrypt"
)

func Encode(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func Check(inputPassword, hashedPasswordFromDB string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPasswordFromDB),
		[]byte(inputPassword),
	)
	return err == nil
}
