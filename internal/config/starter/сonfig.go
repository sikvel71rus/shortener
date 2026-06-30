package starter

import (
	"flag"
	"os"
)

// Config stores runtime settings for the shortener server.
type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	AuthSecret      string
	AuditFile       string
	AuditURL        string
}

// Parse reads configuration values from flags and environment variables.
func Parse() Config {
	cfg := Config{}
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address to run HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "base address for shortened URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "/tmp/url-storage.json", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database connection params")
	flag.StringVar(&cfg.AuthSecret, "secret", "secretkey", "secret key for auth cookie signing")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "file path for audit log receiver")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "URL for remote audit log receiver")
	flag.Parse()

	if envServerAddr := os.Getenv("SERVER_ADDRESS"); envServerAddr != "" {
		cfg.ServerAddress = envServerAddr
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FileStoragePath = envFilePath
	}

	if envDBDSN := os.Getenv("DATABASE_DSN"); envDBDSN != "" {
		cfg.DatabaseDSN = envDBDSN
	}

	if envAuthSecret := os.Getenv("AUTH_SECRET"); envAuthSecret != "" {
		cfg.AuthSecret = envAuthSecret
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	return cfg
}
