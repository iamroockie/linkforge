package logger_test

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iamroockie/linkforge/internal/platform/logger"
)

type (
	ctxKeyUser  struct{}
	ctxKeyTrace struct{}
)

func userAttr(ctx context.Context) slog.Attr {
	if user, ok := ctx.Value(ctxKeyUser{}).(string); ok {
		return slog.String("user", user)
	}

	return slog.Attr{}
}

func traceAttr(ctx context.Context) slog.Attr {
	if trace, ok := ctx.Value(ctxKeyTrace{}).(string); ok {
		return slog.String("trace_id", trace)
	}

	return slog.Attr{}
}

func constAttr(key string) logger.AttrFromContext {
	return func(context.Context) slog.Attr { return slog.Bool(key, true) }
}

func newJSONHandler() (slog.Handler, *bytes.Buffer) {
	var buf bytes.Buffer

	return slog.NewJSONHandler(&buf, nil), &buf
}

func decodeLogs(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	logs := make([]map[string]any, 0)
	for line := range strings.Lines(buf.String()) {
		var rec map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &rec))
		logs = append(logs, rec)
	}

	return logs
}

func TestHandler_ContextExtractors(t *testing.T) {
	jsonHandler, buf := newJSONHandler()
	log := slog.New(logger.NewHandler(jsonHandler).WithContextExtractors(userAttr, traceAttr))
	traceCtx := context.WithValue(t.Context(), ctxKeyTrace{}, "abc")
	fullCtx := context.WithValue(traceCtx, ctxKeyUser{}, "alice")

	log.InfoContext(fullCtx, "all attrs")
	log.InfoContext(traceCtx, "trace only")
	log.InfoContext(t.Context(), "no attrs")

	logs := decodeLogs(t, buf)
	require.Len(t, logs, 3)
	assert.Equal(t, "alice", logs[0]["user"])
	assert.Equal(t, "abc", logs[0]["trace_id"])
	assert.NotContains(t, logs[1], "user")
	assert.Equal(t, "abc", logs[1]["trace_id"])
	assert.NotContains(t, logs[2], "user")
	assert.NotContains(t, logs[2], "trace_id")
}

func TestHandler_WithoutExtractors(t *testing.T) {
	jsonHandler, buf := newJSONHandler()
	log := slog.New(logger.NewHandler(jsonHandler))

	log.InfoContext(t.Context(), "plain", "key", "value")

	logs := decodeLogs(t, buf)
	require.Len(t, logs, 1)
	assert.Equal(t, "plain", logs[0]["msg"])
	assert.Equal(t, "value", logs[0]["key"])
}

func TestHandler_WithAttrs(t *testing.T) {
	jsonHandler, buf := newJSONHandler()
	h := logger.NewHandler(jsonHandler).WithContextExtractors(userAttr)
	ctx := context.WithValue(t.Context(), ctxKeyUser{}, "alice")

	slog.New(h).With("env", "test").InfoContext(ctx, "with attrs")

	logs := decodeLogs(t, buf)
	require.Len(t, logs, 1)
	assert.Equal(t, "test", logs[0]["env"])
	assert.Equal(t, "alice", logs[0]["user"])
}

func TestHandler_WithGroup(t *testing.T) {
	jsonHandler, buf := newJSONHandler()
	h := logger.NewHandler(jsonHandler).WithContextExtractors(userAttr)
	ctx := context.WithValue(t.Context(), ctxKeyUser{}, "alice")

	slog.New(h).WithGroup("db").InfoContext(ctx, "grouped", "query", "select 1")

	logs := decodeLogs(t, buf)
	require.Len(t, logs, 1)
	group, ok := logs[0]["db"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "select 1", group["query"])
	assert.Equal(t, "alice", group["user"])
}

func TestHandler_WithContextExtractors_Immutable(t *testing.T) {
	jsonHandler, buf := newJSONHandler()
	base := logger.NewHandler(jsonHandler).
		WithContextExtractors(constAttr("a"), constAttr("b")).
		WithContextExtractors(constAttr("c"))
	left := base.WithContextExtractors(constAttr("left"))
	right := base.WithContextExtractors(constAttr("right"))

	slog.New(base).InfoContext(t.Context(), "base")
	slog.New(left).InfoContext(t.Context(), "left")
	slog.New(right).InfoContext(t.Context(), "right")

	logs := decodeLogs(t, buf)
	require.Len(t, logs, 3)
	for _, rec := range logs {
		assert.Contains(t, rec, "a")
		assert.Contains(t, rec, "b")
		assert.Contains(t, rec, "c")
	}
	assert.NotContains(t, logs[0], "left")
	assert.NotContains(t, logs[0], "right")
	assert.Contains(t, logs[1], "left")
	assert.NotContains(t, logs[1], "right")
	assert.Contains(t, logs[2], "right")
	assert.NotContains(t, logs[2], "left")
}

func TestHandler_Enabled(t *testing.T) {
	jsonHandler := slog.NewJSONHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn})
	h := logger.NewHandler(jsonHandler).WithContextExtractors(userAttr)

	assert.False(t, h.Enabled(t.Context(), slog.LevelInfo))
	assert.True(t, h.Enabled(t.Context(), slog.LevelWarn))
}
