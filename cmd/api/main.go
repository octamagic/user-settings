package main

import (
	"context"
	"fmt"
	"net"

	"github.com/bakins/twirpotel"
	api "github.com/soundcloud/user-settings/internal/api"
	"github.com/soundcloud/user-settings/internal/appserver"
	"github.com/soundcloud/user-settings/internal/config"
	genapi "github.com/soundcloud/user-settings/internal/generated/user_settings/api"
	"github.com/soundcloud/user-settings/internal/handlers"
	"github.com/soundcloud/user-settings/internal/mappers"
	"github.com/twitchtv/twirp"
)

func main() {
	ctx := context.Background()

	// Initialize the app context (which sets up DynamoDB client, repository, and service)
	appserver.InitContext(ctx)

	// Initialize the API with the service from the context
	api.InitService(appserver.Context().UserSettingsService, mappers.ProtoMapper{})

	// Create the Twirp server for user settings
	ras := genapi.NewUserSettingsServiceServer(
		&handlers.UserSettingsService{},
		twirp.WithServerInterceptors(twirpotel.ServerInterceptor()),
	)

	router := appserver.NewRouter(&ras)

	appserver.Start(ctx, &appserver.StartParams{
		Router: router,
		Addr:   net.JoinHostPort(config.Config.Host, fmt.Sprintf("%d", config.Config.Port)),
	})
}
