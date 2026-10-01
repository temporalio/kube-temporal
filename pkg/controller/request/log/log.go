package log

import (
	"strings"

	"github.com/go-logr/logr"

	"github.com/temporalio/kube-temporal/pkg/types"
)

// requestLogger is a wrapper around a logr.Logger that writes log messages
// contextualized for a single reconcile request.
//
// Implements [types.Logger]
type requestLogger struct {
	log        logr.Logger
	res        types.Resource
	blockDepth int
}

// DebugEnabled returns true when the underlying logger is configured to write
// debug messages, false otherwise.
func (rl *requestLogger) DebugEnabled() bool {
	return rl.log.V(1).Enabled()
}

// WithValues adapts the internal logger with a set of additional values
func (rl *requestLogger) WithValues(
	values ...interface{},
) {
	rl.log = rl.log.WithValues(values...)
}

// Debug writes a supplied log message at debug log level.
func (rl *requestLogger) Debug(
	msg string,
	additionalValues ...interface{},
) {
	if !rl.DebugEnabled() {
		return
	}
	vals := expandResourceFields(rl.res, additionalValues...)
	rl.log.V(1).Info(msg, vals...)
}

// Info writes a supplied log message at info log level.
func (rl *requestLogger) Info(
	msg string,
	additionalValues ...interface{},
) {
	vals := expandResourceFields(rl.res, additionalValues...)
	rl.log.V(0).Info(msg, vals...)
}

// Enter logs an entry to a function or code block
func (rl *requestLogger) Enter(
	name string, // name of the function or code block we're entering
	additionalValues ...interface{},
) {
	if !rl.DebugEnabled() {
		return
	}
	rl.blockDepth++
	depth := strings.Repeat(">", rl.blockDepth)
	msg := depth + " " + name
	vals := expandResourceFields(rl.res, additionalValues...)
	rl.log.V(1).Info(msg, vals...)
}

// Exit logs an exit from a function or code block
func (rl *requestLogger) Exit(
	name string, // name of the function or code block we're exiting
	err error,
	additionalValues ...interface{},
) {
	if !rl.DebugEnabled() {
		return
	}
	depth := strings.Repeat("<", rl.blockDepth)
	msg := depth + " " + name
	if err != nil {
		additionalValues = append(additionalValues, "error")
		additionalValues = append(additionalValues, err)
	}
	vals := expandResourceFields(rl.res, additionalValues...)
	rl.log.V(1).Info(msg, vals...)
	rl.blockDepth--
}

// Trace logs an entry to a function or code block and returns a functor
// that can be called to log the exit of the function or code block
func (rl *requestLogger) Trace(
	name string,
	additionalValues ...interface{},
) types.TraceExiter {
	rl.Enter(name, additionalValues...)
	f := func(err error, args ...interface{}) {
		rl.Exit(name, err, args...)
	}
	return f
}

// expandResourceFields returns the key/value pairs for a resource that should
// be used as structured data in log messages about the resource.
func expandResourceFields(
	res types.Resource,
	additionalValues ...interface{},
) []interface{} {
	co := res.ClientObject()
	generation := co.GetGeneration()
	vals := []interface{}{
		"generation", generation,
	}
	if len(additionalValues) > 0 {
		vals = append(vals, additionalValues...)
	}
	return vals
}
