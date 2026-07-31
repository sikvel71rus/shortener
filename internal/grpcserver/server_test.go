package grpcserver

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/pkg/shortenerpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

type mockURLService struct {
	getOriginalURL func(ctx context.Context, id string) (string, error)
	shortenURL     func(ctx context.Context, url string, userID string) (string, error)
	getUserURLs    func(ctx context.Context, userID string) ([]model.UserURL, error)
}

func (m mockURLService) GetOriginalURL(ctx context.Context, id string) (string, error) {
	return m.getOriginalURL(ctx, id)
}

func (m mockURLService) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	return m.shortenURL(ctx, url, userID)
}

func (m mockURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return m.getUserURLs(ctx, userID)
}

func TestShortenerServiceGRPC(t *testing.T) {
	require.NoError(t, auth.SetSecret("grpc-test-secret"))

	var savedUserID string
	client, cleanup := newTestClient(t, mockURLService{
		shortenURL: func(ctx context.Context, url string, userID string) (string, error) {
			savedUserID = userID
			assert.Equal(t, "https://example.com/long", url)
			return "http://localhost:8080/abc123", nil
		},
		getOriginalURL: func(ctx context.Context, id string) (string, error) {
			assert.Equal(t, "abc123", id)
			return "https://example.com/long", nil
		},
		getUserURLs: func(ctx context.Context, userID string) ([]model.UserURL, error) {
			assert.Equal(t, savedUserID, userID)
			return []model.UserURL{
				{ShortURL: "http://localhost:8080/abc123", OriginalURL: "https://example.com/long"},
			}, nil
		},
	})
	defer cleanup()

	var header metadata.MD
	shortenResp, err := client.ShortenURL(context.Background(),
		&shortenerpb.URLShortenRequest{URL: "https://example.com/long"},
		grpc.Header(&header),
	)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080/abc123", shortenResp.Result)

	authHeaders := header.Get(authorizationHeader)
	require.Len(t, authHeaders, 1)
	require.NotEmpty(t, savedUserID)

	ctx := metadata.AppendToOutgoingContext(context.Background(), authorizationHeader, authHeaders[0])

	expandResp, err := client.ExpandURL(ctx, &shortenerpb.URLExpandRequest{ID: "abc123"})
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/long", expandResp.Result)

	urlsResp, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	require.NoError(t, err)
	require.Len(t, urlsResp.URL, 1)
	assert.Equal(t, "http://localhost:8080/abc123", urlsResp.URL[0].ShortURL)
	assert.Equal(t, "https://example.com/long", urlsResp.URL[0].OriginalURL)
}

func TestListUserURLsInvalidAuthorization(t *testing.T) {
	require.NoError(t, auth.SetSecret("grpc-test-secret"))

	client, cleanup := newTestClient(t, mockURLService{
		getUserURLs: func(ctx context.Context, userID string) ([]model.UserURL, error) {
			return nil, errors.New("should not be called")
		},
	})
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), authorizationHeader, "bad-token")
	_, err := client.ListUserURLs(ctx, &emptypb.Empty{})
	require.Error(t, err)
	assert.Equal(t, "Unauthenticated", status.Code(err).String())
}

func TestListUserURLsEmptyForNewUserAndNoURLs(t *testing.T) {
	require.NoError(t, auth.SetSecret("grpc-test-secret"))

	client, cleanup := newTestClient(t, mockURLService{
		getUserURLs: func(ctx context.Context, userID string) ([]model.UserURL, error) {
			return nil, repository.ErrNoUserURLs
		},
	})
	defer cleanup()

	var header metadata.MD
	resp, err := client.ListUserURLs(context.Background(), &emptypb.Empty{}, grpc.Header(&header))
	require.NoError(t, err)
	assert.Empty(t, resp.URL)
	assert.Len(t, header.Get(authorizationHeader), 1)
}

func TestExpandURLIssuesAuthHeaderForNewUser(t *testing.T) {
	require.NoError(t, auth.SetSecret("grpc-test-secret"))

	client, cleanup := newTestClient(t, mockURLService{
		getOriginalURL: func(ctx context.Context, id string) (string, error) {
			assert.Equal(t, "abc123", id)
			return "https://example.com/long", nil
		},
	})
	defer cleanup()

	var header metadata.MD
	resp, err := client.ExpandURL(context.Background(), &shortenerpb.URLExpandRequest{ID: "abc123"}, grpc.Header(&header))
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/long", resp.Result)
	assert.Len(t, header.Get(authorizationHeader), 1)
}

func newTestClient(t *testing.T, svc mockURLService) (shortenerpb.ShortenerServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(bufSize)
	server := grpc.NewServer()
	shortenerpb.RegisterShortenerServiceServer(server, New(svc))

	go func() {
		_ = server.Serve(listener)
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, address string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		server.Stop()
		listener.Close()
	}

	return shortenerpb.NewShortenerServiceClient(conn), cleanup
}
