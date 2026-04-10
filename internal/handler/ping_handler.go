package handler

import (
	"context"
	"net/http"
	"time"
)

func (h *URLHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.srv.Ping(ctx); err != nil {
		http.Error(w, "Failed to connect to the db", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
