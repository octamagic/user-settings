package e2e

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/octamagic/user-settings/internal/app"
	"github.com/octamagic/user-settings/internal/config"
	commonv1 "github.com/octamagic/user-settings/internal/generated/common/v1"
	usersettingsv1 "github.com/octamagic/user-settings/internal/generated/usersettings/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// testServer is the application under test, served on an ephemeral port.
var testServer *httptest.Server

// startServer builds the app from the environment and serves it in-process.
func startServer(ctx context.Context) (*app.App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	service, err := app.New(ctx, cfg, slog.New(slog.DiscardHandler))
	if err != nil {
		return nil, err
	}
	testServer = httptest.NewServer(service.Handler)
	return service, nil
}

func sendRequest(t *testing.T, method string, reqBody proto.Message) *http.Response {
	t.Helper()
	reqBytes, err := protojson.Marshal(reqBody)
	require.NoError(t, err, "marshal request body")

	url := testServer.URL + usersettingsv1.UserSettingsServicePathPrefix + method
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(reqBytes))
	require.NoError(t, err, "create request")
	req.Header.Set("Content-Type", "application/json")

	res, err := testServer.Client().Do(req)
	require.NoError(t, err, "send request")
	t.Cleanup(func() { _ = res.Body.Close() })
	return res
}

func decodeSuccessResponse(t *testing.T, res *http.Response, resBody proto.Message) {
	t.Helper()
	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err, "read response body")
	require.Equal(t, http.StatusOK, res.StatusCode, "unexpected status, body: %s", bodyBytes)
	require.NoError(t, protojson.Unmarshal(bodyBytes, resBody), "unmarshal response")
}

func userSession(userUrn string) *commonv1.UserSession {
	return &commonv1.UserSession{UserUrn: &wrapperspb.StringValue{Value: userUrn}}
}
