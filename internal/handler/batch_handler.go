package handler

import (
	"encoding/json"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"net/http"
)

// BatchHandler handles POST /api/shorten/batch requests with multiple URLs.
func (h *URLHandler) BatchHandler(w http.ResponseWriter, r *http.Request) {
	var req []model.BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.ensureUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := h.srv.ShortenBatch(r.Context(), req, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
