package api

import (
	"context"

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

// GetSettings handles the GetSettings endpoint.
func GetSettings(ctx context.Context, request *proto.GetSettingsRequest) (*proto.GetSettingsResponse, error) {
	// Retrieve the setting using the provided setting IDs.
	userUrn := request.UserSession.UserUrn.Value
	repoSettings, err := settingsService.GetSettings(ctx, userUrn, request.SettingNames)
	if err != nil {
		// Convert "setting not found" to a Not Found error
		if err.Error() == "setting not found" {
			return nil, twirp.NotFoundError("Setting not found")
		}
		// Otherwise return an internal error
		return nil, twirp.InternalErrorWith(err)
	}

	var protoSettings []*proto.Setting
	for _, item := range repoSettings {
		protoSettings = append(protoSettings, mapper.SettingToProto(&item))
	}

	return &proto.GetSettingsResponse{Settings: protoSettings}, nil
}
func GetPrivacySettings(ctx context.Context, request *proto.GetPrivacySettingsRequest) (*proto.GetPrivacySettingsResponse, error) {
	// Retrieve the setting using the provided setting IDs.
	if len(request.UserUrns) == 0 {
		return nil, twirp.InvalidArgumentError("request.UserUrns", "userUrns should not be empty")
	}
	repoPrivacySettings, err := settingsService.GetPrivacySettings(ctx, request.UserUrns)
	if err != nil {
		// Convert "setting not found" to a Not Found error
		if err.Error() == "setting not found" {
			return nil, twirp.NotFoundError("Setting not found")
		}
		// Otherwise return an internal error
		return nil, twirp.InternalErrorWith(err)
	}

	var protoUserSettings []*proto.UserSettings
	for _, userSettings := range repoPrivacySettings {
		var userPrivacySettings []*proto.Setting
		for _, privacySettings := range userSettings.Settings {
			userPrivacySettings = append(userPrivacySettings, mapper.SettingToProto(&privacySettings))
		}
		protoUserSettings = append(protoUserSettings, &proto.UserSettings{
			UserUrn:  userSettings.UserUrn,
			Settings: userPrivacySettings,
		})
	}

	return &proto.GetPrivacySettingsResponse{Settings: protoUserSettings}, nil
}
func PutSetting(ctx context.Context, request *proto.PutSettingRequest) (*proto.PutSettingResponse, error) {
	// Basic validation: check that a user session is provided.
	if request.UserSession == nil {
		return nil, twirp.InvalidArgumentError("user_session", "user session is required")
	}

	// Validate the setting
	if request.Setting == nil {
		return nil, twirp.InvalidArgumentError("setting", "setting is required")
	}
	if request.Setting.Name == "" {
		return nil, twirp.InvalidArgumentError("setting.name", "name is required")
	}
	if request.Setting.Value == "" {
		return nil, twirp.InvalidArgumentError("setting.value", "value is required")
	}

	// Convert the proto.Setting into our internal representation.
	repoSetting, err := mapper.PutSettingToSetting(request.UserSession.UserUrn.Value, request.Setting)
	if err != nil {
		// Handle specific error for invalid JSON
		if err.Error() != "" && (err.Error() == "invalid JSON in setting value" || err.Error() != "" && err.Error()[0:13] == "invalid JSON ") {
			return nil, twirp.InvalidArgumentError("setting.value", "invalid JSON format")
		}
		return nil, twirp.InvalidArgumentError("setting", err.Error())
	}

	// Write (or overwrite) the setting in the repository.
	if err := settingsService.PutSetting(ctx, *repoSetting); err != nil {
		// Catch JSON validation errors from service layer
		if err.Error() != "" && (err.Error() == "invalid JSON in setting value" || err.Error() != "" && err.Error()[0:13] == "invalid JSON ") {
			return nil, twirp.InvalidArgumentError("setting.value", "invalid JSON format")
		}
		return nil, twirp.InternalErrorWith(err)
	}

	// Return the updated setting.
	return &proto.PutSettingResponse{Setting: mapper.SettingToProto(repoSetting)}, nil
}

// FetchUsersPrivacyAttributes returns privacy attributes for given users
func FetchUsersPrivacyAttributes(ctx context.Context, request *proto.UsersPrivacyAttributesRequest) (*proto.UsersPrivacyAttributesResponse, error) {
	if request.UserSession == nil {
		return nil, twirp.InvalidArgumentError("user_session", "user session is required")
	}
	if len(request.UserUrns) == 0 {
		return nil, twirp.InvalidArgumentError("user_urns", "user_urns should not be empty")
	}

	privacyAttributes, err := settingsService.FetchUsersPrivacyAttributes(ctx, request.UserUrns)
	if err != nil {
		// Convert "setting not found" to a Not Found error
		if err.Error() == "setting not found" {
			return nil, twirp.NotFoundError("Setting not found")
		}
		return nil, twirp.InternalErrorWith(err)
	}

	// Convert the internal model to proto model
	protoPrivacyAttributes := mapper.MapPrivacyAttributes(privacyAttributes)

	return &proto.UsersPrivacyAttributesResponse{PrivacyAttributes: protoPrivacyAttributes}, nil
}
