// Command api runs the user settings Twirp service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/octamagic/user-settings/internal/app"
	"github.com/octamagic/user-settings/internal/config"
	"github.com/octamagic/user-settings/internal/httpserver"
	"github.com/octamagic/user-settings/internal/logging"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("service exited with error", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := logging.New(cfg.LogLevel, cfg.LogPretty)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	service, err := app.New(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("starting app: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := service.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown failed", slog.Any("err", err))
		}
	}()

	return httpserver.Run(ctx, cfg.Addr(), service.Handler, logger)
}
