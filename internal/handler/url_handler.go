package handler

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/service"

	"github.com/sikvel71rus/shortener.git/internal/auth"
	"go.uber.org/zap"
)

// URLHandler serves HTTP requests for URL-shortener endpoints.
type URLHandler struct {
	srv                     service.URLFacade
	audit                   audit.Publisher
	trustedSubnet           netip.Prefix
	trustedSubnetConfigured bool
}

// NewURLHandler creates a URL handler with an optional audit publisher.
func NewURLHandler(srv service.URLFacade, publishers ...audit.Publisher) *URLHandler {
	return newURLHandler(srv, "", publishers...)
}

// NewURLHandlerWithTrustedSubnet creates a URL handler with a trusted subnet for internal endpoints.
func NewURLHandlerWithTrustedSubnet(srv service.URLFacade, trustedSubnet string, publishers ...audit.Publisher) *URLHandler {
	return newURLHandler(srv, trustedSubnet, publishers...)
}

func newURLHandler(srv service.URLFacade, trustedSubnet string, publishers ...audit.Publisher) *URLHandler {
	var publisher audit.Publisher
	if len(publishers) > 0 {
		publisher = publishers[0]
	}

	h := &URLHandler{
		srv:   srv,
		audit: publisher,
	}

	if trustedSubnet = strings.TrimSpace(trustedSubnet); trustedSubnet != "" {
		if prefix, err := netip.ParsePrefix(trustedSubnet); err == nil {
			h.trustedSubnet = prefix
			h.trustedSubnetConfigured = true
		}
	}

	return h
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
