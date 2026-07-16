package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func main() {
	printBuildInfo()

	starterCfg := starter.Parse()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	if err := auth.SetSecret(starterCfg.AuthSecret); err != nil {
		log.Fatalf("Ошибка инициализации секрета авторизации: %v", err)
	}

	var repo repository.URLRepo
	var err error

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

	defer repo.Close()

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

	log.Printf("Сервер запущен на %s, базовый адрес: %s", starterCfg.ServerAddress, starterCfg.BaseURL)

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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
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
