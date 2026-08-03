package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuditPublisher struct {
	publishFunc func(ctx context.Context, event audit.Event) error
}

func (m *mockAuditPublisher) Publish(ctx context.Context, event audit.Event) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, event)
	}
	return nil
}

type mockURLService struct {
	getFunc          func(ctx context.Context, id string) (string, error)
	shortenFunc      func(ctx context.Context, url string, userID string) (string, error)
	shortenBatchFunc func(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	getUserURLsFunc  func(ctx context.Context, userID string) ([]model.UserURL, error)
	countURLsFunc    func(ctx context.Context) (int, error)
	countUsersFunc   func(ctx context.Context) (int, error)
	deleteUserURLsFn func(ctx context.Context, userID string, shortIDs []string) error
	pingFunc         func(ctx context.Context) error
}

func init() {
	_ = auth.SetSecret("test-secret")
}

func (m *mockURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return m.getFunc(ctx, id)
}

func (m *mockURLService) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	return m.shortenFunc(ctx, url, userID)
}

func (m *mockURLService) ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
	if m.shortenBatchFunc != nil {
		return m.shortenBatchFunc(ctx, batch, userID)
	}
	return nil, nil
}

func (m *mockURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	if m.getUserURLsFunc != nil {
		return m.getUserURLsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockURLService) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	if m.deleteUserURLsFn != nil {
		return m.deleteUserURLsFn(ctx, userID, shortIDs)
	}
	return nil
}

func (m *mockURLService) CountURLs(ctx context.Context) (int, error) {
	if m.countURLsFunc != nil {
		return m.countURLsFunc(ctx)
	}
	return 0, nil
}

func (m *mockURLService) CountUsers(ctx context.Context) (int, error) {
	if m.countUsersFunc != nil {
		return m.countUsersFunc(ctx)
	}
	return 0, nil
}

func (m *mockURLService) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func TestURLHandler_GetHandler(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name    string
		request string
		mockRes string
		mockErr error
		want    want
		wantURL string
	}{
		{
			name:    "positive GET test #1",
			request: "/aaBBBaa",
			mockRes: "https://practicum.yandex.ru/",
			mockErr: nil,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru/",
			},
			wantURL: "https://practicum.yandex.ru/",
		},
		{
			name:    "negative error not found GET test #2",
			request: "/CJDcdcdcd",
			mockRes: "",
			mockErr: errors.New("not found"),
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var events []audit.Event
			srv := &mockURLService{
				getFunc: func(ctx context.Context, id string) (string, error) {
					return tt.mockRes, tt.mockErr
				},
			}
			h := NewURLHandler(srv, &mockAuditPublisher{
				publishFunc: func(ctx context.Context, event audit.Event) error {
					events = append(events, event)
					return nil
				},
			})

			r := chi.NewRouter()
			r.Get("/{id}", h.GetURLHandler)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			if tt.want.statusCode == http.StatusTemporaryRedirect {
				require.Len(t, events, 1)
				assert.Equal(t, audit.ActionFollow, events[0].Action)
				assert.Equal(t, tt.wantURL, events[0].URL)
			} else {
				assert.Empty(t, events)
			}
		})
	}
}

func TestURLHandler_PostHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		body        string
	}
	tests := []struct {
		name    string
		body    string
		mockID  string
		mockErr error
		want    want
	}{
		{
			name:    "positive POST test #1",
			body:    "https://practicum.yandex.ru/",
			mockID:  "http://localhost:8080/aaBBBaa",
			mockErr: nil,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				body:        "http://localhost:8080/aaBBBaa",
			},
		},
		{
			name:    "negative POST test #2 (service error)",
			body:    "https://practicum.yandex.ru/",
			mockID:  "",
			mockErr: errors.New("database connection lost"),
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "",
				body:        "",
			},
		},
		{
			name:    "negative POST test #3 (conflict)",
			body:    "https://practicum.yandex.ru/",
			mockID:  "http://localhost:8080/alreadyExists",
			mockErr: repository.ErrConflict,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain",
				body:        "http://localhost:8080/alreadyExists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var events []audit.Event
			srv := &mockURLService{
				shortenFunc: func(ctx context.Context, url string, userID string) (string, error) {
					require.NotEmpty(t, userID)
					return tt.mockID, tt.mockErr
				},
			}
			h := NewURLHandler(srv, &mockAuditPublisher{
				publishFunc: func(ctx context.Context, event audit.Event) error {
					events = append(events, event)
					return nil
				},
			})

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.PostURLHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.NotEmpty(t, result.Cookies())

			respBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(respBody))
			if tt.want.statusCode == http.StatusCreated {
				require.Len(t, events, 1)
				assert.Equal(t, audit.ActionShorten, events[0].Action)
				assert.Equal(t, tt.body, events[0].URL)
				assert.NotEmpty(t, events[0].UserID)
			} else {
				assert.Empty(t, events)
			}
		})
	}
}

