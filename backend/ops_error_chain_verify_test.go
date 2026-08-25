package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOpsDuplicateCreateHTTP409 verifies a duplicate create maps to 409.
func TestOpsDuplicateCreateHTTP409(t *testing.T) {
	h := newOpsHandler(newOpsService(opsSeed()))
	r := httptest.NewRecorder()
	body := `{"id":"batch-1001","subject":"流感疫苗批次冷链复核","owner":"auditor-lin","status":"queued","priority":"high","labels":{"site":"东城接种门诊"}}`
	h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/ops/records", bytes.NewBufferString(body)))
	if r.Code != http.StatusConflict {
		t.Fatalf("duplicate create got %d, want 409 (body=%s)", r.Code, r.Body.String())
	}
}

// TestOpsInvalidTransitionHTTP422 verifies an illegal state move maps to 422.
func TestOpsInvalidTransitionHTTP422(t *testing.T) {
	h := newOpsHandler(newOpsService(opsSeed()))
	r := httptest.NewRecorder()
	// batch-1003 is paused; paused -> queued is not a legal move.
	body := `{"expected":1,"target":"queued"}`
	h.ServeHTTP(r, httptest.NewRequest(http.MethodPost, "/api/ops/records/batch-1003/transition", bytes.NewBufferString(body)))
	if r.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid transition got %d, want 422 (body=%s)", r.Code, r.Body.String())
	}
}

// TestOpsDuplicateCreateKeepsChain verifies the create path error stays
// chainable to ErrOpsConflict.
func TestOpsDuplicateCreateKeepsChain(t *testing.T) {
	svc := newOpsService(opsSeed())
	ctx := context.Background()
	rec := OpsRecord{ID: "batch-1001", Subject: "流感疫苗批次冷链复核", Owner: "auditor-lin", Status: OpsStatusQueued, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "东城接种门诊"}}
	if _, err := svc.Create(ctx, rec); err == nil {
		t.Fatal("duplicate create accepted")
	} else if !errors.Is(err, ErrOpsConflict) {
		t.Fatalf("duplicate create error not chainable to ErrOpsConflict: %v", err)
	}
}

// TestOpsMoveErrorKeepsChain verifies the state machine error stays chainable
// to ErrOpsTransition.
func TestOpsMoveErrorKeepsChain(t *testing.T) {
	svc := newOpsService(opsSeed())
	err := svc.state.Move(OpsStatusClosed, OpsStatusActive, "probe")
	if err == nil {
		t.Fatal("illegal move accepted")
	}
	if !errors.Is(err, ErrOpsTransition) {
		t.Fatalf("move error not chainable to ErrOpsTransition: %v", err)
	}
}

// TestOpsWrappedErrorKeepsChain verifies wrapOps preserves the sentinel chain.
func TestOpsWrappedErrorKeepsChain(t *testing.T) {
	err := wrapOps("transition", "store.update", ErrOpsConflict)
	if !errors.Is(err, ErrOpsConflict) {
		t.Fatalf("wrapOps lost ErrOpsConflict: %v", err)
	}
	var typed *OpsError
	if !errors.As(err, &typed) || typed.Code != "transition" {
		t.Fatalf("wrapOps result not an unwrap-able OpsError: %v", err)
	}
}

// TestOpsHTTPStatusMapping verifies the status mapping for each error class.
func TestOpsHTTPStatusMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{wrapOps("get", "store.get", ErrOpsNotFound), http.StatusNotFound},
		{wrapOps("create", "store.put", ErrOpsConflict), http.StatusConflict},
		{wrapOps("validate", "status.check", ErrOpsInvalid), http.StatusBadRequest},
		{wrapOps("move", "state.move", ErrOpsTransition), http.StatusUnprocessableEntity},
		{wrapOps("create", "policy.check", ErrOpsPolicy), http.StatusBadRequest},
	}
	for _, tc := range cases {
		if got := opsHTTPStatus(tc.err); got != tc.want {
			t.Fatalf("opsHTTPStatus(%v)=%d want %d", tc.err, got, tc.want)
		}
	}
}
