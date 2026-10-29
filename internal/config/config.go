package config

import (
	"context"
	"log/slog"

	"github.com/soundcloud/gokit/v2/baseconfig"
	"github.com/soundcloud/gokit/v2/logger"
)

type config struct {
	baseconfig.Base

	AWSRegion             string `envconfig:"AWS_REGION" required:"true"`
	DynamoDBEndpoint      string `envconfig:"DYNAMODB_ENDPOINT"`
	UserSettingsTableName string `envconfig:"USER_SETTINGS_TABLE_NAME" required:"true"`
	OkidokiBaseURL        string `envconfig:"OKIDOKI_BASE_URL" required:"true"`
}

var Config config

func Load(ctx context.Context) {
	err := baseconfig.Load(&Config)
	if err != nil {
		logger.Get(ctx).Error("failed loading app config", slog.Any("err", err))
		panic("failed loading app config")
	}
}
