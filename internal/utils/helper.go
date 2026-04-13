package utils

import (
	"time"
	"golang.org/x/crypto/bcrypt"
)

func Ms(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}

func PasswordHashing(str string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashPassword), nil
}