package log

import (
	"context"
	"log/slog"
	"strings"

	"github.com/temporalio/kube-temporal/pkg/types"
)

// Logger is a wrapper around a logr.Logger that writes log messages
// contextualized for a single reconcile request.
//
// Implements [types.Logger]
type Logger struct {
	logger     *slog.Logger
	res        types.Resource
	blockDepth int
}

// DebugEnabled returns true when the undelying logger is configured to write
// debug messages, false otherwise.
func (l *Logger) DebugEnabled() bool {
	return l.logger.Enabled(context.Background(), slog.LevelDebug)
}

// Debug writes a supplied log message at debug log level.
func (l *Logger) Debug(
	msg string,
	args ...slog.Attr,
) {
	if !l.DebugEnabled() {
		return
	}
	ctx := context.Background()
	attrs := expandResourceFields(l.res, args...)
	l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Info writes a supplied log message at info log level.
func (l *Logger) Info(
	msg string,
	args ...slog.Attr,
) {
	ctx := context.Background()
	attrs := expandResourceFields(l.res, args...)
	l.logger.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// Enter logs an entry to a function or code block
func (l *Logger) Enter(
	name string, // name of the function or code block we're entering
	args ...slog.Attr,
) {
	if !l.DebugEnabled() {
		return
	}
	ctx := context.Background()
	l.blockDepth++
	depth := strings.Repeat(">", l.blockDepth)
	msg := depth + " " + name
	attrs := expandResourceFields(l.res, args...)
	l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// Exit logs an exit from a function or code block
func (l *Logger) Exit(
	name string, // name of the function or code block we're exiting
	err error,
	args ...slog.Attr,
) {
	if !l.DebugEnabled() {
		return
	}
	ctx := context.Background()
	depth := strings.Repeat("<", l.blockDepth)
	msg := depth + " " + name
	if err != nil {
		args = append(args, slog.String("error", err.Error()))
	}
	attrs := expandResourceFields(l.res, args...)
	l.logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
	l.blockDepth--
}

// Trace logs an entry to a function or code block and returns a functor
// that can be called to log the exit of the function or code block
func (l *Logger) Trace(
	name string,
	args ...slog.Attr,
) func(error, ...slog.Attr) {
	l.Enter(name, args...)
	f := func(err error, args ...slog.Attr) {
		l.Exit(name, err, args...)
	}
	return f
}

// expandResourceFields returns slog Attrs for a resource that should be used
// as structured data in log messages about the resource.
func expandResourceFields(
	res types.Resource,
	args ...slog.Attr,
) []slog.Attr {
	co := res.ClientObject()
	generation := co.GetGeneration()
	attrs := []slog.Attr{
		slog.Int64("generation", generation),
	}
	if len(args) > 0 {
		attrs = append(attrs, args...)
	}
	return attrs
}
