package service

import (
	"context"

	"github.com/sikvel71rus/shortener.git/internal/model"
)

type URLGRPCFacade interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error)
}

type URLFacade interface {
	URLGRPCFacade

	Ping(ctx context.Context) error
	ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error
	CountURLs(ctx context.Context) (int, error)
	CountUsers(ctx context.Context) (int, error)
}
