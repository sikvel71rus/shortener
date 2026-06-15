package model

// ShortenRequest describes the JSON payload for creating a shortened URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse describes the JSON response with a shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}
