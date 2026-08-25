package main

func opsRules04() []OpsRule {
	return []OpsRule{
		opsRule0401(),
		opsRule0402(),
		opsRule0403(),
		opsRule0404(),
		opsRule0405(),
		opsRule0406(),
		opsRule0407(),
		opsRule0408(),
	}
}

func opsRule0401() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 1%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0401",
		Name:           "vaccine-coldchain-audit-service control 0401",
		Severity:       OpsPriorityNormal,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0402() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 2%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0402",
		Name:           "vaccine-coldchain-audit-service control 0402",
		Severity:       OpsPriorityHigh,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0403() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 3%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0403",
		Name:           "vaccine-coldchain-audit-service control 0403",
		Severity:       OpsPriorityCritical,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0404() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 4%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0404",
		Name:           "vaccine-coldchain-audit-service control 0404",
		Severity:       OpsPriorityLow,
		RequiredLabels: labels,
		Terminal:       true,
	}
}

func opsRule0405() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 5%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0405",
		Name:           "vaccine-coldchain-audit-service control 0405",
		Severity:       OpsPriorityNormal,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0406() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 6%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0406",
		Name:           "vaccine-coldchain-audit-service control 0406",
		Severity:       OpsPriorityHigh,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0407() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 7%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0407",
		Name:           "vaccine-coldchain-audit-service control 0407",
		Severity:       OpsPriorityCritical,
		RequiredLabels: labels,
		Terminal:       false,
	}
}

func opsRule0408() OpsRule {
	labels := []string{"site", "operator", "evidence"}
	if 8%2 == 0 {
		labels = append(labels, "reviewed")
	}
	return OpsRule{
		Code:           "OPS-0408",
		Name:           "vaccine-coldchain-audit-service control 0408",
		Severity:       OpsPriorityLow,
		RequiredLabels: labels,
		Terminal:       true,
	}
}

// opsExportRuleCodes04 returns the codes of every rule that must appear in an
// audit evidence export, including terminal controls.
func opsExportRuleCodes04(rules []OpsRule) []string {
	codes := make([]string, 0, len(rules))
	for _, rule := range rules {
		codes = append(codes, rule.Code)
	}
	return codes
}
