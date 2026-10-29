//go:generate mockgen -source=definitions.go -destination=mock_definitions.go -package=definitions
package definitions

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ErrDefinitionNotFound is returned when no definition exists for a setting name.
var ErrDefinitionNotFound = errors.New("setting definition not found")

type SettingDefaults struct {
	Global string `json:"global"`
	Minor  string `json:"minor"`
}

type SettingDefinition struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Privacy     bool            `json:"privacy"`
	Defaults    SettingDefaults `json:"defaults"`
	Description string          `json:"description"`
}

type Repository interface {
	GetDefinition(settingName string) (*SettingDefinition, error)
	GetSettingsNames() []string
	GetPrivacySettingsNames() []string
}

// Repository in-memory store for the settings definitions
type definitionsRepository struct {
	settings             []SettingDefinition
	settingsNames        []string
	privacySettingsNames []string
}

func NewDefinitionsRepository() Repository {
	settings := getSettings()
	return &definitionsRepository{
		settings:             settings,
		settingsNames:        getSettingNames(settings),
		privacySettingsNames: getPrivacySettingNames(settings),
	}
}

func (repo *definitionsRepository) GetDefinition(settingName string) (*SettingDefinition, error) {
	for _, setting := range repo.settings {
		if setting.Name == settingName {
			return &setting, nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrDefinitionNotFound, settingName)
}

func (repo *definitionsRepository) GetSettingsNames() []string {
	return repo.settingsNames
}

func (repo *definitionsRepository) GetPrivacySettingsNames() []string {
	return repo.privacySettingsNames
}

func getSettings() []SettingDefinition {
	var settings []SettingDefinition
	err := json.Unmarshal([]byte(Definitions), &settings)
	if err != nil {
		panic(err)
	}

	return settings
}

func getSettingNames(definitions []SettingDefinition) []string {
	names := make([]string, 0)
	for _, setting := range definitions {
		names = append(names, setting.Name)
	}
	return names
}

func getPrivacySettingNames(definitions []SettingDefinition) []string {
	names := make([]string, 0)
	for _, setting := range definitions {
		if setting.Privacy {
			names = append(names, setting.Name)
		}
	}
	return names
}
