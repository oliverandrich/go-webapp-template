// Copyright 2025 Oliver Andrich
// Licensed under the EUPL-1.2

package email

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"codeberg.org/oliverandrich/go-webapp-template/internal/config"
	"codeberg.org/oliverandrich/go-webapp-template/internal/i18n"
	"github.com/wneessen/go-mail"
)

const (
	// TokenLength is the number of random bytes for verification tokens.
	TokenLength = 32
	// TokenExpiry is how long verification tokens are valid.
	TokenExpiry = 24 * time.Hour
)

// Service handles email sending and verification token management.
type Service struct {
	cfg     *config.SMTPConfig
	baseURL string
}

// NewService creates a new email service.
func NewService(cfg *config.SMTPConfig, baseURL string) (*Service, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("SMTP host is required")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("SMTP from address is required")
	}

	return &Service{
		cfg:     cfg,
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}, nil
}

// GenerateToken generates a new verification token.
// Returns (plaintext token, SHA256 hash for storage, expiry time, error).
func (s *Service) GenerateToken() (string, string, time.Time, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	plaintext := hex.EncodeToString(bytes)
	hash := HashToken(plaintext)
	expiresAt := time.Now().Add(TokenExpiry)

	return plaintext, hash, expiresAt, nil
}

// HashToken computes the SHA256 hash of a token.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// SendVerification sends a verification email with the given token.
func (s *Service) SendVerification(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.baseURL, token)

	subject := i18n.T(ctx, "email_verification_subject")
	body := i18n.TData(ctx, "email_verification_body", map[string]any{
		"VerifyURL": verifyURL,
	})

	return s.send(toEmail, subject, body)
}

// send sends an email via SMTP using go-mail.
func (s *Service) send(to, subject, body string) error {
	msg := mail.NewMsg()

	if s.cfg.FromName != "" {
		if err := msg.FromFormat(s.cfg.FromName, s.cfg.From); err != nil {
			return fmt.Errorf("setting from address: %w", err)
		}
	} else {
		if err := msg.From(s.cfg.From); err != nil {
			return fmt.Errorf("setting from address: %w", err)
		}
	}

	if err := msg.To(to); err != nil {
		return fmt.Errorf("setting to address: %w", err)
	}

	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	// Build client options
	opts := []mail.Option{
		mail.WithPort(s.cfg.Port),
	}

	// Configure TLS based on config and port
	if s.cfg.TLS {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
		// Use implicit TLS (SSL) for port 465, STARTTLS for others
		if s.cfg.Port == 465 {
			opts = append(opts, mail.WithSSL())
		}
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.NoTLS))
	}

	// Add authentication if credentials are provided
	if s.cfg.Username != "" && s.cfg.Password != "" {
		opts = append(opts,
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(s.cfg.Username),
			mail.WithPassword(s.cfg.Password),
		)
	}

	client, err := mail.NewClient(s.cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("creating mail client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("sending email: %w", err)
	}

	return nil
}
