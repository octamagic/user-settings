package mappers

import (
	"encoding/json"
	"fmt"
	"time"

	proto "github.com/soundcloud/user-settings/internal/generated/user_settings/api"
	"github.com/soundcloud/user-settings/internal/prefsrepo"
	"github.com/soundcloud/user-settings/internal/prefsservice/domain"
)

type ProtoMapper struct {
}

func (s *ProtoMapper) MapUserSettings(repoSetting []prefsrepo.Setting) []*proto.UserSettings {
	protoSettings := make([]*proto.UserSettings, len(repoSetting))
	for i, setting := range repoSetting {
		protoSettings[i] = &proto.UserSettings{
			UserUrn:  setting.UserUrn,
			Settings: []*proto.Setting{s.SettingToProto(&setting)},
		}
	}
	return protoSettings
}

// SettingToProto converts an internal Setting to a proto Setting
func (s *ProtoMapper) SettingToProto(setting *prefsrepo.Setting) *proto.Setting {
	return &proto.Setting{
		UserUrn:   setting.UserUrn,
		Name:      setting.Name,
		Type:      s.settingTypeToProto(setting.SettingType),
		UpdatedAt: setting.UpdatedAt.Unix(),
		Value:     setting.Value,
	}
}

// SettingFromProto creates a Setting from a proto Setting
func (s *ProtoMapper) SettingFromProto(protoSetting *proto.Setting) (*prefsrepo.Setting, error) {
	internalType, err := settingTypeFromProto(protoSetting.Type)
	if err != nil {
		return nil, err
	}

	return &prefsrepo.Setting{
		UserUrn:     protoSetting.UserUrn,
		Name:        protoSetting.Name,
		UpdatedAt:   time.Unix(protoSetting.UpdatedAt, 0),
		SettingType: internalType,
		Value:       protoSetting.Value,
	}, nil
}

// PutSettingToSetting creates a Setting from a proto PutSetting
func (s *ProtoMapper) PutSettingToSetting(userUrn string, protoSetting *proto.PutSetting) (*prefsrepo.Setting, error) {
	internalType, err := settingTypeFromProto(protoSetting.Type)
	if err != nil {
		return nil, err
	}

	// Validate the JSON format if a value is provided
	if protoSetting.Value != "" {
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(protoSetting.Value), &jsonObj); err != nil {
			return nil, fmt.Errorf("invalid JSON in setting value: %w", err)
		}
	}

	return &prefsrepo.Setting{
		UserUrn:     userUrn,
		Name:        protoSetting.Name,
		UpdatedAt:   time.Now(),
		SettingType: internalType,
		Value:       protoSetting.Value,
	}, nil
}

func (s *ProtoMapper) MapPrivacyAttributes(privacyAttributes []domain.UserPrivacyAttributes) []*proto.UserPrivacyAttributes {
	var result []*proto.UserPrivacyAttributes
	for _, userPrivacyAttribute := range privacyAttributes {
		result = append(result, &proto.UserPrivacyAttributes{
			UserUrn: userPrivacyAttribute.UserUrn,
			IsMinor: userPrivacyAttribute.IsMinor,
		})
	}
	return result
}

// settingTypeToProto maps internal string type to proto enum
func (s *ProtoMapper) settingTypeToProto(internalType string) proto.SettingType {
	switch internalType {
	case prefsrepo.SettingTypeString.String():
		return proto.SettingType_STRING
	case prefsrepo.SettingTypeBoolean.String():
		return proto.SettingType_BOOLEAN
	case proto.SettingType_NUMBER.String():
		return proto.SettingType_NUMBER
	default:
		return proto.SettingType_STRING
	}
}

// settingTypeFromProto maps proto enum to internal string type
func settingTypeFromProto(protoType proto.SettingType) (string, error) {
	switch protoType {
	case proto.SettingType_STRING:
		return prefsrepo.SettingTypeString.String(), nil
	case proto.SettingType_BOOLEAN:
		return prefsrepo.SettingTypeBoolean.String(), nil
	case proto.SettingType_NUMBER:
		return prefsrepo.SettingTypeNumber.String(), nil
	default:
		return prefsrepo.SettingTypeString.String(), fmt.Errorf("unsupported setting type: %v", protoType)
	}
}
