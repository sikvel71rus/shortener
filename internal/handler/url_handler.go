package handler

type URLService interface {
	GetOriginalURL(id string) (string, error)
	ShortenURL(url string) string
}

type URLHandler struct {
	srv URLService
}

func NewURLHandler(srv URLService) *URLHandler {
	return &URLHandler{srv: srv}
}
