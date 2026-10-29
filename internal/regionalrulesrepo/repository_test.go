package regionalrulesrepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewRegionalRulesRepository()

	t.Run("Returns age restriction rules for GB", func(t *testing.T) {
		countryCode := "GB"
		rules, err := repo.GetRules(countryCode)

		assert.Nil(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 1)
		assert.Equal(t, rules[0].ID, countryCode)
		assert.Equal(t, rules[0].Value, "{\"age\": 18}")
		assert.True(t, rules[0].Type == RuleAgeRestriction)
	})

	t.Run("Returns age restriction rules for US", func(t *testing.T) {
		countryCode := "US"
		rules, err := repo.GetRules(countryCode)

		assert.Nil(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 1)
		assert.Equal(t, rules[0].ID, countryCode)
		assert.Equal(t, rules[0].Value, "{\"age\": 18}")
		assert.True(t, rules[0].Type == RuleAgeRestriction)
	})

	t.Run("Returns nothing for non present country", func(t *testing.T) {
		countryCode := "RU"
		rules, err := repo.GetRules(countryCode)

		assert.Nil(t, err)
		assert.Nil(t, rules)
	})
}
