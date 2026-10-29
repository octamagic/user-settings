package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/soundcloud/gokit/v2/clients/okidoki"
	"github.com/soundcloud/user-settings/internal/rulesservice"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOkidokiClient := NewMockOkidokiClient(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	mockSettingsRepo := NewMockSettingsRepository(ctrl)
	processor := NewMinorSettingProcessor(mockOkidokiClient, mockRulesService, mockClock, mockSettingsRepo)

	// Test context
	ctx := context.Background()

	// Test setting
	testUserUrn := "soundcloud:users:123456"
	testUserCountryCode := "US"
	testUserDateOfBirth := "2000/01/01"
	testDate := time.Now()
	okidokiResponse := okidoki.OkidokiRawResponse{
		Self: okidoki.Self{
			URN: testUserUrn,
		},
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth,
	}

	t.Run("ProcessMinorProtectionSettings returns setting for user", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponse,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, nil)
		mockSettingsRepo.EXPECT().PutSetting(ctx, gomock.Any()).Return(nil)

		setting, err := processor.ProcessMinorProtectionSettings(ctx, []string{testUserUrn})

		assert.Nil(t, err)
		assert.NotNil(t, setting)
	})

	t.Run("ProcessMinorProtectionSettings returns error if setting creation failed", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponse,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, nil)
		mockSettingsRepo.EXPECT().PutSetting(ctx, gomock.Any()).Return(errors.New("error"))

		setting, err := processor.ProcessMinorProtectionSettings(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		assert.NotNil(t, err)
	})

	t.Run("ProcessMinorProtectionSettings returns error if rules returned error", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponse,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, errors.New("error"))

		setting, err := processor.ProcessMinorProtectionSettings(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		assert.NotNil(t, err)
	})

	t.Run("ProcessMinorProtectionSettings returns error if okidoki did not return users", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return(nil, errors.New("okidoki error"))

		setting, err := processor.ProcessMinorProtectionSettings(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		assert.NotNil(t, err)
	})
}

func TestFetchUsersPrivacyAttributesDobCalculation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOkidokiClient := NewMockOkidokiClient(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	mockSettingsRepo := NewMockSettingsRepository(ctrl)
	processor := NewMinorSettingProcessor(mockOkidokiClient, mockRulesService, mockClock, mockSettingsRepo)

	// Test context
	ctx := context.Background()

	// Test setting
	testDate := time.Now()
	eighteenYearsAgo := time.Now().AddDate(-18, 0, 0)
	eighteenYearsAgoMinusMonth := time.Now().AddDate(-18, -1, 0)

	testUserUrn := "soundcloud:users:123456"
	testUserCountryCode := "US"
	testUserDateOfBirth := fmt.Sprintf("%d/%d/%d", eighteenYearsAgo.Year(), eighteenYearsAgo.Month(), eighteenYearsAgo.Day())
	testUserDateOfBirth2 := fmt.Sprintf("%d/%d/%d", eighteenYearsAgoMinusMonth.Year(), eighteenYearsAgoMinusMonth.Month(), eighteenYearsAgoMinusMonth.Day())
	okidokiResponseEighteen := okidoki.OkidokiRawResponse{
		Self: okidoki.Self{
			URN: testUserUrn,
		},
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth,
	}

	okidokiResponseEighteenMinusMonth := okidoki.OkidokiRawResponse{
		Self: okidoki.Self{
			URN: testUserUrn,
		},
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth2,
	}

	t.Run("Restrict user Age if DOB is 18 years ago", func(t *testing.T) {
		const ExpectedSeventeen = 17

		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponseEighteen,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(ExpectedSeventeen, gomock.Any()).Return(rulesservice.ApplyRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, err)
		assert.NotNil(t, setting)
	})

	t.Run("Not Restrict user Age if DOB is 18 years ago + 1 month", func(t *testing.T) {
		const ExpectedEighteen = 18
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponseEighteenMinusMonth,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(ExpectedEighteen, gomock.Any()).Return(rulesservice.NoRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, err)
		assert.NotNil(t, setting)
	})
}

func TestFetchUsersPrivacyAttributes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOkidokiClient := NewMockOkidokiClient(ctrl)
	mockRulesService := NewMockRulesService(ctrl)
	mockClock := NewMockClock(ctrl)
	mockSettingsRepo := NewMockSettingsRepository(ctrl)
	processor := NewMinorSettingProcessor(mockOkidokiClient, mockRulesService, mockClock, mockSettingsRepo)

	// Test context
	ctx := context.Background()

	// Test setting
	testUserUrn := "soundcloud:users:123456"
	testUserCountryCode := "US"
	testUserDateOfBirth := "2000/01/01"
	testDate := time.Now()
	okidokiResponse := okidoki.OkidokiRawResponse{
		Self: okidoki.Self{
			URN: testUserUrn,
		},
		CountryCode: &testUserCountryCode,
		DateOfBirth: &testUserDateOfBirth,
	}

	t.Run("FetchUsersPrivacyAttributes returns user attributes", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponse,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, nil)

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, err)
		assert.NotNil(t, setting)
	})

	t.Run("FetchUsersPrivacyAttributes returns error if rulesService fails", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return([]okidoki.OkidokiRawResponse{
			okidokiResponse,
		}, nil)

		mockClock.EXPECT().Now().Return(testDate)
		mockRulesService.EXPECT().ApplyRules(gomock.Any(), gomock.Any()).Return(rulesservice.ApplyRestriction, errors.New("error"))

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		assert.NotNil(t, err)
	})

	t.Run("FetchUsersPrivacyAttributes returns error if okidoki fails", func(t *testing.T) {
		mockOkidokiClient.EXPECT().FetchUsers(ctx, gomock.Any()).Return(nil, errors.New("error"))

		setting, err := processor.FetchUsersPrivacyAttributes(ctx, []string{testUserUrn})

		assert.Nil(t, setting)
		assert.NotNil(t, err)
	})
}
