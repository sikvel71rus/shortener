package handler

import (
	"context"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"net/http"

	"github.com/sikvel71rus/shortener.git/internal/auth"
)

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
	Ping(ctx context.Context) error
	ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error)
}

type URLHandler struct {
	srv URLService
}

func NewURLHandler(srv URLService) *URLHandler {
	return &URLHandler{srv: srv}
}

func (h *URLHandler) ensureUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return h.issueNewCookie(w)
		}
		return "", err
	}

	userID, err := auth.ParseUserID(cookie.Value)
	if err != nil {
		return h.issueNewCookie(w)
	}

	return userID, nil
}

func (h *URLHandler) getUserIDFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(auth.CookieName)
	if err != nil {
		return "", err
	}

	return auth.ParseUserID(cookie.Value)
}

func (h *URLHandler) issueNewCookie(w http.ResponseWriter) (string, error) {
	cookie, userID, err := auth.NewSignedCookie()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, cookie)
	return userID, nil
}
