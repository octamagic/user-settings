//go:generate mockgen -source=repository.go -destination=mock_repository.go -package=prefsrepo
package prefsrepo

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	ex "github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type SettingsRepository interface {
	PutSetting(ctx context.Context, setting Setting) error
	GetSettings(ctx context.Context, userUrn string, names []string) ([]Setting, error)
}

// settingsRepository stores settings in a single DynamoDB table keyed by
// user URN (pk) and setting name (sk).
type settingsRepository struct {
	db        *dynamodb.Client
	tableName string
}

// NewSettingsRepository creates a new SettingsRepository given a DynamoDB client and table name.
func NewSettingsRepository(dynamo *dynamodb.Client, tableName string) SettingsRepository {
	return &settingsRepository{db: dynamo, tableName: tableName}
}

func (sr *settingsRepository) PutSetting(ctx context.Context, setting Setting) error {
	item, err := attributevalue.MarshalMap(setting)
	if err != nil {
		return err
	}
	_, err = sr.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(sr.tableName),
		Item:      item,
	})
	return err
}

// GetSettings retrieves a setting by user URN and setting ID.
func (sr *settingsRepository) GetSettings(ctx context.Context, userUrn string, names []string) ([]Setting, error) {
	exprBuilder := ex.NewBuilder().WithKeyCondition(ex.Key(pkName).Equal(ex.Value(userUrn)))

	expr, err := exprBuilder.Build()
	if err != nil {
		return nil, err
	}

	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(sr.tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	res, err := sr.db.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	var settings []Setting
	if err = attributevalue.UnmarshalListOfMaps(res.Items, &settings); err != nil {
		return nil, err
	}

	if len(names) != 0 {
		settings = filterSettingsByName(settings, names)
	}

	return settings, nil
}

func filterSettingsByName(settings []Setting, names []string) []Setting {
	filteredSettings := make([]Setting, 0)
	for _, setting := range settings {
		if slices.Contains(names, setting.Name) {
			filteredSettings = append(filteredSettings, setting)
		}
	}
	return filteredSettings
}
