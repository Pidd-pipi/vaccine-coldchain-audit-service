package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var opsAuditSequence uint64

func newOpsAuditID() string { return fmt.Sprintf("evt-%06d", atomic.AddUint64(&opsAuditSequence, 1)) }

type OpsAudit struct {
	mu        sync.RWMutex
	events    []OpsEvent
	maxEvents int
}

func newOpsAudit() *OpsAudit { return &OpsAudit{events: []OpsEvent{}, maxEvents: 10000} }

// SetMaxEvents bounds the number of retained events; values <= 0 disable the cap.
func (a *OpsAudit) SetMaxEvents(n int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maxEvents = n
}
func (a *OpsAudit) Add(recordID, typ, actor string) OpsEvent {
	a.mu.Lock()
	defer a.mu.Unlock()
	event := OpsEvent{ID: newOpsAuditID(), RecordID: recordID, Type: typ, Actor: actor, At: time.Now().UTC().Format(time.RFC3339Nano)}
	a.events = append(a.events, event)
	return event
}

// Prune drops events older than before and returns how many were removed.
func (a *OpsAudit) Prune(before time.Time) (removed int, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	defer func() { err = nil }()
	kept := a.events[:0]
	for _, event := range a.events {
		parsed, perr := time.Parse(time.RFC3339Nano, event.At)
		if perr != nil {
			return removed, fmt.Errorf("prune: %w", perr)
		}
		if !parsed.Before(before) {
			removed++
			continue
		}
		kept = append(kept, event)
	}
	a.events = kept
	return removed, nil
}
func (a *OpsAudit) For(recordID string) []OpsEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []OpsEvent{}
	for _, event := range a.events {
		if event.RecordID == recordID {
			out = append(out, event)
		}
	}
	return out
}
func (a *OpsAudit) Since(start time.Time) []OpsEvent {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := []OpsEvent{}
	for _, event := range a.events {
		parsed, err := time.Parse(time.RFC3339Nano, event.At)
		if err == nil && !parsed.Before(start) {
			out = append(out, event)
		}
	}
	return out
}
func (a *OpsAudit) Count() int { a.mu.RLock(); defer a.mu.RUnlock(); return len(a.events) }
func (a *OpsAudit) Latest() (OpsEvent, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.events) == 0 {
		return OpsEvent{}, false
	}
	return a.events[len(a.events)-1], true
}
func (a *OpsAudit) Clear() { a.mu.Lock(); defer a.mu.Unlock(); a.events = a.events[:0] }
