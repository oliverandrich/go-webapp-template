// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"golang.org/x/crypto/acme/autocert"
)

// TLSMode represents the resolved TLS mode.
type TLSMode string

const (
	TLSModeOff        TLSMode = "off"
	TLSModeACME       TLSMode = "acme"
	TLSModeSelfSigned TLSMode = "selfsigned"
	TLSModeManual     TLSMode = "manual"
)

// TLSResult contains the resolved TLS configuration.
type TLSResult struct {
	TLSConfig   *tls.Config
	CertManager *autocert.Manager // nil unless ACME mode
	HTTPHandler http.Handler      // For HTTP→HTTPS redirect (ACME only)
	Mode        TLSMode
}

// SetupTLS configures TLS based on the configuration.
func SetupTLS(cfg *config.Config) (*TLSResult, error) {
	mode := resolveTLSMode(cfg)

	switch mode {
	case TLSModeOff:
		slog.Info("TLS mode: localhost (no TLS required)")
		return &TLSResult{Mode: TLSModeOff}, nil

	case TLSModeACME:
		// Validate ACME requirements
		if err := validateACME(cfg); err != nil {
			return nil, err
		}
		slog.Info("TLS mode: acme (Let's Encrypt)",
			"host", cfg.Server.Host,
			"email", cfg.TLS.Email,
		)
		return setupACME(cfg)

	case TLSModeSelfSigned:
		slog.Info("TLS mode: selfsigned")
		return setupSelfSigned(cfg)

	case TLSModeManual:
		slog.Info("TLS mode: manual",
			"cert", cfg.TLS.CertFile,
			"key", cfg.TLS.KeyFile,
		)
		return setupManual(cfg)

	default:
		return nil, fmt.Errorf("unknown TLS mode: %s", mode)
	}
}

// resolveTLSMode determines the best TLS mode based on configuration and environment.
func resolveTLSMode(cfg *config.Config) TLSMode {
	host := cfg.Server.Host
	mode := strings.ToLower(cfg.TLS.Mode)

	// Explicit mode takes precedence
	switch mode {
	case "off":
		return TLSModeOff
	case "acme":
		return TLSModeACME
	case "selfsigned":
		return TLSModeSelfSigned
	case "manual":
		return TLSModeManual
	case "auto", "":
		// Fall through to auto-detection
	default:
		slog.Warn("unknown TLS mode, using auto", "mode", mode)
	}

	// Auto-detection logic
	if config.IsLocalhost(host) {
		return TLSModeOff
	}

	// If cert files are provided, use manual mode
	if cfg.TLS.CertFile != "" && cfg.TLS.KeyFile != "" {
		return TLSModeManual
	}

	if canUseACME(cfg) {
		return TLSModeACME
	}

	return TLSModeSelfSigned
}

// validateACME checks requirements when ACME mode is explicitly selected.
func validateACME(cfg *config.Config) error {
	// Warn if configured port is not 443
	if cfg.Server.Port != 443 {
		slog.Warn("ACME mode uses port 443, configured port will be ignored",
			"configured_port", cfg.Server.Port,
		)
	}

	// Validate email is provided
	if cfg.TLS.Email == "" {
		return fmt.Errorf("ACME mode requires TLS_EMAIL to be set")
	}

	// Check if port 80 is available (required for HTTP-01 challenge)
	if !isPortAvailable(80) {
		return fmt.Errorf("ACME mode requires port 80 for HTTP-01 challenge (port in use)")
	}

	// Check if port 443 is available
	if !isPortAvailable(443) {
		return fmt.Errorf("ACME mode requires port 443 for HTTPS (port in use)")
	}

	return nil
}

// canUseACME checks if ACME mode is available (for auto-detection).
func canUseACME(cfg *config.Config) bool {
	host := cfg.Server.Host

	// Must not be localhost
	if config.IsLocalhost(host) {
		return false
	}

	// Must not be an IP address (Let's Encrypt doesn't issue certs for IPs)
	if net.ParseIP(host) != nil {
		slog.Debug("ACME disabled: host is an IP address")
		return false
	}

	// Must have ACME email configured
	if cfg.TLS.Email == "" {
		slog.Debug("ACME disabled: no email configured")
		return false
	}

	// Check if port 80 is available (required for HTTP-01 challenge)
	if !isPortAvailable(80) {
		slog.Debug("ACME disabled: port 80 not available")
		return false
	}

	// Check if port 443 is available
	if !isPortAvailable(443) {
		slog.Debug("ACME disabled: port 443 not available")
		return false
	}

	return true
}

// isPortAvailable checks if a port is available for binding.
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	lc := &net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// setupACME configures Let's Encrypt with autocert.
func setupACME(cfg *config.Config) (*TLSResult, error) {
	certDir := filepath.Join(cfg.TLS.CertDir, "acme")
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create ACME cert directory: %w", err)
	}

	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Email:      cfg.TLS.Email,
		Cache:      autocert.DirCache(certDir),
		HostPolicy: autocert.HostWhitelist(cfg.Server.Host),
	}

	tlsConfig := manager.TLSConfig()
	tlsConfig.MinVersion = tls.VersionTLS12

	slog.Info("Using Let's Encrypt for domain", "host", cfg.Server.Host)

	return &TLSResult{
		Mode:        TLSModeACME,
		TLSConfig:   tlsConfig,
		CertManager: manager,
		HTTPHandler: manager.HTTPHandler(nil),
	}, nil
}

