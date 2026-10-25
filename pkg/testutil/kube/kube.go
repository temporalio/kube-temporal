package kube

import (
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlrt "sigs.k8s.io/controller-runtime"
	kcrfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	ResourceIgnoredFields = []string{
		// fake client doesn't properly handle resource version updating...
		"ObjectMeta.ResourceVersion",
	}
)

var (
	TestResourceGroupResource = schema.GroupResource{
		Group:    "test",
		Resource: "resource",
	}
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = corev1.AddToScheme(scheme)
}

// NewFakeClientBuilder returns a controller-runtime ClientBuilder with the
// tcorev1 runtime.Scheme.
//
// Source code repositories that consume kube-temporal/pkg/testutil/kube are
// expected to call NewFakeClientBuilder() and add their own runtime.Scheme and
// then return a client from the ClientBuilder by calling ClientBuilder.Build()
func NewFakeClientBuilder() *kcrfake.ClientBuilder {
	return kcrfake.NewClientBuilder().
		WithGlobalResourceVersionCounter().
		WithScheme(scheme)
}

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
