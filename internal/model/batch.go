package model

// BatchRequest describes one batch item for URL shortening.
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse describes one batch shortening result.
type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// BatchRecord stores internal batch data before persisting it.
type BatchRecord struct {
	ShortID     string
	OriginalURL string
}
