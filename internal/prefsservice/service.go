//go:generate mockgen -source=service.go -destination=mock_service.go -package=prefsservice
package prefsservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo/definitions"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
)

type SettingsRepository interface {
	PutSetting(ctx context.Context, setting prefsrepo.Setting) error
	GetSettings(ctx context.Context, userUrn string, names []string) ([]prefsrepo.Setting, error)
}

type SettingsService interface {
	GetSettings(ctx context.Context, userUrn string, names []string) ([]prefsrepo.Setting, error)
	GetPrivacySettings(ctx context.Context, userUrns []string) ([]UserPrivacySettings, error)
	PutSetting(ctx context.Context, setting prefsrepo.Setting) error
	FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]domain.UserPrivacyAttributes, error)
}

type UserPrivacySettings struct {
	UserUrn  string
	Settings []prefsrepo.Setting
}

// settingsService implements the SettingsService interface
type settingsService struct {
	settingsRepo           SettingsRepository
	minorSettingProcessor  domain.MinorSettingProcessor
	settingsDefinitionRepo definitions.Repository
}

func NewSettingsService(
	settingsRepository SettingsRepository,
	minorSettingProcessor domain.MinorSettingProcessor,
	settingsDefinitionRepo definitions.Repository,
) SettingsService {
	// Now accepts any type that implements SettingsRepository
	return &settingsService{
		settingsRepo:           settingsRepository,
		minorSettingProcessor:  minorSettingProcessor,
		settingsDefinitionRepo: settingsDefinitionRepo,
	}
}

// GetSettings retrieves a settings by user and conditionally filter by the names
func (s *settingsService) GetSettings(ctx context.Context, userUrn string, names []string) ([]prefsrepo.Setting, error) {
	var settings, err = s.settingsRepo.GetSettings(ctx, userUrn, names)

	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (s *settingsService) GetPrivacySettings(ctx context.Context, userUrns []string) ([]UserPrivacySettings, error) {
	privacySettingNames := s.settingsDefinitionRepo.GetPrivacySettingsNames()
	usersAttributes, err := s.minorSettingProcessor.FetchUsersPrivacyAttributes(ctx, userUrns)
	if err != nil {
		logging.FromContext(ctx).Warn("missing minor attributes for the users")
	}

	// this needs to be parallel ???
	var privacySettings []UserPrivacySettings
	for _, userUrn := range userUrns {
		var settings, err = s.settingsRepo.GetSettings(ctx, userUrn, privacySettingNames)
		if err != nil {
			logging.FromContext(ctx).Warn("missing privacy settings for the user: ", slog.String("userUrn", userUrn))
			continue
		}

		var userAttribute *domain.UserPrivacyAttributes
		for _, attribute := range usersAttributes {
			if attribute.UserUrn == userUrn {
				userAttribute = &attribute
				break
			}
		}
		if userAttribute == nil {
			logging.FromContext(ctx).Warn("missing userPrivacy attributes for user: ", slog.String("userUrn", userUrn))
			continue
		}

		if len(settings) != len(privacySettingNames) {
			missingSettingNames := s.getMissingSettingNames(privacySettingNames, settings)
			createdSettings := s.settingsFromDefaults(ctx, missingSettingNames, userUrn, userAttribute)
			settings = append(settings, createdSettings...)
		}

		privacySettings = append(privacySettings, UserPrivacySettings{
			UserUrn:  userUrn,
			Settings: settings,
		})
	}

	return privacySettings, nil
}

// PutSetting creates or updates the setting
func (s *settingsService) PutSetting(ctx context.Context, setting prefsrepo.Setting) error {
	// Validate that the value is valid JSON
	if setting.Value != "" {
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(setting.Value), &jsonObj); err != nil {
			return fmt.Errorf("invalid JSON in setting value: %w", err)
		}
	}

	return s.settingsRepo.PutSetting(ctx, setting)
}

func (s *settingsService) FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]domain.UserPrivacyAttributes, error) {
	return s.minorSettingProcessor.FetchUsersPrivacyAttributes(ctx, userUrns)
}

func (s *settingsService) settingsFromDefaults(ctx context.Context, missingSettingNames []string, userUrn string, userAttributes *domain.UserPrivacyAttributes) []prefsrepo.Setting {
	var settings []prefsrepo.Setting

	for _, missingSettingName := range missingSettingNames {
		settingDefinition, err := s.settingsDefinitionRepo.GetDefinition(missingSettingName)
		if err != nil {
			logging.FromContext(ctx).Warn("missing privacy setting definition: ", slog.String("settingName", missingSettingName))
			continue
		}

		createdSetting := s.createSettingFromDefinition(settingDefinition, userAttributes, userUrn, missingSettingName)
		settings = append(settings, createdSetting)
	}

	return settings
}

func (s *settingsService) createSettingFromDefinition(settingDefinition *definitions.SettingDefinition, userAttributes *domain.UserPrivacyAttributes, userUrn string, missingSettingName string) prefsrepo.Setting {
	var defaultValue = settingDefinition.Defaults.Global
	if settingDefinition.Defaults.Minor != "" {
		if userAttributes.IsMinor {
			defaultValue = settingDefinition.Defaults.Minor
		} else {
			defaultValue = settingDefinition.Defaults.Global
		}
	}

	return prefsrepo.Setting{
		UserUrn:     userUrn,
		Name:        missingSettingName,
		Value:       defaultValue,
		SettingType: settingDefinition.Type,
		UpdatedAt:   time.Now(),
	}
}

func (s *settingsService) getMissingSettingNames(settingNames []string, settings []prefsrepo.Setting) []string {
	missingSettingNames := make([]string, 0)
	for _, settingName := range settingNames {
		if !slices.ContainsFunc(settings, func(setting prefsrepo.Setting) bool {
			return setting.Name == settingName
		}) {
			missingSettingNames = append(missingSettingNames, settingName)
		}
	}
	return missingSettingNames
}
