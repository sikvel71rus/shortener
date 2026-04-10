package handler

import "context"

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string) (string, error)
	Ping(ctx context.Context) error
}

type URLHandler struct {
	srv URLService
}

func NewURLHandler(srv URLService) *URLHandler {
	return &URLHandler{srv: srv}
}
