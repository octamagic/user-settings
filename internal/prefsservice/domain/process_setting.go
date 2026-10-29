//go:generate mockgen -source=process_setting.go -destination=mock_process_setting.go -package=domain
package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/patrickisgreat/user-settings/internal/userdirectory"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/rulesservice"
)

type UserDirectory interface {
	FetchUsers(ctx context.Context, userUrns []string) ([]userdirectory.User, error)
}

type RulesService interface {
	ApplyRules(age int, countryCode string) (rulesservice.RuleApplication, error)
}

type SettingsRepository interface {
	PutSetting(ctx context.Context, setting prefsrepo.Setting) error
}

type UserPrivacyAttributes struct {
	UserUrn string
	IsMinor bool
}

type minorSettingsProcessor struct {
	userDirectory UserDirectory
	rulesService  RulesService
	clock         Clock
	settingsRepo  SettingsRepository
}

type MinorSettingProcessor interface {
	ProcessMinorProtectionSettings(ctx context.Context, userUrns []string) ([]prefsrepo.Setting, error)
	FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]UserPrivacyAttributes, error)
}

func NewMinorSettingProcessor(userDirectory UserDirectory, rulesService RulesService, clock Clock, settingsRepo SettingsRepository) MinorSettingProcessor {
	return &minorSettingsProcessor{
		userDirectory: userDirectory,
		rulesService:  rulesService,
		clock:         clock,
		settingsRepo:  settingsRepo,
	}
}

func (m *minorSettingsProcessor) ProcessMinorProtectionSettings(ctx context.Context, userUrns []string) (
	[]prefsrepo.Setting,
	error,
) {
	directoryUsers, err := m.userDirectory.FetchUsers(ctx, userUrns)
	if err != nil {
		return nil, err
	}

	// Match the requested URNs against the directory response
	users, err := findUsers(directoryUsers, userUrns)
	if err != nil {
		return nil, err
	}

	var settings []prefsrepo.Setting

	for _, user := range users {
		ruleForUser, err := getRulesForUser(user, m.clock, m.rulesService)
		if err != nil {
			return nil, err
		}
		tempSetting, err := createSettingWithRule(ctx, user.URN, ruleForUser, m.settingsRepo)
		if err != nil {
			return nil, err
		}
		settings = append(settings, *tempSetting)
	}

	return settings, nil
}

func (m *minorSettingsProcessor) FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]UserPrivacyAttributes, error) {
	directoryUsers, err := m.userDirectory.FetchUsers(ctx, userUrns)
	if err != nil {
		return nil, err
	}

	// Match the requested URNs against the directory response
	users, err := findUsers(directoryUsers, userUrns)
	if err != nil {
		return nil, err
	}

	var privacyAttributes []UserPrivacyAttributes

	for _, user := range users {
		ruleForUser, err := getRulesForUser(user, m.clock, m.rulesService)
		if err != nil {
			return nil, err
		}

		privacyAttributes = append(privacyAttributes, UserPrivacyAttributes{
			UserUrn: user.URN,
			IsMinor: rulesservice.ApplyRestriction == ruleForUser,
		})
	}

	return privacyAttributes, nil
}

// mapDateOfBirth parses a raw date of birth string in format "YYYY/MM"
func mapDateOfBirth(rawDateOfBirth *string) (time.Time, error) {
	if rawDateOfBirth == nil || *rawDateOfBirth == "" {
		return time.Time{}, nil
	}

	parts := strings.Split(*rawDateOfBirth, "/")
	if len(parts) != 3 {
		return time.Time{}, nil
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, nil
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, nil
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, nil
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

func getRulesForUser(user userdirectory.User, clock Clock, rulesService RulesService) (rulesservice.RuleApplication, error) {
	age := calculateAgeFromDateOfBirth(user.DateOfBirth, clock)

	country := ""
	if user.CountryCode != nil {
		country = *user.CountryCode
	}

	ruleApplication, err := rulesService.ApplyRules(age, country)
	if err != nil {
		return rulesservice.ApplyRestriction, err
	}

	return ruleApplication, err
}

func createSettingWithRule(ctx context.Context, userUrn string, rule rulesservice.RuleApplication, settingsRepo SettingsRepository) (*prefsrepo.Setting, error) {
	// Create minor protection setting
	tempSetting, err := createMinorProtectionSetting(userUrn, rule == rulesservice.ApplyRestriction)

	if err != nil {
		return nil, err
	}

	err = settingsRepo.PutSetting(ctx, *tempSetting)
	if err != nil {
		return nil, err
	}

	return tempSetting, nil
}

// calculateAgeFromDateOfBirth calculates age from a date of birth string
func calculateAgeFromDateOfBirth(rawDateOfBirth *string, clock Clock) int {
	dob, err := mapDateOfBirth(rawDateOfBirth)
	if err != nil {
		return 0 // Default age when date of birth is invalid
	}

	// to be on safe side due to not having sometime correct day we add 1 month to keep user protected
	dob = dob.AddDate(0, 1, 0)
	difference := clock.Now().Sub(dob)

	years := int(difference.Hours() / 24 / 365)

	return years
}

// findUsers returns the directory record for every requested URN, in request order
func findUsers(directoryUsers []userdirectory.User, userUrns []string) ([]userdirectory.User, error) {
	var users []userdirectory.User
	for _, userUrn := range userUrns {
		user, err := findUser(directoryUsers, userUrn)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, nil
}

func findUser(directoryUsers []userdirectory.User, userUrn string) (*userdirectory.User, error) {
	for _, user := range directoryUsers {
		if user.URN == userUrn {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found in directory response: %s", userUrn)
}

// createMinorProtectionSetting creates a minor protection setting based on rule application
func createMinorProtectionSetting(userUrn string, isRestricted bool) (*prefsrepo.Setting, error) {
	return &prefsrepo.Setting{
		UserUrn:     userUrn,
		Name:        "minor_protection",
		SettingType: prefsrepo.SettingTypeBoolean.String(),
		Value:       strconv.FormatBool(isRestricted),
		UpdatedAt:   time.Now(),
	}, nil
}