func TestURLHandler_ShortenJSONHandlerPublishesAudit(t *testing.T) {
	var events []audit.Event

	srv := &mockURLService{
		shortenFunc: func(ctx context.Context, url string, userID string) (string, error) {
			require.Equal(t, "https://practicum.yandex.ru/", url)
			require.NotEmpty(t, userID)
			return "http://localhost:8080/short", nil
		},
	}
	h := NewURLHandler(srv, &mockAuditPublisher{
		publishFunc: func(ctx context.Context, event audit.Event) error {
			events = append(events, event)
			return nil
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://practicum.yandex.ru/"}`))
	w := httptest.NewRecorder()

	h.ShortenJSONHandler(w, request)

	result := w.Result()
	defer result.Body.Close()

	assert.Equal(t, http.StatusCreated, result.StatusCode)
	require.Len(t, events, 1)
	assert.Equal(t, audit.ActionShorten, events[0].Action)
	assert.Equal(t, "https://practicum.yandex.ru/", events[0].URL)
	assert.NotEmpty(t, events[0].UserID)
}

func TestURLHandler_BatchHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}
	tests := []struct {
		name    string
		body    string
		mockRes []model.BatchResponse
		mockErr error
		want    want
	}{
		{
			name: "positive Batch test #1",
			body: `[
				{"correlation_id": "1", "original_url": "https://yandex.ru"},
				{"correlation_id": "2", "original_url": "https://google.com"}
			]`,
			mockRes: []model.BatchResponse{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/short1"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/short2"},
			},
			mockErr: nil,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:    "negative Batch test #2 (empty batch)",
			body:    `[]`,
			mockRes: nil,
			mockErr: nil,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
			},
		},
		{
			name:    "negative Batch test #3 (invalid JSON)",
			body:    `{"invalid": "json"}`,
			mockRes: nil,
			mockErr: nil,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockURLService{
				shortenBatchFunc: func(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
					require.NotEmpty(t, userID)
					return tt.mockRes, tt.mockErr
				},
			}
			h := NewURLHandler(srv)

			request := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.BatchHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			if tt.want.statusCode == http.StatusCreated {
				assert.NotEmpty(t, result.Cookies())
			}

			if tt.want.statusCode == http.StatusCreated {
				var respBody []model.BatchResponse
				err := json.NewDecoder(result.Body).Decode(&respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.mockRes, respBody)
				assert.Len(t, respBody, len(tt.mockRes))
			}
		})
	}
}

func TestURLHandler_GetHandlerDeleted(t *testing.T) {
	srv := &mockURLService{
		getFunc: func(ctx context.Context, id string) (string, error) {
			return "", repository.ErrDeleted
		},
	}
	h := NewURLHandler(srv)

	r := chi.NewRouter()
	r.Get("/{id}", h.GetURLHandler)

	req := httptest.NewRequest(http.MethodGet, "/deleted", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusGone, res.StatusCode)
}

func TestURLHandler_DeleteUserURLsHandler(t *testing.T) {
	token, err := auth.BuildToken("user-1")
	require.NoError(t, err)

	srv := &mockURLService{
		deleteUserURLsFn: func(ctx context.Context, userID string, shortIDs []string) error {
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, []string{"abc123", "def456"}, shortIDs)
			return nil
		},
	}
	h := NewURLHandler(srv)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["abc123","def456"]`))
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()

	h.DeleteUserURLsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusAccepted, res.StatusCode)
}

func TestURLHandler_StatsHandler(t *testing.T) {
	tests := []struct {
		name          string
		trustedSubnet string
		realIP        string
		countURLsErr  error
		countUsersErr error
		wantStatus    int
		wantBody      string
	}{
		{
			name:          "trusted IP",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.42",
			wantStatus:    http.StatusOK,
			wantBody:      `{"urls":5,"users":3}` + "\n",
		},
		{
			name:          "IP outside trusted subnet",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "10.0.0.1",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:       "empty trusted subnet",
			realIP:     "192.168.1.42",
			wantStatus: http.StatusForbidden,
		},
		{
			name:          "missing real IP",
			trustedSubnet: "192.168.1.0/24",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "invalid real IP",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "not-an-ip",
			wantStatus:    http.StatusForbidden,
		},
		{
			name:          "URL counter error",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.42",
			countURLsErr:  errors.New("storage error"),
			wantStatus:    http.StatusInternalServerError,
		},
		{
			name:          "user counter error",
			trustedSubnet: "192.168.1.0/24",
			realIP:        "192.168.1.42",
			countUsersErr: errors.New("storage error"),
			wantStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockURLService{
				countURLsFunc: func(ctx context.Context) (int, error) {
					return 5, tt.countURLsErr
				},
				countUsersFunc: func(ctx context.Context) (int, error) {
					return 3, tt.countUsersErr
				},
			}
			h := NewURLHandlerWithTrustedSubnet(srv, tt.trustedSubnet)

			req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			if tt.realIP != "" {
				req.Header.Set(realIPHeader, tt.realIP)
			}
			w := httptest.NewRecorder()

			h.StatsHandler(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatus, res.StatusCode)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
