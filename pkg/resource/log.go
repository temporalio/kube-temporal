package resource

import (
	"log/slog"

	"github.com/temporalio/kube-temporal/pkg/types"
)

const (
	logKeyGroup     = "group"
	logKeyVersion   = "version"
	logKeyKind      = "kind"
	logKeyNamespace = "namespace"
	logKeyName      = "name"
)

// LogValue returns the standardized [slog.LogValue] for the supplied
// [types.Resource].
func LogValue(r types.Resource) slog.Value {
	co := r.ClientObject()
	name := co.GetName()
	ns := co.GetNamespace()
	gvk := r.GroupVersionKind()
	kind := gvk.Kind
	groupVersion := gvk.GroupVersion()
	return slog.GroupValue(
		slog.String(logKeyGroup, groupVersion.Group),
		slog.String(logKeyVersion, groupVersion.Version),
		slog.String(logKeyKind, kind),
		slog.String(logKeyNamespace, ns),
		slog.String(logKeyName, name),
	)
}
