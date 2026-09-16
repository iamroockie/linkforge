package rest

import (
	"time"

	"github.com/iamroockie/linkforge/internal/link"
)

type LinkResponse struct {
	Alias     string     `json:"alias"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func LinkResponseFromModel(l link.Link) LinkResponse {
	return LinkResponse{
		Alias:     l.Alias,
		URL:       l.URL,
		ExpiresAt: l.ExpiresAt,
	}
}
