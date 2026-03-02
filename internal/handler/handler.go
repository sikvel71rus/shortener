package handler

import (
	"github.com/sikvel71rus/shortener.git/internal/service"
)

type URLHandler struct {
	srv service.Shortener
	URL string
}

func NewURLHandler(srv service.Shortener, URL string) *URLHandler {
	return &URLHandler{srv: srv, URL: URL}
}
