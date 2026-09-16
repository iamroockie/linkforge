package httpx

import (
	"context"
	"errors"
	"sync"
)

type ctxKeyErrorReporter struct{}

type errorReporter struct {
	mu   sync.RWMutex
	errs []error
}

func ReportErrorCtx(ctx context.Context) context.Context {
	if getErrorReporter(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKeyErrorReporter{}, &errorReporter{})
}

func ReportError(ctx context.Context, err error) {
	rep := getErrorReporter(ctx)
	if rep == nil || err == nil {
		return
	}
	rep.mu.Lock()
	defer rep.mu.Unlock()
	rep.errs = append(rep.errs, err)
}

func GetReportedError(ctx context.Context) error {
	rep := getErrorReporter(ctx)
	if rep == nil {
		return nil
	}
	rep.mu.RLock()
	defer rep.mu.RUnlock()
	return errors.Join(rep.errs...)
}

func getErrorReporter(ctx context.Context) *errorReporter {
	rep, ok := ctx.Value(ctxKeyErrorReporter{}).(*errorReporter)
	if !ok {
		return nil
	}
	return rep
}
