package api

import (
	"context"
	"errors"

	commonv1 "github.com/patrickisgreat/user-settings/internal/generated/common/v1"
	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/mappers"
	"github.com/patrickisgreat/user-settings/internal/prefsservice"
	"github.com/twitchtv/twirp"
)

var settingsService prefsservice.SettingsService
var mapper mappers.ProtoMapper

// InitService sets the service instance for the API handlers
func InitService(service prefsservice.SettingsService, protoMapper mappers.ProtoMapper) {
	settingsService = service
	mapper = protoMapper
}

// GetSettings returns the caller's settings, optionally filtered by name.
func GetSettings(ctx context.Context, request *proto.GetSettingsRequest) (*proto.GetSettingsResponse, error) {
	userUrn, err := callerUrn(request.GetUserSession())
	if err != nil {
		return nil, err
	}

	repoSettings, err := settingsService.GetSettings(ctx, userUrn, request.GetSettingNames())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	protoSettings := make([]*proto.Setting, 0, len(repoSettings))
	for i := range repoSettings {
		protoSettings = append(protoSettings, mapper.SettingToProto(&repoSettings[i]))
	}

	return &proto.GetSettingsResponse{Settings: protoSettings}, nil
}

// GetPrivacySettings returns the privacy-relevant settings of the given users.
func GetPrivacySettings(ctx context.Context, request *proto.GetPrivacySettingsRequest) (*proto.GetPrivacySettingsResponse, error) {
	if len(request.GetUserUrns()) == 0 {
		return nil, twirp.InvalidArgumentError("user_urns", "user_urns should not be empty")
	}

	repoPrivacySettings, err := settingsService.GetPrivacySettings(ctx, request.GetUserUrns())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	protoUserSettings := make([]*proto.UserSettings, 0, len(repoPrivacySettings))
	for _, userSettings := range repoPrivacySettings {
		userPrivacySettings := make([]*proto.Setting, 0, len(userSettings.Settings))
		for i := range userSettings.Settings {
			userPrivacySettings = append(userPrivacySettings, mapper.SettingToProto(&userSettings.Settings[i]))
		}
		protoUserSettings = append(protoUserSettings, &proto.UserSettings{
			UserUrn:  userSettings.UserUrn,
			Settings: userPrivacySettings,
		})
	}

	return &proto.GetPrivacySettingsResponse{Settings: protoUserSettings}, nil
}

// PutSetting creates or overwrites one setting for the caller.
func PutSetting(ctx context.Context, request *proto.PutSettingRequest) (*proto.PutSettingResponse, error) {
	userUrn, err := callerUrn(request.GetUserSession())
	if err != nil {
		return nil, err
	}

	setting := request.GetSetting()
	switch {
	case setting == nil:
		return nil, twirp.InvalidArgumentError("setting", "setting is required")
	case setting.GetName() == "":
		return nil, twirp.InvalidArgumentError("setting.name", "name is required")
	case setting.GetValue() == "":
		return nil, twirp.InvalidArgumentError("setting.value", "value is required")
	}

	repoSetting, err := mapper.PutSettingToSetting(userUrn, setting)
	if err != nil {
		return nil, twirp.InvalidArgumentError("setting.type", err.Error())
	}

	if err := settingsService.PutSetting(ctx, *repoSetting); err != nil {
		if errors.Is(err, prefsservice.ErrInvalidValue) {
			return nil, twirp.InvalidArgumentError("setting.value", "value must be a valid JSON document")
		}
		return nil, twirp.InternalErrorWith(err)
	}

	return &proto.PutSettingResponse{Setting: mapper.SettingToProto(repoSetting)}, nil
}

// FetchUsersPrivacyAttributes reports whether each given user is a minor.
func FetchUsersPrivacyAttributes(ctx context.Context, request *proto.UsersPrivacyAttributesRequest) (*proto.UsersPrivacyAttributesResponse, error) {
	if _, err := callerUrn(request.GetUserSession()); err != nil {
		return nil, err
	}
	if len(request.GetUserUrns()) == 0 {
		return nil, twirp.InvalidArgumentError("user_urns", "user_urns should not be empty")
	}

	privacyAttributes, err := settingsService.FetchUsersPrivacyAttributes(ctx, request.GetUserUrns())
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &proto.UsersPrivacyAttributesResponse{PrivacyAttributes: mapper.MapPrivacyAttributes(privacyAttributes)}, nil
}

// callerUrn extracts the authenticated user's URN from the session, or returns
// an InvalidArgument error when the session or URN is missing. The generated
// getters are nil-safe, so a missing session never dereferences a nil pointer.
func callerUrn(session *commonv1.UserSession) (string, error) {
	if session == nil {
		return "", twirp.InvalidArgumentError("user_session", "user session is required")
	}
	userUrn := session.GetUserUrn().GetValue()
	if userUrn == "" {
		return "", twirp.InvalidArgumentError("user_session.user_urn", "user_urn is required")
	}
	return userUrn, nil
}
