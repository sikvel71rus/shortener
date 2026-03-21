package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/logger"
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

	repo := repository.NewMapURLRepo()
	srv := service.NewURLService(repo, starterCfg.BaseURL)
	h := handler.NewURLHandler(srv)

	r := chi.NewRouter()

	log.Printf("Сервер запущен на %s, базовый адрес: %s", starterCfg.ServerAddress, starterCfg.BaseURL)

	r.Use(logger.RequestLogger)

	r.Post("/", h.PostURLHandler)
	r.Get("/{id}", h.GetURLHandler)

	err := http.ListenAndServe(starterCfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
