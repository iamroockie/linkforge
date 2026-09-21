package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"uuid"

	"github.com/joho/godotenv"

	"github.com/iamroockie/linkforge/internal/app/api"
	"github.com/iamroockie/linkforge/internal/platform/config"
	"github.com/iamroockie/linkforge/internal/platform/httpx/middleware"
	"github.com/iamroockie/linkforge/internal/platform/logger"
	"github.com/iamroockie/linkforge/internal/platform/pg"
)

func main() {
	if err := run(); err != nil {
		slog.Error("failed to run", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	slogHandler := logger.NewHandler(jsonHandler).WithContextExtractors(requestIDAttr)
	log := slog.New(slogHandler).With("env", cfg.Env)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pg.NewPool(ctx, cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	app, err := api.NewApp(pool, log, api.Options{
		RateLimit:      cfg.RateLimit,
		TrustedProxies: cfg.TrustedProxies,
	})
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	svr := &http.Server{
		Addr:              cfg.HTTP.Addr(),
		Handler:           api.NewRouter(app),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       time.Minute,
	}

	svrErr := make(chan error, 1)
	go func() {
		log.Info("http server running", "addr", svr.Addr)
		if err := svr.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			svrErr <- err
		}
	}()

	select {
	case err := <-svrErr:
		return fmt.Errorf("http serve: %w", err)
	case <-ctx.Done():
		stop()
	}

	log.Info("shutdown started")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := svr.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	log.Info("shutdown completed")

	return nil
}

func requestIDAttr(ctx context.Context) slog.Attr {
	if id := middleware.GetRequestID(ctx); id != uuid.Nil() {
		return slog.Any("request_id", id)
	}

	return slog.Attr{}
}
