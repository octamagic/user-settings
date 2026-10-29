package appserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickisgreat/user-settings/internal/logging"
)

const Ten time.Duration = 10 * time.Second

type StartParams struct {
	Addr   string
	Router *http.ServeMux
}

func Start(ctx context.Context, sp *StartParams) {
	log := logging.FromContext(ctx)

	if appctx == nil {
		err := errors.New("server context must be initialized before start")
		log.Error("failed to start server", slog.Any("err", err))
		panic(err)
	}

	log.Info("Starting server on", "addr", sp.Addr)
	server := &http.Server{
		Addr:              sp.Addr,
		Handler:           sp.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error("Failed to start server", slog.Any("err", err))
			panic(err)
		}
	}()

	log.Info("Server started", "addr", sp.Addr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Wait for SIGINT or SIGTERM
	<-stop

	// Give server ten seconds to shutdown gracefully
	ctxWithTimeout, cancel := context.WithTimeout(ctx, Ten)
	defer cancel()

	if err := server.Shutdown(ctxWithTimeout); err != nil {
		log.Error("Failed to shut down server", slog.Any("err", err))
	}
	err := shutdownOtel(ctxWithTimeout)
	if err != nil {
		log.Error("Failed to shut down otel", slog.Any("err", err))
	}

	log.Info("Server stopped")
}
