package service

import (
	"context"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	logger "hris-backend/config/log"
	"hris-backend/internal/struct/model"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushService interface {
	SendPush(ctx context.Context, sub model.PushSubscription, payload []byte) error
}

type pushService struct {
	vapidPrivateKey string
	vapidPublicKey  string
	vapidSubject    string
}

// deriveVAPIDPublicKey derives the VAPID public key from the private key
// using P-256 elliptic curve scalar base multiplication.
// This is the same math webpush-go uses internally in generateVAPIDHeaderKeys.
func deriveVAPIDPublicKey(privateKeyBase64 string) (string, error) {
	// Decode private key (support both padded and unpadded base64url)
	privKeyBytes, err := base64.RawURLEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		privKeyBytes, err = base64.URLEncoding.DecodeString(privateKeyBase64)
		if err != nil {
			return "", fmt.Errorf("failed to decode VAPID private key: %w", err)
		}
	}

	// Derive public key point (X, Y) = privateKey × G
	curve := elliptic.P256()
	x, y := curve.ScalarBaseMult(privKeyBytes)

	// Marshal to uncompressed format: 0x04 || X || Y
	pubKeyBytes := elliptic.Marshal(curve, x, y)

	return base64.RawURLEncoding.EncodeToString(pubKeyBytes), nil
}

func NewPushService(privateKey, publicKey, subject string) PushService {
	if privateKey == "" {
		logger.Warn("push service initialized without VAPID private key — push notifications will fail", map[string]any{
			"service": "push",
		})
	} else if publicKey == "" {
		// Derive public key from private key
		derived, err := deriveVAPIDPublicKey(privateKey)
		if err != nil {
			logger.Error("failed to derive VAPID public key from private key", map[string]any{
				"service": "push",
				"error":   err.Error(),
			})
		} else {
			publicKey = derived
			logger.Info("VAPID public key derived from private key", map[string]any{
				"service": "push",
			})
		}
	}

	return &pushService{
		vapidPrivateKey: privateKey,
		vapidPublicKey:  publicKey,
		vapidSubject:    subject,
	}
}

func (s *pushService) SendPush(ctx context.Context, sub model.PushSubscription, payload []byte) error {
	if s.vapidPrivateKey == "" || s.vapidPublicKey == "" {
		return fmt.Errorf("VAPID keys not configured — cannot send push notification")
	}

	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	resp, err := webpush.SendNotificationWithContext(ctx, payload, subscription, &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             60,
	})
	if err != nil {
		return fmt.Errorf("push send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("push subscription expired (status %d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("push rejected by endpoint (status %d): %s", resp.StatusCode, string(body))
	}

	logger.Info("push sent successfully", map[string]any{
		"service":     "push",
		"status_code": resp.StatusCode,
	})
	return nil
}

// BuildPushPayload — buat payload JSON untuk service worker
type PushPayload struct {
	Title   string              `json:"title"`
	Body    string              `json:"body"`
	Icon    string              `json:"icon,omitempty"`
	Badge   string              `json:"badge,omitempty"`
	Tag     string              `json:"tag,omitempty"`
	Data    PushPayloadData     `json:"data"`
	Actions []PushPayloadAction `json:"actions,omitempty"`
}

type PushPayloadData struct {
	URL            string `json:"url"`
	Tab            string `json:"tab,omitempty"`
	NotificationID uint   `json:"notification_id"`
}

type PushPayloadAction struct {
	Action string `json:"action"`
	Title  string `json:"title"`
}

func BuildPushPayload(title, body, actionURL, actionTab string, notificationID uint) []byte {
	payload := PushPayload{
		Title: title,
		Body:  body,
		Icon:  "/images/icons/android-icon-192.png",
		Badge: "/images/icons/android-icon-96.png",
		Tag:   "wafa-hris",
		Data: PushPayloadData{
			URL:            actionURL,
			Tab:            actionTab,
			NotificationID: notificationID,
		},
	}

	b, _ := json.Marshal(payload)
	return b
}
