package httpx_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/httpx"
)

func TestReportError(t *testing.T) {
	err := errors.New("test")
	ctx := httpx.ReportErrorCtx(t.Context())

	httpx.ReportError(ctx, err)

	require.ErrorIs(t, httpx.GetReportedError(ctx), err)
}

func TestReportError_CtxWithoutReporter(t *testing.T) {
	err := errors.New("test")
	ctx := t.Context()

	httpx.ReportError(ctx, err)

	require.NoError(t, httpx.GetReportedError(ctx))
}

func TestReportError_NilError(t *testing.T) {
	ctx := httpx.ReportErrorCtx(t.Context())

	httpx.ReportError(ctx, nil)

	assert.NoError(t, httpx.GetReportedError(ctx))
}
