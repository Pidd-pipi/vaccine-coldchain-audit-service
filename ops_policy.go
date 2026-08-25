package main

import "fmt"

// opsRuleIndex indexes rules by their code.
func opsRuleIndex(rules []OpsRule) map[string]OpsRule {
	if len(rules) == 0 {
		return nil
	}
	index := make(map[string]OpsRule, len(rules))
	for _, rule := range rules {
		index[rule.Code] = rule
	}
	return index
}

// opsLookupRule returns the rule registered under code, or nil when absent.
func opsLookupRule(index map[string]OpsRule, code string) *OpsRule {
	rule, ok := index[code]
	if !ok {
		return nil
	}
	return &rule
}

// opsEvaluateRule checks a record against one rule and records the first
// missing required label into the violations map.
func opsEvaluateRule(rule OpsRule, record OpsRecord, violations map[string]string) {
	for _, label := range rule.RequiredLabels {
		if record.LabelValue(label) == "" {
			violations[rule.Code] = fmt.Sprintf("missing label %s", label)
			return
		}
	}
}

// opsEvaluateRecords runs every rule in the index against a record.
func opsEvaluateRecords(index map[string]OpsRule, record OpsRecord) map[string]string {
	var violations map[string]string
	for _, rule := range index {
		opsEvaluateRule(rule, record, violations)
	}
	return violations
}

// OpsRuleProvider looks rules up by code behind an interface so callers can
// swap the backing index.
type OpsRuleProvider interface {
	Lookup(code string) (*OpsRule, bool)
}

type opsRuleMap struct {
	index map[string]OpsRule
}

func (m opsRuleMap) Lookup(code string) (*OpsRule, bool) {
	rule, ok := m.index[code]
	if !ok {
		return nil, false
	}
	return &rule, true
}

// opsRuleProviderOf adapts a rule index to the provider interface.
func opsRuleProviderOf(index map[string]OpsRule) OpsRuleProvider {
	if index == nil {
		return (*opsRuleMap)(nil)
	}
	return opsRuleMap{index: index}
}
