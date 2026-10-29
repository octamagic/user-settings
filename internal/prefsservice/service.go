//go:generate mockgen -source=service.go -destination=mock_service.go -package=prefsservice
package prefsservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo/definitions"
	"github.com/patrickisgreat/user-settings/internal/prefsservice/domain"
	"golang.org/x/sync/errgroup"
)

// ErrInvalidValue is returned when a setting value is not a valid JSON document.
var ErrInvalidValue = errors.New("setting value is not valid JSON")

type SettingsRepository interface {
	PutSetting(ctx context.Context, setting prefsrepo.Setting) error
	GetSettings(ctx context.Context, userUrn string, names []string) ([]prefsrepo.Setting, error)
}

// maxConcurrentUserLookups bounds the fan-out in GetPrivacySettings.
const maxConcurrentUserLookups = 10

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
		logging.FromContext(ctx).Warn("missing minor attributes for the users", slog.Any("err", err))
	}
	attributesByUrn := make(map[string]domain.UserPrivacyAttributes, len(usersAttributes))
	for _, attribute := range usersAttributes {
		attributesByUrn[attribute.UserUrn] = attribute
	}

	// Each user lives in its own DynamoDB partition, so the lookups are
	// independent. Fan out with a bound so a large batch cannot exhaust
	// connections; results keep the request order.
	results := make([]*UserPrivacySettings, len(userUrns))
	var group errgroup.Group
	group.SetLimit(maxConcurrentUserLookups)
	for i, userUrn := range userUrns {
		group.Go(func() error {
			results[i] = s.privacySettingsForUser(ctx, userUrn, privacySettingNames, attributesByUrn)
			return nil
		})
	}
	_ = group.Wait() // workers never return an error; failures are logged and skipped

	privacySettings := make([]UserPrivacySettings, 0, len(userUrns))
	for _, result := range results {
		if result != nil {
			privacySettings = append(privacySettings, *result)
		}
	}
	return privacySettings, nil
}

// privacySettingsForUser loads one user's privacy settings and fills in
// definition defaults for any that are unset. It returns nil, after logging,
// when the user cannot be served, so one bad user does not fail the batch.
func (s *settingsService) privacySettingsForUser(
	ctx context.Context,
	userUrn string,
	privacySettingNames []string,
	attributesByUrn map[string]domain.UserPrivacyAttributes,
) *UserPrivacySettings {
	logger := logging.FromContext(ctx).With(slog.String("userUrn", userUrn))

	settings, err := s.settingsRepo.GetSettings(ctx, userUrn, privacySettingNames)
	if err != nil {
		logger.Warn("missing privacy settings for the user", slog.Any("err", err))
		return nil
	}

	userAttribute, ok := attributesByUrn[userUrn]
	if !ok {
		logger.Warn("missing privacy attributes for the user")
		return nil
	}

	if len(settings) != len(privacySettingNames) {
		missingSettingNames := s.getMissingSettingNames(privacySettingNames, settings)
		settings = append(settings, s.settingsFromDefaults(ctx, missingSettingNames, userUrn, &userAttribute)...)
	}

	return &UserPrivacySettings{UserUrn: userUrn, Settings: settings}
}

// PutSetting creates or updates the setting
func (s *settingsService) PutSetting(ctx context.Context, setting prefsrepo.Setting) error {
	// Validate that the value is valid JSON
	if setting.Value != "" {
		var jsonObj interface{}
		if err := json.Unmarshal([]byte(setting.Value), &jsonObj); err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidValue, err)
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
