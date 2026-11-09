package fixtures

import (
	"github.com/temporalio/kube-temporal/pkg/kube/annotation"
	"github.com/temporalio/kube-temporal/pkg/resource/base"
	"github.com/temporalio/kube-temporal/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tcorev1 "github.com/temporalio/kube-temporal/api/core/v1"
)

var (
	connGVK = tcorev1.SchemeGroupVersion.WithKind("Connection")
)

// connRes wraps a core.temporal.io/Connection CR and implements the types.Resource
// interface.
type connRes struct {
	base.Resource
	// cr is the Kubernetes-native custom resource.
	cr *tcorev1.Connection
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

// CR returns the underlying custom resource.
func (r connRes) CR() any {
	return r.cr
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

// DeepCopy returns a copy of the resource
func (r connRes) DeepCopy() types.Resource {
	crCopy := r.cr.DeepCopy()

	// NOTE(jaypipes): we can safely ignore the error returned from base.New()
	// since we know our CR struct pointer always implements client.Object and
	// runtime.Object.
	b, _ := base.New(connGVK, crCopy)

	return &connRes{
		Resource: *b,
		cr:       crCopy,
	}
}

// ResourceFromConnection returns a new [types.Resource] for the supplied
// [tcorev1.Connection].
func ResourceFromConnection(c *tcorev1.Connection) types.Resource {
	// NOTE(jaypipes): we can safely ignore the error returned from base.New()
	// since we know our CR struct pointer always implements client.Object and
	// runtime.Object.
	b, _ := base.New(connGVK, c)

	return &connRes{
		Resource: *b,
		cr:       c,
	}
}

type NewConnectionOption func(*tcorev1.Connection)

// ConnectionWithKubeNamespace sets the Connection's Kubernetes namespace.
func ConnectionWithKubeNamespace(ns string) NewConnectionOption {
	return func(c *tcorev1.Connection) {
		c.Namespace = ns
	}
}

// ConnectionWithConditions sets the Connection's Status.Conditions to
// supplied Conditions.
func ConnectionWithConditions(
	conds ...*metav1.Condition,
) NewConnectionOption {
	return func(c *tcorev1.Connection) {
		c.Status.Conditions = conds
	}
}

// ConnectionWithAnnotations set the Connection's Annotations to the supplied
// Annotations.
func ConnectionWithAnnotations(
	annos annotation.Annotations,
) NewConnectionOption {
	return func(c *tcorev1.Connection) {
		c.Annotations = annos
	}
}

// NewConnection returns a new [tcorev1.Connection] with the supplied name and
// options.
func NewConnection(
	name string,
	opts ...NewConnectionOption,
) types.Resource {
	cr := tcorev1.Connection{}
	cr.Name = name
	for _, opt := range opts {
		opt(&cr)
	}

	// NOTE(jaypipes): we can safely ignore the error returned from base.New()
	// since we know our CR struct pointer always implements client.Object and
	// runtime.Object.
	b, _ := base.New(connGVK, &cr)

	return &connRes{
		Resource: *b,
		cr:       &cr,
	}
}
