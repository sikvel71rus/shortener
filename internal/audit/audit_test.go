package audit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileObserverNotify(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	observer := NewFileObserver(path)
	event := Event{
		Timestamp: 1710000000,
		Action:    ActionShorten,
		UserID:    "user-1",
		URL:       "https://example.com/long",
	}

	err := observer.Notify(context.Background(), event)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var actual Event
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(string(data))), &actual))
	assert.Equal(t, event, actual)
}

func TestHTTPObserverNotify(t *testing.T) {
	var actual Event
	observer := &HTTPObserver{
		target: "https://audit.example.test",
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				require.NoError(t, json.NewDecoder(r.Body).Decode(&actual))

				return &http.Response{
					StatusCode: http.StatusAccepted,
					Status:     "202 Accepted",
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	event := Event{
		Timestamp: 1710000000,
		Action:    ActionFollow,
		URL:       "https://example.com/original",
	}

	err := observer.Notify(context.Background(), event)
	require.NoError(t, err)
	assert.Equal(t, event, actual)
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
