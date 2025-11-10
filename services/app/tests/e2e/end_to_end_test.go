package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/soundcloud/gokit/v2/logger"
	"github.com/soundcloud/user-settings/tests/dbutils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wiremock/wiremock-testcontainers-go" //nolint:goimports
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dynamodb := setupDynamoDBInTestcontainers(ctx)
	okidoki := setupWiremockContainer(ctx)

	defer func() {
		if err := testcontainers.TerminateContainer(dynamodb); err != nil {
			fmt.Printf("Error terminating DynamoDB container: %v\n", err)
		}
		if err := testcontainers.TerminateContainer(okidoki); err != nil {
			fmt.Printf("Error terminating Okidoki container: %v\n", err)
		}
	}()

	cancel := startServer()
	createTestTables(ctx)

	exitCode := m.Run()

	cancel()
	os.Exit(exitCode)
}

func setupWiremockContainer(ctx context.Context) testcontainers.Container {
	wiremockContainer, _ := testcontainers_wiremock.RunContainer(ctx,
		testcontainers_wiremock.WithMappingFile("okidoki", filepath.Join("wiremock", "okidoki", "mappings", "okidoki.json")),
		testcontainers_wiremock.WithFile("user2.json", filepath.Join("wiremock", "okidoki", "__files", "user2.json")),
	)

	endpoint, endpointErr := wiremockContainer.Endpoint(ctx, "")
	if endpointErr != nil {
		fmt.Println("Error getting endpoint:", endpointErr)
		os.Exit(1)
	}

	okidokiEndpoint := fmt.Sprintf("http://%s", endpoint)
	fmt.Println("baseURL for okidoki: ", okidokiEndpoint)
	if err := os.Setenv("OKIDOKI_BASE_URL", okidokiEndpoint); err != nil {
		fmt.Println("Error setting OKIDOKI_BASE_URL environment variable:", err)
		os.Exit(1)
	}

	return wiremockContainer
}

func setupDynamoDBInTestcontainers(ctx context.Context) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		ExposedPorts: []string{"8000"},
		WaitingFor:   wait.ForExposedPort().WithPollInterval(100 * time.Millisecond).WithStartupTimeout(60 * time.Second),
	}
	dynamodb, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	endpoint, endpointErr := dynamodb.Endpoint(ctx, "")
	if endpointErr != nil {
		fmt.Println("Error getting endpoint:", endpointErr)
		os.Exit(1)
	}

	dynamodbEndpoint := fmt.Sprintf("http://%s", endpoint)
	if err := os.Setenv("DYNAMODB_ENDPOINT", dynamodbEndpoint); err != nil {
		fmt.Println("Error setting DYNAMODB_ENDPOINT environment variable:", err)
		os.Exit(1)
	}

	return dynamodb
}

func createTestTables(ctx context.Context) {
	if err := dbutils.CreateTables(); err != nil {
		logger.Get(ctx).Error("Error creating database:", slog.Any("err", err))
		os.Exit(1)
	}
}
