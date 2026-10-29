package prefsrepo

import (
	"time"
)

const (
	// Key attribute names used by DynamoDB.
	pkName = "pk"
	skName = "sk"
)

type SettingType int

const (
	SettingTypeString SettingType = iota
	SettingTypeBoolean
	SettingTypeNumber
)

var settingTypeName = map[SettingType]string{
	SettingTypeString:  "s",
	SettingTypeBoolean: "b",
	SettingTypeNumber:  "n",
}

func (ss SettingType) String() string {
	return settingTypeName[ss]
}

// Setting represents a user setting stored in DynamoDB.
type Setting struct {
	UserUrn     string    `dynamodbav:"pk"`
	Name        string    `dynamodbav:"sk"`
	SettingType string    `dynamodbav:"type"`
	Value       string    `dynamodbav:"value"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt,unixtime"`
}

// NewSetting creates a new Setting with the current time and proper entity.
func NewSetting(userUrn string, name string, settingType SettingType, value string) Setting {
	return Setting{
		UserUrn:     userUrn,
		Name:        name,
		SettingType: settingType.String(),
		Value:       value,
		UpdatedAt:   time.Now(),
	}
}