// setupSelfSigned generates or loads a self-signed certificate.
func setupSelfSigned(cfg *config.Config) (*TLSResult, error) {
	certDir := filepath.Join(cfg.TLS.CertDir, "selfsigned")
	if err := os.MkdirAll(certDir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create self-signed cert directory: %w", err)
	}

	certFile := filepath.Join(certDir, "cert.pem")
	keyFile := filepath.Join(certDir, "key.pem")

	// Check if cert exists and is valid
	if certExists(certFile, keyFile) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err == nil && !isCertExpiringSoon(&cert) {
			slog.Info("Using existing self-signed certificate")
			logCertFingerprint(&cert)
			logSelfSignedWarning()
			return &TLSResult{
				Mode:      TLSModeSelfSigned,
				TLSConfig: createTLSConfig(&cert),
			}, nil
		}
		if err != nil {
			slog.Warn("existing certificate invalid, generating new one", "error", err)
		} else {
			slog.Info("existing certificate expiring soon, generating new one")
		}
	}

	// Generate new certificate
	slog.Info("Generating new self-signed certificate")
	cert, err := generateSelfSignedCert(cfg, certFile, keyFile)
	if err != nil {
		return nil, err
	}

	logCertFingerprint(cert)
	logSelfSignedWarning()

	return &TLSResult{
		Mode:      TLSModeSelfSigned,
		TLSConfig: createTLSConfig(cert),
	}, nil
}

// setupManual loads user-provided certificate files.
func setupManual(cfg *config.Config) (*TLSResult, error) {
	certFile := cfg.TLS.CertFile
	keyFile := cfg.TLS.KeyFile

	// Validate that both files are provided
	if certFile == "" || keyFile == "" {
		return nil, fmt.Errorf("manual TLS mode requires both cert-file and key-file")
	}

	// Check if files exist
	if _, err := os.Stat(certFile); err != nil {
		return nil, fmt.Errorf("certificate file not found: %w", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		return nil, fmt.Errorf("key file not found: %w", err)
	}

	// Load certificate
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	slog.Info("Using manual certificate", "cert", certFile, "key", keyFile)
	logCertFingerprint(&cert)

	return &TLSResult{
		Mode:      TLSModeManual,
		TLSConfig: createTLSConfig(&cert),
	}, nil
}

// generateSelfSignedCert creates a new self-signed certificate with ECDSA P-256.
func generateSelfSignedCert(cfg *config.Config, certFile, keyFile string) (*tls.Certificate, error) {
	// Generate ECDSA P-256 key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Self-Signed"},
			CommonName:   cfg.Server.Host,
		},
		NotBefore:             now,
		NotAfter:              now.Add(365 * 24 * time.Hour), // 1 year validity
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs (Subject Alternative Names)
	host := cfg.Server.Host
	if ip := net.ParseIP(host); ip != nil {
		template.IPAddresses = []net.IP{ip}
	} else {
		template.DNSNames = []string{host}
	}

	// Also add localhost for local access
	template.DNSNames = append(template.DNSNames, "localhost")
	template.IPAddresses = append(template.IPAddresses, net.ParseIP("127.0.0.1"), net.ParseIP("::1"))

	// Generate certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write cert file
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	if certWriteErr := os.WriteFile(certFile, certPEM, 0o600); certWriteErr != nil {
		return nil, fmt.Errorf("failed to write cert file: %w", certWriteErr)
	}

	// Write key file
	keyDER, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if writeErr := os.WriteFile(keyFile, keyPEM, 0o600); writeErr != nil {
		return nil, fmt.Errorf("failed to write key file: %w", writeErr)
	}

	// Load and return
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load generated cert: %w", err)
	}

	return &cert, nil
}

// certExists checks if both cert and key files exist.
func certExists(certFile, keyFile string) bool {
	_, certErr := os.Stat(certFile)
	_, keyErr := os.Stat(keyFile)
	return certErr == nil && keyErr == nil
}

// isCertExpiringSoon checks if certificate expires within 30 days.
func isCertExpiringSoon(cert *tls.Certificate) bool {
	if len(cert.Certificate) == 0 {
		return true
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return true
	}
	return time.Until(x509Cert.NotAfter) < 30*24*time.Hour
}

// logCertFingerprint logs the SHA256 fingerprint of the certificate.
func logCertFingerprint(cert *tls.Certificate) {
	if len(cert.Certificate) == 0 {
		return
	}
	fingerprint := sha256.Sum256(cert.Certificate[0])
	// Format as colon-separated hex string
	hexParts := make([]string, len(fingerprint))
	for i, b := range fingerprint {
		hexParts[i] = fmt.Sprintf("%02X", b)
	}
	slog.Info("Certificate fingerprint", "sha256", strings.Join(hexParts, ":"))
}

// logSelfSignedWarning logs a warning about accepting the certificate.
func logSelfSignedWarning() {
	slog.Warn("Accept the certificate in your browser on first visit")
}

// createTLSConfig creates a TLS config with the given certificate.
func createTLSConfig(cert *tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion:   tls.VersionTLS12,
	}
}
