package api

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	sessionproto "github.com/patrickisgreat/user-settings/internal/generated/common/v1"
	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsservice"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twitchtv/twirp"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func setupTestAPI(ctrl *gomock.Controller) (*Server, *prefsservice.MockSettingsService) {
	mockService := prefsservice.NewMockSettingsService(ctrl)
	return NewServer(mockService), mockService
}

const SettingName = "testSetting"

func createTestSetting(userUrn string) *prefsrepo.Setting {
	if userUrn == "" {
		userUrn = "urn:user:123456"
	}
	return &prefsrepo.Setting{
		UserUrn:     userUrn,
		Name:        SettingName,
		SettingType: prefsrepo.SettingTypeBoolean.String(),
		UpdatedAt:   time.Now(),
		Value:       `true`,
	}
}

func createTestProtoPutSetting() *proto.PutSetting {
	return &proto.PutSetting{
		Type:  proto.SettingType_BOOLEAN,
		Value: `true`,
		Name:  SettingName,
	}
}

func createTestUserSession() *sessionproto.UserSession {
	return &sessionproto.UserSession{
		UserUrn:  wrapperspb.String("urn:user:123456"),
		AgentUrn: wrapperspb.String("urn:agent:web"),
		Geo: &sessionproto.Geo{
			CountryCode: "US",
		},
		Scopes:   []string{"read", "write"},
		Features: []string{},
	}
}

func TestGetSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockService := setupTestAPI(ctrl)
	ctx := context.Background()

	t.Run("Valid request", func(t *testing.T) {
		// Add a test setting to the mock service
		testSetting := createTestSetting("")
		mockService.EXPECT().GetSettings(ctx, gomock.Any(), gomock.Any()).Return([]prefsrepo.Setting{*testSetting}, nil)

		// Create a valid request
		req := &proto.GetSettingsRequest{
			UserSession: createTestUserSession(),
		}

		// Call the API
		resp, err := server.GetSettings(ctx, req)
		// Verify response
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Settings, 1)
		assert.Equal(t, testSetting.UserUrn, resp.Settings[0].UserUrn)
		assert.Equal(t, testSetting.Name, resp.Settings[0].Name)
	})

	t.Run("Service error", func(t *testing.T) {
		mockService.EXPECT().GetSettings(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("error in repo"))
		req := &proto.GetSettingsRequest{
			UserSession: createTestUserSession(),
		}

		_, err := server.GetSettings(ctx, req)
		require.Error(t, err)
	})
}

func TestGetPrivacySettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockService := setupTestAPI(ctrl)
	ctx := context.Background()

	t.Run("Valid request", func(t *testing.T) {
		// Add a test setting to the mock service
		testSetting := createTestSetting("")
		expectedUserPrivacySettings := []prefsservice.UserPrivacySettings{
			{
				UserUrn:  testSetting.UserUrn,
				Settings: []prefsrepo.Setting{*testSetting},
			},
		}
		mockService.EXPECT().GetPrivacySettings(ctx, gomock.Any()).Return(expectedUserPrivacySettings, nil)

		// Create a valid request
		req := &proto.GetPrivacySettingsRequest{
			UserSession: createTestUserSession(),
			UserUrns:    []string{testSetting.UserUrn},
		}

		// Call the API
		resp, err := server.GetPrivacySettings(ctx, req)
		// Verify response
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Settings, 1)
		userSettings := resp.Settings[0]
		assert.Len(t, userSettings.Settings, 1)
		assert.Equal(t, testSetting.UserUrn, userSettings.Settings[0].UserUrn)
		assert.Equal(t, testSetting.Name, userSettings.Settings[0].Name)
	})

	t.Run("Service error", func(t *testing.T) {
		mockService.EXPECT().GetPrivacySettings(ctx, gomock.Any()).Return(nil, errors.New("error in repo"))
		req := &proto.GetPrivacySettingsRequest{
			UserSession: createTestUserSession(),
			UserUrns:    []string{"userUrn"},
		}

		_, err := server.GetPrivacySettings(ctx, req)
		require.Error(t, err)
	})
}

func TestPutSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockService := setupTestAPI(ctrl)
	ctx := context.Background()

	t.Run("Valid request", func(t *testing.T) {
		mockService.EXPECT().PutSetting(ctx, gomock.Any()).Return(nil)
		// Create a valid request
		session := createTestUserSession()
		protoSetting := createTestProtoPutSetting()

		req := &proto.PutSettingRequest{
			UserSession: session,
			Setting:     protoSetting,
		}

		// Call the API
		resp, err := server.PutSetting(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, resp.Setting.Name, protoSetting.Name)
		assert.Equal(t, resp.Setting.Value, protoSetting.Value)
		assert.Equal(t, resp.Setting.UserUrn, session.UserUrn.Value)
	})

	t.Run("Missing user_session", func(t *testing.T) {
		// Create a request without user session
		req := &proto.PutSettingRequest{
			UserSession: nil,
			Setting:     createTestProtoPutSetting(),
		}

		_, err := server.PutSetting(ctx, req)
		require.Error(t, err)

		// Should be a Twirp InvalidArgument error
		var twerr twirp.Error
		ok := errors.As(err, &twerr)
		assert.True(t, ok)
		assert.Equal(t, twirp.InvalidArgument, twerr.Code())
	})

	t.Run("Missing setting Name", func(t *testing.T) {
		// Create a request with missing setting ID
		invalidSetting := createTestProtoPutSetting()
		invalidSetting.Name = ""

		req := &proto.PutSettingRequest{
			UserSession: createTestUserSession(),
			Setting:     invalidSetting,
		}

		_, err := server.PutSetting(ctx, req)
		require.Error(t, err)

		// Should be a Twirp InvalidArgument error
		var twerr twirp.Error
		ok := errors.As(err, &twerr)
		assert.True(t, ok)
		assert.Equal(t, twirp.InvalidArgument, twerr.Code())
	})

	t.Run("Missing setting Value", func(t *testing.T) {
		// Create a request with missing setting ID
		invalidSetting := createTestProtoPutSetting()
		invalidSetting.Value = ""

		req := &proto.PutSettingRequest{
			UserSession: createTestUserSession(),
			Setting:     invalidSetting,
		}

		_, err := server.PutSetting(ctx, req)
		require.Error(t, err)

		// Should be a Twirp InvalidArgument error
		var twerr twirp.Error
		ok := errors.As(err, &twerr)
		assert.True(t, ok)
		assert.Equal(t, twirp.InvalidArgument, twerr.Code())
	})
}

func TestFetchUsersPrivacyAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	server, mockService := setupTestAPI(ctrl)
	ctx := context.Background()

	const TestUserUrn = "userUrn"

	t.Run("Valid request", func(t *testing.T) {
		expectedAttributes := []domain.UserPrivacyAttributes{
			{
				UserUrn: TestUserUrn,
				IsMinor: false,
			},
		}
		mockService.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return(expectedAttributes, nil)
		// Create a valid request
		session := createTestUserSession()

		req := &proto.UsersPrivacyAttributesRequest{
			UserSession: session,
			UserUrns:    []string{TestUserUrn},
		}

		// Call the API
		resp, err := server.FetchUsersPrivacyAttributes(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.PrivacyAttributes, 1)
		assert.Equal(t, TestUserUrn, resp.PrivacyAttributes[0].UserUrn)
		assert.False(t, resp.PrivacyAttributes[0].IsMinor)
	})

	t.Run("Service return error", func(t *testing.T) {
		mockService.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return(nil, errors.New("service error"))
		session := createTestUserSession()

		req := &proto.UsersPrivacyAttributesRequest{
			UserSession: session,
			UserUrns:    []string{TestUserUrn},
		}

		// Call the API
		resp, err := server.FetchUsersPrivacyAttributes(ctx, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func assertTwirpCode(t *testing.T, err error, want twirp.ErrorCode) {
	t.Helper()
	var twerr twirp.Error
	if assert.ErrorAs(t, err, &twerr, "expected a twirp error, got %v", err) {
		assert.Equal(t, want, twerr.Code())
	}
}

func TestGetSettingsValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	server, _ := setupTestAPI(ctrl)
	ctx := context.Background()

	t.Run("missing user_session is InvalidArgument, not a panic", func(t *testing.T) {
		_, err := server.GetSettings(ctx, &proto.GetSettingsRequest{})
		assertTwirpCode(t, err, twirp.InvalidArgument)
	})

	t.Run("missing user_urn is InvalidArgument", func(t *testing.T) {
		_, err := server.GetSettings(ctx, &proto.GetSettingsRequest{UserSession: &sessionproto.UserSession{}})
		assertTwirpCode(t, err, twirp.InvalidArgument)
	})
}

func TestPutSettingErrorMapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	server, mockService := setupTestAPI(ctrl)
	ctx := context.Background()

	t.Run("invalid JSON value from the service is InvalidArgument", func(t *testing.T) {
		mockService.EXPECT().PutSetting(ctx, gomock.Any()).Return(fmt.Errorf("%w: unexpected end of input", prefsservice.ErrInvalidValue))

		_, err := server.PutSetting(ctx, &proto.PutSettingRequest{UserSession: createTestUserSession(), Setting: createTestProtoPutSetting()})
		assertTwirpCode(t, err, twirp.InvalidArgument)
	})

	t.Run("any other service error is Internal", func(t *testing.T) {
		mockService.EXPECT().PutSetting(ctx, gomock.Any()).Return(errors.New("dynamodb unavailable"))

		_, err := server.PutSetting(ctx, &proto.PutSettingRequest{UserSession: createTestUserSession(), Setting: createTestProtoPutSetting()})
		assertTwirpCode(t, err, twirp.Internal)
	})

	t.Run("unsupported setting type is InvalidArgument", func(t *testing.T) {
		setting := createTestProtoPutSetting()
		setting.Type = proto.SettingType(99)

		_, err := server.PutSetting(ctx, &proto.PutSettingRequest{UserSession: createTestUserSession(), Setting: setting})
		assertTwirpCode(t, err, twirp.InvalidArgument)
	})
}
