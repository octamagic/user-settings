package definitions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSettings(t *testing.T) {
	repo := NewDefinitionsRepository()

	t.Run("Get privacy settings names", func(t *testing.T) {
		names := repo.GetPrivacySettingsNames()
		assert.Contains(t, names, "liked_by_playlist")
	})

	t.Run("Get definition", func(t *testing.T) {
		definition, err := repo.GetDefinition("liked_by_playlist")
		require.NoError(t, err)
		assert.Equal(t, "liked_by_playlist", definition.Name)
		assert.Equal(t, "b", definition.Type)
		assert.True(t, definition.Privacy)
		assert.Equal(t, "true", definition.Defaults.Global)
		assert.Equal(t, "false", definition.Defaults.Minor)
	})
}
