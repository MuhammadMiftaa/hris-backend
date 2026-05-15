package dto

type ExternalPrayerTimeAPIResponse struct {
	IsSuccess bool                     `json:"is_success"`
	Message   string                   `json:"message"`
	Data    []ExternalPrayerTimeItem `json:"data"`
}

type ExternalPrayerTimeItem struct {
	Date   string `json:"date"`   // Format: YYYY-MM-DD
	Dzuhur string `json:"dzuhur"` // Format: HH:mm
	Ashr   string `json:"ashr"`   // Format: HH:mm
}
