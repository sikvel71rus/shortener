package main

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"syscall"
	"testing"
)

func TestNewTLSConfig(t *testing.T) {
	cfg, err := newTLSConfig("127.0.0.1:0")
	if err != nil {
		t.Fatalf("newTLSConfig returned error: %v", err)
	}

	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected TLS min version %d, got %d", tls.VersionTLS12, cfg.MinVersion)
	}

	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected one certificate, got %d", len(cfg.Certificates))
	}

	cert, err := x509.ParseCertificate(cfg.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatalf("failed to parse generated certificate: %v", err)
	}

	if !containsIP(cert.IPAddresses, net.ParseIP("127.0.0.1")) {
		t.Fatal("expected certificate to contain 127.0.0.1")
	}
}

func containsIP(ips []net.IP, target net.IP) bool {
	for _, ip := range ips {
		if ip.Equal(target) {
			return true
		}
	}

	return false
}

func TestShutdownSignals(t *testing.T) {
	expected := []os.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT}

	for _, signal := range expected {
		if !containsSignal(shutdownSignals, signal) {
			t.Fatalf("expected shutdown signals to include %v", signal)
		}
	}
}

func containsSignal(signals []os.Signal, target os.Signal) bool {
	for _, signal := range signals {
		if signal == target {
			return true
		}
	}

	return false
}
