package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestOpsSnapshotConcurrentLabelsRace drives a reader (Get of Labels) and a
// writer (Update that mutates Labels in place) against the same record. The
// store must never hand out references sharing the internal Labels map.
func TestOpsSnapshotConcurrentLabelsRace(t *testing.T) {
	svc := newOpsService(opsSeed())
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 3000; i++ {
			r, err := svc.store.Get(context.Background(), "batch-1002")
			if err == nil {
				_ = r.Labels["site"]
			}
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 3000; i++ {
			r, err := svc.store.Get(context.Background(), "batch-1002")
			if err == nil {
				r.Status = OpsStatusActive
				_ = svc.store.Update(context.Background(), r, 0)
			}
		}
	}()
	close(start)
	wg.Wait()
}

// TestOpsGetReturnsDeepCopy verifies Get never shares the internal Labels map.
func TestOpsGetReturnsDeepCopy(t *testing.T) {
	svc := newOpsService(opsSeed())
	r1, err := svc.store.Get(context.Background(), "batch-1001")
	if err != nil {
		t.Fatal(err)
	}
	r1.Labels["mutated"] = "yes"
	r2, err := svc.store.Get(context.Background(), "batch-1001")
	if err != nil {
		t.Fatal(err)
	}
	if r2.Labels["mutated"] != "" {
		t.Fatalf("Get returned a shared Labels map: %v", r2.Labels)
	}
}

// TestOpsListReturnsDeepCopy verifies List items do not share the store map.
func TestOpsListReturnsDeepCopy(t *testing.T) {
	svc := newOpsService(opsSeed())
	items, err := svc.store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("empty list")
	}
	items[0].Labels["hacked"] = "x"
	again, err := svc.store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range again {
		if item.Labels["hacked"] != "" {
			t.Fatalf("List leaked a mutated label for %s: %v", item.ID, item.Labels)
		}
	}
}

// TestOpsUpdateDoesNotAddOperator verifies a status update never writes new
// labels into the stored record.
func TestOpsUpdateDoesNotAddOperator(t *testing.T) {
	svc := newOpsService(nil)
	ctx := context.Background()
	rec := OpsRecord{ID: "batch-9001", Subject: "新批次", Owner: "auditor-x", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "north", "evidence": "log-1"}}
	if _, err := svc.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	got, err := svc.store.Get(ctx, "batch-9001")
	if err != nil {
		t.Fatal(err)
	}
	got.Status = OpsStatusActive
	if err := svc.store.Update(ctx, got, 0); err != nil {
		t.Fatal(err)
	}
	check, err := svc.store.Get(ctx, "batch-9001")
	if err != nil {
		t.Fatal(err)
	}
	if check.Labels["operator"] != "" {
		t.Fatalf("Update wrote an operator label into the stored record: %v", check.Labels)
	}
}

// TestOpsHousekeepingStopsOnCancel verifies the housekeeping goroutine exits
// once its context is canceled (no goroutine leak).
func TestOpsHousekeepingStopsOnCancel(t *testing.T) {
	svc := newOpsService(opsSeed())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		opsHousekeeping(ctx, svc.store, time.Millisecond)
		close(done)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("housekeeping goroutine did not stop after cancel (leak)")
	}
}
