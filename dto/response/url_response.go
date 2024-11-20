package response

import "time"

type URLResponse struct {
	OriginalURL    string    `json:"originalUrl"`
	ShortLink      string    `json:"shortLink"`
	ExpirationDate time.Time `json:"expirationDate"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
