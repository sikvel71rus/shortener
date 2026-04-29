package handler

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *URLHandler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := h.srv.GetOriginalURL(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
