package appserver

import (
	"fmt"
	"log/slog"
	"net/http"

	api "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/logging"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func NewRouter(s *api.TwirpServer, logger *slog.Logger) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle(
		(*s).PathPrefix(),
		otelhttp.NewHandler(
			logging.Middleware(logger)(*s),
			"UserSettings",
			otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
		),
	)

	return addHealthCheckRoute(router)
}

func addHealthCheckRoute(router *http.ServeMux) *http.ServeMux {
	router.Handle("/-/health", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := fmt.Fprint(w, "OK")
		if err != nil {
			// Check if the response writer implements the Error method
			if errWriter, ok := w.(interface{ Error(string, int) }); ok {
				errWriter.Error("health check failed", http.StatusInternalServerError)
			} else {
				http.Error(w, "health check failed", http.StatusInternalServerError)
			}
		}
	}))
	return router
}
