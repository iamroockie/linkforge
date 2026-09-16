package logger

import (
	"context"
	"log/slog"
	"slices"
)

type AttrFromContext func(ctx context.Context) slog.Attr

type Handler struct {
	next              slog.Handler
	contextExtractors []AttrFromContext
}

func NewHandler(next slog.Handler) *Handler {
	return &Handler{next: next}
}

func (h *Handler) WithContextExtractors(extractors ...AttrFromContext) *Handler {
	return &Handler{
		next:              h.next,
		contextExtractors: slices.Concat(h.contextExtractors, extractors),
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, rec slog.Record) error {
	rec = rec.Clone()
	for _, extract := range h.contextExtractors {
		if attr := extract(ctx); !attr.Equal(slog.Attr{}) {
			rec.AddAttrs(attr)
		}
	}

	return h.next.Handle(ctx, rec)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{next: h.next.WithAttrs(attrs), contextExtractors: h.contextExtractors}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{next: h.next.WithGroup(name), contextExtractors: h.contextExtractors}
}
