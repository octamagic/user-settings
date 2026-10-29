package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/bakins/twirpotel"
	api "github.com/patrickisgreat/user-settings/internal/api"
	"github.com/patrickisgreat/user-settings/internal/appserver"
	"github.com/patrickisgreat/user-settings/internal/config"
	genapi "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/handlers"
	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/mappers"
	"github.com/twitchtv/twirp"
)

func main() {
	ctx := context.Background()

	// Initialize the app context (which sets up DynamoDB client, repository, and service)
	appserver.InitContext(ctx)

	logger := logging.New(config.Config.LogLevel, config.Config.LogPretty)
	slog.SetDefault(logger)
	ctx = logging.WithContext(ctx, logger)

	// Initialize the API with the service from the context
	api.InitService(appserver.Context().UserSettingsService, mappers.ProtoMapper{})

	// Create the Twirp server for user settings
	ras := genapi.NewUserSettingsServiceServer(
		&handlers.UserSettingsService{},
		twirp.WithServerInterceptors(twirpotel.ServerInterceptor()),
	)

	router := appserver.NewRouter(&ras, logger)

	appserver.Start(ctx, &appserver.StartParams{
		Router: router,
		Addr:   net.JoinHostPort(config.Config.Host, fmt.Sprintf("%d", config.Config.Port)),
	})
}
