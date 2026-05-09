package dto

import "time"

// NotificationListParams — query params untuk list notifikasi
type NotificationListParams struct {
	PaginationParams
	IsRead *bool `query:"is_read"`
}

// NotificationResponse — response single notification
type NotificationResponse struct {
	ID                uint       `json:"id"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	ActionURL         string     `json:"action_url"`
	ActionTab         string     `json:"action_tab"`
	IsRead            bool       `json:"is_read"`
	ReadAt            *time.Time `json:"read_at"`
	PushStatus        string     `json:"push_status"`
	RelatedEntityType string     `json:"related_entity_type"`
	RelatedEntityID   *uint      `json:"related_entity_id"`
	CreatedAt         time.Time  `json:"created_at"`
}

// MarkAsReadRequest — body untuk mark as read (kosong = mark all)
type MarkAsReadRequest struct {
	IDs []uint `json:"ids"`
}

// UnreadCountResponse — response unread count
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}
