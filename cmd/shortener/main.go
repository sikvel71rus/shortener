package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/config/flag"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"net/http"
)

func main() {

	flagCfg := flag.Parse()

	repo := repository.NewMapURLRepo()
	srv := service.NewURLService(repo, flagCfg.BaseURL)
	h := handler.NewURLHandler(srv)

	r := chi.NewRouter()

	r.Post("/", h.PostUrlHandler)
	r.Get("/{id}", h.GetURLHandler)

	err := http.ListenAndServe(flagCfg.ServerAddress, r)
	if err != nil {
		panic(err)
	}
}
