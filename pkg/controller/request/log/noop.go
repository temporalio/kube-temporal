package log

import (
	"github.com/temporalio/kube-temporal/pkg/types"
)

var (
	// NoopLogger is useful for testing/mocking
	NoopLogger types.Logger = &noopLogger{}
)

// noopLogger implements Logger but does nothing. Useful for
// testing and mocking...
type noopLogger struct{}

func (l *noopLogger) DebugEnabled() bool           { return false }
func (l *noopLogger) WithValues(...interface{})    {}
func (l *noopLogger) Info(string, ...interface{})  {}
func (l *noopLogger) Debug(string, ...interface{}) {}
func (l *noopLogger) Trace(name string, additionalValues ...interface{}) types.TraceExiter {
	f := func(err error, args ...interface{}) {
		l.Exit(name, err, args...)
	}
	return f
}
func (l *noopLogger) Enter(string, ...interface{})       {}
func (l *noopLogger) Exit(string, error, ...interface{}) {}
