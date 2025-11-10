package db

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type DynamoDBConfig struct {
	Region   string
	Endpoint string
	Log      *slog.Logger
}

type DynamoDBRepository struct {
	TableName *string
	DB        *ddb.Client
}

type DynamoDBPrimaryKey struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk" default:""`
}

type resolverV2 struct{ config *DynamoDBConfig }

func (r *resolverV2) ResolveEndpoint(ctx context.Context, params ddb.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	if r.config.Endpoint != "" {
		u, err := url.Parse(r.config.Endpoint)
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	// fallback to default
	return ddb.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

func NewDynamoDBClient(ctx context.Context, config *DynamoDBConfig) *ddb.Client {
	var dynamodbConfig aws.Config
	if config.Endpoint == "" {
		dynamodbConfig = getAWSConfig(ctx, config)
		return ddb.NewFromConfig(dynamodbConfig)
	}

	dynamodbConfig = getLocalConfig(ctx, config)
	return ddb.NewFromConfig(dynamodbConfig, func(o *ddb.Options) {
		o.EndpointResolverV2 = &resolverV2{config: config}
	})
}

func NewDynamoDBRepository(tableName string, dynamo *ddb.Client) *DynamoDBRepository {
	return &DynamoDBRepository{
		TableName: aws.String(tableName),
		DB:        dynamo,
	}
}

func getLocalConfig(ctx context.Context, config *DynamoDBConfig) aws.Config {
	conf, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(config.Region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		config.Log.Error("Failed to load DynamoDB config", slog.Any("err", err))
		panic(err)
	}
	return conf
}

func getAWSConfig(ctx context.Context, config *DynamoDBConfig) aws.Config {
	conf, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(config.Region))
	otelaws.AppendMiddlewares(&conf.APIOptions)
	if err != nil {
		config.Log.Error("Failed to load DynamoDB config", slog.Any("err", err))
		panic(err)
	}
	return conf
}
