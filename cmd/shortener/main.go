package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"net/http"
)

func main() {

	repo := repository.NewMemoryRepo()
	srv := service.NewURLService(repo)
	h := handler.NewURLHandler(srv, "http://localhost:8080")

	r := chi.NewRouter()

	r.Post("/", h.PostHandler)
	r.Get("/{id}", h.GetHandler)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
