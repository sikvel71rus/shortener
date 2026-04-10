package service

import (
	"context"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"math/rand"
	"strings"
)

type URLRepo interface {
	SaveURL(ctx context.Context, id string, originalURL string) error
	GetURL(ctx context.Context, id string) (string, error)
	CheckIfURLExist(ctx context.Context, id string) (bool, error)
	Ping(ctx context.Context) error
	SaveBatch(ctx context.Context, records []model.BatchRecord) error
	Close() error
}

type URLService struct {
	repo    URLRepo
	baseURL string
}

func NewURLService(repo URLRepo, baseURL string) *URLService {
	return &URLService{repo: repo, baseURL: baseURL}
}

func (s *URLService) ShortenURL(ctx context.Context, url string) (string, error) {
	id := ""
	for {
		id = generateID()
		exists, err := s.repo.CheckIfURLExist(ctx, id)
		if err != nil {
			return "", err
		}
		if !exists {
			break
		}
	}

	if err := s.repo.SaveURL(ctx, id, url); err != nil {
		return "", err
	}

	return s.baseURL + "/" + id, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.GetURL(ctx, id)
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func generateID() string {
	length := 6
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func (s *URLService) ShortenBatch(ctx context.Context, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	records := make([]model.BatchRecord, 0, len(batch))
	result := make([]model.BatchResponse, 0, len(batch))

	for _, req := range batch {
		id := generateID()

		records = append(records, model.BatchRecord{
			ShortID:     id,
			OriginalURL: req.OriginalURL,
		})

		result = append(result, model.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      s.baseURL + "/" + id,
		})
	}

	if err := s.repo.SaveBatch(ctx, records); err != nil {
		return nil, err
	}

	return result, nil
}
