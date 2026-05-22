package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sikvel71rus/shortener.git/internal/model"
)

func (h *URLHandler) DeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req model.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.srv.DeleteUserURLs(r.Context(), userID, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
