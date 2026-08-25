package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// opsExportColumns returns the CSV columns produced for one rule.
func opsExportColumns(rule OpsRule) []string {
	columns := []string{"code", "name", "severity"}
	if rule.Terminal {
		columns = append(columns, "terminal")
	}
	for _, label := range rule.RequiredLabels {
		columns = append(columns, "label:"+label)
	}
	return columns
}

// opsExportAudit writes one evidence file per event under dest and returns the
// number of files written. The open function is injectable so callers can
// control how files are created (e.g. in tests).
func opsExportAudit(dest string, events []OpsEvent, rules []OpsRule, open func(string) (io.WriteCloser, error)) (written int, err error) {
	if open == nil {
		open = func(path string) (io.WriteCloser, error) {
			return newEvidenceFile(path)
		}
	}
	written = 0
	for _, event := range events {
		path := filepath.Join(dest, sanitizeEventName(event.ID)+".jsonl")
		f, oerr := open(path)
		if oerr != nil {
			return written, oerr
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				err = cerr
			}
		}()
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n", event.ID, event.RecordID, event.Type, event.Actor)
		if _, werr := io.WriteString(f, line); werr != nil {
			return written, werr
		}
		written++
	}
	return written, nil
}

// newEvidenceFile opens a file for appending evidence lines.
func newEvidenceFile(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
}

// sanitizeEventName keeps only safe filename characters.
func sanitizeEventName(name string) string {
	return strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' {
			return '_'
		}
		return r
	}, name)
}
