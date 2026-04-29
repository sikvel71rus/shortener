package handler

import (
	"context"
	"github.com/sikvel71rus/shortener.git/internal/model"
)

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string) (string, error)
	Ping(ctx context.Context) error
	ShortenBatch(ctx context.Context, batch []model.BatchRequest) ([]model.BatchResponse, error)
}

type URLHandler struct {
	srv URLService
}

func NewURLHandler(srv URLService) *URLHandler {
	return &URLHandler{srv: srv}
}
