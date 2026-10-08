package log

import (
	"context"
	"log/slog"

	"github.com/temporalio/kube-temporal/pkg/types"
)

type contextKey string

const (
	contextKeyRequestLogger contextKey = "request.logger"
)

// FromContext returns a [types.Logger] from a saved key in the request
// context
func FromContext(ctx context.Context) *Logger {
	if v := ctx.Value(contextKeyRequestLogger); v != nil {
		return v.(*Logger)
	}
	return nil
}

// ToContext ensures that a request logger adapted for the supplied
// [types.Resource] is saved to the supplied context, returning the adapted
// context.
func ToContext(
	ctx context.Context,
	logger *slog.Logger,
	res types.Resource,
) context.Context {
	rl := FromContext(ctx)
	if rl != nil {
		return ctx
	}
	logger = logger.With(
		"resource", res,
	)
	rl = &Logger{
		logger:     logger,
		res:        res,
		blockDepth: 0,
	}
	return context.WithValue(ctx, contextKeyRequestLogger, rl)
}
