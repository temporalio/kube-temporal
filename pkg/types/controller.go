package types

import (
	"github.com/go-logr/logr"
	ctrlrt "sigs.k8s.io/controller-runtime"
	ctrlrtrec "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Controller can bind to a [ctrlrt.Manager]. Implements the upstream
// controller-runtime Reconciler interface.
type Controller interface {
	ctrlrtrec.Reconciler
	// BindManager binds the Controller to the supplied [ctrlrt.Manager].
	BindManager(ctrlrt.Manager) error
	// Logger returns the Controller's Logger.
	Logger() logr.Logger
	// SetLogger sets the Controller's Logger to the supplied logger.
	SetLogger(logr.Logger)
}

// ControllerWithOption modifies a Controller
type ControllerWithOption func(Controller)
