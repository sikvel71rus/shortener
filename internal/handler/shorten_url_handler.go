package handler

import (
	"encoding/json"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"net/http"
)

func (h *URLHandler) ShortenJSONHandler(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.ensureUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	url, err := h.srv.ShortenURL(r.Context(), req.URL, userID)

	resp := model.ShortenResponse{
		Result: url,
	}

	if errors.Is(err, repository.ErrConflict) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
