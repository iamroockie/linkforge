package rest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func TestMapError_UnknownError(t *testing.T) {
	err := errors.New("unknown")

	e := mapError(err)

	require.ErrorIs(t, e, err)
}

func TestMapError_ValidationError(t *testing.T) {
	tests := map[string]struct {
		err       error
		wantField string
		wantCode  httpx.FieldCode
	}{
		"url required": {
			err:       link.ErrURLRequired,
			wantField: "url",
			wantCode:  httpx.FieldCodeRequired,
		},
		"url malformed": {
			err:       link.ErrURLMalformed,
			wantField: "url",
			wantCode:  httpx.FieldCodeInvalidFormat,
		},
		"url missing host": {
			err:       link.ErrURLMissingHost,
			wantField: "url",
			wantCode:  httpx.FieldCodeInvalidFormat,
		},
		"url unsupported scheme": {
			err:       link.ErrURLUnsupportedScheme,
			wantField: "url",
			wantCode:  FieldCodeUnsupportedScheme,
		},
		"url too long": {
			err:       link.ErrURLTooLong,
			wantField: "url",
			wantCode:  httpx.FieldCodeTooLong,
		},
		"ttl non-positive": {
			err:       link.ErrTTLNonPositive,
			wantField: "ttl",
			wantCode:  httpx.FieldCodeNonPositive,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := mapError(test.err)

			e, ok := errors.AsType[httpx.Error](err)
			require.True(t, ok)
			require.Len(t, e.Details, 1)
			assert.Equal(t, test.wantField, e.Details[0].Field)
			assert.Equal(t, test.wantCode, e.Details[0].Code)
		})
	}
}
