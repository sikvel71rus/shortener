package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/middleware"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
)

func main() {
	starterCfg := starter.Parse()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	var repo repository.URLRepo
	var err error

	if starterCfg.DatabaseDSN != "" {
		repo, err = repository.NewPostgresRepo(starterCfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Ошибка инициализации БД: %v", err)
		}
		log.Println("Используется хранилище: PostgreSQL")

	} else if starterCfg.FileStoragePath != "" {
		repo, err = repository.NewMapURLRepo(starterCfg.FileStoragePath)
		if err != nil {
			log.Fatalf("Ошибка инициализации файлового хранилища: %v", err)
		}
		log.Println("Используется хранилище: Файл")

	} else {
		repo, err = repository.NewMapURLRepo("")
		if err != nil {
			log.Fatalf("Ошибка инициализации in-memory хранилища: %v", err)
		}
		log.Println("Используется хранилище: In-Memory")
	}

	defer repo.Close()

	srv := service.NewURLService(repo, starterCfg.BaseURL)
	h := handler.NewURLHandler(srv)

	r := chi.NewRouter()

	log.Printf("Сервер запущен на %s, базовый адрес: %s", starterCfg.ServerAddress, starterCfg.BaseURL)

	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.PostURLHandler)
	r.Post("/api/shorten", h.ShortenJSONHandler)
	r.Get("/{id}", h.GetURLHandler)
	r.Get("/ping", h.PingHandler)
	r.Post("/api/shorten/batch", h.BatchHandler)
	log.Fatal(http.ListenAndServe(starterCfg.ServerAddress, r))
}
