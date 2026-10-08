package types

import (
	ctrlrt "sigs.k8s.io/controller-runtime"
	ctrlrtrec "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controller can bind to a [ctrlrt.Manager]. Implements the upstream
// controller-runtime Reconciler interface.
type Controller interface {
	ctrlrtrec.Reconciler
	// BindManager binds the Controller to the supplied [ctrlrt.Manager].
	BindManager(ctrlrt.Manager) error
}
