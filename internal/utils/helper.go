package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"hris-backend/internal/utils/data"

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

func IsPasswordMatch(hashPassword, reqPassword string) bool {
	hash, pass := []byte(hashPassword), []byte(reqPassword)

	err := bcrypt.CompareHashAndPassword(hash, pass)
	return err == nil
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	if length <= 0 {
		length = 16
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func ParseDate(s string) (time.Time, error) {
	for _, layout := range data.DateFormats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("parseDate: format tidak dikenali untuk %q", s)
}

func ParseTimestamp(s string) (time.Time, error) {
	for _, layout := range data.TimestampFormats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("parseTimestamp: format tidak dikenali untuk %q", s)
}

func ParseAuto(s string) (time.Time, error) {
	if t, err := ParseTimestamp(s); err == nil {
		return t, nil
	}
	if t, err := ParseDate(s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("parseAuto: tidak bisa mengurai %q sebagai DATE maupun TIMESTAMP", s)
}

func GenerateEmail(name string) string {
	name = strings.Split(name, " ")[0]
	return fmt.Sprintf("%s@wafa.id", name)
}
