// Package httpserver wires the Twirp handler into an HTTP server with
// logging, tracing and a health check, and runs it until shutdown.
package httpserver

import (
	"fmt"
	"log/slog"
	"net/http"

	usersettingsv1 "github.com/octamagic/user-settings/internal/generated/usersettings/v1"
	"github.com/octamagic/user-settings/internal/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewRouter mounts the Twirp server under its path prefix with request logging
// and OpenTelemetry instrumentation, plus a /-/health endpoint.
func NewRouter(twirpServer usersettingsv1.TwirpServer, logger *slog.Logger) http.Handler {
	router := http.NewServeMux()
	router.Handle(
		twirpServer.PathPrefix(),
		otelhttp.NewHandler(
			logging.Middleware(logger)(twirpServer),
			"UserSettings",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		),
	)
	addHealthCheckRoute(router)
	return router
}

func addHealthCheckRoute(router *http.ServeMux) {
	router.HandleFunc("/-/health", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, "OK"); err != nil {
			http.Error(w, "health check failed", http.StatusInternalServerError)
		}
	})
}
