package types

// TraceExiter demarcates the end of a traced code block or function
type TraceExiter func(err error, additionalValues ...interface{})

// Tracer is responsible for tracing entrance and exit from blocks of code
type Tracer interface {
	// Trace logs an entry to a function or code block and returns a functor
	// that can be called to log the exit of the function or code block
	Trace(name string, additionalValues ...interface{}) TraceExiter
}
