package service

import (
	"context"
	"encoding/json"
	"fmt"
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

func NewPushService(privateKey, publicKey, subject string) PushService {
	return &pushService{
		vapidPrivateKey: privateKey,
		vapidPublicKey:  publicKey,
		vapidSubject:    subject,
	}
}

func (s *pushService) SendPush(ctx context.Context, sub model.PushSubscription, payload []byte) error {
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
		return fmt.Errorf("push rejected by endpoint (status %d)", resp.StatusCode)
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
