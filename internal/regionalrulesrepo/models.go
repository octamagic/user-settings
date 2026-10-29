package regionalrulesrepo

import (
	"encoding/json"
	"fmt"
)

type RuleType int

const (
	RuleAgeRestriction RuleType = iota
)

type Rule struct {
	ID    string
	Type  RuleType
	Value string
}

type RuleAgeRestrictionValue struct {
	Age int `json:"age"`
}

func (r *Rule) UnwrapAgeRestrictionValue() (*RuleAgeRestrictionValue, error) {
	var value RuleAgeRestrictionValue
	if err := json.Unmarshal([]byte(r.Value), &value); err != nil {
		return nil, fmt.Errorf("failed to parse rule value: %w", err)
	}
	return &value, nil
}
