package appserver

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/soundcloud/gokit/v2/clients/okidoki"
	"github.com/soundcloud/gokit/v2/logger"
	"github.com/soundcloud/gokit/v2/opentelemetry/otelsetup"
	"github.com/soundcloud/user-settings/internal/config"
	"github.com/soundcloud/user-settings/internal/prefsrepo"
	"github.com/soundcloud/user-settings/internal/prefsrepo/definitions"
	"github.com/soundcloud/user-settings/internal/prefsservice"
	"github.com/soundcloud/user-settings/internal/prefsservice/domain"
	"github.com/soundcloud/user-settings/internal/regionalrulesrepo"
	"github.com/soundcloud/user-settings/internal/rulesservice"
	db "github.com/soundcloud/user-settings/pkg/dynamodb"
)

type AppContext struct {
	DynamoDBClient          *dynamodb.Client
	UserSettingsRepository  prefsrepo.SettingsRepository
	RegionalRulesRepository regionalrulesrepo.RegionalRulesRepository
	UserSettingsService     prefsservice.SettingsService
	RulesService            rulesservice.RulesService
	OkidokiClient           *okidoki.Client
}

var appctx *AppContext
var shutdownOtel func(context.Context) error
var initOnce sync.Once

// InitContext initializes the application context.
func InitContext(ctx context.Context) {
	initOnce.Do(func() {
		var err error
		shutdownOtel, err = otelsetup.SetupSDK(ctx, otelsetup.WithECSDetector())
		if err != nil {
			log.Fatalf("Failed to set up otel: %v", err)
		}
		config.Load(ctx)
		logger.Init(logger.LoggerConfig{Config: &config.Config.Base})
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
		Log:      logger.Get(ctx),
	})

	// Create clients
	okidokiClient := okidoki.New(config.Config.OkidokiBaseURL)

	// Create repository
	settingsRepository := prefsrepo.NewSettingsRepository(ddbClient, config.Config.UserSettingsTableName)
	regionalRulesRepo := regionalrulesrepo.NewRegionalRulesRepository()

	// Create services
	rulesService := rulesservice.NewRulesService(regionalRulesRepo)
	minorSettingProcessor := domain.NewMinorSettingProcessor(okidokiClient, rulesService, &domain.DefaultClock{}, settingsRepository)
	definitionsRepo := definitions.NewDefinitionsRepository()
	userSettingService := prefsservice.NewSettingsService(settingsRepository, minorSettingProcessor, definitionsRepo)

	return &AppContext{
		DynamoDBClient:          ddbClient,
		UserSettingsRepository:  settingsRepository,
		RegionalRulesRepository: regionalRulesRepo,
		UserSettingsService:     userSettingService,
		RulesService:            rulesService,
		OkidokiClient:           okidokiClient,
	}
}
