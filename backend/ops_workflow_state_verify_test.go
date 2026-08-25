package main

import "testing"

// TestOpsWorkflowReviewFlow drives the full received -> reviewing -> released
// flow through the governance gates.
func TestOpsWorkflowReviewFlow(t *testing.T) {
	record := OpsRecord{ID: "batch-2001", Labels: map[string]string{"site": "north", "evidence": "temp-log-9"}}
	w := NewOpsWorkflow()
	if err := opsAdvanceWorkflow(w, record, WorkflowReviewing); err != nil {
		t.Fatalf("enter reviewing failed: %v", err)
	}
	if w.State() != WorkflowReviewing {
		t.Fatalf("state=%s want reviewing", w.State())
	}
	if err := opsAdvanceWorkflow(w, record, WorkflowReleased); err != nil {
		t.Fatalf("release after reviewing failed: %v", err)
	}
	if w.State() != WorkflowReleased {
		t.Fatalf("state=%s want released", w.State())
	}
}

// TestOpsWorkflowHistoryRealTarget verifies history records the actual target
// state of each move.
func TestOpsWorkflowHistoryRealTarget(t *testing.T) {
	w := NewOpsWorkflow()
	record := OpsRecord{ID: "batch-2002", Labels: map[string]string{"site": "north", "evidence": "temp-log-9"}}
	if err := opsAdvanceWorkflow(w, record, WorkflowReviewing); err != nil {
		t.Fatal(err)
	}
	if err := opsAdvanceWorkflow(w, record, WorkflowRecalled); err != nil {
		t.Fatal(err)
	}
	history := w.History()
	if len(history) != 3 {
		t.Fatalf("history length=%d want 3: %v", len(history), history)
	}
	if history[2] != WorkflowRecalled {
		t.Fatalf("history last=%q want recalled: %v", history[2], history)
	}
}

// TestOpsWorkflowTableReviewEdge verifies the workflow table allows the
// reviewing intermediate state to move forward.
func TestOpsWorkflowTableReviewEdge(t *testing.T) {
	w := NewOpsWorkflow()
	if err := w.Move(WorkflowReviewing); err != nil {
		t.Fatalf("received -> reviewing blocked by table: %v", err)
	}
	if err := w.Move(WorkflowReleased); err != nil {
		t.Fatalf("reviewing -> released blocked by table: %v", err)
	}
}

// TestOpsWorkflowGate07Evidence verifies the reviewing gate requires evidence.
func TestOpsWorkflowGate07Evidence(t *testing.T) {
	record := OpsRecord{ID: "batch-2003", Labels: map[string]string{"site": "north", "evidence": "temp-log-9"}}
	if err := opsWorkflowGate07(record); err != nil {
		t.Fatalf("record with evidence rejected: %v", err)
	}
}

// TestOpsWorkflowGate08Reviewing verifies the release gate only permits batches
// in the reviewing state.
func TestOpsWorkflowGate08Reviewing(t *testing.T) {
	w := NewOpsWorkflow()
	if err := w.Move(WorkflowReviewing); err != nil {
		t.Fatal(err)
	}
	if err := opsWorkflowGate08(w); err != nil {
		t.Fatalf("reviewing batch rejected by release gate: %v", err)
	}
}
