// Package app composes the service from its configuration: external clients,
// repositories, domain services and the HTTP handler. It is the single place
// where concrete implementations meet the interfaces that consume them.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/bakins/twirpotel"
	"github.com/patrickisgreat/user-settings/internal/api"
	"github.com/patrickisgreat/user-settings/internal/awsdynamo"
	"github.com/patrickisgreat/user-settings/internal/config"
	usersettingsv1 "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/httpserver"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo/definitions"
	"github.com/patrickisgreat/user-settings/internal/prefsservice"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/patrickisgreat/user-settings/internal/regionalrulesrepo"
	"github.com/patrickisgreat/user-settings/internal/rulesservice"
	"github.com/patrickisgreat/user-settings/internal/telemetry"
	"github.com/patrickisgreat/user-settings/internal/userdirectory"
	"github.com/twitchtv/twirp"
)

// App is a fully wired service instance.
type App struct {
	Config   config.Config
	Logger   *slog.Logger
	DynamoDB *dynamodb.Client
	Settings prefsservice.SettingsService
	// Handler serves the Twirp API and the health check.
	Handler http.Handler

	shutdownTelemetry func(context.Context) error
}

// New builds an App from cfg. It returns an error rather than exiting so the
// caller decides how a failed start is reported.
func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	shutdownTelemetry, err := telemetry.Setup(ctx, cfg.AppName, cfg.OTelSDKDisabled)
	if err != nil {
		return nil, fmt.Errorf("setting up telemetry: %w", err)
	}

	ddb, err := awsdynamo.NewClient(ctx, awsdynamo.Config{Region: cfg.AWSRegion, Endpoint: cfg.DynamoDBEndpoint})
	if err != nil {
		return nil, errors.Join(fmt.Errorf("creating DynamoDB client: %w", err), shutdownTelemetry(ctx))
	}

	settingsRepo := prefsrepo.NewSettingsRepository(ddb, cfg.UserSettingsTableName)
	rulesService := rulesservice.NewRulesService(regionalrulesrepo.NewRegionalRulesRepository())
	minorProcessor := domain.NewMinorSettingProcessor(
		userdirectory.New(cfg.UserDirectoryBaseURL),
		rulesService,
		domain.DefaultClock{},
	)
	settings := prefsservice.NewSettingsService(settingsRepo, minorProcessor, definitions.NewDefinitionsRepository())

	twirpServer := usersettingsv1.NewUserSettingsServiceServer(
		api.NewServer(settings),
		twirp.WithServerInterceptors(twirpotel.ServerInterceptor()),
	)

	return &App{
		Config:            cfg,
		Logger:            logger,
		DynamoDB:          ddb,
		Settings:          settings,
		Handler:           httpserver.NewRouter(twirpServer, logger),
		shutdownTelemetry: shutdownTelemetry,
	}, nil
}

// Shutdown flushes telemetry. Call it after the HTTP server has stopped.
func (a *App) Shutdown(ctx context.Context) error {
	return a.shutdownTelemetry(ctx)
}
