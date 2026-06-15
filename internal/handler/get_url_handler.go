package handler

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"net/http"
)

func (h *URLHandler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	url, err := h.srv.GetOriginalURL(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		userID = ""
	}

	h.publishAuditEvent(r.Context(), "follow", userID, url)
}
