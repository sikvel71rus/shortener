package handler

import (
	"fmt"
	"io"
	"net/http"
)

import (
	"github.com/sikvel71rus/shortener.git/internal/service"
)

type URLHandler struct {
	srv service.Shortener
	URL string
}

func NewURLHandler(srv service.Shortener, URL string) *URLHandler {
	return &URLHandler{srv: srv, URL: URL}
}

func (h *URLHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := h.srv.ShortenURL(string(body))
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s/%s", h.URL, id)
}

func (h *URLHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, err := h.srv.GetOriginalURL(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
