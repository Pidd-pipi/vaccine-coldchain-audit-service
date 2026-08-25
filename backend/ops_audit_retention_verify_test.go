package main

import (
	"testing"
	"time"
)

// TestAuditPruneRemovesOldKeepsRecent verifies Prune drops old events and keeps
// recent ones.
func TestAuditPruneRemovesOldKeepsRecent(t *testing.T) {
	a := newOpsAudit()
	a.mu.Lock()
	a.events = []OpsEvent{
		{ID: "evt-old", RecordID: "batch-1", At: "2026-01-01T00:00:00Z"},
		{ID: "evt-recent", RecordID: "batch-1", At: time.Now().UTC().Format(time.RFC3339Nano)},
	}
	a.mu.Unlock()
	removed, err := a.Prune(time.Now().Add(-24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed=%d want 1", removed)
	}
	for _, ev := range a.events {
		if ev.ID == "evt-recent" {
			return
		}
	}
	t.Fatalf("recent event was pruned: %+v", a.events)
}

// TestAuditPruneErrorPreserved verifies an unparseable event timestamp surfaces
// as an error instead of being swallowed.
func TestAuditPruneErrorPreserved(t *testing.T) {
	a := newOpsAudit()
	a.mu.Lock()
	a.events = []OpsEvent{{ID: "evt-bad", RecordID: "batch-1", At: "not-a-timestamp"}}
	a.mu.Unlock()
	if _, err := a.Prune(time.Now()); err == nil {
		t.Fatal("expected prune error for unparseable timestamp")
	}
}

// TestAuditBoundedAfterManyAdds verifies the event log stays bounded by the cap.
func TestAuditBoundedAfterManyAdds(t *testing.T) {
	a := newOpsAudit()
	a.SetMaxEvents(10)
	for i := 0; i < 100; i++ {
		a.Add("batch-1", "status_changed", "auditor-lin")
	}
	if n := a.Count(); n > 10 {
		t.Fatalf("audit grew to %d events (cap 10)", n)
	}
}

// TestAuditPrunerStopsOnShutdown verifies the pruner goroutine exits when the
// stop signal is closed.
func TestAuditPrunerStopsOnShutdown(t *testing.T) {
	a := newOpsAudit()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		runAuditPruner(a, time.Millisecond, stop)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	close(stop)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("audit pruner did not stop after signal (goroutine leak)")
	}
}

// TestAuditPrunerTrimsOldEvents verifies the background pruner actually trims
// events older than the retention window.
func TestAuditPrunerTrimsOldEvents(t *testing.T) {
	a := newOpsAudit()
	a.mu.Lock()
	a.events = []OpsEvent{
		{ID: "evt-old-1", RecordID: "batch-1", At: time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339Nano)},
		{ID: "evt-old-2", RecordID: "batch-2", At: time.Now().Add(-96 * time.Hour).UTC().Format(time.RFC3339Nano)},
	}
	a.mu.Unlock()
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		startAuditPruner(a, 5*time.Millisecond, stop)
		close(done)
	}()
	time.Sleep(60 * time.Millisecond)
	close(stop)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("audit pruner did not stop")
	}
	if n := a.Count(); n != 0 {
		t.Fatalf("old events were not trimmed: %d remaining", n)
	}
}
