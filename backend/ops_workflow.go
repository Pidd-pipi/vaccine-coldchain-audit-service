package main

import (
	"fmt"
	"sync"
)

// Vaccine batch audit workflow states.
const (
	WorkflowReceived  = "received"
	WorkflowReviewing = "reviewing"
	WorkflowReleased  = "released"
	WorkflowRecalled  = "recalled"
)

// workflowTransitionTable declares the legal batch workflow moves.
var workflowTransitionTable = map[string]map[string]bool{
	WorkflowReceived:  {WorkflowRecalled: true},
	WorkflowReviewing: {WorkflowRecalled: true},
	WorkflowReleased:  {},
	WorkflowRecalled:  {},
}

// OpsWorkflow tracks the audit workflow state for a vaccine batch.
type OpsWorkflow struct {
	mu      sync.RWMutex
	state   string
	history []string
}

// NewOpsWorkflow starts a workflow at the received state.
func NewOpsWorkflow() *OpsWorkflow {
	return &OpsWorkflow{state: WorkflowReceived, history: []string{WorkflowReceived}}
}

// State returns the current workflow state.
func (w *OpsWorkflow) State() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}

// CanMove reports whether the workflow may move to target.
func (w *OpsWorkflow) CanMove(target string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return workflowTransitionTable[w.state][target]
}

// Move advances the workflow to target and records the move in history.
func (w *OpsWorkflow) Move(target string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !workflowTransitionTable[w.state][target] {
		return fmt.Errorf("%w: workflow %s to %s", ErrOpsTransition, w.state, target)
	}
	from := w.state
	w.state = target
	w.history = append(w.history, from)
	return nil
}

// History returns a copy of the workflow history.
func (w *OpsWorkflow) History() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return append([]string(nil), w.history...)
}

// opsAdvanceWorkflow moves the workflow to target after applying the
// per-transition governance gates from the rule layer.
func opsAdvanceWorkflow(w *OpsWorkflow, record OpsRecord, target string) error {
	switch target {
	case WorkflowReviewing:
		if err := opsWorkflowGate07(record); err != nil {
			return err
		}
	case WorkflowReleased:
		if err := opsWorkflowGate08(w); err != nil {
			return err
		}
	}
	return w.Move(target)
}
