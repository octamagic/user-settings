package appserver

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/patrickisgreat/user-settings/internal/config"
	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo/definitions"
	"github.com/patrickisgreat/user-settings/internal/prefsservice"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/patrickisgreat/user-settings/internal/regionalrulesrepo"
	"github.com/patrickisgreat/user-settings/internal/rulesservice"
	"github.com/patrickisgreat/user-settings/internal/telemetry"
	"github.com/patrickisgreat/user-settings/internal/userdirectory"
	db "github.com/patrickisgreat/user-settings/pkg/dynamodb"
)

type AppContext struct {
	DynamoDBClient          *dynamodb.Client
	UserSettingsRepository  prefsrepo.SettingsRepository
	RegionalRulesRepository regionalrulesrepo.RegionalRulesRepository
	UserSettingsService     prefsservice.SettingsService
	RulesService            rulesservice.RulesService
	UserDirectoryClient     *userdirectory.Client
}

var appctx *AppContext
var shutdownOtel func(context.Context) error
var initOnce sync.Once

// InitContext initializes the application context.
func InitContext(ctx context.Context) {
	initOnce.Do(func() {
		if err := config.Load(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		var err error
		shutdownOtel, err = telemetry.Setup(ctx, config.Config.AppName, config.Config.OTelSDKDisabled)
		if err != nil {
			log.Fatalf("Failed to set up otel: %v", err)
		}
		appctx = newContext(ctx)
	})
}

// Context returns the initialized application context.
func Context() *AppContext {
	if appctx == nil {
		log.Fatal("Server context was not initialized")
	}
	return appctx
}

// newContext creates a new application context
func newContext(ctx context.Context) *AppContext {
	ddbClient := db.NewDynamoDBClient(ctx, &db.DynamoDBConfig{
		Region:   config.Config.AWSRegion,
		Endpoint: config.Config.DynamoDBEndpoint,
		Log:      logging.FromContext(ctx),
	})

	// Create clients
	userDirectory := userdirectory.New(config.Config.UserDirectoryBaseURL)

	// Create repository
	settingsRepository := prefsrepo.NewSettingsRepository(ddbClient, config.Config.UserSettingsTableName)
	regionalRulesRepo := regionalrulesrepo.NewRegionalRulesRepository()

	// Create services
	rulesService := rulesservice.NewRulesService(regionalRulesRepo)
	minorSettingProcessor := domain.NewMinorSettingProcessor(userDirectory, rulesService, &domain.DefaultClock{}, settingsRepository)
	definitionsRepo := definitions.NewDefinitionsRepository()
	userSettingService := prefsservice.NewSettingsService(settingsRepository, minorSettingProcessor, definitionsRepo)

	return &AppContext{
		DynamoDBClient:          ddbClient,
		UserSettingsRepository:  settingsRepository,
		RegionalRulesRepository: regionalRulesRepo,
		UserSettingsService:     userSettingService,
		RulesService:            rulesService,
		UserDirectoryClient:     userDirectory,
	}
}
