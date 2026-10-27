package types

import (
	"context"

	"github.com/temporalio/kube-temporal/pkg/controller/config"
	ctrlrt "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconciler understands how to reconcile a [Resource]. Implements the
// upstream controller-runtime Reconciler interface.
//
//mockery:generate: true
type Reconciler interface {
	// SetConfig sets the Reconciler's configuration.
	SetConfig(cfg config.Config)
	// SetKubeClient sets the Reconciler's Kubernetes controller-runtime
	// Client.
	SetKubeClient(cli client.Client)
	// SetKubeReader sets the Reconciler's Kubernetes controller-runtime
	// Reader.
	SetKubeReader(reader client.Reader)
	// AdaptBuilder returns a [*ctrlrt.Builder] adapted from the supplied
	// [*ctrlrt.Builder], typically calling the [ctrlrt.Builder.For] method
	// with the Reconciler's managed Kind prototype.
	AdaptBuilder(*ctrlrt.Builder) *ctrlrt.Builder

	// ResourceFromRequest translates a controller-runtime Request into a
	// [Resource]
	ResourceFromRequest(
		context.Context,
		ctrlrt.Request,
	) (Resource, error)

	// ResourceSetLatest mutates the supplied [Resource] with the latest
	// observed state of the externally-managed resource.
	//
	// Any error returned from this function is considered a runtime error and
	// will cause the resource to contain a Stalled Condition with Status=True.
	ResourceSetLatest(
		context.Context,
		Resource, // callers expected to mutate this parameter
	) error

	// ResourceSync synchronizes any managed resources to match the supplied
	// desired [Resource] state.
	//
	// Returns nil when the synchronization was successful, otherwise returns
	// an error derived from `pkg/controller/errors.ErrResourceSync`.
	//
	// `desired` is NEVER mutated.
	// `latest` is NEVER mutated.
	//
	// This method is only called when there was a difference detected between
	// the desired and latest [Resource].
	ResourceSync(
		context.Context,
		Resource, // desired
		Resource, // latest
	) error

	// ResourceDelete deletes any managed external resources for the supplied
	// [Resource]. Note that the Reconciler does NOT need to remove the
	// finalizer on the associated [client.Object]. The Controller handles
	// that. All the Reconciler needs to do is delete any owned resources or
	// externally-managed resources associated with the supplied [Resource].
	//
	// This method returns a copy of the supplied [Resource] with any status
	// mutations applied to the copy. If the external resource deletion
	// operations are long-running or failed in a non-terminal fashion, the
	// implementer should return a copy of the supplied [Resource] with
	// any saved status mutations (for instance to the resource's Status.Conditions
	// collection).
	//
	// `original` should NOT be mutated. The caller of `ResourceDelete` will consider
	// the returned [Resource] as the latest observed state of the resource
	// undergoing reconciliation.
	ResourceDelete(
		context.Context,
		Resource, // original
	) (Resource, error)
}
