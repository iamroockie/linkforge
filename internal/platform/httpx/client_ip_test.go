package httpx_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func TestClientIPResolver(t *testing.T) {
	tests := map[string]struct {
		peer    string
		headers []string
		trusted []string
		want    string
	}{
		"direct": {
			peer: "192.0.2.1:123",
			want: "192.0.2.1",
		},
		"ipv6": {
			peer: "[2001:db8::1]:123",
			want: "2001:db8::1",
		},
		"mapped": {
			peer: "[::ffff:192.0.2.1]:123",
			want: "192.0.2.1",
		},
		"untrusted": {
			peer:    "192.0.2.1:123",
			headers: []string{"203.0.113.1"},
			want:    "192.0.2.1",
		},
		"proxy": {
			peer:    "10.0.0.1:123",
			headers: []string{"203.0.113.1"},
			trusted: []string{"10.0.0.0/8"},
			want:    "203.0.113.1",
		},
		"multiple": {
			peer:    "10.0.0.1:123",
			headers: []string{"192.0.2.1, 203.0.113.1", " 10.1.1.1 "},
			trusted: []string{"10.0.0.0/8"},
			want:    "203.0.113.1",
		},
		"all trusted": {
			peer:    "10.0.0.1:123",
			headers: []string{"10.1.1.1, 10.2.2.2"},
			trusted: []string{"10.0.0.0/8"},
			want:    "10.1.1.1",
		},
		"missing": {
			peer:    "10.0.0.1:123",
			trusted: []string{"10.0.0.0/8"},
			want:    "10.0.0.1",
		},
		"invalid prefix ignored": {
			peer:    "10.0.0.1:123",
			headers: []string{"bad, 203.0.113.1"},
			trusted: []string{"10.0.0.0/8"},
			want:    "203.0.113.1",
		},
		"invalid earlier header ignored": {
			peer:    "10.0.0.1:123",
			headers: []string{"bad", "203.0.113.1, 10.1.1.1"},
			trusted: []string{"10.0.0.0/8"},
			want:    "203.0.113.1",
		},
		"invalid header from untrusted peer ignored": {
			peer:    "192.0.2.1:123",
			headers: []string{"bad"},
			trusted: []string{"10.0.0.0/8"},
			want:    "192.0.2.1",
		},
		"mapped proxy": {
			peer:    "[::ffff:10.0.0.1]:123",
			headers: []string{"::ffff:192.0.2.1"},
			trusted: []string{"::ffff:10.0.0.0/104"},
			want:    "192.0.2.1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resolver, err := httpx.NewClientIPResolver(tt.trusted)
			require.NoError(t, err)
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.peer
			for _, header := range tt.headers {
				req.Header.Add("X-Forwarded-For", header)
			}
			ip, err := resolver.Resolve(req)
			require.NoError(t, err)
			assert.Equal(t, tt.want, ip.String())
		})
	}
}

func TestClientIPResolverInvalid(t *testing.T) {
	_, err := httpx.NewClientIPResolver([]string{"bad"})
	require.Error(t, err)
	var resolver httpx.ClientIPResolver
	req := httptest.NewRequest("GET", "/", nil)
	for _, remote := range []string{"invalid", "invalid:80"} {
		req.RemoteAddr = remote
		_, err = resolver.Resolve(req)
		require.Error(t, err)
	}
	req.RemoteAddr = "192.0.2.1:80"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	ip, err := resolver.Resolve(req)
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.1", ip.String())
}

func TestClientIPResolver_InvalidForwardedChain(t *testing.T) {
	tests := map[string][]string{
		"invalid rightmost":                         {"203.0.113.1, bad"},
		"empty header":                              {""},
		"empty element":                             {"203.0.113.1,"},
		"address with port":                         {"203.0.113.1:80"},
		"address with zone":                         {"fe80::1%eth0"},
		"invalid before trusted hop":                {"bad, 10.1.1.1"},
		"invalid earlier header behind trusted hop": {"bad", "10.1.1.1"},
	}
	for name, headers := range tests {
		t.Run(name, func(t *testing.T) {
			resolver, err := httpx.NewClientIPResolver([]string{"10.0.0.0/8"})
			require.NoError(t, err)
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "10.0.0.1:80"
			for _, header := range headers {
				req.Header.Add("X-Forwarded-For", header)
			}

			ip, err := resolver.Resolve(req)

			require.Error(t, err)
			assert.False(t, ip.IsValid())
		})
	}
}
