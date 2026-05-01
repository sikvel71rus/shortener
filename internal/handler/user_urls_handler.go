package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/repository"
)

func (h *URLHandler) UserURLsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			if _, issueErr := h.issueNewCookie(w); issueErr != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.ParseUserID(cookie.Value)
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
