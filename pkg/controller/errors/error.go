package errors

import (
	"fmt"
)

var (
	// ErrNotImplemented is returned when a code path isn't implemented yet
	ErrNotImplemented = fmt.Errorf("not implemented")
	// NotFound is returned when an expected resource was not found
	ErrNotFound = fmt.Errorf("resource not found")
	// ErrResourceSync is returned from the ResourceSyncFunc callback when the
	// reconciler fails to synchronize the desired state of a resource.
	ErrResourceSync = fmt.Errorf("failed resource sync")
)

// TerminalError defines an error that should be considered terminal, causing a
// [kstatus.ConditionStalled] to be added to a resource.
type TerminalError struct {
	err error
}

func NewTerminalError(terminalError error) *TerminalError {
	return &TerminalError{err: terminalError}
}

func (e TerminalError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e TerminalError) Unwrap() error {
	return e.err
}

var _ error = &TerminalError{}
