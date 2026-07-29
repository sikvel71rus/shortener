package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/sikvel71rus/shortener.git/internal/audit"
	"github.com/sikvel71rus/shortener.git/internal/auth"
	"github.com/sikvel71rus/shortener.git/internal/config/starter"
	"github.com/sikvel71rus/shortener.git/internal/handler"
	"github.com/sikvel71rus/shortener.git/internal/logger"
	"github.com/sikvel71rus/shortener.git/internal/middleware"
	"github.com/sikvel71rus/shortener.git/internal/repository"
	"github.com/sikvel71rus/shortener.git/internal/service"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

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

	h := handler.NewURLHandler(srv, audit.NewBroadcaster(auditObservers...))

	r := chi.NewRouter()

	protocol := "HTTP"
	if starterCfg.EnableHTTPS {
		protocol = "HTTPS"
	}
	log.Printf("Сервер запущен на %s по %s, базовый адрес: %s", starterCfg.ServerAddress, protocol, starterCfg.BaseURL)

	r.Use(logger.RequestLogger)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.PostURLHandler)
	r.Post("/api/shorten", h.ShortenJSONHandler)
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

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- serve(server, starterCfg.EnableHTTPS)
	}()

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	case <-ctx.Done():
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Ошибка graceful shutdown: %v", err)
		}

		if err := <-serverErrCh; err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка остановки сервера: %v", err)
		}
	}
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
