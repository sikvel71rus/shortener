package starter

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
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
	EnableHTTPS     bool
}

type fileConfig struct {
	ServerAddress   *string `json:"server_address"`
	BaseURL         *string `json:"base_url"`
	FileStoragePath *string `json:"file_storage_path"`
	DatabaseDSN     *string `json:"database_dsn"`
	AuthSecret      *string `json:"auth_secret"`
	AuditFile       *string `json:"audit_file"`
	AuditURL        *string `json:"audit_url"`
	EnableHTTPS     *bool   `json:"enable_https"`
}

// Parse reads configuration values from a file, flags and environment variables.
func Parse() (Config, error) {
	cfg := defaultConfig()
	configPath := ""

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "address to run HTTP server")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base address for shortened URL")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database connection params")
	flag.StringVar(&cfg.AuthSecret, "secret", cfg.AuthSecret, "secret key for auth cookie signing")
	flag.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "file path for audit log receiver")
	flag.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "URL for remote audit log receiver")
	flag.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "enable HTTPS server")
	flag.StringVar(&configPath, "c", "", "config file path")
	flag.StringVar(&configPath, "config", "", "config file path")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return Config{}, err
	}

	flagsCfg := cfg
	definedFlags := visitedFlags()
	cfg = defaultConfig()

	if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
	}

	if configPath != "" {
		if err := applyFileConfig(&cfg, configPath); err != nil {
			return Config{}, err
		}
	}

	applyDefinedFlags(&cfg, flagsCfg, definedFlags)
	applyEnv(&cfg)

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		ServerAddress:   "localhost:8080",
		BaseURL:         "http://localhost:8080",
		FileStoragePath: "/tmp/url-storage.json",
		AuthSecret:      "secretkey",
	}
}

func visitedFlags() map[string]bool {
	definedFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		definedFlags[f.Name] = true
	})

	return definedFlags
}

func applyFileConfig(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}

	var fileCfg fileConfig
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return fmt.Errorf("parse config file %q: %w", path, err)
	}

	if fileCfg.ServerAddress != nil {
		cfg.ServerAddress = *fileCfg.ServerAddress
	}
	if fileCfg.BaseURL != nil {
		cfg.BaseURL = *fileCfg.BaseURL
	}
	if fileCfg.FileStoragePath != nil {
		cfg.FileStoragePath = *fileCfg.FileStoragePath
	}
	if fileCfg.DatabaseDSN != nil {
		cfg.DatabaseDSN = *fileCfg.DatabaseDSN
	}
	if fileCfg.AuthSecret != nil {
		cfg.AuthSecret = *fileCfg.AuthSecret
	}
	if fileCfg.AuditFile != nil {
		cfg.AuditFile = *fileCfg.AuditFile
	}
	if fileCfg.AuditURL != nil {
		cfg.AuditURL = *fileCfg.AuditURL
	}
	if fileCfg.EnableHTTPS != nil {
		cfg.EnableHTTPS = *fileCfg.EnableHTTPS
	}

	return nil
}

func applyDefinedFlags(cfg *Config, flagsCfg Config, definedFlags map[string]bool) {
	if definedFlags["a"] {
		cfg.ServerAddress = flagsCfg.ServerAddress
	}
	if definedFlags["b"] {
		cfg.BaseURL = flagsCfg.BaseURL
	}
	if definedFlags["f"] {
		cfg.FileStoragePath = flagsCfg.FileStoragePath
	}
	if definedFlags["d"] {
		cfg.DatabaseDSN = flagsCfg.DatabaseDSN
	}
	if definedFlags["secret"] {
		cfg.AuthSecret = flagsCfg.AuthSecret
	}
	if definedFlags["audit-file"] {
		cfg.AuditFile = flagsCfg.AuditFile
	}
	if definedFlags["audit-url"] {
		cfg.AuditURL = flagsCfg.AuditURL
	}
	if definedFlags["s"] {
		cfg.EnableHTTPS = flagsCfg.EnableHTTPS
	}
}

func applyEnv(cfg *Config) {
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

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		enableHTTPS, err := strconv.ParseBool(envEnableHTTPS)
		cfg.EnableHTTPS = err != nil || enableHTTPS
	}
}
