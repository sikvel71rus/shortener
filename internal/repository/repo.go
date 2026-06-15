package repository

import (
	"context"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
)

// URLRepo describes storage operations required by the URL service.
type URLRepo interface {
	SaveURL(ctx context.Context, id string, originalURL string, userID string) error
	GetURL(ctx context.Context, id string) (string, error)
	GetShortIDByOriginalURL(ctx context.Context, originalURL string) (string, error)
	SaveBatch(ctx context.Context, records []model.BatchRecord, userID string) error
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error)
	DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error
	CountURLs(ctx context.Context) (int, error)
	Ping(ctx context.Context) error
	Close() error
}

// ErrConflict indicates that the original URL already exists in storage.
var ErrConflict = errors.New("URL already exists")

// ErrNoUserURLs indicates that no URLs were found for the user.
var ErrNoUserURLs = errors.New("user has no urls")

// ErrDeleted indicates that a short URL exists but is marked as deleted.
var ErrDeleted = errors.New("url deleted")
