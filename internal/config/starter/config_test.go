package starter

import (
	"flag"
	"io"
	"os"
	"testing"
)

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
	} {
		t.Setenv(key, "")
	}

	for key, value := range env {
		t.Setenv(key, value)
	}

	return Parse()
}
