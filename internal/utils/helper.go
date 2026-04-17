package utils

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"regexp"
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
	name = strings.ToLower(strings.Split(name, " ")[0])
	// Remove special characters and symbols
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	name = reg.ReplaceAllString(name, "")
	return fmt.Sprintf("%s@wafa.id", name)
}

func TodayDate() string {
	return time.Now().Format("2006-01-02")
}

// haversineDistance menghitung jarak dua koordinat dalam meter
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meter
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

// parseTimeString parse "HH:MM:SS" menjadi time.Time di tanggal yang diberikan
func ParseTimeString(t string, date string) (time.Time, error) {
	combined := fmt.Sprintf("%s %s", date, t)
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", combined, time.Local)
	if err != nil {
		// coba format HH:MM
		parsed, err = time.ParseInLocation("2006-01-02 15:04", combined[:len(date)+6], time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("parse time %q: %w", t, err)
		}
	}
	return parsed, nil
}

// UploadToPresignedURL performs an HTTP PUT to a presigned MinIO URL with the given data.
// Used for server-side uploads (e.g., profile photo from base64 payload).
func UploadToPresignedURL(presignedURL string, data []byte, contentType string) error {
	req, err := http.NewRequest(http.MethodPut, presignedURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(data))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}
	return nil
}

