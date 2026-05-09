package dto

// PushSubscribeRequest — request dari frontend saat subscribe push
type PushSubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// PushSubscriptionResponse — response status subscription
type PushSubscriptionResponse struct {
	IsSubscribed bool `json:"is_subscribed"`
}
