// Package config loads service configuration from environment variables.
package config

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

// Config holds every runtime setting the service reads from its environment.
type Config struct {
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

// Load reads the process environment into a Config.
func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("loading config: %w", err)
	}
	return cfg, nil
}

// Addr returns the host:port the HTTP server should listen on.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
