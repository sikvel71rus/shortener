package handler

import (
	"context"
	"encoding/json"
	"errors"
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

type mockURLService struct {
	getFunc          func(ctx context.Context, id string) (string, error)
	shortenFunc      func(ctx context.Context, url string) (string, error)
	shortenBatchFunc func(ctx context.Context, batch []model.BatchRequest) ([]model.BatchResponse, error)
	pingFunc         func(ctx context.Context) error
}

func (m *mockURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return m.getFunc(ctx, id)
}

func (m *mockURLService) ShortenURL(ctx context.Context, url string) (string, error) {
	return m.shortenFunc(ctx, url)
}

func (m *mockURLService) ShortenBatch(ctx context.Context, batch []model.BatchRequest) ([]model.BatchResponse, error) {
	if m.shortenBatchFunc != nil {
		return m.shortenBatchFunc(ctx, batch)
	}
	return nil, nil
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
			srv := &mockURLService{
				getFunc: func(ctx context.Context, id string) (string, error) {
					return tt.mockRes, tt.mockErr
				},
			}
			h := NewURLHandler(srv)

			r := chi.NewRouter()
			r.Get("/{id}", h.GetURLHandler)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
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
				contentType: "text/plain; charset=utf-8",
				body:        "Internal Server Error\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockURLService{
				shortenFunc: func(ctx context.Context, url string) (string, error) {
					return tt.mockID, tt.mockErr
				},
			}
			h := NewURLHandler(srv)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.PostURLHandler(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			respBody, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, string(respBody))
		})
	}
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
				shortenBatchFunc: func(ctx context.Context, batch []model.BatchRequest) ([]model.BatchResponse, error) {
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
				var respBody []model.BatchResponse
				err := json.NewDecoder(result.Body).Decode(&respBody)
				require.NoError(t, err)
				assert.Equal(t, tt.mockRes, respBody)
				assert.Len(t, respBody, len(tt.mockRes))
			}
		})
	}
}
