package repository

import (
	"context"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
)

type URLRepo interface {
	SaveURL(ctx context.Context, id string, originalURL string) error
	GetURL(ctx context.Context, id string) (string, error)
	GetShortIDByOriginalURL(ctx context.Context, originalURL string) (string, error)
	SaveBatch(ctx context.Context, records []model.BatchRecord) error
	Ping(ctx context.Context) error
	Close() error
}

var ErrConflict = errors.New("URL already exists")
