package grpcserver

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/model"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/pkg/shortenerpb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const authorizationHeader = "authorization"

type URLService interface {
	GetOriginalURL(ctx context.Context, id string) (string, error)
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error)
}

type Server struct {
	srv   URLService
	audit audit.Publisher
}

func New(srv URLService, publishers ...audit.Publisher) *Server {
	var publisher audit.Publisher
	if len(publishers) > 0 {
		publisher = publishers[0]
	}

	return &Server{srv: srv, audit: publisher}
}

func (s *Server) ShortenURL(ctx context.Context, req *shortenerpb.URLShortenRequest) (*shortenerpb.URLShortenResponse, error) {
	if req == nil || strings.TrimSpace(req.Url) == "" {
		return nil, status.Error(codes.InvalidArgument, "empty url")
	}

	userID, err := ensureUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to issue auth token")
	}

	result, err := s.srv.ShortenURL(ctx, req.Url, userID)
	if err != nil && !errors.Is(err, repository.ErrConflict) {
		return nil, status.Error(codes.Internal, "failed to shorten url")
	}

	s.publishAuditEvent(ctx, audit.ActionShorten, userID, req.Url)

	return &shortenerpb.URLShortenResponse{Result: result}, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *shortenerpb.URLExpandRequest) (*shortenerpb.URLExpandResponse, error) {
	if req == nil || strings.TrimSpace(req.Id) == "" {
		return nil, status.Error(codes.InvalidArgument, "empty id")
	}

	result, err := s.srv.GetOriginalURL(ctx, req.Id)
	if err != nil {
		if errors.Is(err, repository.ErrDeleted) {
			return nil, status.Error(codes.FailedPrecondition, "url deleted")
		}
		return nil, status.Error(codes.NotFound, "url not found")
	}

	userID, _ := userIDFromMetadata(ctx)
	s.publishAuditEvent(ctx, audit.ActionFollow, userID, result)

	return &shortenerpb.URLExpandResponse{Result: result}, nil
}

func (s *Server) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*shortenerpb.UserURLsResponse, error) {
	userID, err := userIDFromMetadata(ctx)
	if err != nil {
		if errors.Is(err, errMissingAuthorization) {
			if _, _, issueErr := issueAuthHeader(ctx); issueErr != nil {
				return nil, status.Error(codes.Internal, "failed to issue auth token")
			}
			return &shortenerpb.UserURLsResponse{}, nil
		}
		return nil, status.Error(codes.Unauthenticated, "invalid authorization")
	}

	urls, err := s.srv.GetUserURLs(ctx, userID)
	if errors.Is(err, repository.ErrNoUserURLs) {
		return &shortenerpb.UserURLsResponse{}, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list user urls")
	}

	response := &shortenerpb.UserURLsResponse{Url: make([]*shortenerpb.URLData, 0, len(urls))}
	for _, item := range urls {
		response.Url = append(response.Url, &shortenerpb.URLData{
			ShortUrl:    item.ShortURL,
			OriginalUrl: item.OriginalURL,
		})
	}

	return response, nil
}

func ensureUserID(ctx context.Context) (string, error) {
	userID, err := userIDFromMetadata(ctx)
	if err == nil {
		return userID, nil
	}

	_, userID, issueErr := issueAuthHeader(ctx)
	return userID, issueErr
}

var errMissingAuthorization = errors.New("missing authorization")

func userIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingAuthorization
	}

	values := md.Get(authorizationHeader)
	if len(values) == 0 {
		return "", errMissingAuthorization
	}

	token := strings.TrimSpace(values[0])
	if token == "" {
		return "", errMissingAuthorization
	}
	token = strings.TrimPrefix(token, "Bearer ")

	return auth.ParseUserID(token)
}

func issueAuthHeader(ctx context.Context) (string, string, error) {
	token, userID, err := auth.NewToken()
	if err != nil {
		return "", "", err
	}

	if err := grpc.SetHeader(ctx, metadata.Pairs(authorizationHeader, token)); err != nil {
		return "", "", err
	}

	return token, userID, nil
}

func (s *Server) publishAuditEvent(ctx context.Context, action string, userID string, url string) {
	if s.audit == nil {
		return
	}

	event := audit.Event{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    userID,
		URL:       url,
	}

	if err := s.audit.Publish(ctx, event); err != nil {
		logger.Log.Error("failed to publish audit event",
			zap.String("action", action),
			zap.String("url", url),
			zap.Error(err),
		)
	}
}
