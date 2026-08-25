package main

// opsRulesSummaryCount returns the number of rules surfaced by the compliance
// summary view. Codes that collide in the rule index collapse into one entry.
func opsRulesSummaryCount() int {
	return len(opsRuleIndex(opsRules()))
}
