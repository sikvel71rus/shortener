package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"net/http"
)

func main() {

	starterCfg := starter.Parse()

	repo := repository.NewMapURLRepo()
	srv := service.NewURLService(repo, starterCfg.BaseURL)
	h := handler.NewURLHandler(srv)

	r := chi.NewRouter()

	r.Post("/", h.PostURLHandler)
	r.Get("/{id}", h.GetURLHandler)

	err := http.ListenAndServe(starterCfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
