package rest

import (
	"net/http"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

const AliasPathKey = "alias"

func RedirectLink(provider RedirectURLProvider) http.Handler {
	return httpx.Handle(func(w http.ResponseWriter, r *http.Request) (*httpx.Response, error) {
		target, err := provider.Get(r.Context(), r.PathValue(AliasPathKey))
		if err != nil {
			return nil, mapError(err)
		}

		//nolint:gosec // G710: Redirect is main goal
		http.Redirect(w, r, target, http.StatusFound)

		return nil, nil
	})
}
