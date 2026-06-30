package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/sikvel71rus/shortener.git/internal/model"
)

func BenchmarkMapURLRepoSaveURL(b *testing.B) {
	repo, err := NewMapURLRepo("")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = repo.Close()
	})

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		if err := repo.SaveURL(ctx, fmt.Sprintf("id%06d", i), fmt.Sprintf("https://example.com/%d", i), "user-1"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapURLRepoGetUserURLs(b *testing.B) {
	repo, err := NewMapURLRepo("")
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() {
		_ = repo.Close()
	})

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
		urls, err := repo.GetUserURLs(ctx, "user-1")
		if err != nil {
			b.Fatal(err)
		}
		if len(urls) != len(records) {
			b.Fatalf("unexpected urls count: got %d want %d", len(urls), len(records))
		}
	}
}
