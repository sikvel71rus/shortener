package service

import (
	"context"
	"errors"
	"math/rand"
	"sync"

	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
)

const shortIDLength = 6
const shortIDAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var errServiceClosed = errors.New("service is closed")

// URLService provides business logic for shortened URLs.
type URLService struct {
	repo          repository.URLRepo
	baseURL       string
	shortBase     string
	deleteCh      chan deleteTask
	closing       chan struct{}
	closeOnce     sync.Once
	closeMu       sync.Mutex
	closed        bool
	deleteSenders sync.WaitGroup
	wg            sync.WaitGroup
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
		closing:   make(chan struct{}),
	}

	svc.wg.Add(1)
	go svc.processDeleteQueue()

	return svc
}

// ShortenURL creates a short URL for the provided original URL.
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

// GetOriginalURL resolves a short ID to its original URL.
func (s *URLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.repo.GetURL(ctx, id)
}

// Ping checks that the underlying storage is available.
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

// ShortenBatch creates shortened URLs for all items in the batch.
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

// GetUserURLs returns all non-deleted URLs belonging to the user.
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

// DeleteUserURLs schedules deletion of user-owned shortened URLs.
func (s *URLService) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	idsCopy := append([]string(nil), shortIDs...)
	task := deleteTask{userID: userID, shortIDs: idsCopy}

	s.closeMu.Lock()
	if s.closed {
		s.closeMu.Unlock()
		return errServiceClosed
	}
	s.deleteSenders.Add(1)
	s.closeMu.Unlock()
	defer s.deleteSenders.Done()

	select {
	case s.deleteCh <- task:
		return nil
	case <-s.closing:
		return errServiceClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}

// CountURLs returns the total number of stored shortened URLs.
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

// Close gracefully stops background workers owned by the service.
func (s *URLService) Close() {
	s.closeOnce.Do(func() {
		s.closeMu.Lock()
		s.closed = true
		close(s.closing)
		s.closeMu.Unlock()

		s.deleteSenders.Wait()
		close(s.deleteCh)
		s.wg.Wait()
	})
}
