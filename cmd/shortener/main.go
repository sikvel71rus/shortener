package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/middleware"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"log"
	"net/http"
)

func main() {

	starterCfg := starter.Parse()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	var db *sql.DB
	if starterCfg.DatabaseDSN != "" {
		var err error
		db, err = sql.Open("pgx", starterCfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
		defer db.Close()
	}

	repo, err := repository.NewMapURLRepo(starterCfg.FileStoragePath)
	if err != nil {
		panic(err)
	}

	srv := service.NewURLService(repo, starterCfg.BaseURL)
	h := handler.NewURLHandler(srv)
	pingHandler := handler.PingHandler(db)

	defer repo.Close()
	r := chi.NewRouter()

	log.Printf("Сервер запущен на %s, базовый адрес: %s", starterCfg.ServerAddress, starterCfg.BaseURL)

	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.PostURLHandler)
	r.Post("/api/shorten", h.ShortenJSONHandler)
	r.Get("/{id}", h.GetURLHandler)
	r.Get("/ping", pingHandler)

	err = http.ListenAndServe(starterCfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
