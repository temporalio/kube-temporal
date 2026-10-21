package base

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/temporalio/kube-temporal/pkg/controller/config"
)

// Reconciler is a base struct meant to be embedded in structs that implement
// [types.Reconciler] to reduce the amount of copy/paste necessary in
// constructing new operator repositories.
type Reconciler struct {
	// ks is the controller-runtime client.Client associated with the
	// Reconciler
	kc client.Client
	// kr is the direct, non-caching API reader associated with Reconciler.
	kr client.Reader
	// cfg contains the configuration settings used when creating the
	// Reconciler.
	cfg config.Config
}

// SetConfig sets the Reconciler's configuration.
func (r *Reconciler) SetConfig(cfg config.Config) {
	r.cfg = cfg
}

// Config returns the Reconciler's configuration
func (r Reconciler) Config() config.Config {
	return r.cfg
}

// SetKubeClient sets the Reconciler's Kubernetes controller-runtime Client.
func (r *Reconciler) SetKubeClient(cli client.Client) {
	r.kc = cli
}

// KubeClient returns the Reconciler's Kubernetes controller-runtime Client.
func (r Reconciler) KubeClient() client.Client {
	return r.kc
}

// SetKubeReader sets the Reconciler's Kubernetes controller-runtime Reader.
func (r *Reconciler) SetKubeReader(reader client.Reader) {
	r.kr = reader
}

// KubeReader returns the Reconciler's Kubernetes controller-runtime Reader.
// The Reader is a direct, non-caching API reader.
func (r Reconciler) KubeReader() client.Reader {
	return r.kr
}

// KubeReadLatestInto populates the supplied [client.Object] with the latest
// representation of the resource pointed to by the supplied
// [types.NamespacdeName]. This method always uses the direct, non-caching API
// client.Reader, never the caching client.Client in order to prevent stale
// reads.
func (r *Reconciler) KubeReadLatestInto(
	ctx context.Context,
	nn types.NamespacedName,
	obj client.Object,
) error {
	if r.kr == nil {
		return fmt.Errorf("nil kube client")
	}
	return r.kr.Get(ctx, nn, obj)
}
