package handler

import (
	"encoding/json"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"net/http"
)

func (h *URLHandler) ShortenJSONHandler(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	url, err := h.srv.ShortenURL(r.Context(), req.URL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp := model.ShortenResponse{
		Result: url,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
