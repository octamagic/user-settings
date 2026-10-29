package definitions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSettings(t *testing.T) {
	repo := NewDefinitionsRepository()

	t.Run("Get privacy settings names", func(t *testing.T) {
		names := repo.GetPrivacySettingsNames()
		assert.Contains(t, names, "liked_by_playlist")
	})

	t.Run("Get definition", func(t *testing.T) {
		definition, err := repo.GetDefinition("liked_by_playlist")
		assert.Nil(t, err)
		assert.Equal(t, definition.Name, "liked_by_playlist")
		assert.Equal(t, definition.Type, "b")
		assert.Equal(t, definition.Privacy, true)
		assert.Equal(t, definition.Defaults.Global, "true")
		assert.Equal(t, definition.Defaults.Minor, "false")
	})
}
