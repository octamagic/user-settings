package regionalrulesrepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewRegionalRulesRepository()

	t.Run("Returns age restriction rules for GB", func(t *testing.T) {
		countryCode := "GB"
		rules, err := repo.GetRules(countryCode)

		require.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 1)
		assert.Equal(t, countryCode, rules[0].ID)
		assert.Equal(t, "{\"age\": 18}", rules[0].Value)
		assert.Equal(t, RuleAgeRestriction, rules[0].Type)
	})

	t.Run("Returns age restriction rules for US", func(t *testing.T) {
		countryCode := "US"
		rules, err := repo.GetRules(countryCode)

		require.NoError(t, err)
		assert.NotNil(t, rules)
		assert.Len(t, rules, 1)
		assert.Equal(t, countryCode, rules[0].ID)
		assert.Equal(t, "{\"age\": 18}", rules[0].Value)
		assert.Equal(t, RuleAgeRestriction, rules[0].Type)
	})

	t.Run("Returns nothing for non present country", func(t *testing.T) {
		countryCode := "RU"
		rules, err := repo.GetRules(countryCode)

		require.NoError(t, err)
		assert.Nil(t, rules)
	})
}
