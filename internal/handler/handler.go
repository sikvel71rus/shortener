package handler

type URLService interface {
	GetOriginalURL(id string) (string, error)
	ShortenURL(url string) string
}

type URLHandler struct {
	srv URLService
	URL string
}

func NewURLHandler(srv URLService, URL string) *URLHandler {
	return &URLHandler{srv: srv, URL: URL}
}
