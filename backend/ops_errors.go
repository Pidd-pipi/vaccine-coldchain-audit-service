package main

import (
	"errors"
	"fmt"
)

var (
	ErrOpsNotFound   = errors.New("operations record not found")
	ErrOpsConflict   = errors.New("operations revision conflict")
	ErrOpsInvalid    = errors.New("operations request is invalid")
	ErrOpsTransition = errors.New("operations status transition is not allowed")
	ErrOpsPolicy     = errors.New("operations policy rejected the request")
)

type OpsError struct {
	Code      string
	Operation string
	Cause     error
}

func (e *OpsError) Error() string {
	if e.Cause == nil {
		return e.Code + ": " + e.Operation
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Operation, e.Cause)
}
func (e *OpsError) Unwrap() error { return e.Cause }

// wrapOps wraps cause in an *OpsError so callers can inspect the underlying
// sentinel with errors.Is. Returning the typed error (rather than a plain
// fmt.Errorf string) is what keeps the HTTP status classification correct.
func wrapOps(code, operation string, cause error) error {
	return &OpsError{Code: code, Operation: operation, Cause: cause}
}

// opsCode classifies an operations error into a stable code for HTTP mapping.
// It must use errors.Is (not ==) so that errors produced via fmt.Errorf("%w", ...)
// or wrapOps still resolve to their sentinel category instead of "internal".
func opsCode(err error) string {
	switch {
	case opsIsNotFound(err):
		return "not_found"
	case opsIsConflict(err):
		return "conflict"
	case opsIsInvalid(err):
		return "invalid"
	case opsIsTransition(err):
		return "transition"
	case opsIsPolicy(err):
		return "policy"
	default:
		return "internal"
	}
}
func opsIsNotFound(err error) bool   { return errors.Is(err, ErrOpsNotFound) }
func opsIsConflict(err error) bool   { return errors.Is(err, ErrOpsConflict) }
func opsIsInvalid(err error) bool    { return errors.Is(err, ErrOpsInvalid) }
func opsIsTransition(err error) bool { return errors.Is(err, ErrOpsTransition) }
func opsIsPolicy(err error) bool     { return errors.Is(err, ErrOpsPolicy) }
