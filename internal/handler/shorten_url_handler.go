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
	json.NewDecoder(r.Body).Decode(&req)

	url, err := h.srv.ShortenURL(r.Context(), req.URL)

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
