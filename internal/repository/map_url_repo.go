package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/sikvel71rus/shortener.git/internal/storage"
)

type MapURLRepo struct {
	mu       sync.RWMutex
	urls     map[string]string
	producer *storage.Producer
	counter  int
}

func NewMapURLRepo(filePath string) (*MapURLRepo, error) {
	repo := &MapURLRepo{
		urls: make(map[string]string),
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
			repo.urls[record.ShortURL] = record.OriginalURL
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

func (r *MapURLRepo) SaveURL(ctx context.Context, id, originalURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[id] = originalURL
	r.counter++

	if r.producer != nil {
		record := &storage.Record{
			UUID:        strconv.Itoa(r.counter),
			ShortURL:    id,
			OriginalURL: originalURL,
		}
		if err := r.producer.WriteEvent(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}

func (r *MapURLRepo) GetURL(ctx context.Context, id string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[id]
	if !ok {
		return "", errors.New("not found")
	}
	return url, nil
}

func (r *MapURLRepo) CheckIfURLExist(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.urls[id]
	return ok, nil
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
