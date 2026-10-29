//go:generate mockgen -source=repository.go -destination=mock_repository.go -package=regionalrulesrepo
package regionalrulesrepo

type RegionalRulesRepository interface {
	GetRules(countryCode string) ([]Rule, error)
}

// RegionalRulesRepository in-memory store for the custom rules
type regionalRulesRepository struct {
}

func NewRegionalRulesRepository() RegionalRulesRepository {
	return &regionalRulesRepository{}
}

func (rulesRepo *regionalRulesRepository) GetRules(countryCode string) ([]Rule, error) {
	rules := []Rule{{
		ID:    "GB",
		Type:  RuleAgeRestriction,
		Value: "{\"age\": 18}",
	}, {
		ID:    "US",
		Type:  RuleAgeRestriction,
		Value: "{\"age\": 18}",
	}}

	var rulesToApply []Rule
	for _, rule := range rules {
		if rule.ID == countryCode {
			rulesToApply = append(rulesToApply, rule)
		}
	}
	return rulesToApply, nil
}
