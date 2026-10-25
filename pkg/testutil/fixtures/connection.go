package fixtures

import (
	"log/slog"

	"github.com/temporalio/kube-temporal/pkg/kube/annotation"
	"github.com/temporalio/kube-temporal/pkg/kube/transform"
	"github.com/temporalio/kube-temporal/pkg/resource"
	"github.com/temporalio/kube-temporal/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
)

// connRes wraps a core.temporal.io/Connection CR and implements the types.Resource
// interface.
type connRes struct {
	// cr is the Kubernetes-native custom resource.
	cr *tcorev1.Connection
}

// CR returns the underlying custom resource.
func (r connRes) CR() any {
	return r.cr
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestamp
func (r connRes) IsBeingDeleted() bool {
	return !r.cr.DeletionTimestamp.IsZero()
}

// ClientObject returns the Kubernetes controller-runtime Client representation
// of the Resource
func (r connRes) ClientObject() client.Object {
	return r.cr
}

// RuntimeObject returns the Kubernetes apimachinery/runtime representation of
// the Resource
func (r connRes) RuntimeObject() runtime.Object {
	return r.cr
}

// GroupVersionKind returns the GroupVersionKind that the Resource
// represents.
func (r connRes) GroupVersionKind() schema.GroupVersionKind {
	return tcorev1.SchemeGroupVersion.WithKind("Connection")
}

// Conditions returns the Resource's collection of Conditions from its status
// field. Satisfies the `pkg/condition.Manager` interface.
func (r connRes) Conditions() []*metav1.Condition {
	return r.cr.Status.Conditions
}

// ReplaceConditions replaces the Resource's collection of Conditions with the
// supplied collection. Satisfies the `pkg/condition.Manager` interface.
func (r *connRes) ReplaceConditions(conditions []*metav1.Condition) {
	r.cr.Status.Conditions = conditions
}

// Annotations returns the set of Annotations.
func (r *connRes) Annotations() annotation.Annotations {
	return r.cr.ObjectMeta.Annotations
}

// ReplaceAnnotations replaces the resource's set of Annotations with
// the supplied Annotations.
func (r *connRes) ReplaceAnnotations(annos annotation.Annotations) {
	r.cr.ObjectMeta.Annotations = annos
}

// DeepCopy returns a copy of the resource
func (r connRes) DeepCopy() types.Resource {
	crCopy := r.cr.DeepCopy()
	return &connRes{crCopy}
}

// Unstructured returns the resource as a [unstructured.Unstructured]
func (r connRes) Unstructured() (*unstructured.Unstructured, error) {
	return transform.ToUnstructured(r.RuntimeObject())
}

// LogValue implements [slog.LogValuer] to format the Resource as a group.
func (r *connRes) LogValue() slog.Value {
	return resource.LogValue(r)
}

// ResourceFromConnection returns a new [types.Resource] for the supplied
// [tcorev1.Connection].
func ResourceFromConnection(c *tcorev1.Connection) types.Resource {
	return &connRes{c}
}

type NewConnectionOption func(*tcorev1.Connection)

// ConnectionWithKubeNamespace sets the Connection's Kubernetes namespace.
func ConnectionWithKubeNamespace(ns string) NewConnectionOption {
	return func(c *tcorev1.Connection) {
		c.Namespace = ns
	}
}

// NewConnection returns a new [tcorev1.Connection] with the supplied name and
// options.
func NewConnection(
	name string,
	opts ...NewConnectionOption,
) *tcorev1.Connection {
	c := tcorev1.Connection{}
	c.Name = name
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}
