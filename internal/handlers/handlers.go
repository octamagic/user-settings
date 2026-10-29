package handlers

import (
	"context"

	"github.com/soundcloud/user-settings/internal/api"
	proto "github.com/soundcloud/user-settings/internal/generated/user_settings/api"
)

type UserSettingsService struct{}

func (h *UserSettingsService) GetSettings(ctx context.Context, req *proto.GetSettingsRequest) (*proto.GetSettingsResponse, error) {
	return api.GetSettings(ctx, req)
}

func (h *UserSettingsService) GetPrivacySettings(ctx context.Context, req *proto.GetPrivacySettingsRequest) (*proto.GetPrivacySettingsResponse, error) {
	return api.GetPrivacySettings(ctx, req)
}

func (h *UserSettingsService) PutSetting(ctx context.Context, req *proto.PutSettingRequest) (*proto.PutSettingResponse, error) {
	return api.PutSetting(ctx, req)
}

func (h *UserSettingsService) FetchUsersPrivacyAttributes(ctx context.Context, req *proto.UsersPrivacyAttributesRequest) (*proto.UsersPrivacyAttributesResponse, error) {
	return api.FetchUsersPrivacyAttributes(ctx, req)
}
