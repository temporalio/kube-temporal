package types

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HasClientObject describes something that returns a controller-runtime
// client.Object and group version kind information.
//
//mockery:generate: true
type HasClientObject interface {
	// ClientObject returns the Kubernetes controller-runtime client.Object
	// representation of the Resource
	ClientObject() client.Object
	// GroupVersionKind returns the GroupVersionKind that the Resource
	// represents.
	GroupVersionKind() schema.GroupVersionKind
}
