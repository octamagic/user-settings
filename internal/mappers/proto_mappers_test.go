package mappers

import (
	"testing"
	"time"

	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/stretchr/testify/assert"
)

func TestSettingConversion(t *testing.T) {
	mapper := ProtoMapper{}

	// Test data
	now := time.Now()
	repoSetting := &prefsrepo.Setting{
		UserUrn:     "soundcloud:users:123456",
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
		assert.Equal(t, protoSetting.Type, proto.SettingType_BOOLEAN)
		assert.Equal(t, protoSetting.Value, repoSetting.Value)
	})

	// Test SettingFromProto
	t.Run("SettingFromProto", func(t *testing.T) {
		protoSetting := &proto.Setting{
			UserUrn:   "soundcloud:users:123456",
			Name:      "testSetting",
			Type:      proto.SettingType_BOOLEAN,
			UpdatedAt: now.Unix(),
			Value:     `true`,
		}

		repoSetting, err := mapper.SettingFromProto(protoSetting)

		assert.Nil(t, err)
		assert.Equal(t, protoSetting.UserUrn, repoSetting.UserUrn)
		assert.Equal(t, protoSetting.UpdatedAt, repoSetting.UpdatedAt.Unix())
		assert.Equal(t, protoSetting.Name, repoSetting.Name)
		assert.Equal(t, prefsrepo.SettingTypeBoolean.String(), repoSetting.SettingType)
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

		assert.Equal(t, len(privacyAttributes), len(protoAttributes))
		assert.Equal(t, privacyAttributes[0].UserUrn, protoAttributes[0].UserUrn)
		assert.Equal(t, privacyAttributes[0].IsMinor, protoAttributes[0].IsMinor)
	})
}
