//go:generate mockgen -source=process_setting.go -destination=mock_process_setting.go -package=domain
package domain

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/patrickisgreat/user-settings/internal/logging"
	"github.com/patrickisgreat/user-settings/internal/prefsrepo"
	"github.com/patrickisgreat/user-settings/internal/rulesservice"
	"github.com/patrickisgreat/user-settings/internal/userdirectory"
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
		ruleForUser, err := getRulesForUser(ctx, user, m.clock, m.rulesService)
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
		ruleForUser, err := getRulesForUser(ctx, user, m.clock, m.rulesService)
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

// errUnknownDateOfBirth marks a directory record whose date of birth is
// missing or malformed. Callers treat such users as minors (fail closed).
var errUnknownDateOfBirth = errors.New("date of birth is missing or malformed")

// parseDateOfBirth parses the directory's "YYYY/MM/DD" format. Any missing,
// non-numeric or impossible date (such as 2000/02/31) is an error.
func parseDateOfBirth(raw *string) (time.Time, error) {
	if raw == nil || *raw == "" {
		return time.Time{}, errUnknownDateOfBirth
	}
	parts := strings.Split(*raw, "/")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("%w: %q", errUnknownDateOfBirth, *raw)
	}
	fields := make([]int, 3)
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %q", errUnknownDateOfBirth, *raw)
		}
		fields[i] = n
	}
	year, month, day := fields[0], time.Month(fields[1]), fields[2]
	dob := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	// time.Date normalises out-of-range values, so compare back to catch them.
	if dob.Year() != year || dob.Month() != month || dob.Day() != day {
		return time.Time{}, fmt.Errorf("%w: %q", errUnknownDateOfBirth, *raw)
	}
	return dob, nil
}

// ageAt returns the number of whole years between dob and now.
func ageAt(dob, now time.Time) int {
	years := now.Year() - dob.Year()
	birthdayThisYear := dob.AddDate(years, 0, 0)
	if now.Before(birthdayThisYear) {
		years--
	}
	return max(years, 0)
}

// ageFromDateOfBirth computes a user's age for rule evaluation. An unknown
// date of birth yields 0, the most protective assumption, together with the
// parse error so the caller can log it.
func ageFromDateOfBirth(raw *string, now time.Time) (int, error) {
	dob, err := parseDateOfBirth(raw)
	if err != nil {
		return 0, err
	}
	// The directory's day-of-month is not always reliable. Shifting the date
	// of birth one month later makes a user count as younger around their
	// birthday, which errs on the side of protection.
	dob = dob.AddDate(0, 1, 0)
	return ageAt(dob, now), nil
}

func getRulesForUser(ctx context.Context, user userdirectory.User, clock Clock, rulesService RulesService) (rulesservice.RuleApplication, error) {
	age, err := ageFromDateOfBirth(user.DateOfBirth, clock.Now())
	if err != nil {
		logging.FromContext(ctx).Warn("treating user with unknown date of birth as a minor",
			slog.String("userUrn", user.URN), slog.Any("err", err))
	}

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
