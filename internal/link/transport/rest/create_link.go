package rest

import (
	"net/http"
	"time"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

type CreateLinkRequest struct {
	URL string  `json:"url"`
	TTL *string `json:"ttl"`
}

func CreateLink(creator LinkCreator) http.Handler {
	return httpx.Handle(func(w http.ResponseWriter, r *http.Request) (*httpx.Response, error) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<10) // 1KB

		req, err := httpx.ParseRequestJSON[CreateLinkRequest](r)
		if err != nil {
			return nil, err
		}
		ttl, err := parseDuration(req.TTL)
		if err != nil {
			return nil, httpx.BadRequestError()
		}

		p := link.CreateLinkParams{
			URL: req.URL,
			TTL: ttl,
		}
		l, err := creator.Create(r.Context(), p)
		if err != nil {
			return nil, mapError(err)
		}

		return httpx.NewResponse(http.StatusCreated, LinkResponseFromModel(l)), nil
	})
}

func parseDuration(raw *string) (*time.Duration, error) {
	if raw == nil {
		return nil, nil
	}
	d, err := time.ParseDuration(*raw)
	if err != nil {
		return nil, err
	}
	return &d, nil
}
