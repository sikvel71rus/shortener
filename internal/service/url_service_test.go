package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/sikvel71rus/shortener.git/internal/model"
)

type blockingDeleteRepo struct {
	started     chan struct{}
	release     chan struct{}
	startedOnce sync.Once
}

func (r *blockingDeleteRepo) SaveURL(ctx context.Context, id string, originalURL string, userID string) error {
	return nil
}

func (r *blockingDeleteRepo) GetURL(ctx context.Context, id string) (string, error) {
	return "", errors.New("not implemented")
}

func (r *blockingDeleteRepo) GetShortIDByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	return "", errors.New("not implemented")
}

func (r *blockingDeleteRepo) SaveBatch(ctx context.Context, records []model.BatchRecord, userID string) error {
	return nil
}

func (r *blockingDeleteRepo) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return nil, errors.New("not implemented")
}

func (r *blockingDeleteRepo) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	r.startedOnce.Do(func() {
		close(r.started)
	})
	<-r.release
	return nil
}

func (r *blockingDeleteRepo) CountURLs(ctx context.Context) (int, error) {
	return 0, nil
}

func (r *blockingDeleteRepo) Ping(ctx context.Context) error {
	return nil
}

func (r *blockingDeleteRepo) Close() error {
	return nil
}

func TestURLServiceCloseWaitsForQueuedDeletes(t *testing.T) {
	repo := &blockingDeleteRepo{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := NewURLService(repo, "http://localhost")

	if err := svc.DeleteUserURLs(context.Background(), "user-1", []string{"abc123"}); err != nil {
		t.Fatalf("DeleteUserURLs returned error: %v", err)
	}

	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("delete task was not started")
	}

	closed := make(chan struct{})
	go func() {
		svc.Close()
		close(closed)
	}()

	select {
	case <-closed:
		t.Fatal("Close returned before delete task finished")
	case <-time.After(20 * time.Millisecond):
	}

	close(repo.release)

	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not return after delete task finished")
	}
}

func TestURLServiceCloseUnblocksPendingDeleteSend(t *testing.T) {
	repo := &blockingDeleteRepo{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := NewURLService(repo, "http://localhost")

	if err := svc.DeleteUserURLs(context.Background(), "user-1", []string{"started"}); err != nil {
		t.Fatalf("DeleteUserURLs returned error: %v", err)
	}

	select {
	case <-repo.started:
	case <-time.After(time.Second):
		t.Fatal("delete task was not started")
	}

	for i := 0; i < cap(svc.deleteCh); i++ {
		if err := svc.DeleteUserURLs(context.Background(), "user-1", []string{"queued"}); err != nil {
			t.Fatalf("DeleteUserURLs returned error while filling queue: %v", err)
		}
	}

	pendingDone := make(chan error, 1)
	go func() {
		pendingDone <- svc.DeleteUserURLs(context.Background(), "user-1", []string{"pending"})
	}()

	select {
	case err := <-pendingDone:
		t.Fatalf("pending DeleteUserURLs returned before Close: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	closeDone := make(chan struct{})
	go func() {
		svc.Close()
		close(closeDone)
	}()

	select {
	case err := <-pendingDone:
		if !errors.Is(err, errServiceClosed) {
			t.Fatalf("expected service closed error, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("pending DeleteUserURLs was not unblocked by Close")
	}

	close(repo.release)

	select {
	case <-closeDone:
	case <-time.After(time.Second):
		t.Fatal("Close did not return after queued deletes were released")
	}
}
