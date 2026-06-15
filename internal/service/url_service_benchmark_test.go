package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
)

func BenchmarkGenerateID(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_ = generateID()
	}
}

func BenchmarkURLServiceShortenURL(b *testing.B) {
	repo, err := repository.NewMapURLRepo("")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = repo.Close()
	})

	svc := NewURLService(repo, "http://localhost:8080")
	b.Cleanup(svc.Close)

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		url := fmt.Sprintf("https://example.com/%d", i)
		if _, err := svc.ShortenURL(ctx, url, "user-1"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkURLServiceGetUserURLs(b *testing.B) {
	repo, err := repository.NewMapURLRepo("")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = repo.Close()
	})

	svc := NewURLService(repo, "http://localhost:8080")
	b.Cleanup(svc.Close)

	ctx := context.Background()
	records := make([]model.BatchRecord, 0, 1024)
	for i := 0; i < 1024; i++ {
		records = append(records, model.BatchRecord{
			ShortID:     fmt.Sprintf("id%06d", i),
			OriginalURL: fmt.Sprintf("https://example.com/%d", i),
		})
	}

	if err := repo.SaveBatch(ctx, records, "user-1"); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		urls, err := svc.GetUserURLs(ctx, "user-1")
		if err != nil {
			b.Fatal(err)
		}
		if len(urls) != len(records) {
			b.Fatalf("unexpected urls count: got %d want %d", len(urls), len(records))
		}
	}
}
