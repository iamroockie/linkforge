package rest

import (
	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

const (
	FieldCodeUnsupportedScheme httpx.FieldCode = "unsupported_scheme"
)

func mapError(err error) error {
	if fieldErrors := httpx.MatchFieldErrors(err, fieldRules()); len(fieldErrors) > 0 {
		return httpx.ValidationError(fieldErrors)
	}

	return err
}

func fieldRules() []httpx.FieldRule {
	return []httpx.FieldRule{
		{Sentinel: link.ErrTTLNonPositive, Field: "ttl", Code: httpx.FieldCodeNonPositive},
		{Sentinel: link.ErrURLMalformed, Field: "url", Code: httpx.FieldCodeInvalidFormat},
		{Sentinel: link.ErrURLMissingHost, Field: "url", Code: httpx.FieldCodeInvalidFormat},
		{Sentinel: link.ErrURLRequired, Field: "url", Code: httpx.FieldCodeRequired},
		{Sentinel: link.ErrURLTooLong, Field: "url", Code: httpx.FieldCodeTooLong},
		{Sentinel: link.ErrURLUnsupportedScheme, Field: "url", Code: FieldCodeUnsupportedScheme},
	}
}
