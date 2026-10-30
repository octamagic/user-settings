// Package awsdynamo builds the DynamoDB client used by the repositories.
package awsdynamo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/tracing/smithyoteltracing"
	"go.opentelemetry.io/otel"
)

// Config selects the region and, for local development, a custom endpoint.
type Config struct {
	Region string
	// Endpoint, when set, points the client at DynamoDB Local with static
	// credentials instead of the default AWS credential chain.
	Endpoint string
}

// NewClient returns a DynamoDB client instrumented with OpenTelemetry.
func NewClient(ctx context.Context, cfg Config) (*dynamodb.Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{awsconfig.WithRegion(cfg.Region)}
	if cfg.Endpoint != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("local", "local", ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}
	return dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		// The SDK traces every operation through smithy-go; adapt the global
		// provider that telemetry.Setup installs.
		o.TracerProvider = smithyoteltracing.Adapt(otel.GetTracerProvider())
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	}), nil
}
