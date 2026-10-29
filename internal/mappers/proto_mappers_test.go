package mappers

import (
	"testing"
	"time"

	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingConversion(t *testing.T) {
	mapper := ProtoMapper{}

	// Test data
	now := time.Now()
	repoSetting := &prefsrepo.Setting{
		UserUrn:     "urn:user:123456",
		UpdatedAt:   now,
		Name:        "testSetting",
		SettingType: prefsrepo.SettingTypeBoolean.String(),
		Value:       `true`,
	}

	// Test SettingToProto
	t.Run("SettingToProto", func(t *testing.T) {
		protoSetting := mapper.SettingToProto(repoSetting)

		// Verify conversion
		assert.Equal(t, protoSetting.UserUrn, repoSetting.UserUrn)
		assert.Equal(t, protoSetting.UpdatedAt, repoSetting.UpdatedAt.Unix())
		assert.Equal(t, protoSetting.Name, repoSetting.Name)
		assert.Equal(t, proto.SettingType_BOOLEAN, protoSetting.Type)
		assert.Equal(t, protoSetting.Value, repoSetting.Value)
	})

	t.Run("PrivacyAttributes to proto", func(t *testing.T) {
		privacyAttributes := []domain.UserPrivacyAttributes{
			{
				UserUrn: "urn",
				IsMinor: true,
			},
		}

		protoAttributes := mapper.MapPrivacyAttributes(privacyAttributes)

		assert.Len(t, protoAttributes, len(privacyAttributes))
		assert.Equal(t, privacyAttributes[0].UserUrn, protoAttributes[0].UserUrn)
		assert.Equal(t, privacyAttributes[0].IsMinor, protoAttributes[0].IsMinor)
	})
}

func TestSettingTypeToProto(t *testing.T) {
	mapper := ProtoMapper{}

	tests := []struct {
		name string
		code string
		want proto.SettingType
	}{
		{name: "string code maps to STRING", code: prefsrepo.SettingTypeString.String(), want: proto.SettingType_STRING},
		{name: "boolean code maps to BOOLEAN", code: prefsrepo.SettingTypeBoolean.String(), want: proto.SettingType_BOOLEAN},
		{name: "number code maps to NUMBER", code: prefsrepo.SettingTypeNumber.String(), want: proto.SettingType_NUMBER},
		{name: "unknown code falls back to STRING", code: "x", want: proto.SettingType_STRING},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mapper.settingTypeToProto(tt.code))
		})
	}
}

func TestSettingTypeRoundTrip(t *testing.T) {
	mapper := ProtoMapper{}
	for _, protoType := range []proto.SettingType{proto.SettingType_STRING, proto.SettingType_BOOLEAN, proto.SettingType_NUMBER} {
		t.Run(protoType.String(), func(t *testing.T) {
			code, err := settingTypeFromProto(protoType)
			require.NoError(t, err)
			assert.Equal(t, protoType, mapper.settingTypeToProto(code))
		})
	}
}
