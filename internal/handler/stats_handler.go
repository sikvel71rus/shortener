package handler

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"strings"

	"github.com/sikvel71rus/shortener.git/internal/model"
)

const realIPHeader = "X-Real-IP"

// StatsHandler handles GET /api/internal/stats requests.
func (h *URLHandler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	if !h.isTrustedRequest(r) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	urls, err := h.srv.CountURLs(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users, err := h.srv.CountUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(model.StatsResponse{URLs: urls, Users: users}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) isTrustedRequest(r *http.Request) bool {
	if !h.trustedSubnetConfigured {
		return false
	}

	clientIP := strings.TrimSpace(r.Header.Get(realIPHeader))
	if clientIP == "" {
		return false
	}

	addr, err := netip.ParseAddr(clientIP)
	if err != nil {
		return false
	}

	return h.trustedSubnet.Contains(addr)
}
