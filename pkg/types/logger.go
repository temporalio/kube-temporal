package types

// Logger is responsible for writing log messages
type Logger interface {
	Tracer
	// DebugEnabled returns true when the underlying logger is configured to
	// write debug messages, false otherwise.
	DebugEnabled() bool
	// WithValues adapts the internal logger with a set of additional key/value
	// data
	WithValues(...interface{})
	// Debug writes a supplied log message at debug log level.
	Debug(msg string, additionalValues ...interface{})
	// Info writes a supplied log message at info log level.
	Info(msg string, additionalValues ...interface{})
	// Enter logs an entry to a function or code block
	Enter(name string, additionalValues ...interface{})
	// Exit logs an exit from a function or code block
	Exit(name string, err error, additionalValues ...interface{})
}
