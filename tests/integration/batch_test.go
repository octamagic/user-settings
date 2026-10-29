package integration

import (
	"context"
	"testing"

	"github.com/octamagic/user-settings/internal/prefsrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSettingsFiltersByName(t *testing.T) {
	setupDynamoDB(t)
	defer cleanupDynamoDB(t)

	ctx := context.Background()
	userUrn := "urn:user:12345"
	settings := []prefsrepo.Setting{
		prefsrepo.NewSetting(userUrn, "first", prefsrepo.SettingTypeBoolean, `true`),
		prefsrepo.NewSetting(userUrn, "second", prefsrepo.SettingTypeNumber, `2`),
		prefsrepo.NewSetting(userUrn, "third", prefsrepo.SettingTypeString, `"three"`),
	}
	for _, setting := range settings {
		require.NoError(t, testRepo.PutSetting(ctx, setting))
	}

	t.Run("a subset of names returns only those settings", func(t *testing.T) {
		fetched, err := testRepo.GetSettings(ctx, userUrn, []string{"first", "third"})
		require.NoError(t, err)
		names := make([]string, 0, len(fetched))
		for _, setting := range fetched {
			names = append(names, setting.Name)
		}
		assert.ElementsMatch(t, []string{"first", "third"}, names)
	})

	t.Run("no names returns every setting for the user", func(t *testing.T) {
		fetched, err := testRepo.GetSettings(ctx, userUrn, nil)
		require.NoError(t, err)
		assert.Len(t, fetched, 3)
	})

	t.Run("another user's partition is not visible", func(t *testing.T) {
		fetched, err := testRepo.GetSettings(ctx, "urn:user:other", nil)
		require.NoError(t, err)
		assert.Empty(t, fetched)
	})
}
