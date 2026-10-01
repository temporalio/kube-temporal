package requeue

import (
	"time"
)

const (
	DefaultRequeueAfterDuration time.Duration = 30 * time.Second
)

// None returns a new NoRequeue to instruct the reconciler to not requeue the
// processing item.
func None(err error) *NoRequeue {
	return &NoRequeue{
		err: err,
	}
}

// Needed returns a new RequeueNeeded to instruct the reconciler to requeue the
// processing item.
func Needed(err error) *RequeueNeeded {
	return &RequeueNeeded{
		err: err,
	}
}

// NeededAfter returns a new RequeueNeededAfter to instruct the reconciler to
// requeue the processing item after specified duration.
func NeededAfter(
	err error,
	duration time.Duration,
) *RequeueNeededAfter {
	return &RequeueNeededAfter{
		RequeueNeeded{
			err: err,
		},
		duration,
	}
}

// NoRequeue instructs the reconciler to process an error, but not requeue the
// object that raised it. This should be used when there was a non-terminal
// error, but one that cannot be fixed by requeuing.
type NoRequeue struct {
	err error
}

func (e *NoRequeue) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *NoRequeue) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Ensure NoRequeue implements the error interface
var _ error = &NoRequeue{}

// RequeueNeeded instructs the reconciler to process a non-terminal error and
// requeue the processing item.
type RequeueNeeded struct {
	err error
}

func (e *RequeueNeeded) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *RequeueNeeded) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Ensure RequeueNeeded implements the error interface
var _ error = &RequeueNeeded{}

// RequeueNeededAfter instructs the reconciler to process a non-terminal error
// and requeue the processing item after the specified duration.
type RequeueNeededAfter struct {
	RequeueNeeded
	duration time.Duration
}

func (e *RequeueNeededAfter) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *RequeueNeededAfter) Duration() time.Duration {
	if e == nil {
		return time.Duration(0) * time.Second
	}
	return e.duration
}

func (e *RequeueNeededAfter) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Ensure RequeueNeededAfter implements the error interface
var _ error = &RequeueNeededAfter{}
