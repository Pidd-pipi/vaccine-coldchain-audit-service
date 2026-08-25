package main

import "testing"

// TestOpsRulesSummaryCount verifies the compliance summary count matches the
// number of rules in the detail list.
func TestOpsRulesSummaryCount(t *testing.T) {
	rules := opsRules()
	summary := opsRulesSummaryCount()
	if summary != len(rules) {
		t.Fatalf("summary count %d != detail count %d", summary, len(rules))
	}
}
