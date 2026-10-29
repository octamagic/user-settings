package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/octamagic/user-settings/internal/rulesservice"
	"github.com/octamagic/user-settings/internal/userdirectory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFetchUsersPrivacyAttributesDobCalculation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDirectory := NewMockUserDirectory(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	processor := NewMinorSettingProcessor(mockUserDirectory, mockRulesService, mockClock)

	// Test context
	ctx := context.Background()

	// Test setting
	testDate := time.Now()
	eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
	eighteenYearsAgoMinusMonth := time.Now().AddDate(-18, -1, 0)

	testUserUrn := "urn:user:123456"
	testUserCountryCode := "US"
	testUserDateOfBirth := fmt.Sprintf("%d/%d/%d", eighteenYearsAgo.Year(), eighteenYearsAgo.Month(), eighteenYearsAgo.Day())
	testUserDateOfBirth2 := fmt.Sprintf("%d/%d/%d", eighteenYearsAgoMinusMonth.Year(), eighteenYearsAgoMinusMonth.Month(), eighteenYearsAgoMinusMonth.Day())
	directoryUserEighteen := userdirectory.User{
		URN:         testUserUrn,
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth,
	}

	directoryUserEighteenMinusMonth := userdirectory.User{
		URN:         testUserUrn,
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth2,
	}

	t.Run("Restrict user Age if DOB is 18 years ago", func(t *testing.T) {
		const ExpectedSeventeen = 17

		mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]userdirectory.User{
			directoryUserEighteen,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(ExpectedSeventeen, gomock.Any()).Return(rulesservice.ApplyRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		require.NoError(t, err)
		assert.NotNil(t, setting)
	})

	t.Run("Not Restrict user Age if DOB is 18 years ago + 1 month", func(t *testing.T) {
		const ExpectedEighteen = 18
		mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]userdirectory.User{
			directoryUserEighteenMinusMonth,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(ExpectedEighteen, gomock.Any()).Return(rulesservice.NoRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		require.NoError(t, err)
		assert.NotNil(t, setting)
	})
}

func TestFetchUsersPrivacyAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserDirectory := NewMockUserDirectory(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	processor := NewMinorSettingProcessor(mockUserDirectory, mockRulesService, mockClock)

	// Test context
	ctx := context.Background()

	// Test setting
	testUserUrn := "urn:user:123456"
	testUserCountryCode := "US"
	testUserDateOfBirth := "2000/01/01"
	testDate := time.Now()
	directoryUser := userdirectory.User{
		URN:         testUserUrn,
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth,
	}

	t.Run("FetchUsersPrivacyAttributes returns user attributes", func(t *testing.T) {
		mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]userdirectory.User{
			directoryUser,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		require.NoError(t, err)
		assert.NotNil(t, setting)
	})

	t.Run("FetchUsersPrivacyAttributes returns error if rulesService fails", func(t *testing.T) {
		mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]userdirectory.User{
			directoryUser,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, errors.New("error"))

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		require.Error(t, err)
	})

	t.Run("FetchUsersPrivacyAttributes returns error if the directory fails", func(t *testing.T) {
		mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return(nil, errors.New("error"))

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		require.Error(t, err)
	})
}

func TestFetchUsersPrivacyAttributesFailsClosedWithoutDateOfBirth(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockUserDirectory := NewMockUserDirectory(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	processor := NewMinorSettingProcessor(mockUserDirectory, mockRulesService, mockClock)
	ctx := context.Background()
	country := "GB"

	for name, dob := range map[string]*string{"missing": nil, "malformed": ptr("not-a-date")} {
		t.Run(name+" date of birth is evaluated as age 0", func(t *testing.T) {
			mockUserDirectory.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]userdirectory.User{
				{URN: "urn:user:1", CountryCode: &country, DateOfBirth: dob},
			}, nil)
			mockClock.EXPECT().Now().Return(time.Now())
			mockRulesService.EXPECT().ApplyRules(0, country).Return(rulesservice.ApplyRestriction, nil)

			attrs, err := processor.FetchUsersPrivacyAttributes(ctx, []string{"urn:user:1"})

			require.NoError(t, err)
			assert.True(t, attrs[0].IsMinor)
		})
	}
}
