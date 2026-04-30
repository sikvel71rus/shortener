package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"strconv"
	"sync"

	"github.com/sikvel71rus/shortener.git/internal/storage"
)

type MapURLRepo struct {
	mu       sync.RWMutex
	urls     map[string]string
	original map[string]string
	userURLs map[string]map[string]struct{}
	deleted  map[string]bool
	producer *storage.Producer
	counter  int
}

func NewMapURLRepo(filePath string) (*MapURLRepo, error) {
	repo := &MapURLRepo{
		urls:     make(map[string]string),
		original: make(map[string]string),
		userURLs: make(map[string]map[string]struct{}),
		deleted:  make(map[string]bool),
	}

	if filePath != "" {
		consumer, err := storage.NewConsumer(filePath)
		if err != nil {
			return nil, err
		}
		defer consumer.Close()

		for {
			record, err := consumer.ReadEvent()
			if err != nil {
				break
			}
			if record == nil {
				break
			}
			if record.OriginalURL != "" {
				repo.urls[record.ShortURL] = record.OriginalURL
				repo.original[record.OriginalURL] = record.ShortURL
			}
			repo.bindUserURL(record.UserID, record.ShortURL)
			repo.deleted[record.ShortURL] = record.IsDeleted
			id, _ := strconv.Atoi(record.UUID)
			if id > repo.counter {
				repo.counter = id
			}
		}

		producer, err := storage.NewProducer(filePath)
		if err != nil {
			return nil, err
		}
		repo.producer = producer
	}

	return repo, nil
}

func (r *MapURLRepo) SaveURL(ctx context.Context, id, originalURL, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingID, ok := r.original[originalURL]; ok {
		r.bindUserURL(userID, existingID)
		return ErrConflict
	}

	r.urls[id] = originalURL
	r.original[originalURL] = id
	r.bindUserURL(userID, id)
	r.counter++

	if r.producer != nil {
		record := &storage.Record{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    id,
			OriginalURL: originalURL,
			UserID:      userID,
		}
		if err := r.producer.WriteEvent(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (r *MapURLRepo) GetShortIDByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.original[originalURL]; ok {
		return id, nil
	}
	return "", errors.New("not found")
}

func (r *MapURLRepo) GetURL(ctx context.Context, id string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[id]
	if !ok {
		return "", errors.New("not found")
	}
	if r.deleted[id] {
		return "", ErrDeleted
	}
	return url, nil
}

func (r *MapURLRepo) Ping(ctx context.Context) error {
	return nil
}

func (r *MapURLRepo) Close() error {
	if r.producer != nil {
		return r.producer.Close()
	}
	return nil
}

func (r *MapURLRepo) SaveBatch(ctx context.Context, records []model.BatchRecord, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, rec := range records {
		if existingID, ok := r.original[rec.OriginalURL]; ok {
			r.bindUserURL(userID, existingID)
			continue
		}

		r.urls[rec.ShortID] = rec.OriginalURL
		r.original[rec.OriginalURL] = rec.ShortID
		r.bindUserURL(userID, rec.ShortID)
		r.counter++

		if r.producer != nil {
			record := &storage.Record{
				UUID:        strconv.Itoa(r.counter),
				ShortURL:    rec.ShortID,
				OriginalURL: rec.OriginalURL,
				UserID:      userID,
			}
			if err := r.producer.WriteEvent(record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MapURLRepo) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shortIDs, ok := r.userURLs[userID]
	if !ok || len(shortIDs) == 0 {
		return nil, ErrNoUserURLs
	}

	result := make([]model.UserURL, 0, len(shortIDs))
	for shortID := range shortIDs {
		if r.deleted[shortID] {
			continue
		}
		result = append(result, model.UserURL{
			ShortURL:    shortID,
			OriginalURL: r.urls[shortID],
		})
	}

	if len(result) == 0 {
		return nil, ErrNoUserURLs
	}

	return result, nil
}

func (r *MapURLRepo) CountURLs(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.urls), nil
}

func (r *MapURLRepo) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	owned := r.userURLs[userID]
	for _, shortID := range shortIDs {
		if _, ok := owned[shortID]; ok {
			r.deleted[shortID] = true
			if r.producer != nil {
				r.counter++
				record := &storage.Record{
					UUID:      strconv.Itoa(r.counter),
					ShortURL:  shortID,
					UserID:    userID,
					IsDeleted: true,
				}
				if err := r.producer.WriteEvent(record); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *MapURLRepo) bindUserURL(userID, shortID string) {
	if userID == "" {
		return
	}

	if _, ok := r.userURLs[userID]; !ok {
		r.userURLs[userID] = make(map[string]struct{})
	}

	r.userURLs[userID][shortID] = struct{}{}
}
