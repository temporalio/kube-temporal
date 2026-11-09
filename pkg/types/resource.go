package types

import (
	"log/slog"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/temporalio/kube-temporal/pkg/kube/annotation"
	"github.com/temporalio/kube-temporal/pkg/kube/condition"
	"github.com/temporalio/kube-temporal/pkg/kube/finalizer"
)

// Resource wraps the custom resource object from the Kubernetes API with some
// helper methods.
//
//mockery:generate: true
type Resource interface {
	condition.Manager
	annotation.Manager
	finalizer.Manager
	slog.LogValuer
	HasClientObject
	// CR returns a pointer to the underlying custom resource.
	CR() any
	// IsBeingDeleted returns true if the Kubernetes resource has a non-zero
	// deletion timestamp.
	IsBeingDeleted() bool
	// SetDeletionTimestamp sets the time that the resource was marked for
	// deletion.
	SetDeletionTimestamp(time.Time)
	// Unstructured returns the Resource as an unstructured.Unstructured.
	Unstructured() (*unstructured.Unstructured, error)
	// RuntimeObject returns the apimachinery runtime.Object representation for
	// the resource.
	RuntimeObject() runtime.Object
	// DeepCopy will return a copy of the resource.
	DeepCopy() Resource
}
