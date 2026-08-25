package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type countingCloser struct {
	open   int
	closed int
	max    int
}

func (c *countingCloser) writer() func(string) (io.WriteCloser, error) {
	return func(string) (io.WriteCloser, error) {
		c.open++
		if c.open-c.closed > c.max {
			c.max = c.open - c.closed
		}
		return &countingWriteCloser{c: c}, nil
	}
}

type countingWriteCloser struct {
	c *countingCloser
}

func (w *countingWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (w *countingWriteCloser) Close() error {
	w.c.closed++
	return nil
}

// TestOpsExportClosesFilesAsItGoes verifies evidence files are closed as the
// export progresses instead of accumulating open handles.
func TestOpsExportClosesFilesAsItGoes(t *testing.T) {
	events := make([]OpsEvent, 0, 200)
	for i := 0; i < 200; i++ {
		events = append(events, OpsEvent{ID: "evt-export-test", RecordID: "batch-1001", Type: "status_changed", Actor: "auditor-lin"})
	}
	c := &countingCloser{}
	written, err := opsExportAudit(t.TempDir(), events, nil, c.writer())
	if err != nil {
		t.Fatal(err)
	}
	if written != len(events) {
		t.Fatalf("written=%d want %d", written, len(events))
	}
	if c.max > 2 {
		t.Fatalf("export held %d files open at once (defer accumulation leak)", c.max)
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) { return 0, errors.New("disk write failed") }
func (failingWriter) Close() error                { return errors.New("close failed") }

// TestOpsExportErrorPreservesOriginal verifies a write failure is returned as
// the export error instead of being overwritten by a close error.
func TestOpsExportErrorPreservesOriginal(t *testing.T) {
	events := []OpsEvent{{ID: "evt-1", RecordID: "batch-1001", Type: "status_changed", Actor: "auditor-lin"}}
	open := func(string) (io.WriteCloser, error) { return failingWriter{}, nil }
	_, err := opsExportAudit(t.TempDir(), events, nil, open)
	if err == nil {
		t.Fatal("expected an export error")
	}
	if !strings.Contains(err.Error(), "disk write failed") {
		t.Fatalf("original write error lost, got: %v", err)
	}
	if strings.Contains(err.Error(), "close failed") {
		t.Fatalf("close error overwrote the original write error: %v", err)
	}
}

// TestOpsExportCreatesDestDir verifies the export creates the destination
// directory when it does not exist.
func TestOpsExportCreatesDestDir(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "nested", "evidence")
	events := []OpsEvent{{ID: "evt-1", RecordID: "batch-1001", Type: "status_changed", Actor: "auditor-lin"}}
	if _, err := opsExportAudit(dest, events, nil, nil); err != nil {
		t.Fatalf("export into missing dir failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "evt-1.jsonl")); err != nil {
		t.Fatalf("evidence file not created: %v", err)
	}
}
