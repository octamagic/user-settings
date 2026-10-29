package prefsservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo/definitions"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const SettingName = "testSetting"

func createTestSetting(userUrn string) *prefsrepo.Setting {
	if userUrn == "" {
		userUrn = "soundcloud:users:123456"
	}
	return &prefsrepo.Setting{
		UserUrn:     userUrn,
		Name:        SettingName,
		SettingType: prefsrepo.SettingTypeBoolean.String(),
		UpdatedAt:   time.Now(),
		Value:       `true`,
	}
}

func TestGetSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := NewMockSettingsRepository(ctrl)
	rulesService := domain.NewMockRulesService(ctrl)
	userDirectory := domain.NewMockUserDirectory(ctrl)
	settingsDefinitionRepo := definitions.NewMockRepository(ctrl)

	minorSettingProcessor := domain.NewMinorSettingProcessor(userDirectory, rulesService, domain.DefaultClock{})
	service := NewSettingsService(mockRepo, minorSettingProcessor, settingsDefinitionRepo)

	// Test context
	ctx := context.Background()

	// Test setting
	testUserUrn := "soundcloud:users:123456"
	testSetting := createTestSetting(testUserUrn)

	// Test case 1: Get existing setting
	t.Run("Get existing settings", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, gomock.Any(), gomock.Any()).Return([]prefsrepo.Setting{*testSetting}, nil)

		setting, err := service.GetSettings(ctx, testUserUrn, []string{testSetting.Name})
		require.NoError(t, err)
		assert.NotNil(t, setting)
		assert.Equal(t, setting[0].Name, testSetting.Name)
		assert.Equal(t, setting[0].UserUrn, testUserUrn)
	})

	// Test case 2: Get non-existent setting
	t.Run("Get non-existent setting", func(t *testing.T) {
		name := "nonexistent"
		mockRepo.EXPECT().GetSettings(ctx, gomock.Any(), []string{name}).Return(nil, errors.New("not existing"))

		_, err := service.GetSettings(ctx, testUserUrn, []string{name})
		require.Error(t, err)
	})

	// Test case 3: Get with empty setting IDs
	t.Run("Get with empty setting names, returns all settings", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, gomock.Any(), gomock.Any()).Return(nil, errors.New("not existing"))

		_, err := service.GetSettings(ctx, testUserUrn, []string{})
		require.Error(t, err)
	})

	// Test case 4: Repository error
	t.Run("Repository error", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, testUserUrn, gomock.Any()).Return(nil, errors.New("mock rules service error"))

		_, err := service.GetSettings(ctx, testUserUrn, []string{testSetting.Name})
		require.Error(t, err)
	})
}

func TestGetPrivacySettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := NewMockSettingsRepository(ctrl)
	definitionsRepo := definitions.NewMockRepository(ctrl)
	minorSettingProcessor := domain.NewMockMinorSettingProcessor(ctrl)

	service := NewSettingsService(mockRepo, minorSettingProcessor, definitionsRepo)

	// Test context
	ctx := context.Background()

	// Test setting
	testUserUrn := "soundcloud:users:123456"
	testSetting := createTestSetting(testUserUrn)

	// Test case 1: Get existing setting
	t.Run("Get existing settings", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, gomock.Any(), gomock.Any()).Return([]prefsrepo.Setting{*testSetting}, nil)
		definitionsRepo.EXPECT().GetPrivacySettingsNames().Return([]string{testSetting.Name})
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return([]domain.UserPrivacyAttributes{
			{
				UserUrn: testUserUrn,
				IsMinor: false,
			},
		}, nil)

		usersPrivacySettings, err := service.GetPrivacySettings(ctx, []string{testUserUrn})
		require.NoError(t, err)
		assert.NotNil(t, usersPrivacySettings)
		assert.Len(t, usersPrivacySettings, 1)
		privacySettings := usersPrivacySettings[0].Settings
		assert.Len(t, privacySettings, 1)
		assert.Equal(t, privacySettings[0].Name, testSetting.Name)
		assert.Equal(t, privacySettings[0].UserUrn, testUserUrn)
	})

	t.Run("When setting was not in the repo, create it from definition", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, testUserUrn, gomock.Any()).Return([]prefsrepo.Setting{}, nil)
		definitionsRepo.EXPECT().GetPrivacySettingsNames().Return([]string{testSetting.Name})
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return([]domain.UserPrivacyAttributes{
			{
				UserUrn: testUserUrn,
				IsMinor: false,
			},
		}, nil)
		definitionsRepo.EXPECT().GetDefinition(testSetting.Name).Return(&definitions.SettingDefinition{
			Name: testSetting.Name,
			Type: testSetting.SettingType,
			Defaults: definitions.SettingDefaults{
				Global: "true",
				Minor:  "false",
			},
		}, nil)

		privacySettings, err := service.GetPrivacySettings(ctx, []string{testUserUrn})
		require.NoError(t, err)
		assert.NotNil(t, privacySettings)
		assert.Len(t, privacySettings, 1)
		assert.Equal(t, privacySettings[0].UserUrn, testUserUrn)
		assert.Equal(t, privacySettings[0].Settings[0].Name, testSetting.Name)
		assert.Equal(t, "true", privacySettings[0].Settings[0].Value)
		assert.Equal(t, privacySettings[0].Settings[0].SettingType, testSetting.SettingType)
	})

	t.Run("When setting was not in the repo and minor is true, create it from definition", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, testUserUrn, gomock.Any()).Return([]prefsrepo.Setting{}, nil)
		definitionsRepo.EXPECT().GetPrivacySettingsNames().Return([]string{testSetting.Name})
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return([]domain.UserPrivacyAttributes{
			{
				UserUrn: testUserUrn,
				IsMinor: true,
			},
		}, nil)
		definitionsRepo.EXPECT().GetDefinition(testSetting.Name).Return(&definitions.SettingDefinition{
			Name: testSetting.Name,
			Type: testSetting.SettingType,
			Defaults: definitions.SettingDefaults{
				Global: "true",
				Minor:  "false",
			},
		}, nil)

		privacySettings, err := service.GetPrivacySettings(ctx, []string{testUserUrn})
		require.NoError(t, err)
		assert.NotNil(t, privacySettings)
		assert.Len(t, privacySettings, 1)
		assert.Equal(t, privacySettings[0].UserUrn, testUserUrn)
		assert.Equal(t, privacySettings[0].Settings[0].Name, testSetting.Name)
		assert.Equal(t, "false", privacySettings[0].Settings[0].Value)
		assert.Equal(t, privacySettings[0].Settings[0].SettingType, testSetting.SettingType)
	})

	t.Run("Repository error, produce empty privacy settings", func(t *testing.T) {
		mockRepo.EXPECT().GetSettings(ctx, testUserUrn, gomock.Any()).Return(nil, errors.New("mock rules service error"))
		definitionsRepo.EXPECT().GetPrivacySettingsNames().Return([]string{testSetting.Name})
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return([]domain.UserPrivacyAttributes{
			{
				UserUrn: testUserUrn,
				IsMinor: true,
			},
		}, nil)

		privacySettings, err := service.GetPrivacySettings(ctx, []string{testUserUrn})
		require.NoError(t, err)
		assert.Empty(t, privacySettings)
	})
}

func TestPutSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := NewMockSettingsRepository(ctrl)
	minorSettingProcessor := domain.NewMockMinorSettingProcessor(ctrl)
	settingsDefinitionsRepo := definitions.NewMockRepository(ctrl)

	service := NewSettingsService(mockRepo, minorSettingProcessor, settingsDefinitionsRepo)

	// Test context
	ctx := context.Background()

	// Test data
	testUserUrn := "soundcloud:users:123456"
	testSetting := createTestSetting(testUserUrn)

	// Test case 1: Put new setting
	t.Run("Put new setting", func(t *testing.T) {
		mockRepo.EXPECT().PutSetting(ctx, gomock.Any()).Return(nil)

		err := service.PutSetting(ctx, *testSetting)
		require.NoError(t, err)
	})

	t.Run("Invalid JSON value is rejected before reaching the repository", func(t *testing.T) {
		invalid := *testSetting
		invalid.Value = "{not json"

		err := service.PutSetting(ctx, invalid)
		assert.ErrorIs(t, err, ErrInvalidValue)
	})

	// Test case 3: Repository error
	t.Run("Repository error", func(t *testing.T) {
		mockRepo.EXPECT().PutSetting(ctx, gomock.Any()).Return(errors.New("put failed"))

		err := service.PutSetting(ctx, *testSetting)
		require.Error(t, err)
	})
}

func TestFetchUsersPrivacyAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock repository
	mockRepo := NewMockSettingsRepository(ctrl)
	minorSettingProcessor := domain.NewMockMinorSettingProcessor(ctrl)
	definitionsRepo := definitions.NewMockRepository(ctrl)

	service := NewSettingsService(mockRepo, minorSettingProcessor, definitionsRepo)

	ctx := context.Background()

	testUserUrn := "soundcloud:users:123456"

	t.Run("Fetch attributes for user", func(t *testing.T) {
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return([]domain.UserPrivacyAttributes{
			{
				UserUrn: testUserUrn,
				IsMinor: false,
			},
		}, nil)

		privacyAttributes, err := service.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		assert.Len(t, privacyAttributes, 1)
	})

	t.Run("Returns error if error happen", func(t *testing.T) {
		minorSettingProcessor.EXPECT().FetchUsersPrivacyAttributes(ctx, gomock.Any()).Return(nil, errors.New("error happen"))

		privacyAttributes, err := service.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, privacyAttributes)
		require.Error(t, err)
	})

}
