package types

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/temporalio/kube-temporal/pkg/condition"
)

// Resource wraps the custom resource object from the Kubernetes API with some
// helper methods.
type Resource interface {
	condition.Manager
	// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
	// deletion timestamp.
	IsBeingDeleted() bool
	// Unstructured returns the Resource as an unstructured.Unstructured.
	Unstructured() (*unstructured.Unstructured, error)
	// ClientObject returns the Kubernetes controller-runtime Client
	// representation of the Resource
	ClientObject() client.Object
	// RuntimeObject returns the controller-runtime client.Object for the
	// resource.
	RuntimeObject() runtime.Object
	// GroupVersionKind returns the GroupVersionKind that the Resource
	// represents.
	GroupVersionKind() schema.GroupVersionKind
	// GroupVersion returns the
	// DeepCopy will return a copy of the resource.
	DeepCopy() Resource
}
