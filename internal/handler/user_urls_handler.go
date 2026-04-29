package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sikvel71rus/shortener.git/internal/repository"
)

func (h *URLHandler) UserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.srv.GetUserURLs(r.Context(), userID)
	if errors.Is(err, repository.ErrNoUserURLs) {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
