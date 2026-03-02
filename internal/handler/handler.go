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

func (h *URLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Path[1:]
		url, err := h.srv.GetOriginalURL(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		id := h.srv.ShortenURL(string(body))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "%s/%s", h.URL, id)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
