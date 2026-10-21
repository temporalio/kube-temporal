package testutil

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlrt "sigs.k8s.io/controller-runtime"
)

var (
	TestResourceGroupResource = schema.GroupResource{
		Group:    "test",
		Resource: "resource",
	}
)

// NewResourceRequest returns a controller-runtime.Request object for the given
// namespace and name.
func NewResourceRequest(
	ns string,
	name string,
) ctrlrt.Request {
	return ctrlrt.Request{
		NamespacedName: types.NamespacedName{
			Namespace: ns,
			Name:      name,
		},
	}
}

// NewNotFoundError returns an apimachinery api/errors NotFound object.
func NewNotFoundError(
	name string,
) *apierrors.StatusError {
	return apierrors.NewNotFound(TestResourceGroupResource, name)
}
