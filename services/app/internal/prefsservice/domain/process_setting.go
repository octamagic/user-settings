package domain

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/soundcloud/gokit/v2/clients/okidoki"
	"github.com/soundcloud/user-settings/internal/prefsrepo"
	"github.com/soundcloud/user-settings/internal/rulesservice"
)

type OkidokiClient interface {
	FetchUsers(ctx context.Context, userUrns []string) ([]okidoki.OkidokiRawResponse, error)
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
	okidokiClient OkidokiClient
	rulesService  RulesService
	clock         Clock
	settingsRepo  SettingsRepository
}

type MinorSettingProcessor interface {
	ProcessMinorProtectionSettings(ctx context.Context, userUrns []string) ([]prefsrepo.Setting, error)
	FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]UserPrivacyAttributes, error)
}

func NewMinorSettingProcessor(okidokiClient OkidokiClient, rulesService RulesService, clock Clock, settingsRepo SettingsRepository) MinorSettingProcessor {
	return &minorSettingsProcessor{
		okidokiClient: okidokiClient,
		rulesService:  rulesService,
		clock:         clock,
		settingsRepo:  settingsRepo,
	}
}

func (m *minorSettingsProcessor) ProcessMinorProtectionSettings(ctx context.Context, userUrns []string) (
	[]prefsrepo.Setting,
	error,
) {
	okidokiResponse, err := m.okidokiClient.FetchUsers(ctx, userUrns)
	if err != nil {
		return nil, err
	}

	// Search for the user in okidokiResponse
	okidokiUsers, err := findUsersInOkidokiResponse(okidokiResponse, userUrns)
	if err != nil {
		return nil, err
	}

	var settings []prefsrepo.Setting

	for _, okidokiUser := range okidokiUsers {
		ruleForUser, err := getRulesForUser(okidokiUser, m.clock, m.rulesService)
		if err != nil {
			return nil, err
		}
		tempSetting, err := createSettingWithRule(ctx, okidokiUser.Self.URN, ruleForUser, m.settingsRepo)
		if err != nil {
			return nil, err
		}
		settings = append(settings, *tempSetting)
	}

	return settings, nil
}

func (m *minorSettingsProcessor) FetchUsersPrivacyAttributes(ctx context.Context, userUrns []string) ([]UserPrivacyAttributes, error) {
	okidokiResponse, err := m.okidokiClient.FetchUsers(ctx, userUrns)
	if err != nil {
		return nil, err
	}

	// Search for the user in okidokiResponse
	okidokiUsers, err := findUsersInOkidokiResponse(okidokiResponse, userUrns)
	if err != nil {
		return nil, err
	}

	var privacyAttributes []UserPrivacyAttributes

	for _, okidokiUser := range okidokiUsers {
		ruleForUser, err := getRulesForUser(okidokiUser, m.clock, m.rulesService)
		if err != nil {
			return nil, err
		}

		privacyAttributes = append(privacyAttributes, UserPrivacyAttributes{
			UserUrn: okidokiUser.Self.URN,
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

func getRulesForUser(okidokiUser okidoki.OkidokiRawResponse, clock Clock, rulesService RulesService) (rulesservice.RuleApplication, error) {
	age := calculateAgeFromDateOfBirth(okidokiUser.DateOfBirth, clock)

	country := ""
	if okidokiUser.CountryCode != nil {
		country = *okidokiUser.CountryCode
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

// findUsersInOkidokiResponse searches for users in the okidoki response by userUrn
func findUsersInOkidokiResponse(okidokiResponse []okidoki.OkidokiRawResponse, userUrns []string) ([]okidoki.OkidokiRawResponse, error) {
	var okidokiUsers []okidoki.OkidokiRawResponse
	for _, userUrn := range userUrns {
		okidokiUser, err := findUserInOkidokiResponse(okidokiResponse, userUrn)
		if err != nil {
			return nil, err
		}
		okidokiUsers = append(okidokiUsers, *okidokiUser)
	}
	return okidokiUsers, nil
}

func findUserInOkidokiResponse(okidokiResponse []okidoki.OkidokiRawResponse, userUrn string) (*okidoki.OkidokiRawResponse, error) {
	for _, user := range okidokiResponse {
		if user.Self.URN == userUrn {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found in okidoki response: %s", userUrn)
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
