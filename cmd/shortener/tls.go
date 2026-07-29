package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func newTLSListener(serverAddress string) (net.Listener, error) {
	tlsConfig, err := newTLSConfig(serverAddress)
	if err != nil {
		return nil, err
	}

	return tls.Listen("tcp", serverAddress, tlsConfig)
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
