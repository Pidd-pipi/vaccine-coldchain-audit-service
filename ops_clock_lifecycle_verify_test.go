package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// TestOpsDelayHonorsCancel verifies a canceled context interrupts opsDelay.
func TestOpsDelayHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := opsDelay(ctx, 5*time.Second); err == nil {
		t.Fatal("opsDelay with canceled context returned nil error")
	}
	if time.Since(start) > time.Second {
		t.Fatalf("opsDelay ignored cancellation, waited %s", time.Since(start))
	}
}

// TestOpsContextNilParentNoPanic verifies opsContext tolerates a nil parent.
func TestOpsContextNilParentNoPanic(t *testing.T) {
	ctx, cancel := opsContext(nil, 3*time.Second)
	defer cancel()
	if ctx == nil {
		t.Fatal("opsContext(nil) returned nil context")
	}
}

// TestOpsShutdownContextHasDeadline verifies the shutdown context is bounded.
func TestOpsShutdownContextHasDeadline(t *testing.T) {
	ctx, cancel := opsShutdownContext()
	defer cancel()
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("shutdown context has no deadline (shutdown can hang forever)")
	}
}

// TestServerTimeoutsConfigured verifies the HTTP server carries read/write
// timeouts so slow requests cannot pin connections forever.
func TestServerTimeoutsConfigured(t *testing.T) {
	srv := newEnterpriseServer("127.0.0.1:0", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if srv.ReadTimeout <= 0 {
		t.Fatalf("server ReadTimeout not configured: %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout <= 0 {
		t.Fatalf("server WriteTimeout not configured: %v", srv.WriteTimeout)
	}
	if srv.ReadHeaderTimeout <= 0 {
		t.Fatalf("server ReadHeaderTimeout not configured: %v", srv.ReadHeaderTimeout)
	}
}
