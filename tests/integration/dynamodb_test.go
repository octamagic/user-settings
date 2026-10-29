package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/stretchr/testify/assert"
)

var (
	testTableName       = "user-settings-test"
	testEndpoint        = "http://localhost:9000" // DynamoDB Local endpoint
	testRegion          = "eu-central-1"
	dynamoDBClient      *dynamodb.Client
	testRepo            prefsrepo.SettingsRepository
	dynamoDBInitialized = false
)

// setupDynamoDB initializes the DynamoDB client and creates a test table
func setupDynamoDB(t *testing.T) {
	if dynamoDBInitialized {
		return
	}

	ctx := context.Background()

	// Load AWS config without the deprecated endpoint resolver
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(testRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		t.Fatalf("Failed to load AWS config: %v", err)
	}

	// Create DynamoDB client with the latest endpoint approach
	dynamoDBClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(testEndpoint)
	})

	// Create test table
	_, err = dynamoDBClient.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(testTableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("pk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("sk"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("pk"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("sk"),
				KeyType:       types.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	// Ignore error if table already exists
	if err != nil && !strings.Contains(err.Error(), "ResourceInUseException") {
		t.Fatalf("Failed to create DynamoDB table: %v", err)
	}

	// Initialize repository
	testRepo = prefsrepo.NewSettingsRepository(dynamoDBClient, testTableName)
	dynamoDBInitialized = true
}

// cleanupDynamoDB deletes all data from the test table after tests
func cleanupDynamoDB(t *testing.T) {
	ctx := context.Background()

	// Get all items in the table
	scanResult, err := dynamoDBClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(testTableName),
	})
	if err != nil {
		t.Fatalf("Failed to scan DynamoDB table: %v", err)
	}

	// Delete each item
	for _, item := range scanResult.Items {
		_, err = dynamoDBClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(testTableName),
			Key: map[string]types.AttributeValue{
				"pk": item["pk"],
				"sk": item["sk"],
			},
		})
		if err != nil {
			t.Fatalf("Failed to delete item from DynamoDB: %v", err)
		}
	}
}

func TestSettingsRepositoryIntegration(t *testing.T) {
	setupDynamoDB(t)
	defer cleanupDynamoDB(t)

	ctx := context.Background()

	// Test data
	userUrn := "soundcloud:users:54321"
	name := "testSetting"
	settingType := prefsrepo.SettingTypeBoolean
	value := `true`

	// Create a new setting
	setting := prefsrepo.NewSetting(userUrn, name, settingType, value)

	// Test PutSetting (update existing setting)
	t.Run("PutSetting", func(t *testing.T) {
		err := testRepo.PutSetting(ctx, setting)
		assert.Nil(t, err)

		_, err = testRepo.GetSettings(ctx, userUrn, []string{setting.Name})
		assert.Nil(t, err)
	})

	// Test GetSettings
	t.Run("GetSettings", func(t *testing.T) {
		fetchedSetting, err := testRepo.GetSettings(ctx, userUrn, []string{})
		assert.Nil(t, err)
		assert.Equal(t, len(fetchedSetting), 1)
		assert.Equal(t, fetchedSetting[0].UserUrn, setting.UserUrn)
		assert.Equal(t, fetchedSetting[0].Name, setting.Name)
		assert.Equal(t, fetchedSetting[0].SettingType, setting.SettingType)
		assert.Equal(t, fetchedSetting[0].Value, setting.Value)
		assert.Equal(t, fetchedSetting[0].UpdatedAt.Unix(), setting.UpdatedAt.Unix())

		// Try to get a non-existent setting
		fetchedSetting, err = testRepo.GetSettings(ctx, userUrn, []string{"nonexistent"})
		assert.Nil(t, err)
		assert.Len(t, fetchedSetting, 0)
	})
}
