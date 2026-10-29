package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/bakins/twirpotel"
	"github.com/patrickisgreat/user-settings/internal/api"
	"github.com/patrickisgreat/user-settings/internal/appserver"
	"github.com/patrickisgreat/user-settings/internal/config"
	common "github.com/patrickisgreat/user-settings/internal/generated/common/v1"
	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/handlers"
	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/mappers"
	"github.com/stretchr/testify/assert"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/encoding/protojson"
	pbproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var httpClient = &http.Client{}

func startServer() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	appserver.InitContext(ctx)
	api.InitService(appserver.Context().UserSettingsService, mappers.ProtoMapper{})
	ras := proto.NewUserSettingsServiceServer(
		&handlers.UserSettingsService{},
		twirp.WithServerInterceptors(twirpotel.ServerInterceptor()),
	)
	router := appserver.NewRouter(&ras, logging.New(config.Config.LogLevel, config.Config.LogPretty))
	addr := net.JoinHostPort(config.Config.Host, fmt.Sprintf("%d", config.Config.Port))
	go appserver.Start(ctx, &appserver.StartParams{
		Router: router,
		Addr:   addr,
	})
	return cancel
}

func sendRequest(t *testing.T, method string, reqBody pbproto.Message) *http.Response {
	baseURL := fmt.Sprintf("http://%s:%d/twirp/usersettings.v1.", config.Config.Host, config.Config.Port)
	reqBytes, err := protojson.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}
	fmt.Println("Request Body:", string(reqBytes))

	req, err := http.NewRequest("POST", baseURL+method, bytes.NewBuffer(reqBytes))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	return res
}

func decodeSuccessResponse(t *testing.T, res *http.Response, resBody pbproto.Message) pbproto.Message {
	assert.Equal(t, 200, res.StatusCode)
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	fmt.Println("Response Body:", string(bodyBytes))

	if err := protojson.Unmarshal(bodyBytes, resBody); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	return resBody
}

//func decodeErrorResponse(t *testing.T, res *http.Response) string {
//	bodyBytes, err := io.ReadAll(res.Body)
//	if err != nil {
//		t.Fatalf("Failed to read response body: %v", err)
//	}
//	fmt.Println("Response Body:", string(bodyBytes))
//	return string(bodyBytes)
//}

func userSession(userUrn string) *common.UserSession {
	return &common.UserSession{UserUrn: &wrapperspb.StringValue{Value: userUrn}}
}
