package link

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/iamroockie/linkforge/internal/platform/clock"
)

type Link struct {
	Alias     string
	URL       string
	ExpiresAt *time.Time
}

type CreateLinkParams struct {
	URL string
	TTL *time.Duration
}

func NewLink(params CreateLinkParams) (Link, error) {
	errs := make([]error, 0, 2)

	if err := validateURL(params.URL); err != nil {
		errs = append(errs, err)
	}

	var expiresAt *time.Time
	if params.TTL != nil {
		if *params.TTL <= 0 {
			errs = append(errs, ErrTTLNonPositive)
		}
		expiresAt = new(clock.Now().Add(*params.TTL))
	}

	if len(errs) > 0 {
		return Link{}, errors.Join(errs...)
	}

	alias, err := generateAlias()
	if err != nil {
		return Link{}, fmt.Errorf("generate alias: %w", err)
	}

	return Link{
		Alias:     alias,
		URL:       params.URL,
		ExpiresAt: expiresAt,
	}, nil
}

func validateURL(raw string) error {
	if raw == "" {
		return ErrURLRequired
	}
	if len(raw) > 300 {
		return ErrURLTooLong
	}
	u, err := url.Parse(raw)
	if err != nil {
		return ErrURLMalformed
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrURLUnsupportedScheme
	}
	if u.Hostname() == "" {
		return ErrURLMissingHost
	}

	return nil
}

func generateAlias() (string, error) {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	length := len(charset)
	for i, n := range buf {
		buf[i] = charset[int(n)%length]
	}

	return string(buf), nil
}
