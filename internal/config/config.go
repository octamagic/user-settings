// Package config loads service configuration from environment variables.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config holds every runtime setting the service reads from its environment.
type Settings struct {
	AppName     string `envconfig:"APP_NAME" default:"user-settings"`
	Environment string `envconfig:"ENVIRONMENT" default:"local"`
	Host        string `envconfig:"HOST" default:"0.0.0.0"`
	Port        int    `envconfig:"PORT" default:"8000"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	LogPretty   bool   `envconfig:"LOG_PRETTY" default:"false"`

	// OTelSDKDisabled mirrors the standard OTEL_SDK_DISABLED variable; when
	// true no exporter is configured and tracing is a no-op.
	OTelSDKDisabled bool `envconfig:"OTEL_SDK_DISABLED" default:"false"`

	AWSRegion             string `envconfig:"AWS_REGION" required:"true"`
	DynamoDBEndpoint      string `envconfig:"DYNAMODB_ENDPOINT"`
	UserSettingsTableName string `envconfig:"USER_SETTINGS_TABLE_NAME" required:"true"`
	UserDirectoryBaseURL  string `envconfig:"USER_DIRECTORY_BASE_URL" required:"true"`
}

// Config is the process-wide configuration populated by Load.
var Config Settings

// Load reads the environment into Config.
func Load() error {
	if err := envconfig.Process("", &Config); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	return nil
}
