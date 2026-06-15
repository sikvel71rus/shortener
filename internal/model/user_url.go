package model

// UserURL represents one shortened URL owned by a user.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteRequest contains short URL identifiers to be deleted for a user.
type DeleteRequest []string
