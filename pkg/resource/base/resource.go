package base

import (
	"fmt"
	"log/slog"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/temporalio/kube-temporal/pkg/kube/annotation"
	"github.com/temporalio/kube-temporal/pkg/kube/finalizer"
	"github.com/temporalio/kube-temporal/pkg/kube/transform"
	"github.com/temporalio/kube-temporal/pkg/resource"
)

// Resource is a base struct meant to be embedded in structs that implement
// [types.Resource] to reduce the amount of copy/paste necessary in
// constructing new operator repositories.
type Resource struct {
	gvk schema.GroupVersionKind
	co  client.Object
	ro  runtime.Object
}

// Annotations returns the set of Annotations.
func (r Resource) Annotations() annotation.Annotations {
	return r.ClientObject().GetAnnotations()
}

// ReplaceAnnotations replaces the resource's set of Annotations with
// the supplied Annotations.
func (r *Resource) ReplaceAnnotations(annos annotation.Annotations) {
	r.ClientObject().SetAnnotations(annos)
}

// SetAnnotation adds or replaces the supplied annotation key and value.
func (r *Resource) SetAnnotation(key string, value string) {
	annos := r.Annotations()
	annos[key] = value
	r.ReplaceAnnotations(annos)
}

// Finalizers returns the set of Finalizers.
func (r *Resource) Finalizers() []string {
	return r.ClientObject().GetFinalizers()
}

// ReplaceFinalizers replaces the resource's set of Finalizers with
// the supplied Finalizers.
func (r *Resource) ReplaceFinalizers(finalizers []string) {
	r.ClientObject().SetFinalizers(finalizers)
}

// SetFinalizer adds or replaces the supplied finalizer on the resource.
func (r *Resource) SetFinalizer(f string) {
	finalizer.AddFinalizer(r, f)
}

// LogValue implements [slog.LogValuer] to format the Resource as a group.
func (r *Resource) LogValue() slog.Value {
	return resource.LogValue(r)
}

// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
// deletion timestamp
func (r Resource) IsBeingDeleted() bool {
	return !r.ClientObject().GetDeletionTimestamp().IsZero()
}

// SetDeletionTimestamp sets the time that the resource was marked for
// deletion.
func (r *Resource) SetDeletionTimestamp(t time.Time) {
	dt := metav1.NewTime(t)
	r.ClientObject().SetDeletionTimestamp(&dt)
}

// Unstructured returns the resource as a [unstructured.Unstructured]
func (r Resource) Unstructured() (*unstructured.Unstructured, error) {
	return transform.ToUnstructured(r.RuntimeObject())
}

// ClientObject returns the Kubernetes controller-runtime client.Object
// representation of the Resource
func (r Resource) ClientObject() client.Object {
	return r.co
}

// GroupVersionKind returns the GroupVersionKind that the Resource
// represents.
func (r Resource) GroupVersionKind() schema.GroupVersionKind {
	return r.gvk
}

// RuntimeObject returns the apimachinery runtime.Object representation for the
// resource.
func (r Resource) RuntimeObject() runtime.Object {
	return r.ro
}

// New returns a new base Resource from the supplied GroupVersionKind and
// concrete custom resource.  The supplied custom resource must satisfy both
// apimachinery's runtime.Object interface as well as controller-runtime's
// client.Object interface. If it does not, returns an error.
func New(gvk schema.GroupVersionKind, cr any) (*Resource, error) {
	co, ok := cr.(client.Object)
	if !ok {
		return nil, fmt.Errorf(
			"supplied custom resource %T does not implement "+
				"controller-runtime client.Object",
			cr,
		)
	}
	ro, ok := cr.(runtime.Object)
	if !ok {
		return nil, fmt.Errorf(
			"supplied custom resource %T does not implement "+
				"apimachinery runtime.Object",
			cr,
		)
	}
	return &Resource{
		gvk: gvk,
		ro:  ro,
		co:  co,
	}, nil
}
