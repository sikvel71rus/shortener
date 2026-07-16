package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
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

	tlsConfig, err := newTLSConfig(server.Addr)
	if err != nil {
		return err
	}

	listener, err := tls.Listen("tcp", server.Addr, tlsConfig)
	if err != nil {
		return err
	}

	return server.Serve(listener)
}

func newTLSConfig(serverAddress string) (*tls.Config, error) {
	cert, err := generateSelfSignedCertificate(serverAddress)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func generateSelfSignedCertificate(serverAddress string) (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Shortener"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	addCertificateHosts(&certTemplate, serverAddress)

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func addCertificateHosts(certTemplate *x509.Certificate, serverAddress string) {
	certTemplate.DNSNames = append(certTemplate.DNSNames, "localhost")
	certTemplate.IPAddresses = append(certTemplate.IPAddresses, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))

	host, _, err := net.SplitHostPort(serverAddress)
	if err != nil || host == "" {
		return
	}

	if ip := net.ParseIP(host); ip != nil {
		certTemplate.IPAddresses = append(certTemplate.IPAddresses, ip)
		return
	}

	certTemplate.DNSNames = append(certTemplate.DNSNames, host)
}
