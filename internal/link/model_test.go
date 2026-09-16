package link_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/link"
	"github.com/iamroockie/linkforge/internal/link/linktest"
	"github.com/iamroockie/linkforge/internal/platform/clock"
)

func TestNewLink_OK_WithTTL(t *testing.T) {
	p := linktest.CreateLinkParams()

	start := clock.Now().Add(*p.TTL)
	l, err := link.NewLink(p)
	finish := clock.Now().Add(*p.TTL)

	require.NoError(t, err)
	assert.Len(t, l.Alias, 6)
	assert.Equal(t, p.URL, l.URL)
	require.NotNil(t, l.ExpiresAt)
	assert.GreaterOrEqual(t, *l.ExpiresAt, start)
	assert.LessOrEqual(t, *l.ExpiresAt, finish)
}

func TestNewLink_OK_WithoutTTL(t *testing.T) {
	p := link.CreateLinkParams{URL: "http://test"}

	l, err := link.NewLink(p)

	require.NoError(t, err)
	assert.Len(t, l.Alias, 6)
	assert.Equal(t, p.URL, l.URL)
	require.Nil(t, l.ExpiresAt)
}

func TestNewLink_ValidationError(t *testing.T) {
	validURL := "http://test"
	tests := map[string]struct {
		params  link.CreateLinkParams
		wantErr error
	}{
		"url empty": {
			params:  link.CreateLinkParams{URL: ""},
			wantErr: link.ErrURLRequired,
		},
		"url too long": {
			params:  link.CreateLinkParams{URL: strings.Repeat("u", 301)},
			wantErr: link.ErrURLTooLong,
		},
		"url invalid escape": {
			params:  link.CreateLinkParams{URL: "http://test%"},
			wantErr: link.ErrURLMalformed,
		},
		"url invalid port": {
			params:  link.CreateLinkParams{URL: "http://test:abc"},
			wantErr: link.ErrURLMalformed,
		},
		"url invalid control character": {
			params:  link.CreateLinkParams{URL: "http://te\nst"},
			wantErr: link.ErrURLMalformed,
		},
		"url empty scheme": {
			params:  link.CreateLinkParams{URL: "test.com"},
			wantErr: link.ErrURLUnsupportedScheme,
		},
		"url invalid scheme": {
			params:  link.CreateLinkParams{URL: "tcp://test"},
			wantErr: link.ErrURLUnsupportedScheme,
		},
		"url empty hostname": {
			params:  link.CreateLinkParams{URL: "http://:8080"},
			wantErr: link.ErrURLMissingHost,
		},
		"ttl zero": {
			params:  link.CreateLinkParams{URL: validURL, TTL: new(time.Duration)},
			wantErr: link.ErrTTLNonPositive,
		},
		"ttl negative": {
			params:  link.CreateLinkParams{URL: validURL, TTL: new(-time.Hour)},
			wantErr: link.ErrTTLNonPositive,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := link.NewLink(test.params)

			require.ErrorIs(t, err, test.wantErr)
			assert.Zero(t, got)
		})
	}
}
