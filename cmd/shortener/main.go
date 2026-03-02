package main

import (
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"net/http"
)

func main() {

	repo := repository.NewMemoryRepo()
	srv := service.NewURLService(repo)
	h := handler.NewURLHandler(srv, "http://localhost:8080")

	err := http.ListenAndServe(":8080", h)
	if err != nil {
		panic(err)
	}
}
