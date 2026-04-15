package dto

type APIResponse struct {
	Status     bool   `json:"status"`
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Data       any    `json:"data,omitempty"`
}

type UserData struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type Meta struct {
	ID         any   `json:"id" db:"id"`
	Name       string  `json:"name" db:"name"`
}
