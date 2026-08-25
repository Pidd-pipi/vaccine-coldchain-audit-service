package main

import "testing"

// TestOpsEvaluateRecordsNoNilMapWrite verifies rule evaluation never writes
// into a nil violations map when a record is missing required labels.
func TestOpsEvaluateRecordsNoNilMapWrite(t *testing.T) {
	index := opsRuleIndex(opsRules01())
	if len(index) == 0 {
		t.Fatal("expected default rules to be indexed")
	}
	record := OpsRecord{ID: "rec-nolabels", Subject: "缺失标签批次", Owner: "auditor-x"}
	violations := opsEvaluateRecords(index, record)
	if len(violations) == 0 {
		t.Fatal("expected violations for a record without required labels")
	}
}

// TestOpsRuleIndexEmptyReturnsUsableMap verifies an empty rule set yields a
// usable (non-nil) index.
func TestOpsRuleIndexEmptyReturnsUsableMap(t *testing.T) {
	index := opsRuleIndex(nil)
	if index == nil {
		t.Fatal("opsRuleIndex(nil) returned a nil map")
	}
	if len(index) != 0 {
		t.Fatalf("expected empty index, got %d entries", len(index))
	}
}

// TestOpsLookupProviderTypedNilSafe verifies a provider built from a nil index
// is safe to call.
func TestOpsLookupProviderTypedNilSafe(t *testing.T) {
	provider := opsRuleProviderOf(nil)
	if provider == nil {
		t.Fatal("provider should be non-nil")
	}
	if _, ok := provider.Lookup("OPS-9999"); ok {
		t.Fatal("unexpected rule found in empty provider")
	}
}

// TestOpsRuleGateEmptyConfigNoPanic verifies the gate tolerates a nil provider
// and reports the rule as missing.
func TestOpsRuleGateEmptyConfigNoPanic(t *testing.T) {
	record := OpsRecord{ID: "rec-1", Labels: map[string]string{"site": "north"}}
	err := opsRuleGate01(opsRuleProviderOf(nil), "OPS-0101", record)
	if err == nil {
		t.Fatal("expected a missing-rule error")
	}
}

// TestOpsDefaultRuleSetNonEmpty verifies an empty configuration falls back to
// the default control set.
func TestOpsDefaultRuleSetNonEmpty(t *testing.T) {
	set := opsRuleSetFor(nil)
	if len(set) == 0 {
		t.Fatal("empty configuration produced no default rule set")
	}
}
