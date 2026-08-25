package main

import (
	"context"
	"sync"
	"testing"
)

// TestOpsSearchPageIsolationAcrossCalls verifies a page returned by an earlier
// search is not overwritten by a later search (no shared scratch backing).
func TestOpsSearchPageIsolationAcrossCalls(t *testing.T) {
	svc := newOpsService(opsSeed())
	ctx := context.Background()

	p1, err := svc.Search(ctx, OpsQuery{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]OpsRecord, len(p1.Items))
	for i, item := range p1.Items {
		want[i] = item.Clone()
	}
	if len(want) < 2 {
		t.Fatalf("expected at least 2 items on page 1, got %d", len(want))
	}

	if _, err := svc.Search(ctx, OpsQuery{Subject: "流感", Page: 1, PageSize: 2}); err != nil {
		t.Fatal(err)
	}
	// The first page's items must still hold the original content even though a
	// later search ran.
	if len(p1.Items) != len(want) {
		t.Fatalf("page length changed after another search: want %d got %d", len(want), len(p1.Items))
	}
	for i := range want {
		if want[i].ID != p1.Items[i].ID || want[i].Subject != p1.Items[i].Subject {
			t.Fatalf("page item %d changed after another search: want %+v got %+v", i, want[i], p1.Items[i])
		}
	}
}

// TestOpsSearchConcurrentIsolation verifies concurrent searches do not share a
// mutable scratch slice (data race).
func TestOpsSearchConcurrentIsolation(t *testing.T) {
	svc := newOpsService(opsSeed())
	ctx := context.Background()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			for j := 0; j < 100; j++ {
				q := OpsQuery{Page: 1, PageSize: 2}
				if n%2 == 0 {
					q.Subject = "批次"
				}
				if _, err := svc.Search(ctx, q); err != nil {
					t.Error(err)
					return
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
}

// TestOpsClonePageDeepCopy verifies opsClonePage detaches Items and their
// Labels maps from the source page.
func TestOpsClonePageDeepCopy(t *testing.T) {
	p := OpsPage{Items: []OpsRecord{
		{ID: "a", Subject: "original", Labels: map[string]string{"site": "east"}},
	}}
	clone := opsClonePage(p)
	clone.Items[0].Subject = "changed"
	clone.Items[0].Labels["site"] = "west"
	if p.Items[0].Subject != "original" {
		t.Fatalf("opsClonePage shares item struct: %+v", p.Items[0])
	}
	if p.Items[0].Labels["site"] != "east" {
		t.Fatalf("opsClonePage shares Labels map: %+v", p.Items[0].Labels)
	}
}

// TestOpsCloneDeepCopiesLabels verifies Clone detaches the Labels map.
func TestOpsCloneDeepCopiesLabels(t *testing.T) {
	r := OpsRecord{ID: "b", Labels: map[string]string{"site": "north"}}
	c := r.Clone()
	c.Labels["site"] = "south"
	if r.Labels["site"] != "north" {
		t.Fatalf("Clone shares Labels map: %v", r.Labels)
	}
}

// TestOpsSearchPaginationBounds verifies pagination boundaries remain stable.
func TestOpsSearchPaginationBounds(t *testing.T) {
	svc := newOpsService(opsSeed())
	ctx := context.Background()
	p1, err := svc.Search(ctx, OpsQuery{Page: 1, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	p2, err := svc.Search(ctx, OpsQuery{Page: 2, PageSize: 2})
	if err != nil {
		t.Fatal(err)
	}
	if p1.Total != 3 || p2.Total != 3 {
		t.Fatalf("total mismatch: p1=%d p2=%d", p1.Total, p2.Total)
	}
	if !p1.HasNext {
		t.Fatal("page 1 should have next")
	}
	if p2.HasNext {
		t.Fatal("page 2 should not have next")
	}
	seen := map[string]bool{}
	for _, item := range append(append([]OpsRecord{}, p1.Items...), p2.Items...) {
		if seen[item.ID] {
			t.Fatalf("duplicate item across pages: %s", item.ID)
		}
		seen[item.ID] = true
	}
	if len(seen) != 3 {
		t.Fatalf("expected 3 distinct items across pages, got %d", len(seen))
	}
}
