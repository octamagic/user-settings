package e2e

import (
	"testing"

	proto "github.com/patrickisgreat/user-settings/internal/generated/usersettings/v1"
	"github.com/stretchr/testify/assert"
)

func TestFetchUsersPrivacyAttributes(t *testing.T) {
	userUrn := "urn:user:2"

	t.Run("returns valid response for user", func(t *testing.T) {
		res := sendRequest(t, "UserSettingsService/FetchUsersPrivacyAttributes", &proto.UsersPrivacyAttributesRequest{
			UserSession: userSession(userUrn),
			UserUrns:    []string{userUrn},
		})

		resBody := &proto.UsersPrivacyAttributesResponse{}
		decodeSuccessResponse(t, res, resBody)

		assert.NotNil(t, resBody.PrivacyAttributes)
		assert.Equal(t, userUrn, resBody.PrivacyAttributes[0].UserUrn)
		assert.False(t, resBody.PrivacyAttributes[0].IsMinor)
	})
}
