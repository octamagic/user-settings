// Package e2e exercises the service over HTTP against DynamoDB Local and a
// WireMock stand-in for the user directory, both started via testcontainers.
package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/patrickisgreat/user-settings/tests/dbutils"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	testcontainers_wiremock "github.com/wiremock/wiremock-testcontainers-go"
)

func TestMain(m *testing.M) {
	code, err := runSuite(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e setup failed:", err)
		code = 1
	}
	os.Exit(code)
}

func runSuite(m *testing.M) (int, error) {
	ctx := context.Background()

	dynamo, err := startDynamoDB(ctx)
	if err != nil {
		return 0, err
	}
	defer terminate(dynamo, "dynamodb")

	directory, err := startUserDirectory(ctx)
	if err != nil {
		return 0, err
	}
	defer terminate(directory, "user directory")

	service, err := startServer(ctx)
	if err != nil {
		return 0, err
	}
	defer testServer.Close()
	defer func() { _ = service.Shutdown(ctx) }()

	if err := dbutils.CreateSettingsTable(ctx, service.DynamoDB, service.Config.UserSettingsTableName); err != nil {
		return 0, err
	}

	return m.Run(), nil
}

func startUserDirectory(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers_wiremock.RunContainer(ctx,
		testcontainers_wiremock.WithMappingFile("userdirectory", filepath.Join("wiremock", "userdirectory_mappings.json")),
		testcontainers_wiremock.WithFile("user2.json", filepath.Join("wiremock", "user2.json")),
	)
	if err != nil {
		return nil, fmt.Errorf("starting wiremock: %w", err)
	}
	endpoint, err := container.Endpoint(ctx, "http")
	if err != nil {
		return nil, fmt.Errorf("resolving wiremock endpoint: %w", err)
	}
	return container, os.Setenv("USER_DIRECTORY_BASE_URL", endpoint)
}

func startDynamoDB(ctx context.Context) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "amazon/dynamodb-local:latest",
			ExposedPorts: []string{"8000"},
			WaitingFor:   wait.ForExposedPort().WithPollInterval(100 * time.Millisecond).WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, fmt.Errorf("starting dynamodb local: %w", err)
	}
	endpoint, err := container.Endpoint(ctx, "http")
	if err != nil {
		return nil, fmt.Errorf("resolving dynamodb endpoint: %w", err)
	}
	return container, os.Setenv("DYNAMODB_ENDPOINT", endpoint)
}

func terminate(container testcontainers.Container, name string) {
	if err := testcontainers.TerminateContainer(container); err != nil {
		fmt.Fprintf(os.Stderr, "error terminating %s container: %v\n", name, err)
	}
}
