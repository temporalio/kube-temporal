package log

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/temporalio/kube-temporal/pkg/types"
)

const (
	// contextKey is the string key used to store a logger in a Context
	contextKey = "request.logger"
)

// FromContext returns a [types.Logger] from a saved key in the request
// context
func FromContext(ctx context.Context) types.Logger {
	if v := ctx.Value(contextKey); v != nil {
		return v.(*requestLogger)
	}
	return nil
}

// ToContext ensures that a request logger adapted for the supplied
// [types.Resource] is saved to the supplied context, returning the adapted
// context.
func ToContext(
	ctx context.Context,
	l logr.Logger,
	res types.Resource,
) context.Context {
	rl := FromContext(ctx)
	if rl != nil {
		return ctx
	}
	co := res.ClientObject()
	name := co.GetName()
	ns := co.GetNamespace()
	gvk := res.GroupVersionKind()
	kind := gvk.Kind
	groupVersion := gvk.GroupVersion()
	l = l.WithValues(
		"groupVersion", groupVersion.String(),
		"kind", kind,
		"namespace", ns,
		"name", name,
	)
	rl = &requestLogger{
		log:        l,
		res:        res,
		blockDepth: 0,
	}
	return context.WithValue(ctx, contextKey, rl)
}
