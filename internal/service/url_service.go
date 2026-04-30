package service

import (
	"context"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"math/rand"
	"strings"
)

type URLService struct {
	repo     repository.URLRepo
	baseURL  string
	deleteCh chan deleteTask
}

type deleteTask struct {
	userID   string
	shortIDs []string
}

func NewURLService(repo repository.URLRepo, baseURL string) *URLService {
	svc := &URLService{
		repo:     repo,
		baseURL:  baseURL,
		deleteCh: make(chan deleteTask, 128),
	}

	go svc.processDeleteQueue()

	return svc
}

func (s *URLService) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	id := generateID()
	err := s.repo.SaveURL(ctx, id, url, userID)

	if errors.Is(err, repository.ErrConflict) {
		existingID, getErr := s.repo.GetShortIDByOriginalURL(ctx, url)
		if getErr != nil {
			return "", getErr
		}
		return s.baseURL + "/" + existingID, repository.ErrConflict
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

func (s *URLService) ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
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

	if err := s.repo.SaveBatch(ctx, records, userID); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	urls, err := s.repo.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.UserURL, 0, len(urls))
	for _, item := range urls {
		result = append(result, model.UserURL{
			ShortURL:    s.baseURL + "/" + item.ShortURL,
			OriginalURL: item.OriginalURL,
		})
	}

	return result, nil
}

func (s *URLService) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	idsCopy := append([]string(nil), shortIDs...)

	select {
	case s.deleteCh <- deleteTask{userID: userID, shortIDs: idsCopy}:
	default:
		go s.deleteURLs(idsCopy, userID)
	}

	return nil
}

func (s *URLService) CountURLs(ctx context.Context) (int, error) {
	return s.repo.CountURLs(ctx)
}

func (s *URLService) processDeleteQueue() {
	for task := range s.deleteCh {
		s.deleteURLs(task.shortIDs, task.userID)
	}
}

func (s *URLService) deleteURLs(shortIDs []string, userID string) {
	_ = s.repo.DeleteUserURLs(context.Background(), userID, shortIDs)
}
