package linktest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/link"
)

func CreateLinkParams() link.CreateLinkParams {
	return link.CreateLinkParams{URL: "http://test", TTL: new(24 * time.Hour)}
}

func Link(t testing.TB) link.Link {
	t.Helper()
	l, err := link.NewLink(CreateLinkParams())
	require.NoError(t, err)

	return l
}
