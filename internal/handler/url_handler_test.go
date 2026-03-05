package handler

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockURLService struct {
	getFunc     func(id string) (string, error)
	shortenFunc func(url string) string
}

func (m *mockURLService) GetOriginalURL(id string) (string, error) {
	return m.getFunc(id)
}

func (m *mockURLService) ShortenURL(url string) string {
	return m.shortenFunc(url)
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
				getFunc: func(id string) (string, error) {
					return tt.mockRes, tt.mockErr
				},
			}
			h := NewURLHandler(srv)

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)
			w := httptest.NewRecorder()

			h.GetURLHandler(w, request)

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
		name   string
		body   string
		mockID string
		want   want
	}{
		{
			name:   "positive POST test #1",
			body:   "https://practicum.yandex.ru/",
			mockID: "http://localhost:8080/aaBBBaa",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				body:        "http://localhost:8080/aaBBBaa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &mockURLService{
				shortenFunc: func(url string) string {
					return tt.mockID
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
