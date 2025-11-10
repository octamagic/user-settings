package integration

import (
	"context"
	"testing"

	"github.com/soundcloud/user-settings/internal/prefsrepo"
	"github.com/stretchr/testify/assert"
)

func TestBatchOperations(t *testing.T) {
	setupDynamoDB(t)
	defer cleanupDynamoDB(t)

	ctx := context.Background()

	// Test data
	userUrn := "soundcloud:users:12345"
	name1 := "testSetting"
	name2 := "testSetting2"
	name3 := "testSetting3"
	settingType := prefsrepo.SettingTypeBoolean
	value := `true`

	// Create multiple settings for batch operations
	settings := []prefsrepo.Setting{
		prefsrepo.NewSetting(userUrn, name1, settingType, value),
		prefsrepo.NewSetting(userUrn, name2, settingType, value),
		prefsrepo.NewSetting(userUrn, name3, settingType, value),
	}

	// Test batch creation
	t.Run("BatchPutSettings", func(t *testing.T) {
		// Insert all settings in a batch
		for _, setting := range settings {
			err := testRepo.PutSetting(ctx, setting)
			assert.Nil(t, err)
		}

		// Verify all settings were created
		for _, setting := range settings {
			fetchedSetting, err := testRepo.GetSettings(ctx, userUrn, []string{setting.Name})
			assert.Nil(t, err)
			assert.Equal(t, fetchedSetting[0].Name, setting.Name)
		}
	})

	// Test batch retrieval
	t.Run("BatchGetSettings", func(t *testing.T) {
		// Get settings for multiple IDs
		names := []string{settings[0].Name, settings[1].Name}

		// Note: The repository doesn't have a batch get method yet,
		// so we'll implement a simple loop as a simulation
		for _, name := range names {
			fetchedSetting, err := testRepo.GetSettings(ctx, userUrn, []string{name})
			assert.Nil(t, err)
			assert.Equal(t, fetchedSetting[0].Name, name)
		}
	})
}
