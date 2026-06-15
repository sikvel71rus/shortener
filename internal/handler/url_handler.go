package handler

import (
	"context"
	"errors"
	"time"

	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"net/http"

	"github.com/sikvel71rus/shortener.git/internal/auth"
	"go.uber.org/zap"
)

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
	Ping(ctx context.Context) error
	ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error)
	DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error
	CountURLs(ctx context.Context) (int, error)
}

type URLHandler struct {
	srv   URLService
	audit audit.Publisher
}

func NewURLHandler(srv URLService, publishers ...audit.Publisher) *URLHandler {
	var publisher audit.Publisher
	if len(publishers) > 0 {
		publisher = publishers[0]
	}

	return &URLHandler{
		srv:   srv,
		audit: publisher,
	}
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

func (h *URLHandler) publishAuditEvent(ctx context.Context, action string, userID string, url string) {
	if h.audit == nil {
		return
	}

	event := audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}

	if err := h.audit.Publish(ctx, event); err != nil {
		logger.Log.Error("failed to publish audit event",
			zap.String("action", action),
			zap.String("url", url),
			zap.Error(err),
		)
	}
}
