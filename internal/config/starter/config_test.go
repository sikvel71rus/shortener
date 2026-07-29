package starter

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "config-host:9090",
		"base_url": "https://config.example",
		"file_storage_path": "",
		"database_dsn": "postgres://config",
		"auth_secret": "config-secret",
		"audit_file": "/tmp/config-audit.log",
		"audit_url": "https://audit.example",
		"enable_https": true,
		"trusted_subnet": "192.168.1.0/24"
	}`)

	cfg := parseWithArgsAndEnv(t, []string{"shortener", "-c", configPath}, nil)

	if cfg.ServerAddress != "config-host:9090" {
		t.Fatalf("expected server address from config file, got %q", cfg.ServerAddress)
	}
	if cfg.BaseURL != "https://config.example" {
		t.Fatalf("expected base URL from config file, got %q", cfg.BaseURL)
	}
	if cfg.FileStoragePath != "" {
		t.Fatalf("expected empty file storage path from config file, got %q", cfg.FileStoragePath)
	}
	if cfg.DatabaseDSN != "postgres://config" {
		t.Fatalf("expected database DSN from config file, got %q", cfg.DatabaseDSN)
	}
	if cfg.AuthSecret != "config-secret" {
		t.Fatalf("expected auth secret from config file, got %q", cfg.AuthSecret)
	}
	if cfg.AuditFile != "/tmp/config-audit.log" {
		t.Fatalf("expected audit file from config file, got %q", cfg.AuditFile)
	}
	if cfg.AuditURL != "https://audit.example" {
		t.Fatalf("expected audit URL from config file, got %q", cfg.AuditURL)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("expected HTTPS to be enabled by config file")
	}
	if cfg.TrustedSubnet != "192.168.1.0/24" {
		t.Fatalf("expected trusted subnet from config file, got %q", cfg.TrustedSubnet)
	}
}

func TestParseConfigFileHasLowerPriorityThanFlags(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "config-host:9090",
		"base_url": "https://config.example",
		"enable_https": true,
		"trusted_subnet": "192.168.1.0/24"
	}`)

	cfg := parseWithArgsAndEnv(t, []string{
		"shortener",
		"-config", configPath,
		"-a", "flag-host:8080",
		"-b", "https://flag.example",
		"-s=false",
		"-t", "10.0.0.0/8",
	}, nil)

	if cfg.ServerAddress != "flag-host:8080" {
		t.Fatalf("expected server address from flag, got %q", cfg.ServerAddress)
	}
	if cfg.BaseURL != "https://flag.example" {
		t.Fatalf("expected base URL from flag, got %q", cfg.BaseURL)
	}
	if cfg.EnableHTTPS {
		t.Fatal("expected HTTPS to be disabled by explicit flag")
	}
	if cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Fatalf("expected trusted subnet from flag, got %q", cfg.TrustedSubnet)
	}
}

func TestParseConfigFileHasLowerPriorityThanEnv(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "config-host:9090",
		"base_url": "https://config.example",
		"enable_https": false,
		"trusted_subnet": "192.168.1.0/24"
	}`)

	cfg := parseWithArgsAndEnv(t, []string{"shortener", "-c", configPath}, map[string]string{
		"SERVER_ADDRESS": "env-host:7070",
		"BASE_URL":       "https://env.example",
		"ENABLE_HTTPS":   "true",
		"TRUSTED_SUBNET": "10.0.0.0/8",
	})

	if cfg.ServerAddress != "env-host:7070" {
		t.Fatalf("expected server address from env, got %q", cfg.ServerAddress)
	}
	if cfg.BaseURL != "https://env.example" {
		t.Fatalf("expected base URL from env, got %q", cfg.BaseURL)
	}
	if !cfg.EnableHTTPS {
		t.Fatal("expected HTTPS to be enabled by env")
	}
	if cfg.TrustedSubnet != "10.0.0.0/8" {
		t.Fatalf("expected trusted subnet from env, got %q", cfg.TrustedSubnet)
	}
}

func TestParseConfigPathEnv(t *testing.T) {
	configPath := writeConfigFile(t, `{
		"server_address": "config-host:9090"
	}`)

	cfg := parseWithArgsAndEnv(t, []string{"shortener"}, map[string]string{
		"CONFIG": configPath,
	})

	if cfg.ServerAddress != "config-host:9090" {
		t.Fatalf("expected server address from CONFIG file, got %q", cfg.ServerAddress)
	}
}

func TestParseEnableHTTPSFlag(t *testing.T) {
	cfg := parseWithArgsAndEnv(t, []string{"shortener", "-s"}, nil)

	if !cfg.EnableHTTPS {
		t.Fatal("expected HTTPS to be enabled by -s flag")
	}
}

func TestParseEnableHTTPSEnv(t *testing.T) {
	cfg := parseWithArgsAndEnv(t, []string{"shortener"}, map[string]string{
		"ENABLE_HTTPS": "true",
	})

	if !cfg.EnableHTTPS {
		t.Fatal("expected HTTPS to be enabled by ENABLE_HTTPS")
	}
}

func TestParseEnableHTTPSEnvFalse(t *testing.T) {
	cfg := parseWithArgsAndEnv(t, []string{"shortener", "-s"}, map[string]string{
		"ENABLE_HTTPS": "false",
	})

	if cfg.EnableHTTPS {
		t.Fatal("expected ENABLE_HTTPS=false to disable HTTPS")
	}
}

func parseWithArgsAndEnv(t *testing.T, args []string, env map[string]string) Config {
	t.Helper()

	oldCommandLine := flag.CommandLine
	oldArgs := os.Args
	t.Cleanup(func() {
		flag.CommandLine = oldCommandLine
		os.Args = oldArgs
	})

	flag.CommandLine = flag.NewFlagSet(args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	os.Args = args

	for _, key := range []string{
		"SERVER_ADDRESS",
		"BASE_URL",
		"FILE_STORAGE_PATH",
		"DATABASE_DSN",
		"AUTH_SECRET",
		"AUDIT_FILE",
		"AUDIT_URL",
		"ENABLE_HTTPS",
		"TRUSTED_SUBNET",
		"CONFIG",
	} {
		t.Setenv(key, "")
	}

	for key, value := range env {
		t.Setenv(key, value)
	}

	cfg, err := Parse()
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	return cfg
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	return path
}
