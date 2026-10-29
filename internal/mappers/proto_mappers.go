package mappers

import (
	"fmt"
	"time"

	proto "github.com/octamagic/user-settings/internal/generated/usersettings/v1"
	"github.com/octamagic/user-settings/internal/prefsrepo"
	"github.com/octamagic/user-settings/internal/prefsservice/domain"
)

type ProtoMapper struct {
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

// PutSettingToSetting creates a Setting from a proto PutSetting
func (s *ProtoMapper) PutSettingToSetting(userUrn string, protoSetting *proto.PutSetting) (*prefsrepo.Setting, error) {
	internalType, err := settingTypeFromProto(protoSetting.Type)
	if err != nil {
		return nil, err
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

// settingTypeToProto maps a stored type code to the proto enum. Unknown codes
// fall back to STRING, which is also the proto zero value.
func (s *ProtoMapper) settingTypeToProto(internalType string) proto.SettingType {
	settingType, _ := prefsrepo.ParseSettingType(internalType)
	switch settingType {
	case prefsrepo.SettingTypeBoolean:
		return proto.SettingType_BOOLEAN
	case prefsrepo.SettingTypeNumber:
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
