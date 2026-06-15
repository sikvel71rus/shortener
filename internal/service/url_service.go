package service

import (
	"context"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"math/rand"
	"sync"
)

const shortIDLength = 6
const shortIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type URLService struct {
	repo      repository.URLRepo
	baseURL   string
	shortBase string
	deleteCh  chan deleteTask
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type deleteTask struct {
	userID   string
	shortIDs []string
}

func NewURLService(repo repository.URLRepo, baseURL string) *URLService {
	svc := &URLService{
		repo:      repo,
		baseURL:   baseURL,
		shortBase: baseURL + "/",
		deleteCh:  make(chan deleteTask, 128),
	}

	svc.wg.Add(1)
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
		return s.shortBase + existingID, repository.ErrConflict
	}

	return s.shortBase + id, nil
}

func (s *URLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.GetURL(ctx, id)
}

func (s *URLService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func generateID() string {
	var b [shortIDLength]byte
	for i := range b {
		b[i] = shortIDAlphabet[rand.Intn(len(shortIDAlphabet))]
	}
	return string(b[:])
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
			ShortURL:      s.shortBase + id,
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

	for i := range urls {
		urls[i].ShortURL = s.shortBase + urls[i].ShortURL
	}

	return urls, nil
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
	defer s.wg.Done()

	for task := range s.deleteCh {
		s.deleteURLs(task.shortIDs, task.userID)
	}
}

func (s *URLService) deleteURLs(shortIDs []string, userID string) {
	_ = s.repo.DeleteUserURLs(context.Background(), userID, shortIDs)
}

func (s *URLService) Close() {
	s.closeOnce.Do(func() {
		close(s.deleteCh)
		s.wg.Wait()
	})
}
