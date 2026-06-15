package handler

import (
	"errors"
	"fmt"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"io"
	"net/http"
)

func (h *URLHandler) PostURLHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	userID, err := h.ensureUserID(w, r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := h.srv.ShortenURL(r.Context(), string(body), userID)

	if errors.Is(err, repository.ErrConflict) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(id))
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%s", id)
	h.publishAuditEvent(r.Context(), "shorten", userID, string(body))
}
