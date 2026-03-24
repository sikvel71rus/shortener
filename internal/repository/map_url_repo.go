package repository

import (
	"fmt"
	"github.com/sikvel71rus/shortener.git/internal/storage"
	"strconv"
	"sync"
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
			if record != nil {
				repo.urls[record.ShortURL] = record.OriginalURL

				id, _ := strconv.Atoi(record.UUID)
				if id > repo.counter {
					repo.counter = id
				}
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

func (r *MapURLRepo) SaveURL(id, originalURL string) error {
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

func (r *MapURLRepo) Close() error {
	return r.producer.Close()
}

func (r *MapURLRepo) GetURL(id string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.urls[id]
	return url, ok
}

func (r *MapURLRepo) CheckIfURLExist(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.urls[id]
	return ok
}
