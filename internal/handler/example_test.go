package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/model"
)

type exampleURLService struct {
	getOriginalURL func(ctx context.Context, id string) (string, error)
	shortenURL     func(ctx context.Context, url string, userID string) (string, error)
	ping           func(ctx context.Context) error
	shortenBatch   func(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error)
	getUserURLs    func(ctx context.Context, userID string) ([]model.UserURL, error)
	deleteUserURLs func(ctx context.Context, userID string, shortIDs []string) error
	countURLs      func(ctx context.Context) (int, error)
	countUsers     func(ctx context.Context) (int, error)
}

func (s exampleURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return s.getOriginalURL(ctx, id)
}

func (s exampleURLService) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	return s.shortenURL(ctx, url, userID)
}

func (s exampleURLService) Ping(ctx context.Context) error {
	return s.ping(ctx)
}

func (s exampleURLService) ShortenBatch(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
	return s.shortenBatch(ctx, batch, userID)
}

func (s exampleURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return s.getUserURLs(ctx, userID)
}

func (s exampleURLService) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) error {
	return s.deleteUserURLs(ctx, userID, shortIDs)
}

func (s exampleURLService) CountURLs(ctx context.Context) (int, error) {
	return s.countURLs(ctx)
}

func (s exampleURLService) CountUsers(ctx context.Context) (int, error) {
	return s.countUsers(ctx)
}

func ExampleURLHandler_PostURLHandler() {
	_ = auth.SetSecret("example-secret")

	h := handler.NewURLHandler(exampleURLService{
		shortenURL: func(ctx context.Context, url string, userID string) (string, error) {
			return "http://localhost:8080/abc123", nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com/long"))
	w := httptest.NewRecorder()

	h.PostURLHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 201
	// http://localhost:8080/abc123
}

func ExampleURLHandler_ShortenJSONHandler() {
	_ = auth.SetSecret("example-secret")

	h := handler.NewURLHandler(exampleURLService{
		shortenURL: func(ctx context.Context, url string, userID string) (string, error) {
			return "http://localhost:8080/json01", nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com/long"}`))
	w := httptest.NewRecorder()

	h.ShortenJSONHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 201
	// {"result":"http://localhost:8080/json01"}
}

func ExampleURLHandler_GetURLHandler() {
	h := handler.NewURLHandler(exampleURLService{
		getOriginalURL: func(ctx context.Context, id string) (string, error) {
			return "https://example.com/original", nil
		},
	})

	r := chi.NewRouter()
	r.Get("/{id}", h.GetURLHandler)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(res.Header.Get("Location"))
	// Output:
	// 307
	// https://example.com/original
}

func ExampleURLHandler_BatchHandler() {
	_ = auth.SetSecret("example-secret")

	h := handler.NewURLHandler(exampleURLService{
		shortenBatch: func(ctx context.Context, batch []model.BatchRequest, userID string) ([]model.BatchResponse, error) {
			return []model.BatchResponse{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/a1"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/b2"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(`[{"correlation_id":"1","original_url":"https://a.test"},{"correlation_id":"2","original_url":"https://b.test"}]`))
	w := httptest.NewRecorder()

	h.BatchHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 201
	// [{"correlation_id":"1","short_url":"http://localhost:8080/a1"},{"correlation_id":"2","short_url":"http://localhost:8080/b2"}]
}

func ExampleURLHandler_UserURLsHandler() {
	_ = auth.SetSecret("example-secret")
	token, _ := auth.BuildToken("user-1")

	h := handler.NewURLHandler(exampleURLService{
		getUserURLs: func(ctx context.Context, userID string) ([]model.UserURL, error) {
			return []model.UserURL{
				{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com/long"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()

	h.UserURLsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 200
	// [{"short_url":"http://localhost:8080/abc123","original_url":"https://example.com/long"}]
}

func ExampleURLHandler_DeleteUserURLsHandler() {
	_ = auth.SetSecret("example-secret")
	token, _ := auth.BuildToken("user-1")

	h := handler.NewURLHandler(exampleURLService{
		deleteUserURLs: func(ctx context.Context, userID string, shortIDs []string) error {
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["abc123","def456"]`))
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token})
	w := httptest.NewRecorder()

	h.DeleteUserURLsHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 202
}

func ExampleURLHandler_PingHandler() {
	h := handler.NewURLHandler(exampleURLService{
		ping: func(ctx context.Context) error {
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.PingHandler(w, req)

	res := w.Result()
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
	// Output:
	// 200
}

func ExampleURLHandler_ShortenJSONHandler_responseFormat() {
	_ = auth.SetSecret("example-secret")

	h := handler.NewURLHandler(exampleURLService{
		shortenURL: func(ctx context.Context, url string, userID string) (string, error) {
			return "http://localhost:8080/demo42", nil
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://example.com/demo"}`))
	w := httptest.NewRecorder()

	h.ShortenJSONHandler(w, req)

	var resp model.ShortenResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Println(resp.Result)
	// Output:
	// http://localhost:8080/demo42
}
