package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/grpcserver"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/middleware"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
	"github.com/sikvel71rus/shortener.git/pkg/shortenerpb"
	"google.golang.org/grpc"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	serverCount     = 2
	shutdownTimeout = 10 * time.Second
)

// shutdownSignals is a slice because signal.NotifyContext accepts variadic values.
var shutdownSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT}

func main() {
	printBuildInfo()

	starterCfg, err := starter.Parse()
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации: %v", err)
	}

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	if err := auth.SetSecret(starterCfg.AuthSecret); err != nil {
		log.Fatalf("Ошибка инициализации секрета авторизации: %v", err)
	}

	var repo repository.URLRepo
	if starterCfg.DatabaseDSN != "" {
		repo, err = repository.NewPostgresRepo(starterCfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Ошибка инициализации БД: %v", err)
		}
		log.Println("Используется хранилище: PostgreSQL")

	} else if starterCfg.FileStoragePath != "" {
		repo, err = repository.NewMapURLRepo(starterCfg.FileStoragePath)
		if err != nil {
			log.Fatalf("Ошибка инициализации файлового хранилища: %v", err)
		}
		log.Println("Используется хранилище: Файл")

	} else {
		repo, err = repository.NewMapURLRepo("")
		if err != nil {
			log.Fatalf("Ошибка инициализации in-memory хранилища: %v", err)
		}
		log.Println("Используется хранилище: In-Memory")
	}

	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Ошибка закрытия хранилища: %v", err)
		}
	}()

	srv := service.NewURLService(repo, starterCfg.BaseURL)
	defer srv.Close()

	auditObservers := make([]audit.Observer, 0, 2)
	if observer := audit.NewFileObserver(starterCfg.AuditFile); observer != nil {
		auditObservers = append(auditObservers, observer)
	}

	httpObserver, err := audit.NewHTTPObserver(starterCfg.AuditURL)
	if err != nil {
		log.Fatalf("Ошибка инициализации HTTP-аудита: %v", err)
	}
	if httpObserver != nil {
		auditObservers = append(auditObservers, httpObserver)
	}

	auditPublisher := audit.NewBroadcaster(auditObservers...)
	h := handler.NewURLHandlerWithTrustedSubnet(srv, starterCfg.TrustedSubnet, auditPublisher)
	grpcServer := grpc.NewServer()
	shortenerpb.RegisterShortenerServiceServer(grpcServer, grpcserver.New(srv, auditPublisher))

	r := chi.NewRouter()

	protocol := "HTTP"
	if starterCfg.EnableHTTPS {
		protocol = "HTTPS"
	}
	log.Printf("HTTP-сервер запущен на %s по %s, базовый адрес: %s", starterCfg.ServerAddress, protocol, starterCfg.BaseURL)
	log.Printf("gRPC-сервер запущен на %s", starterCfg.GRPCAddress)

	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.PostURLHandler)
	r.Post("/api/shorten", h.ShortenJSONHandler)
	r.Get("/api/internal/stats", h.StatsHandler)
	r.Get("/{id}", h.GetURLHandler)
	r.Get("/ping", h.PingHandler)
	r.Post("/api/shorten/batch", h.BatchHandler)
	r.Get("/api/user/urls", h.UserURLsHandler)
	r.Delete("/api/user/urls", h.DeleteUserURLsHandler)

	server := &http.Server{
		Addr:    starterCfg.ServerAddress,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), shutdownSignals...)
	defer stop()

	serverErrCh := make(chan serverError, serverCount)
	go func() {
		serverErrCh <- serverError{name: "HTTP", err: serve(server, starterCfg.EnableHTTPS)}
	}()
	go func() {
		serverErrCh <- serverError{name: "gRPC", err: serveGRPC(grpcServer, starterCfg.GRPCAddress, starterCfg.EnableHTTPS)}
	}()

	select {
	case serverErr := <-serverErrCh:
		if !isExpectedServerError(serverErr.err) {
			log.Fatalf("Ошибка запуска %s-сервера: %v", serverErr.name, serverErr.err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Ошибка graceful shutdown: %v", err)
			if closeErr := server.Close(); closeErr != nil {
				log.Printf("Ошибка принудительной остановки сервера: %v", closeErr)
			}
		}

		grpcStopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(grpcStopped)
		}()
		select {
		case <-grpcStopped:
		case <-shutdownCtx.Done():
			grpcServer.Stop()
		}

		for i := 0; i < serverCount; i++ {
			serverErr := <-serverErrCh
			if !isExpectedServerError(serverErr.err) {
				log.Printf("Ошибка остановки %s-сервера: %v", serverErr.name, serverErr.err)
			}
		}
	}
}

type serverError struct {
	name string
	err  error
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildValue(buildVersion))
	fmt.Printf("Build date: %s\n", buildValue(buildDate))
	fmt.Printf("Build commit: %s\n", buildValue(buildCommit))
}

func buildValue(value string) string {
	if value == "" {
		return "N/A"
	}

	return value
}

func serve(server *http.Server, enableHTTPS bool) error {
	if !enableHTTPS {
		return server.ListenAndServe()
	}

	listener, err := newTLSListener(server.Addr)
	if err != nil {
		return err
	}

	return server.Serve(listener)
}

func serveGRPC(server *grpc.Server, address string, enableHTTPS bool) error {
	if enableHTTPS {
		listener, err := newTLSListener(address)
		if err != nil {
			return err
		}

		return server.Serve(listener)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return server.Serve(listener)
}

func isExpectedServerError(err error) bool {
	return err == nil || errors.Is(err, http.ErrServerClosed) || errors.Is(err, grpc.ErrServerStopped)
}
