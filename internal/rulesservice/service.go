//go:generate mockgen -source=service.go -destination=mock_service.go -package=rulesservice
package rulesservice

import (
	"github.com/octamagic/user-settings/internal/regionalrulesrepo"
)

type RuleApplication int

const (
	ApplyRestriction RuleApplication = iota
	NoRestriction
)

type RegionalRulesRepository interface {
	GetRules(countryCode string) ([]regionalrulesrepo.Rule, error)
}

type RulesService interface {
	ApplyRules(age int, countryCode string) (RuleApplication, error)
}

type rulesService struct {
	rulesRepo RegionalRulesRepository
}

func NewRulesService(rulesRepository RegionalRulesRepository) RulesService {
	return &rulesService{
		rulesRepo: rulesRepository,
	}
}

func (s *rulesService) ApplyRules(age int, countryCode string) (RuleApplication, error) {
	rules, err := s.rulesRepo.GetRules(countryCode)
	if err != nil {
		return NoRestriction, err
	}

	if len(rules) == 0 {
		return NoRestriction, nil
	}

	for _, rule := range rules {
		if rule.Type == regionalrulesrepo.RuleAgeRestriction {
			ruleValue, err := rule.UnwrapAgeRestrictionValue()
			if err != nil {
				return NoRestriction, err
			}

			if rule.ID == countryCode && ruleValue.Age > age {
				return ApplyRestriction, nil
			}
		}
	}

	return NoRestriction, nil
}
