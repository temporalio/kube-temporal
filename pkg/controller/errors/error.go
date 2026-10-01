package errors

import (
	"fmt"
)

var (
	// NotImplemented is returned when a code path isn't implemented yet
	NotImplemented = fmt.Errorf("not implemented")
	// NotFound is returned when an expected resource was not found
	NotFound = fmt.Errorf("resource not found")
	// Terminal is returned with resource is in Terminal Condition
	Terminal = fmt.Errorf("resource is in terminal condition")
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
